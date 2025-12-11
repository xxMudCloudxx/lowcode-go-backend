package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	jsonpatch "github.com/evanphx/json-patch/v5"
)

// ========== Actor Model: Room 是完全自治的独立单元 ==========
// clients map 只在 run() 循环内访问，无需锁！

// Room 既包含数据，也包含处理逻辑（Actor Model）
type Room struct {
	ID           string
	CurrentState []byte
	Version      int64

	// 私有 clients map - 只在 run() 内访问，无需锁
	clients map[*Client]bool

	// 事件通道：所有操作都变成消息
	broadcast  chan *RoomBroadcast // 广播消息
	register   chan *Client        // 加入请求
	unregister chan *Client        // 退出请求
	stopChan   chan struct{}       // 停止信号（由 Hub 发送）

	// 状态标志
	stopping    bool         // 是否正在停止
	clientCount int          // 客户端计数（供 Hub 双重检查）
	countMu     sync.RWMutex // 保护 clientCount 和 stopping

	// 状态锁 - 只用于保护 CurrentState/Version 的并发读写
	stateMu sync.RWMutex

	// 刷盘相关
	lastPersistedVersion int64
	flushTicker          *time.Ticker
	pageService          PageService

	// 反向引用：通知 Hub
	hub *Hub
}

// RoomBroadcast 广播消息结构
type RoomBroadcast struct {
	Message    []byte
	Sender     *Client
	IsCritical bool
}

// 刷盘配置
const (
	FlushInterval  = 30 * time.Second
	FlushThreshold = 50
)

// NewRoom 创建房间并启动事件循环
func NewRoom(id string, initialState []byte, pageService PageService, hub *Hub) *Room {
	r := &Room{
		ID:           id,
		CurrentState: initialState,
		Version:      1,
		clients:      make(map[*Client]bool),
		broadcast:    make(chan *RoomBroadcast, 256),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		stopChan:     make(chan struct{}),
		flushTicker:  time.NewTicker(FlushInterval),
		pageService:  pageService,
		hub:          hub,
	}

	go r.run() // 启动房间事件循环

	log.Printf("[Room %s] 🚀 已创建并启动", id)
	return r
}

// run 是房间的主宰，所有逻辑都在这里串行处理，所以 clients map 不需要锁！
func (r *Room) run() {
	defer func() {
		r.flushTicker.Stop()
		r.flushToDB("销毁前")
		log.Printf("[Room %s] 🛑 事件循环已停止", r.ID)
	}()

	for {
		select {
		// 1. 处理客户端注册 (无锁！)
		case client := <-r.register:
			r.clients[client] = true
			client.Room = r
			r.updateClientCount(1)
			r.sendSyncToClient(client)
			log.Printf("[Room %s] 👋 用户 [%s] 加入，当前人数: %d",
				r.ID, client.UserInfo.UserName, len(r.clients))

		// 2. 处理客户端注销 (无锁！)
		case client := <-r.unregister:
			if _, ok := r.clients[client]; ok {
				delete(r.clients, client)
				close(client.send)
				r.updateClientCount(-1)
				log.Printf("[Room %s] 👋 用户 [%s] 离开，剩余人数: %d",
					r.ID, client.UserInfo.UserName, len(r.clients))

				// 房间空了，通知 Hub 请求销毁（不立即退出，等 Hub 确认）
				if len(r.clients) == 0 && r.hub != nil {
					r.hub.NotifyIdle(r)
				}
			}

		// 3. 处理广播 (核心热路径 - 无锁！)
		case msg := <-r.broadcast:
			for client := range r.clients {
				if msg.Sender != nil && client == msg.Sender {
					continue
				}

				select {
				case client.send <- msg.Message:
					// 发送成功
				default:
					// 缓冲区满
					if msg.IsCritical {
						log.Printf("[Room %s] ⚠️ 关键消息阻塞，踢出 [%s]",
							r.ID, client.UserInfo.UserName)
						delete(r.clients, client)
						close(client.send)
					}
					// 非关键消息直接丢弃
				}
			}

		// 4. 定时刷盘
		case <-r.flushTicker.C:
			r.flushToDB("定时")

		// 5. 停止信号
		case <-r.stopChan:
			return
		}
	}
}

// sendSyncToClient 发送全量同步消息给新用户
func (r *Room) sendSyncToClient(client *Client) {
	snapshot, version := r.GetSnapshot()

	// 收集房间内其他用户信息
	users := make([]UserInfo, 0, len(r.clients))
	for c := range r.clients {
		if c != client {
			users = append(users, c.UserInfo)
		}
	}

	syncPayload := SyncPayload{
		Schema:  snapshot,
		Version: version,
		Users:   users,
	}

	payload, _ := json.Marshal(syncPayload)
	msg := WSMessage{
		Type:      TypeSync,
		SenderID:  "server",
		Payload:   payload,
		Timestamp: time.Now().UnixMilli(),
	}

	data, _ := json.Marshal(msg)
	client.send <- data

	log.Printf("[Room %s] 📤 已发送 Sync 给 [%s], 版本: %d",
		r.ID, client.UserInfo.UserName, version)
}

// ========== 对外暴露的接口 ==========

