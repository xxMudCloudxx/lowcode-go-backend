package ws

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	Hub      *Hub
	Conn     *websocket.Conn
	RoomID   string
	UserInfo UserInfo
	Room     *Room // 持有 Room 引用，方便访问
	send     chan []byte
}

func NewClient(hub *Hub, conn *websocket.Conn, roomID string, userInfo UserInfo) *Client {
	return &Client{
		Hub:      hub,
		Conn:     conn,
		RoomID:   roomID,
		UserInfo: userInfo,
		send:     make(chan []byte, 256),
	}
}

func (c *Client) WritePump() {
	for message := range c.send {
		// 把消息写入 WebSocket 发给前端
		c.Conn.WriteMessage(websocket.TextMessage, message)
	}
}

func (c *Client) ReadPump() {
	for {
		// 阻塞等待前端发消息
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break // 连接断开
		}

		// 根据消息类型处理
		var msg WSMessage
		json.Unmarshal(message, &msg)

		switch msg.Type {
		case TypeOpPatch:
			c.handleOpPatch(message)
			// TODO：设计好后取消注释
			// case TypeCursorMove:
			// 	c.handleCursorMove(message)
		}
	}
}

// TODO：重新设计
func (c *Client) handleOpPatch(message []byte) {
	var wsMsg WSMessage
	json.Unmarshal(message, &wsMsg)

	var patchPayload struct {
		Patches json.RawMessage `json:"patches"` // RFC 6902 格式的 Patch 数组
		Version int64           `json:"version"`
	}
	json.Unmarshal(wsMsg.Payload, &patchPayload)

	// 1. 获取房间
	room := c.Hub.GetRoom(c.RoomID)
	if room == nil {
		return
	}

	// 2. 版本冲突检测（乐观锁）
	if patchPayload.Version != room.Version {
		// 版本不一致，拒绝或尝试合并
		c.sendError(ErrVersionConflict, "版本冲突，请刷新")
		return
	}

	// 3. 核心：应用 Patch 到内存状态
	if err := room.ApplyPatch(patchPayload.Patches); err != nil {
		log.Printf("[Client] Patch 应用失败: %v", err)
		c.sendError(ErrPatchFailed, err.Error())
		return
	}

	// 4. 广播给房间内其他人
	c.Hub.Broadcast(c.RoomID, message, c)

	log.Printf("[Client] ✅ 用户 [%s] Patch 已应用，新版本: %d",
		c.UserInfo.UserName, room.Version)
}

// TODP: 需要重新设计
// handleCursorMove 处理光标移动消息
// 光标移动不需要服务器处理，直接广播给房间内其他用户
// func (c *Client) handleCursorMove(message []byte) {
// 	// 直接广播给房间内其他人（不包括自己）
// 	c.Hub.Broadcast(c.RoomID, message, c)

// 	log.Printf("[Client] 光标移动: 用户 [%s] 在房间 [%s]", c.UserInfo.UserName, c.RoomID)
// }

// sendError 发送结构化错误消息
// code: 错误码（前端用于判断逻辑分支）
// message: 错误描述（用于调试/日志）
func (c *Client) sendError(code ErrorCode, message string) {
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
	c.send <- data
}

// TODO List:
// 补丁应用和光标移动的message如何处理？补丁应用的message结构已经设计好了，那光标移动呢？
// 为什么handleOpPatch的message里的Payload会有version？这是前端处理发的还是后端接收到后处理的？
// 前端要如何处理后端发送的错误消息？
// 版本冲突检测（乐观锁）太简单了，需要重新设计！
// 需要重新审查下广播的设计
