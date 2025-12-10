package ws

import (
	"log"
	"sync"
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
	GetPageState(pageID string) ([]byte, int64, error)
	SavePageState(pageID string, state []byte, version int64) error
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

// GetOrCreateRoom 线程安全地获取或创建房间
// 这是外部进入房间的唯一入口
func (h *Hub) GetOrCreateRoom(roomID string) *Room {
	// 先尝试读锁快速路径
	h.mu.RLock()
	room, exists := h.rooms[roomID]
	h.mu.RUnlock()

	if exists && !room.IsStopping() {
		return room
	}

	// 不存在或正在停止，加写锁创建
	h.mu.Lock()
	defer h.mu.Unlock()

	// 双重检查
	room, exists = h.rooms[roomID]
	if exists && !room.IsStopping() {
		return room
	}

	// 加载初始状态
	state, version, err := h.pageService.GetPageState(roomID)
	if err != nil {
		log.Printf("[Hub] ⚠️ 加载页面 %s 失败: %v，使用默认状态", roomID, err)
		state = []byte(`{"rootId":1,"components":{"1":{"id":1,"name":"Page","props":{},"desc":"页面","parentId":null}}}`)
		version = 1
	}

	// 创建房间
	room = NewRoom(roomID, state, h.pageService, h)
	room.Version = version
	h.rooms[roomID] = room

	log.Printf("[Hub] 🏠 创建房间 %s，版本: %d", roomID, version)
	return room
}

// NotifyIdle 供 Room 调用，通知 Hub 房间空闲
func (h *Hub) NotifyIdle(room *Room) {
	h.idleRoom <- room
}