// ErrRoomClosed 房间已关闭错误
var ErrRoomClosed = fmt.Errorf("room is closing")

// Register 注册客户端到房间（非阻塞，防止向已死房间注册）
func (r *Room) Register(client *Client) error {
	select {
	case r.register <- client:
		return nil // 注册成功
	case <-r.stopChan:
		return ErrRoomClosed // 房间已关闭
	}
}

// Unregister 注销客户端（非阻塞）
func (r *Room) Unregister(client *Client) {
	select {
	case r.unregister <- client:
	case <-r.stopChan:
		// 房间已关闭，不需要注销
	}
}

// Broadcast 广播消息
func (r *Room) Broadcast(message []byte, sender *Client, isCritical bool) {
	r.broadcast <- &RoomBroadcast{
		Message:    message,
		Sender:     sender,
		IsCritical: isCritical,
	}
}

// Stop 停止房间（由 Hub 调用）
func (r *Room) Stop() {
	r.countMu.Lock()
	r.stopping = true
	r.countMu.Unlock()
	close(r.stopChan)
}

// StopWithReason 带原因的停止房间（页面被删除时调用）
// reason: 通知客户端的错误码（如 PAGE_DELETED）
func (r *Room) StopWithReason(reason ErrorCode, message string) {
	r.countMu.Lock()
	r.stopping = true
	r.countMu.Unlock()

	// 广播错误消息给所有客户端（最后一条消息）
	r.broadcastError(reason, message)

	// 等一小段时间让消息发出去
	time.Sleep(100 * time.Millisecond)

	close(r.stopChan)
}

// broadcastError 广播错误消息给所有客户端
func (r *Room) broadcastError(code ErrorCode, message string) {
	errPayload, _ := json.Marshal(ErrorPayload{
		Code:    code,
		Message: message,
	})
	msg := WSMessage{
		Type:      TypeError,
		SenderID:  "server",
		Payload:   errPayload,
		Timestamp: time.Now().UnixMilli(),
	}
	data, _ := json.Marshal(msg)

	// 直接发送到 broadcast channel
	r.broadcast <- &RoomBroadcast{
		Message:    data,
		Sender:     nil, // 发给所有人
		IsCritical: true,
	}
}

// ClientCount 返回当前客户端数量（供 Hub 双重检查）
func (r *Room) ClientCount() int {
	r.countMu.RLock()
	defer r.countMu.RUnlock()
	return r.clientCount
}

// IsStopping 返回房间是否正在停止
func (r *Room) IsStopping() bool {
	r.countMu.RLock()
	defer r.countMu.RUnlock()
	return r.stopping
}

// updateClientCount 更新客户端计数（供 run() 内部调用）
func (r *Room) updateClientCount(delta int) {
	r.countMu.Lock()
	r.clientCount += delta
	r.countMu.Unlock()
}

// ========== 需要锁保护的状态操作 ==========

// ApplyPatch 应用 Patch（需要锁保护 CurrentState）
func (r *Room) ApplyPatch(patchBytes []byte, expectedVersion int64) error {
	r.stateMu.Lock()
	defer r.stateMu.Unlock()

	if r.Version != expectedVersion {
		return &VersionConflictError{
			CurrentVersion:  r.Version,
			ExpectedVersion: expectedVersion,
		}
	}

	patch, err := jsonpatch.DecodePatch(patchBytes)
	if err != nil {
		return &PatchError{Reason: fmt.Sprintf("patch 解析失败: %v", err)}
	}

	modified, err := patch.Apply(r.CurrentState)
	if err != nil {
		return &PatchError{Reason: fmt.Sprintf("patch 应用失败: %v", err)}
	}

	r.CurrentState = modified
	r.Version++

	// 阈值刷盘
	if r.Version-r.lastPersistedVersion >= FlushThreshold {
		go r.flushToDB("阈值触发")
	}

	return nil
}

// GetSnapshot 获取当前快照
func (r *Room) GetSnapshot() ([]byte, int64) {
	r.stateMu.RLock()
	defer r.stateMu.RUnlock()

	snapshot := make([]byte, len(r.CurrentState))
	copy(snapshot, r.CurrentState)

	return snapshot, r.Version
}

// flushToDB 刷盘
func (r *Room) flushToDB(reason string) {
	r.stateMu.RLock()
	if r.Version == r.lastPersistedVersion {
		r.stateMu.RUnlock()
		return
	}

	snapshot := make([]byte, len(r.CurrentState))
	copy(snapshot, r.CurrentState)
	currentVersion := r.Version
	lastVersion := r.lastPersistedVersion
	r.stateMu.RUnlock()

	if err := r.pageService.SavePageState(r.ID, snapshot, lastVersion, currentVersion); err != nil {
		log.Printf("[Room %s] ⚠️ %s刷盘失败: %v", r.ID, reason, err)
		return
	}

	r.stateMu.Lock()
	if currentVersion > r.lastPersistedVersion {
		r.lastPersistedVersion = currentVersion
		log.Printf("[Room %s] ✅ %s刷盘, 版本: %d → %d", r.ID, reason, lastVersion, currentVersion)
	}
	r.stateMu.Unlock()
}
