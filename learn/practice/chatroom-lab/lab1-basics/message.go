package lab1

import (
	"fmt"
	"time"
)

// MessageType 表示消息的类型
// 这是一个自定义类型，底层是 int
type MessageType int

// TODO: Task 1 - 定义消息类型常量
// 使用 const + iota 定义以下 4 个常量：
//   MsgJoin   = 0  (用户加入房间)
//   MsgLeave  = 1  (用户离开房间)
//   MsgChat   = 2  (聊天消息)
//   MsgSystem = 3  (系统消息)
//
// 参考: https://gobyexample-cn.github.io/constants
//
// 在下方编写你的代码:
const (
	MsgJoin MessageType = iota
	MsgLeave
	MsgChat
	MsgSystem
)

// String 返回 MessageType 的可读名称
// 实现 fmt.Stringer 接口
//
// TODO: Task 2 - 使用 switch 语句返回对应的字符串
//   MsgJoin   -> "JOIN"
//   MsgLeave  -> "LEAVE"
//   MsgChat   -> "CHAT"
//   MsgSystem -> "SYSTEM"
//   default   -> "UNKNOWN"
//
// 参考: https://gobyexample-cn.github.io/switch
func (t MessageType) String() string {
	// 在这里编写你的代码...
	switch t {
	case MsgJoin:
		return "JOIN"
	case MsgLeave:
		return "LEAVE"
	case MsgChat:
		return "CHAT"
	case MsgSystem:
		return "SYSTEM"
	default:
		return "UNKNOWN"
	}
}

// Message 表示一条聊天消息
// 这个结构体已经定义好了，你不需要修改
type Message struct {
	Type      MessageType // 消息类型
	Username  string      // 发送者用户名
	Content   string      // 消息内容
	Timestamp time.Time   // 发送时间
}

// NewMessage 创建一条新消息
//
// TODO: Task 3 - 返回一个 *Message 指针
// 需要设置所有字段：
//   - Type: 使用参数 msgType
//   - Username: 使用参数 username
//   - Content: 使用参数 content
//   - Timestamp: 使用 time.Now() 获取当前时间
//
// 参考: https://gobyexample-cn.github.io/structs
// 参考: https://gobyexample-cn.github.io/pointers
func NewMessage(msgType MessageType, username, content string) *Message {
	// 在这里编写你的代码...
	message := Message{Type: msgType, Username: username, Content: content, Timestamp: time.Now()}

	return &message
}

// Format 返回格式化的消息字符串
//
// TODO: Task 4 - 根据消息类型返回不同格式
//   MsgJoin:   ">> {Username} joined"     例如: ">> Alice joined"
//   MsgLeave:  "<< {Username} left"       例如: "<< Alice left"
//   MsgChat:   "{Username}: {Content}"    例如: "Alice: Hello!"
//   MsgSystem: "[System] {Content}"       例如: "[System] Welcome"
//
// 提示: 使用 fmt.Sprintf 格式化字符串
// 参考: https://gobyexample-cn.github.io/methods
func (m *Message) Format() string {
	// 在这里编写你的代码...
	switch m.Type {
	case MsgJoin:
		return fmt.Sprintf(">> %s joined", m.Username)
	case MsgLeave:
		return fmt.Sprintf("<< %s left", m.Username)
	case MsgChat:
		return fmt.Sprintf("%s: %s", m.Username, m.Content)
	case MsgSystem:
		return fmt.Sprintf("[System] %s", m.Content)
	default:
		return ""
	}
}
