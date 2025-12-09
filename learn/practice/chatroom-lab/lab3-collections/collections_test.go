package lab3

import (
	"errors"
	"sort"
	"testing"
)

// ============================================================
// 测试 MessageHistory
// ============================================================

func TestMessageHistoryAdd(t *testing.T) {
	h := NewMessageHistory(3)

	h.Add(&Message{Type: MsgChat, Content: "msg1"})
	h.Add(&Message{Type: MsgChat, Content: "msg2"})

	if h.Len() != 2 {
		t.Errorf("Len() = %d, expected 2", h.Len())
	}

	// 添加第三条
	h.Add(&Message{Type: MsgChat, Content: "msg3"})
	if h.Len() != 3 {
		t.Errorf("Len() = %d, expected 3", h.Len())
	}

	// 添加第四条，应该删除最早的
	h.Add(&Message{Type: MsgChat, Content: "msg4"})
	if h.Len() != 3 {
		t.Errorf("Len() = %d after overflow, expected 3", h.Len())
	}

	// 检查最早的消息已被删除
	recent := h.GetRecent(3)
	if recent[0].Content != "msg2" {
		t.Errorf("oldest message should be msg2, got %s", recent[0].Content)
	}
}

func TestMessageHistoryGetRecent(t *testing.T) {
	h := NewMessageHistory(10)
	for i := 1; i <= 5; i++ {
		h.Add(&Message{Type: MsgChat, Content: ""})
	}

	t.Run("get 3 recent", func(t *testing.T) {
		recent := h.GetRecent(3)
		if len(recent) != 3 {
			t.Errorf("len(GetRecent(3)) = %d, expected 3", len(recent))
		}
	})

	t.Run("get more than available", func(t *testing.T) {
		recent := h.GetRecent(10)
		if len(recent) != 5 {
			t.Errorf("len(GetRecent(10)) = %d, expected 5", len(recent))
		}
	})

	t.Run("get 0", func(t *testing.T) {
		recent := h.GetRecent(0)
		if len(recent) != 0 {
			t.Errorf("len(GetRecent(0)) = %d, expected 0", len(recent))
		}
	})

	t.Run("get negative", func(t *testing.T) {
		recent := h.GetRecent(-5)
		if len(recent) != 0 {
			t.Errorf("len(GetRecent(-5)) = %d, expected 0", len(recent))
		}
	})
}

func TestMessageHistoryClear(t *testing.T) {
	h := NewMessageHistory(10)
	h.Add(&Message{})
	h.Add(&Message{})

	h.Clear()

	if h.Len() != 0 {
		t.Errorf("Len() after Clear() = %d, expected 0", h.Len())
	}
}

// ============================================================
// 测试 UserRegistry
// ============================================================

func TestUserRegistryRegister(t *testing.T) {
	r := NewUserRegistry()

	// 正常注册
	err := r.Register("user1", "Alice")
	if err != nil {
		t.Errorf("Register should succeed, got error: %v", err)
	}

	// 检查数量
	if r.Count() != 1 {
		t.Errorf("Count() = %d, expected 1", r.Count())
	}

	// 重复注册
	err = r.Register("user1", "Bob")
	if !errors.Is(err, ErrUserExists) {
		t.Errorf("Register duplicate should return ErrUserExists, got %v", err)
	}
}

func TestUserRegistryUnregister(t *testing.T) {
	r := NewUserRegistry()
	r.Register("user1", "Alice")

	// 正常注销
	err := r.Unregister("user1")
	if err != nil {
		t.Errorf("Unregister should succeed, got error: %v", err)
	}

	if r.Count() != 0 {
		t.Errorf("Count() after Unregister = %d, expected 0", r.Count())
	}

	// 注销不存在的用户
	err = r.Unregister("user999")
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("Unregister non-existent should return ErrUserNotFound, got %v", err)
	}
}

func TestUserRegistryGetUser(t *testing.T) {
	r := NewUserRegistry()
	r.Register("user1", "Alice")

	// 获取存在的用户
	user, exists := r.GetUser("user1")
	if !exists {
		t.Error("GetUser should return exists=true")
	}
	if user == nil {
		t.Fatal("GetUser returned nil user")
	}
	if user.Username != "Alice" {
		t.Errorf("user.Username = %q, expected %q", user.Username, "Alice")
	}

	// 获取不存在的用户
	user, exists = r.GetUser("user999")
	if exists {
		t.Error("GetUser for non-existent should return exists=false")
	}
}

func TestUserRegistryGetUsernames(t *testing.T) {
	r := NewUserRegistry()
	r.Register("1", "Alice")
	r.Register("2", "Bob")
	r.Register("3", "Charlie")

	names := r.GetUsernames()
	if names == nil {
		t.Fatal("GetUsernames returned nil")
	}
	if len(names) != 3 {
		t.Fatalf("len(names) = %d, expected 3", len(names))
	}

	// 排序后比较（因为 map 遍历顺序不确定）
	sort.Strings(names)
	expected := []string{"Alice", "Bob", "Charlie"}
	for i, name := range names {
		if name != expected[i] {
			t.Errorf("names[%d] = %q, expected %q", i, name, expected[i])
		}
	}
}

// ============================================================
// 测试 FilterMessages
// ============================================================

func TestFilterMessages(t *testing.T) {
	messages := []*Message{
		{Type: MsgJoin, Username: "Alice"},
		{Type: MsgChat, Username: "Alice", Content: "Hi"},
		{Type: MsgChat, Username: "Bob", Content: "Hello"},
		{Type: MsgLeave, Username: "Alice"},
		{Type: MsgSystem, Content: "Welcome"},
	}

	t.Run("filter chat messages", func(t *testing.T) {
		result := FilterMessages(messages, MsgChat)
		if result == nil {
			t.Fatal("FilterMessages returned nil")
		}
		if len(result) != 2 {
			t.Fatalf("len(result) = %d, expected 2", len(result))
		}
	})

	t.Run("filter join messages", func(t *testing.T) {
		result := FilterMessages(messages, MsgJoin)
		if len(result) != 1 {
			t.Errorf("len(result) = %d, expected 1", len(result))
		}
	})

	t.Run("filter non-existent type", func(t *testing.T) {
		result := FilterMessages(messages, MessageType(99))
		if result == nil {
			t.Fatal("FilterMessages returned nil for empty result")
		}
		if len(result) != 0 {
			t.Errorf("len(result) = %d, expected 0", len(result))
		}
	})

	t.Run("empty input", func(t *testing.T) {
		result := FilterMessages([]*Message{}, MsgChat)
		if len(result) != 0 {
			t.Errorf("len(result) = %d, expected 0", len(result))
		}
	})
}
