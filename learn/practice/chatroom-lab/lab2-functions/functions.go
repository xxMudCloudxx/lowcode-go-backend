package lab2

import (
	"errors"
	"time"
)

// --- 预定义错误 (不要修改) ---

var (
	ErrNotCommand     = errors.New("input is not a command")
	ErrEmptyInput     = errors.New("input is empty")
	ErrNameTooShort   = errors.New("username too short (min 2 chars)")
	ErrNameTooLong    = errors.New("username too long (max 20 chars)")
	ErrNameHasSpace   = errors.New("username cannot contain spaces")
)

// --- 从 Lab 1 导入的类型 (不要修改) ---

type MessageType int

const (
	MsgJoin MessageType = iota
	MsgLeave
	MsgChat
	MsgSystem
)

type Message struct {
	Type      MessageType
	Username  string
	Content   string
	Timestamp time.Time
}

func (m *Message) Format() string {
	switch m.Type {
	case MsgJoin:
		return ">> " + m.Username + " joined"
	case MsgLeave:
		return "<< " + m.Username + " left"
	case MsgChat:
		return m.Username + ": " + m.Content
	case MsgSystem:
		return "[System] " + m.Content
	default:
		return ""
	}
}

// ============================================================
// TODO: 在下方实现你的函数
// ============================================================

// ParseCommand 解析用户输入的命令
//
// TODO: Task 1 - 解析命令字符串
//
// 输入格式: "/command args" 或 "/command"
// 返回值:
//   - cmd: 命令名 (不含 /)
//   - args: 参数 (如果有)
//   - err: 错误
//
// 规则:
//   - 空字符串 → 返回 ErrEmptyInput
//   - 不以 "/" 开头 → 返回 ErrNotCommand
//   - "/join room1" → cmd="join", args="room1", err=nil
//   - "/quit" → cmd="quit", args="", err=nil
//
// 提示: 使用 strings.TrimPrefix, strings.SplitN
// 参考: https://gobyexample-cn.github.io/multiple-return-values
func ParseCommand(input string) (cmd string, args string, err error) {
	// 在这里编写你的代码...
	return "", "", nil
}

// ValidateUsername 验证用户名是否合法
//
// TODO: Task 2 - 验证用户名
//
// 规则:
//   - 长度 < 2 → 返回 ErrNameTooShort
//   - 长度 > 20 → 返回 ErrNameTooLong
//   - 包含空格 → 返回 ErrNameHasSpace
//   - 通过验证 → 返回 nil
//
// 提示: 使用 len(), strings.Contains
// 参考: https://gobyexample-cn.github.io/errors
func ValidateUsername(name string) error {
	// 在这里编写你的代码...
	return nil
}

// FormatMessages 格式化多条消息
//
// TODO: Task 3 - 使用变参函数处理多条消息
//
// 参数:
//   - prefix: 添加在每条消息前的前缀
//   - messages: 可变数量的消息指针
//
// 返回:
//   - 格式化后的字符串切片
//
// 示例:
//   msg1 := &Message{Type: MsgChat, Username: "A", Content: "Hi"}
//   msg2 := &Message{Type: MsgChat, Username: "B", Content: "Hello"}
//   FormatMessages("> ", msg1, msg2)
//   // 返回 ["> A: Hi", "> B: Hello"]
//
// 参考: https://gobyexample-cn.github.io/variadic-functions
func FormatMessages(prefix string, messages ...*Message) []string {
	// 在这里编写你的代码...
	return nil
}

// CreateCounter 创建一个计数器闭包
//
// TODO: Task 4 - 返回一个闭包函数
//
// 返回的函数每次调用时：
//   1. 将内部计数器加 1
//   2. 返回新的值
//
// 示例:
//   counter := CreateCounter(0)
//   counter() // 返回 1
//   counter() // 返回 2
//   counter() // 返回 3
//
//   counter2 := CreateCounter(10)
//   counter2() // 返回 11
//
// 参考: https://gobyexample-cn.github.io/closures
func CreateCounter(start int) func() int {
	// 在这里编写你的代码...
	return nil
}

// Fibonacci 计算第 n 个斐波那契数
//
// TODO: Task 5 - 使用递归实现
//
// 定义:
//   - Fibonacci(0) = 0
//   - Fibonacci(1) = 1
//   - Fibonacci(n) = Fibonacci(n-1) + Fibonacci(n-2)
//
// 参考: https://gobyexample-cn.github.io/recursion
func Fibonacci(n int) int {
	// 在这里编写你的代码...
	return 0
}
