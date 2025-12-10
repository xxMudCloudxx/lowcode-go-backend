package ws

import "encoding/json"

type MessageType string

const (
	// 核心协同消息
	TypeOpPatch    MessageType = "op-patch"    // 增量编辑补丁
	TypeCursorMove MessageType = "cursor-move" // 光标位置同步

	// 系统消息
	TypeUserJoin  MessageType = "user-join"  // 用户加入房间
	TypeUserLeave MessageType = "user-leave" // 用户离开房间
	TypeSync      MessageType = "sync"       // 全量同步（用于新用户加入）
	TypeAck       MessageType = "ack"        // 确认消息
	TypeError     MessageType = "error"      // 错误消息
)

// WSMessage 统一的 WebSocket 消息结构
type WSMessage struct {
	Type      MessageType     `json:"type"`     // 消息类型
	SenderID  string          `json:"senderId"` // 发送者id
	Payload   json.RawMessage `json:"payload"`  // 消息内容(补丁)
	Timestamp int64           `json:"ts"`       // 时间戳
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

// ========== 错误码系统 ==========
// 前端根据 Code 判断错误类型，而不是匹配 Message 字符串

type ErrorCode string

const (
	ErrVersionConflict ErrorCode = "VERSION_CONFLICT" // 版本冲突
	ErrPatchInvalid    ErrorCode = "PATCH_INVALID"    // Patch 格式错误
	ErrPatchFailed     ErrorCode = "PATCH_FAILED"     // Patch 应用失败
	ErrRoomNotFound    ErrorCode = "ROOM_NOT_FOUND"   // 房间不存在
	ErrUnauthorized    ErrorCode = "UNAUTHORIZED"     // 未授权
	ErrInternalError   ErrorCode = "INTERNAL_ERROR"   // 服务器内部错误
)

// ErrorPayload 错误消息的 payload 结构
type ErrorPayload struct {
	Code    ErrorCode `json:"code"`    // 错误码（前端用于判断逻辑）
	Message string    `json:"message"` // 错误描述（用于调试/日志，可本地化）
}

// ========== 自定义错误类型 ==========
// 使用类型断言判断错误，而非字符串匹配

// VersionConflictError 版本冲突错误
type VersionConflictError struct {
	CurrentVersion  int64
	ExpectedVersion int64
}

func (e *VersionConflictError) Error() string {
	return "version conflict"
}

// PatchError Patch 处理错误（解析或应用失败）
type PatchError struct {
	Reason string
}

func (e *PatchError) Error() string {
	return e.Reason
}
