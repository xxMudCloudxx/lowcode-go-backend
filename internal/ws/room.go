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
	stopChan   chan struct{}       // 停止信号

	// 状态锁 - 只用于保护 CurrentState/Version 的并发读写
	stateMu sync.RWMutex

	// 刷盘相关
	lastPersistedVersion int64
	flushTicker          *time.Ticker
	pageService          PageService

	// 反向引用：房间销毁时通知 Hub
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
		// 通知 Hub 销毁房间
		if r.hub != nil {
			r.hub.destroyRoom <- r
		}
		log.Printf("[Room %s] 🛑 事件循环已停止", r.ID)
	}()

	for {
		select {
		// 1. 处理客户端注册 (无锁！)
		case client := <-r.register:
			r.clients[client] = true
			client.Room = r
			r.sendSyncToClient(client)
			log.Printf("[Room %s] 👋 用户 [%s] 加入，当前人数: %d",
				r.ID, client.UserInfo.UserName, len(r.clients))

		// 2. 处理客户端注销 (无锁！)
		case client := <-r.unregister:
			if _, ok := r.clients[client]; ok {
				delete(r.clients, client)
				close(client.send)
				log.Printf("[Room %s] 👋 用户 [%s] 离开，剩余人数: %d",
					r.ID, client.UserInfo.UserName, len(r.clients))

				// 房间空了，退出循环触发销毁
				if len(r.clients) == 0 {
					return
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

// Register 注册客户端到房间
func (r *Room) Register(client *Client) {
	r.register <- client
}

// Unregister 注销客户端
func (r *Room) Unregister(client *Client) {
	r.unregister <- client
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
	close(r.stopChan)
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
	version := r.Version
	r.stateMu.RUnlock()

	if err := r.pageService.SavePageState(r.ID, snapshot, version); err != nil {
		log.Printf("[Room %s] ⚠️ %s刷盘失败: %v", r.ID, reason, err)
		return
	}

	r.stateMu.Lock()
	if version > r.lastPersistedVersion {
		r.lastPersistedVersion = version
		log.Printf("[Room %s] ✅ %s刷盘, 版本: %d", r.ID, reason, version)
	}
	r.stateMu.Unlock()
}
