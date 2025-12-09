package lab5

import (
	"testing"
	"time"
)

// ============================================================
// 测试 Task 1: Merge
// ============================================================

func TestMerge(t *testing.T) {
	ch1 := make(chan int, 3)
	ch2 := make(chan int, 3)
	ch3 := make(chan int, 3)

	ch1 <- 1
	ch1 <- 2
	close(ch1)

	ch2 <- 3
	ch2 <- 4
	close(ch2)

	ch3 <- 5
	close(ch3)

	merged := Merge(ch1, ch2, ch3)
	if merged == nil {
		t.Fatal("Merge returned nil")
	}

	var results []int
	for v := range merged {
		results = append(results, v)
	}

	if len(results) != 5 {
		t.Errorf("received %d items, expected 5", len(results))
	}

	// 检查所有值都收到了（顺序可能不同）
	sum := 0
	for _, v := range results {
		sum += v
	}
	if sum != 15 { // 1+2+3+4+5
		t.Errorf("sum = %d, expected 15", sum)
	}
}

func TestMergeEmpty(t *testing.T) {
	merged := Merge()
	if merged == nil {
		t.Fatal("Merge returned nil for empty input")
	}

	count := 0
	for range merged {
		count++
	}
	if count != 0 {
		t.Errorf("received %d items from empty Merge, expected 0", count)
	}
}

// ============================================================
// 测试 Task 2: WithTimeout
// ============================================================

func TestWithTimeoutSuccess(t *testing.T) {
	ch := make(chan int, 1)
	ch <- 42

	value, ok := WithTimeout(ch, 100*time.Millisecond)
	if !ok {
		t.Error("WithTimeout returned ok=false, expected true")
	}
	if value != 42 {
		t.Errorf("value = %d, expected 42", value)
	}
}

func TestWithTimeoutExpired(t *testing.T) {
	ch := make(chan int) // 无缓冲，不会有数据

	start := time.Now()
	value, ok := WithTimeout(ch, 50*time.Millisecond)
	elapsed := time.Since(start)

	if ok {
		t.Error("WithTimeout returned ok=true, expected false (timeout)")
	}
	if value != 0 {
		t.Errorf("value = %d, expected 0", value)
	}
	if elapsed < 50*time.Millisecond {
		t.Errorf("returned too early: %v", elapsed)
	}
	if elapsed > 100*time.Millisecond {
		t.Errorf("took too long: %v", elapsed)
	}
}

// ============================================================
// 测试 Task 3: TrySend / TryReceive
// ============================================================

func TestTrySend(t *testing.T) {
	t.Run("buffered channel has space", func(t *testing.T) {
		ch := make(chan int, 1)
		ok := TrySend(ch, 42)
		if !ok {
			t.Error("TrySend returned false for empty buffered channel")
		}
	})

	t.Run("buffered channel is full", func(t *testing.T) {
		ch := make(chan int, 1)
		ch <- 1 // 填满

		ok := TrySend(ch, 42)
		if ok {
			t.Error("TrySend returned true for full channel")
		}
	})

	t.Run("unbuffered channel no receiver", func(t *testing.T) {
		ch := make(chan int)
		ok := TrySend(ch, 42)
		if ok {
			t.Error("TrySend returned true for unbuffered channel without receiver")
		}
	})
}

func TestTryReceive(t *testing.T) {
	t.Run("has data", func(t *testing.T) {
		ch := make(chan int, 1)
		ch <- 42

		value, ok := TryReceive(ch)
		if !ok {
			t.Error("TryReceive returned false when data available")
		}
		if value != 42 {
			t.Errorf("value = %d, expected 42", value)
		}
	})

	t.Run("empty channel", func(t *testing.T) {
		ch := make(chan int, 1)

		value, ok := TryReceive(ch)
		if ok {
			t.Error("TryReceive returned true for empty channel")
		}
		if value != 0 {
			t.Errorf("value = %d, expected 0", value)
		}
	})
}

// ============================================================
// 测试 Task 4: Broadcast
// ============================================================

func TestBroadcast(t *testing.T) {
	ch1 := make(chan int, 1)
	ch2 := make(chan int, 1)
	ch3 := make(chan int, 1)
	ch3 <- 0 // 填满

	count := Broadcast(42, ch1, ch2, ch3)

	if count != 2 {
		t.Errorf("Broadcast returned %d, expected 2 (ch3 is full)", count)
	}

	// 验证 ch1 和 ch2 收到了消息
	if v := <-ch1; v != 42 {
		t.Errorf("ch1 received %d, expected 42", v)
	}
	if v := <-ch2; v != 42 {
		t.Errorf("ch2 received %d, expected 42", v)
	}
}

func TestBroadcastEmpty(t *testing.T) {
	count := Broadcast(42)
	if count != 0 {
		t.Errorf("Broadcast to zero channels returned %d, expected 0", count)
	}
}

// ============================================================
// 测试 Task 5: Drain
// ============================================================

func TestDrain(t *testing.T) {
	ch := make(chan int, 5)
	ch <- 1
	ch <- 2
	ch <- 3
	close(ch)

	result := Drain(ch)
	if result == nil {
		t.Fatal("Drain returned nil")
	}
	if len(result) != 3 {
		t.Fatalf("len(result) = %d, expected 3", len(result))
	}

	expected := []int{1, 2, 3}
	for i, v := range result {
		if v != expected[i] {
			t.Errorf("result[%d] = %d, expected %d", i, v, expected[i])
		}
	}
}

func TestDrainEmpty(t *testing.T) {
	ch := make(chan int)
	close(ch)

	result := Drain(ch)
	if result == nil {
		t.Fatal("Drain returned nil for closed empty channel")
	}
	if len(result) != 0 {
		t.Errorf("len(result) = %d, expected 0", len(result))
	}
}
