package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// ========== 心跳配置 ==========
const (
	// pongWait 允许等待 Pong 的最大时间
	pongWait = 60 * time.Second

	// pingPeriod 发送 Ping 的间隔 (必须小于 pongWait)
	pingPeriod = (pongWait * 9) / 10

	// writeWait 写消息的超时时间
	writeWait = 10 * time.Second

	// maxMessageSize 最大消息大小 (防止恶意攻击)
	maxMessageSize = 512 * 1024 // 512KB
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

// WritePump 负责写消息和发送心跳 Ping
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			// 设置写超时
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				// send channel 已关闭，发送关闭帧
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// 写入消息
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			// 定时发送 Ping（保活）
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return // Ping 发送失败，连接已断
			}
		}
	}
}

// ReadPump 负责读消息和处理心跳 Pong
func (c *Client) ReadPump() {
	defer func() {
		// 读循环退出 = 连接断开，通知 Room 注销
		if c.Room != nil {
			c.Room.Unregister(c)
		}
		c.Conn.Close()
	}()

	// 设置最大消息大小（防止恶意攻击）
	c.Conn.SetReadLimit(maxMessageSize)

	// 设置初始读超时
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))

	// 设置 Pong 处理函数：每次收到 Pong 就重置读超时
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[Client] ⚠️ 连接异常关闭: %v", err)
			}
			break
		}

		// 收到消息也重置读超时（客户端活跃）
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))

		// 解析消息类型
		var msg WSMessage
		json.Unmarshal(message, &msg)

		switch msg.Type {
		case TypeOpPatch:
			c.handleOpPatch(message)
		case TypeCursorMove:
			c.handleCursorMove(message)
		}
	}
}

// handleOpPatch 处理 op-patch 消息
// TODO：未来参考Google Doc的协同编辑实现
func (c *Client) handleOpPatch(message []byte) {
	// 检查房间是否存在
	if c.Room == nil {
		c.sendError(ErrRoomNotFound, c.RoomID)
		return
	}

	var wsMsg WSMessage
	json.Unmarshal(message, &wsMsg)

	var patchPayload struct {
		Patches json.RawMessage `json:"patches"`
		Version int64           `json:"version"`
	}
	json.Unmarshal(wsMsg.Payload, &patchPayload)

	// 2. 应用 Patch（版本检查在 ApplyPatch 内部的锁保护下进行）
	if err := c.Room.ApplyPatch(patchPayload.Patches, patchPayload.Version); err != nil {
		// 使用类型断言判断错误类型，而非字符串匹配
		var versionErr *VersionConflictError
		var patchErr *PatchError

		switch {
		case errors.As(err, &versionErr):
			c.sendError(ErrVersionConflict, fmt.Sprintf("current: %d, expected: %d",
				versionErr.CurrentVersion, versionErr.ExpectedVersion))
		case errors.As(err, &patchErr):
			c.sendError(ErrPatchFailed, patchErr.Reason)
		default:
			c.sendError(ErrInternalError, err.Error())
		}
		log.Printf("[Client] Patch 处理失败: %v", err)
		return
	}

	// 3. 广播给房间内其他人（Patch 是关键消息，阻塞时断开连接）
	// 直接找 Room 广播，不经过 Hub，实现去中心化
	c.Room.Broadcast(message, c, true)
	log.Printf("[Client] ✅ 用户 [%s] Patch 已应用，新版本: %d",
		c.UserInfo.UserName, c.Room.Version)
}

// handleCursorMove 处理光标移动消息
// 光标是非关键消息（Ephemeral），阻塞时静默跳过
func (c *Client) handleCursorMove(message []byte) {
	// 直接找 Room 广播，不经过 Hub
	if c.Room != nil {
		c.Room.Broadcast(message, c, false)
	}
}

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
