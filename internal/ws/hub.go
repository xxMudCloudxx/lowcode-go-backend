package ws

import (
	"errors"
	"log"
	"sync"

	domainErrors "lowercode-go-server/domain/errors"
)

// ========== Actor Model: Hub 是生死的唯一仲裁者 ==========
// Hub 不处理任何业务消息，只管理 Room 的生命周期

// Hub 维护房间目录
type Hub struct {
	rooms       map[string]*Room
	mu          sync.RWMutex
	idleRoom    chan *Room // Room 空闲信号（请求销毁）
	pageService PageService
}

// PageService 接口，用于数据库操作
type PageService interface {
	// GetPageState 返回页面状态，如果页面不存在返回 (nil, 0, ErrPageNotFound)
	GetPageState(pageID string) ([]byte, int64, error)
	// PageExists 检查页面是否存在
	PageExists(pageID string) (bool, error)
	// SavePageState 保存页面状态（支持版本跳跃）
	// oldVersion: 上次持久化的版本（用于乐观锁检查）
	// newVersion: 当前内存中的版本（要写入 DB）
	SavePageState(pageID string, state []byte, oldVersion, newVersion int64) error
}

// NewHub 创建 Hub 实例
func NewHub(pageService PageService) *Hub {
	return &Hub{
		rooms:       make(map[string]*Room),
		idleRoom:    make(chan *Room, 16),
		pageService: pageService,
	}
}

// Run Hub 事件循环
func (h *Hub) Run() {
	log.Println("[Hub] 🚀 Hub 已启动（生死仲裁者）")

	for room := range h.idleRoom {
		h.handleIdleRoom(room)
	}
}

// handleIdleRoom 处理空闲房间（双重检查后决定是否销毁）
func (h *Hub) handleIdleRoom(room *Room) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 双重检查：Room 可能在我们处理期间又有人加入了
	if room.ClientCount() > 0 {
		log.Printf("[Hub] 🔄 房间 %s 已有新用户，取消销毁", room.ID)
		return
	}

	// 确认房间还在 map 中
	if _, exists := h.rooms[room.ID]; !exists {
		return
	}

	// 从 map 中移除
	delete(h.rooms, room.ID)

	// 通知 Room 停止（Room 收到 stopChan 才真正退出）
	room.Stop()

	log.Printf("[Hub] 🗑️ 房间 %s 已销毁", room.ID)
}

// GetRoom 只读获取房间，不创建（供 HTTP GET 请求使用）
// 返回 nil 表示房间不存在于内存中
// ⚠️ 这是解决"观察者效应"问题的关键方法
func (h *Hub) GetRoom(roomID string) *Room {
	h.mu.RLock()
	defer h.mu.RUnlock()

	room, exists := h.rooms[roomID]
	if exists && !room.IsStopping() {
		return room
	}
	return nil
}

// GetOrCreateRoom 线程安全地获取或创建房间
// ⚠️ 只有在数据库中存在的页面才会创建房间（Pre-creation 模式）
// 返回值: (*Room, error) - 如果页面不存在，返回 ErrPageNotFound
func (h *Hub) GetOrCreateRoom(roomID string) (*Room, error) {
	// 先尝试读锁快速路径
	h.mu.RLock()
	room, exists := h.rooms[roomID]
	h.mu.RUnlock()

	if exists && !room.IsStopping() {
		return room, nil
	}

	// 不存在或正在停止，加写锁创建
	h.mu.Lock()
	defer h.mu.Unlock()

	// 双重检查
	room, exists = h.rooms[roomID]
	if exists && !room.IsStopping() {
		return room, nil
	}

	// ⚠️ 关键修复：从数据库加载状态，如果页面不存在，返回错误
	state, version, err := h.pageService.GetPageState(roomID)
	if err != nil {
		if errors.Is(err, domainErrors.ErrPageNotFound) {
			log.Printf("[Hub] ❌ 页面 %s 不存在，拒绝创建房间", roomID)
			return nil, domainErrors.ErrPageNotFound
		}
		// 其他数据库错误
		log.Printf("[Hub] ⚠️ 加载页面 %s 失败: %v", roomID, err)
		return nil, err
	}

	// 创建房间
	room = NewRoom(roomID, state, h.pageService, h)
	room.Version = version
	room.lastPersistedVersion = version
	h.rooms[roomID] = room

	log.Printf("[Hub] 🏠 创建房间 %s，版本: %d", roomID, version)
	return room, nil
}

// NotifyIdle 供 Room 调用，通知 Hub 房间空闲
func (h *Hub) NotifyIdle(room *Room) {
	h.idleRoom <- room
}

// CloseRoom 强制关闭房间（供 API 删除页面时调用）
// ⚠️ 这是流程的第一步：先关闭房间，后删数据库
func (h *Hub) CloseRoom(roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, exists := h.rooms[roomID]
	if !exists {
		log.Printf("[Hub] ℹ️ 房间 %s 不存在于内存中，无需关闭", roomID)
		return
	}

	// 1. 先从 Hub 目录中移除（防止新用户加入）
	delete(h.rooms, roomID)

	// 2. 通知房间内所有用户，页面已被删除
	// 使用 StopWithReason 发送 PAGE_DELETED 错误，让前端显示友好提示
	room.StopWithReason(ErrPageDeleted, "页面已被删除")

	log.Printf("[Hub] 💀 强制关闭房间 %s（页面被删除）", roomID)
}
