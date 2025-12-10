package ws

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

// BroadcastMessage 广播消息结构
type BroadcastMessage struct {
	RoomID  string
	Message []byte
	Sender  *Client
}

// Hub 维护所有活跃房间和客户端连接
type Hub struct {
	// 房间映射改为 map[string]*Room
	// 每个 Room 维护自己的 CurrentState
	rooms     map[string]*Room
	listeners map[*Client]bool

	// Channel 事件通道
	register   chan *Client
	unregister chan *Client
	broadcast  chan *BroadcastMessage

	mu sync.RWMutex
	wg sync.WaitGroup
	// 数据库服务（用于加载初始状态）
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
		listeners:   make(map[*Client]bool),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan *BroadcastMessage, 256),
		pageService: pageService,
	}
}

// Run 启动 Hub 事件循环
func (h *Hub) Run() {
	log.Println("[Hub] 🚀 Hub 事件循环已启动")
	for {
		select {
		case client := <-h.register:
			h.handleRegister(client)
		case client := <-h.unregister:
			h.handleUnregister(client)
		case msg := <-h.broadcast:
			h.handleBroadcast(msg)
		}
	}

}

// handleRegister 处理客户端加入
func (h *Hub) handleRegister(client *Client) {
	// 将客户端加入房间
	roomID := client.RoomID

	h.mu.Lock()
	room, exists := h.rooms[roomID]

	if !exists {
		state, version, err := h.pageService.GetPageState(roomID)

		if err != nil {
			log.Printf("[Hub] ⚠️ 加载页面失败: %v", err)
			state = []byte(`{"rootd":1, "components":{1: {id: 1, name: "Page", props: {}, desc: "页面", parentId: null}}}`)
			version = 1
		}
		room = NewRoom(roomID, state, h.pageService)
		room.Version = version
		h.rooms[roomID] = room
		h.wg.Add(1)
		log.Printf("[Hub]创建房间: %s", roomID)
	}
	h.mu.Unlock()

	room.mu.Lock()
	room.Clients[client] = true
	room.mu.Unlock()
	client.Room = room
	// 发送最新快照给新用户
	h.sendSyncMessage(client, room)
}

// sendSyncMessage 发送全量同步消息给新用户
func (h *Hub) sendSyncMessage(client *Client, room *Room) {
	snapshot, version := room.GetSnapshot()

	room.mu.RLock()
	// 收集房间内其他用户信息
	users := make([]UserInfo, 0, len(room.Clients))
	for c := range room.Clients {
		if c != client {
			users = append(users, c.UserInfo)
		}
	}
	room.mu.RUnlock()

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

	log.Printf("[Hub] 📤 已发送 Sync 消息给 [%s], 版本: %d",
		client.UserInfo.UserName, version)
}

// handleUnregister 处理客户端离开
func (h *Hub) handleUnregister(client *Client) {
	room := client.Room
	if room == nil {
		return
	}

	delete(room.Clients, client)
	close(client.send)

	// ⚠️ 房间空了，必须善后 + 加写锁
	if len(room.Clients) == 0 {
		room.Stop() // 停止 Goroutine

		h.mu.Lock()
		delete(h.rooms, room.ID)
		h.mu.Unlock()

		h.wg.Done() // 计数减一
		log.Printf("[Hub] 🗑️ 房间 %s 已销毁", room.ID)
	}
}

func (h *Hub) GetRoom(roomID string) *Room {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.rooms[roomID]
}

// handleBroadcast 处理广播消息
func (h *Hub) handleBroadcast(msg *BroadcastMessage) {
	h.mu.RLock()
	room := h.rooms[msg.RoomID]
	h.mu.RUnlock()
	if room == nil {
		return
	}

	for client := range room.Clients {
		if msg.Sender != nil && client == msg.Sender {
			continue
		}

		select {
		case client.send <- msg.Message:
		default:
			close(client.send)
			delete(room.Clients, client)
		}
	}
}

// Broadcast 外部调用接口
func (h *Hub) Broadcast(roomID string, message []byte, sender *Client) {
	h.broadcast <- &BroadcastMessage{
		RoomID:  roomID,
		Message: message,
		Sender:  sender,
	}
}
