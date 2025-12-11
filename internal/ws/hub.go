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
		// ✅ 使用 goroutine 避免阻塞 Hub 事件循环
		// 因为 handleIdleRoom 现在会阻塞等待刷盘完成
		go h.handleIdleRoom(room)
	}
}

// handleIdleRoom 处理空闲房间（双重检查后决定是否销毁）
// ⚠️ 关键修复：先刷盘，再从 Hub 移除，并检查指针同一性
func (h *Hub) handleIdleRoom(room *Room) {
	// 双重检查：Room 可能在我们处理期间又有人加入了
	if room.ClientCount() > 0 {
		log.Printf("[Hub] 🔄 房间 %s 已有新用户，取消销毁", room.ID)
		return
	}

	// ✅ 先停止房间（阻塞等待刷盘完成）
	room.Stop()

	// ✅ 安全删除：检查指针同一性，防止误删新创建的房间
	h.mu.Lock()
	defer h.mu.Unlock()

	// ⚠️ 关键：检查 Map 里的房间是不是当初那个房间
	// 防止 GetOrCreateRoom 在刷盘期间创建了新房间，结果被我们删了
	if currentRoom, ok := h.rooms[room.ID]; ok && currentRoom == room {
		delete(h.rooms, room.ID)
		log.Printf("[Hub] 🗑️ 房间 %s 已销毁", room.ID)
	} else {
		log.Printf("[Hub] ⚠️ 房间 %s 销毁时发现已被替换或移除，跳过删除", room.ID)
	}
}

// GetRoom 只读获取房间，不创建（供 HTTP GET 请求使用）
// ✅ 修正：只要房间在内存，就返回它，因为内存数据永远比 DB 新
// 即使房间正在 Stopping，它的 State 仍然是可读的（有 stateMu 保护）
func (h *Hub) GetRoom(roomID string) *Room {
	h.mu.RLock()
	defer h.mu.RUnlock()

	room, exists := h.rooms[roomID]
	// ✅ 只要存在就返回，哪怕正在 stopping
	// stopping 的房间仍持有最新数据，且 GetSnapshot 有 stateMu 保护
	if exists {
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

	if exists {
		// ⚠️ 关键修正：如果房间存在但正在停止，返回错误让客户端重试
		if room.IsStopping() {
			log.Printf("[Hub] ⏳ 房间 %s 正在关闭，请客户端重试", roomID)
			return nil, domainErrors.ErrRoomClosing
		}
		return room, nil
	}

	// 不存在，加写锁创建
	h.mu.Lock()
	defer h.mu.Unlock()

	// 双重检查
	room, exists = h.rooms[roomID]
	if exists {
		// ⚠️ 关键修正：如果房间存在但正在停止，返回错误让客户端重试
		if room.IsStopping() {
			log.Printf("[Hub] ⏳ 房间 %s 正在关闭，请客户端重试", roomID)
			return nil, domainErrors.ErrRoomClosing
		}
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
// ⚠️ 这是"处决"流程的第一步：先关闭房间并刷盘，后删数据库
func (h *Hub) CloseRoom(roomID string) {
	h.mu.Lock()
	room, exists := h.rooms[roomID]
	if !exists {
		h.mu.Unlock()
		log.Printf("[Hub] ℹ️ 房间 %s 不存在于内存中，无需关闭", roomID)
		return
	}
	// 先从 map 中移除（防止新用户加入）
	delete(h.rooms, roomID)
	h.mu.Unlock()

	// ✅ 停止房间并刷盘（StopWithReason 是阻塞的）
	room.StopWithReason(ErrPageDeleted, "页面已被删除")

	log.Printf("[Hub] 💀 强制关闭房间 %s（页面被删除）", roomID)
}
