package ws

import (
	"log"
	"sync"
)

// ========== Actor Model: Hub 只是房间目录管理员 ==========
// Hub 不处理任何业务消息，只管理 Room 的生命周期

// Hub 维护房间目录
type Hub struct {
	rooms       map[string]*Room
	mu          sync.RWMutex
	destroyRoom chan *Room // 接收房间销毁请求
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
		destroyRoom: make(chan *Room, 16),
		pageService: pageService,
	}
}

// Run Hub 事件循环（非常轻量）
func (h *Hub) Run() {
	log.Println("[Hub] 🚀 Hub 已启动（只管理房间目录）")

	for room := range h.destroyRoom {
		h.mu.Lock()
		if _, exists := h.rooms[room.ID]; exists {
			delete(h.rooms, room.ID)
			log.Printf("[Hub] 🗑️ 房间 %s 已从目录移除", room.ID)
		}
		h.mu.Unlock()
	}
}

// GetOrCreateRoom 线程安全地获取或创建房间
// 这是外部进入房间的唯一入口
func (h *Hub) GetOrCreateRoom(roomID string) *Room {
	// 先尝试读锁快速路径
	h.mu.RLock()
	room, exists := h.rooms[roomID]
	h.mu.RUnlock()

	if exists {
		return room
	}

	// 不存在，加写锁创建
	h.mu.Lock()
	defer h.mu.Unlock()

	// 双重检查（可能其他 goroutine 已经创建）
	room, exists = h.rooms[roomID]
	if exists {
		return room
	}

	// 加载初始状态
	state, version, err := h.pageService.GetPageState(roomID)
	if err != nil {
		log.Printf("[Hub] ⚠️ 加载页面 %s 失败: %v，使用默认状态", roomID, err)
		state = []byte(`{"rootId":1,"components":{"1":{"id":1,"name":"Page","props":{},"desc":"页面","parentId":null}}}`)
		version = 1
	}

	// 创建房间（会自动启动事件循环）
	room = NewRoom(roomID, state, h.pageService, h)
	room.Version = version
	h.rooms[roomID] = room

	log.Printf("[Hub] 🏠 创建房间 %s，版本: %d", roomID, version)
	return room
}
