package lab3

import (
	"errors"
	"time"
)

// --- 预定义错误 ---

var (
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
)

// --- 消息类型 (从之前 Lab 引入) ---

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

// ============================================================
// Part 1: MessageHistory - 消息历史记录
// ============================================================

// MessageHistory 存储消息历史
type MessageHistory struct {
	messages []*Message // 消息切片
	maxSize  int        // 最大容量
}

// NewMessageHistory 创建消息历史实例
func NewMessageHistory(maxSize int) *MessageHistory {
	return &MessageHistory{
		messages: make([]*Message, 0, maxSize),
		maxSize:  maxSize,
	}
}

// Add 添加消息到历史记录
//
// TODO: Task 1a - 实现添加消息
//
// 规则:
//   1. 将 msg 添加到 messages 切片末尾
//   2. 如果长度超过 maxSize，删除最早的消息(第一个)
//
// 提示:
//   - 使用 append(slice, item) 添加元素
//   - 使用 slice[1:] 删除第一个元素
//
// 参考: https://gobyexample-cn.github.io/slices
func (h *MessageHistory) Add(msg *Message) {
	// 在这里编写你的代码...
	h.messages = append(h.messages, msg)
	if (len(h.messages) > h.maxSize) {
		h.messages = h.messages[1:]
	}
}

// GetRecent 获取最近 n 条消息
//
// TODO: Task 1b - 返回最近的消息
//
// 规则:
//   - 返回最近的 n 条消息
//   - 如果 n > len(messages)，返回全部
//   - 如果 n <= 0，返回空切片
//
// 提示:
//   - 使用切片语法 slice[start:end]
//   - 最近的消息在切片末尾
//
// 参考: https://gobyexample-cn.github.io/slices
func (h *MessageHistory) GetRecent(n int) []*Message {
	// 在这里编写你的代码...
	if (n > h.Len()) {
		return h.messages;
	}
	if (n <= 0) {
		return h.messages[:0]
	}

	return h.messages[h.Len() - n:]
}

// Len 返回消息数量
func (h *MessageHistory) Len() int {
	return len(h.messages)
}

// Clear 清空所有消息
//
// TODO: Task 1c - 清空消息历史
//
// 提示: 重新赋值一个空切片，或使用 slice[:0]
func (h *MessageHistory) Clear() {
	// 在这里编写你的代码...
	h.messages = h.messages[:0]
}

// ============================================================
// Part 2: UserRegistry - 用户注册表
// ============================================================

// User 用户信息
type User struct {
	ID       string
	Username string
	JoinTime time.Time
}

// UserRegistry 用户注册表
type UserRegistry struct {
	users map[string]*User
}

// NewUserRegistry 创建用户注册表
func NewUserRegistry() *UserRegistry {
	return &UserRegistry{
		users: make(map[string]*User),
	}
}

// Register 注册新用户
//
// TODO: Task 2a - 注册用户到 Map
//
// 规则:
//   - 如果 id 已存在，返回 ErrUserExists
//   - 否则创建新用户并添加到 map，返回 nil
//   - JoinTime 使用 time.Now()
//
// 提示:
//   - 使用 _, exists := map[key] 检查 key 是否存在
//   - 使用 map[key] = value 添加/更新
//
// 参考: https://gobyexample-cn.github.io/maps
func (r *UserRegistry) Register(id, username string) error {
	// 在这里编写你的代码...
	_, exists := r.users[id];
	if (exists) {
		return ErrUserExists
	}
	newUseer := User{
		ID: id,
		Username: username,
		JoinTime: time.Now(),
	}
	r.users[id] = &newUseer
	return nil
}

// Unregister 注销用户
//
// TODO: Task 2b - 从 Map 删除用户
//
// 规则:
//   - 如果 id 不存在，返回 ErrUserNotFound
//   - 否则删除用户，返回 nil
//
// 提示: 使用 delete(map, key)
func (r *UserRegistry) Unregister(id string) error {
	// 在这里编写你的代码...
		_, exists := r.users[id];
	if (!exists) {
		return ErrUserNotFound
	}
	delete(r.users, id)
	return nil
}

// GetUser 获取用户
//
// TODO: Task 2c - 从 Map 获取用户
//
// 返回:
//   - user: 用户指针 (不存在则为 nil)
//   - exists: 是否存在
//
// 提示: value, exists := map[key]
func (r *UserRegistry) GetUser(id string) (*User, bool) {
	// 在这里编写你的代码...
	_, exists := r.users[id]
	if (!exists) {
		return nil, false
	}
	user := r.users[id]
	return user, true
}

// Count 返回用户总数
//
// TODO: Task 2d - 返回 Map 长度
func (r *UserRegistry) Count() int {
	// 在这里编写你的代码...
	return len(r.users)
}

// GetUsernames 获取所有用户名
//
// TODO: Task 3 - 遍历 Map 收集所有用户名
//
// 返回所有用户的 Username 组成的切片
//
// 提示: 使用 for key, value := range map { ... }
//
// 参考: https://gobyexample-cn.github.io/range
func (r *UserRegistry) GetUsernames() []string {
	// 在这里编写你的代码...
	result := []string{}

	for _, user:= range r.users {
		result = append(result, user.Username)
	}
	return result
}

// ============================================================
// Part 3: 消息过滤
// ============================================================

// FilterMessages 按类型过滤消息
//
// TODO: Task 4 - 使用 Range 过滤消息
//
// 参数:
//   - messages: 消息切片
//   - msgType: 要过滤的类型
//
// 返回:
//   - 所有 Type == msgType 的消息
//
// 参考: https://gobyexample-cn.github.io/range
func FilterMessages(messages []*Message, msgType MessageType) []*Message {
	// 在这里编写你的代码..
	result := []*Message{}
	for _, user := range messages {
		if user.Type == msgType {
			result = append(result, user)
		}
	}
	return result
}
