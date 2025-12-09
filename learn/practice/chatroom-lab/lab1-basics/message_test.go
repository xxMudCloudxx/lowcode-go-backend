package lab1

import (
	"strings"
	"testing"
	"time"
)

// ============================================================
// 测试 Task 1: MessageType 常量
// ============================================================

func TestMessageTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		got      MessageType
		expected int
	}{
		{"MsgJoin should be 0", MsgJoin, 0},
		{"MsgLeave should be 1", MsgLeave, 1},
		{"MsgChat should be 2", MsgChat, 2},
		{"MsgSystem should be 3", MsgSystem, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.got) != tt.expected {
				t.Errorf("got %d, expected %d", tt.got, tt.expected)
			}
		})
	}
}

// ============================================================
// 测试 Task 2: MessageType.String() 方法
// ============================================================

func TestMessageTypeString(t *testing.T) {
	tests := []struct {
		msgType  MessageType
		expected string
	}{
		{MsgJoin, "JOIN"},
		{MsgLeave, "LEAVE"},
		{MsgChat, "CHAT"},
		{MsgSystem, "SYSTEM"},
		{MessageType(99), "UNKNOWN"}, // 未知类型
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.msgType.String()
			if got != tt.expected {
				t.Errorf("MessageType(%d).String() = %q, expected %q",
					tt.msgType, got, tt.expected)
			}
		})
	}
}

// ============================================================
// 测试 Task 3: NewMessage() 函数
// ============================================================

func TestNewMessage(t *testing.T) {
	before := time.Now()
	msg := NewMessage(MsgChat, "Alice", "Hello!")
	after := time.Now()

	// 检查是否返回了非 nil 指针
	if msg == nil {
		t.Fatal("NewMessage returned nil, expected *Message")
	}

	// 检查字段值
	if msg.Type != MsgChat {
		t.Errorf("msg.Type = %v, expected MsgChat", msg.Type)
	}
	if msg.Username != "Alice" {
		t.Errorf("msg.Username = %q, expected %q", msg.Username, "Alice")
	}
	if msg.Content != "Hello!" {
		t.Errorf("msg.Content = %q, expected %q", msg.Content, "Hello!")
	}

	// 检查时间戳是否在合理范围内
	if msg.Timestamp.Before(before) || msg.Timestamp.After(after) {
		t.Errorf("msg.Timestamp = %v, should be between %v and %v",
			msg.Timestamp, before, after)
	}
}

func TestNewMessageDifferentTypes(t *testing.T) {
	tests := []struct {
		msgType  MessageType
		username string
		content  string
	}{
		{MsgJoin, "Bob", ""},
		{MsgLeave, "Charlie", ""},
		{MsgSystem, "", "Server restarting"},
	}

	for _, tt := range tests {
		msg := NewMessage(tt.msgType, tt.username, tt.content)
		if msg == nil {
			t.Fatalf("NewMessage returned nil for type %v", tt.msgType)
		}
		if msg.Type != tt.msgType {
			t.Errorf("msg.Type = %v, expected %v", msg.Type, tt.msgType)
		}
	}
}

// ============================================================
// 测试 Task 4: Message.Format() 方法
// ============================================================

func TestMessageFormat(t *testing.T) {
	tests := []struct {
		name     string
		msg      *Message
		expected string
	}{
		{
			name:     "Join message",
			msg:      &Message{Type: MsgJoin, Username: "Alice"},
			expected: ">> Alice joined",
		},
		{
			name:     "Leave message",
			msg:      &Message{Type: MsgLeave, Username: "Bob"},
			expected: "<< Bob left",
		},
		{
			name:     "Chat message",
			msg:      &Message{Type: MsgChat, Username: "Charlie", Content: "Hello!"},
			expected: "Charlie: Hello!",
		},
		{
			name:     "System message",
			msg:      &Message{Type: MsgSystem, Content: "Welcome"},
			expected: "[System] Welcome",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.msg.Format()
			if got != tt.expected {
				t.Errorf("Format() = %q, expected %q", got, tt.expected)
			}
		})
	}
}

func TestMessageFormatChatWithSpecialContent(t *testing.T) {
	msg := &Message{
		Type:     MsgChat,
		Username: "Dev",
		Content:  "Hello, World! 你好！",
	}
	
	got := msg.Format()
	if !strings.Contains(got, "Dev") || !strings.Contains(got, "Hello") {
		t.Errorf("Format() = %q, should contain username and content", got)
	}
}
