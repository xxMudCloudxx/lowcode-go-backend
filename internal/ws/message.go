package ws

import "encoding/json"

type MessageType string

const (
	// 核心协同消息
	TypeOpPatch    MessageType = "op-patch"		// 增量编辑补丁
	TypeCursorMove MessageType = "cursor-move"	// 光标位置同步

	// 系统消息
	TypeUserJoin   MessageType = "user-join"	// 用户加入房间
	TypeUserLeave  MessageType = "user-leave"	// 用户离开房间
	TypeSync       MessageType = "sync"			// 全量同步（用于新用户加入）
	TypeAck        MessageType = "ack"			// 确认消息
	TypeError      MessageType = "error"		// 错误消息
)

// WSMessage 统一的 WebSocket 消息结构
type WSMessage struct {
	Type      MessageType     `json:"type"`		// 消息类型
	SenderID  string          `json:"senderId"`	// 发送者id
	Payload   json.RawMessage `json:"payload"`	// 消息内容(补丁)
	Timestamp int64           `json:"ts"`		// 时间戳
}

// SyncPayload sync 消息的 payload（新用户加入时发送
type SyncPayload struct {
	Schema  json.RawMessage `json:"schema"`
	Version int64           `json:"version"`
	Users   []UserInfo      `json:"users"`
}

// UserInfo 用户基础信息
type UserInfo struct {
	UserID   string `json:"userId"`
	UserName string `json:"userName"`
	Color    string `json:"color,omitempty"`
}