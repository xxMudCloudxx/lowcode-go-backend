package lab2

import (
	"errors"
	"testing"
)

// ============================================================
// 测试 Task 1: ParseCommand
// ============================================================

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedCmd string
		expectedArg string
		expectedErr error
	}{
		{"join with room", "/join room1", "join", "room1", nil},
		{"name with value", "/name Alice", "name", "Alice", nil},
		{"quit no args", "/quit", "quit", "", nil},
		{"help no args", "/help", "help", "", nil},
		{"join with space", "/join my room", "join", "my room", nil},
		{"empty input", "", "", "", ErrEmptyInput},
		{"not a command", "hello world", "", "", ErrNotCommand},
		{"just text", "message", "", "", ErrNotCommand},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, args, err := ParseCommand(tt.input)

			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("ParseCommand(%q) error = %v, expected %v",
					tt.input, err, tt.expectedErr)
			}
			if cmd != tt.expectedCmd {
				t.Errorf("ParseCommand(%q) cmd = %q, expected %q",
					tt.input, cmd, tt.expectedCmd)
			}
			if args != tt.expectedArg {
				t.Errorf("ParseCommand(%q) args = %q, expected %q",
					tt.input, args, tt.expectedArg)
			}
		})
	}
}

// ============================================================
// 测试 Task 2: ValidateUsername
// ============================================================

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name        string
		username    string
		expectedErr error
	}{
		{"valid short", "Al", nil},
		{"valid normal", "Alice", nil},
		{"valid long", "VeryLongUsername123", nil},
		{"too short", "A", ErrNameTooShort},
		{"empty", "", ErrNameTooShort},
		{"too long", "ThisUsernameIsWayTooLongToBeValid", ErrNameTooLong},
		{"has space", "John Doe", ErrNameHasSpace},
		{"only space", " ", ErrNameHasSpace},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUsername(tt.username)
			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("ValidateUsername(%q) = %v, expected %v",
					tt.username, err, tt.expectedErr)
			}
		})
	}
}

// ============================================================
// 测试 Task 3: FormatMessages
// ============================================================

func TestFormatMessages(t *testing.T) {
	msg1 := &Message{Type: MsgChat, Username: "Alice", Content: "Hi"}
	msg2 := &Message{Type: MsgChat, Username: "Bob", Content: "Hello"}
	msg3 := &Message{Type: MsgJoin, Username: "Charlie"}

	t.Run("multiple messages", func(t *testing.T) {
		result := FormatMessages("> ", msg1, msg2)
		if result == nil {
			t.Fatal("FormatMessages returned nil")
		}
		if len(result) != 2 {
			t.Fatalf("len(result) = %d, expected 2", len(result))
		}
		if result[0] != "> Alice: Hi" {
			t.Errorf("result[0] = %q, expected %q", result[0], "> Alice: Hi")
		}
		if result[1] != "> Bob: Hello" {
			t.Errorf("result[1] = %q, expected %q", result[1], "> Bob: Hello")
		}
	})

	t.Run("single message", func(t *testing.T) {
		result := FormatMessages(">> ", msg3)
		if len(result) != 1 {
			t.Fatalf("len(result) = %d, expected 1", len(result))
		}
		if result[0] != ">> >> Charlie joined" {
			t.Errorf("result[0] = %q, expected %q", result[0], ">> >> Charlie joined")
		}
	})

	t.Run("no messages", func(t *testing.T) {
		result := FormatMessages("prefix")
		if result == nil {
			t.Fatal("FormatMessages returned nil, expected empty slice")
		}
		if len(result) != 0 {
			t.Errorf("len(result) = %d, expected 0", len(result))
		}
	})
}

// ============================================================
// 测试 Task 4: CreateCounter
// ============================================================

func TestCreateCounter(t *testing.T) {
	t.Run("start from 0", func(t *testing.T) {
		counter := CreateCounter(0)
		if counter == nil {
			t.Fatal("CreateCounter returned nil")
		}

		if v := counter(); v != 1 {
			t.Errorf("first call = %d, expected 1", v)
		}
		if v := counter(); v != 2 {
			t.Errorf("second call = %d, expected 2", v)
		}
		if v := counter(); v != 3 {
			t.Errorf("third call = %d, expected 3", v)
		}
	})

	t.Run("start from 10", func(t *testing.T) {
		counter := CreateCounter(10)
		if v := counter(); v != 11 {
			t.Errorf("first call = %d, expected 11", v)
		}
	})

	t.Run("independent counters", func(t *testing.T) {
		c1 := CreateCounter(0)
		c2 := CreateCounter(100)

		c1() // 1
		c1() // 2
		v1 := c1() // 3

		v2 := c2() // 101

		if v1 != 3 {
			t.Errorf("c1 third call = %d, expected 3", v1)
		}
		if v2 != 101 {
			t.Errorf("c2 first call = %d, expected 101", v2)
		}
	})
}

// ============================================================
// 测试 Task 5: Fibonacci
// ============================================================

func TestFibonacci(t *testing.T) {
	tests := []struct {
		n        int
		expected int
	}{
		{0, 0},
		{1, 1},
		{2, 1},
		{3, 2},
		{4, 3},
		{5, 5},
		{6, 8},
		{7, 13},
		{10, 55},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := Fibonacci(tt.n)
			if result != tt.expected {
				t.Errorf("Fibonacci(%d) = %d, expected %d",
					tt.n, result, tt.expected)
			}
		})
	}
}
