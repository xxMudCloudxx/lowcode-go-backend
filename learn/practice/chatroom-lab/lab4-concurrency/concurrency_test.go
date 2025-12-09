package lab4

import (
	"testing"
	"time"
)

// ============================================================
// 测试 Task 1: RunAsync
// ============================================================

func TestRunAsync(t *testing.T) {
	counter := &SafeCounter{}

	start := time.Now()
	RunAsync(
		func() {
			time.Sleep(50 * time.Millisecond)
			counter.Inc()
		},
		func() {
			time.Sleep(50 * time.Millisecond)
			counter.Inc()
		},
		func() {
			time.Sleep(50 * time.Millisecond)
			counter.Inc()
		},
	)
	elapsed := time.Since(start)

	// 检查所有任务都执行了
	if counter.Value() != 3 {
		t.Errorf("counter = %d, expected 3 (tasks didn't run)", counter.Value())
	}

	// 检查是并行执行的（< 150ms 说明是并行）
	if elapsed > 150*time.Millisecond {
		t.Errorf("elapsed = %v, expected < 150ms (tasks should run in parallel)", elapsed)
	}
}

func TestRunAsyncEmpty(t *testing.T) {
	// 不应该 panic
	RunAsync()
}

// ============================================================
// 测试 Task 2: Generator
// ============================================================

func TestGenerator(t *testing.T) {
	ch := Generator(1, 2, 3, 4, 5)
	if ch == nil {
		t.Fatal("Generator returned nil channel")
	}

	var results []int
	for num := range ch {
		results = append(results, num)
	}

	if len(results) != 5 {
		t.Fatalf("received %d numbers, expected 5", len(results))
	}

	expected := []int{1, 2, 3, 4, 5}
	for i, v := range results {
		if v != expected[i] {
			t.Errorf("results[%d] = %d, expected %d", i, v, expected[i])
		}
	}
}

func TestGeneratorEmpty(t *testing.T) {
	ch := Generator()
	if ch == nil {
		t.Fatal("Generator returned nil for empty input")
	}

	count := 0
	for range ch {
		count++
	}

	if count != 0 {
		t.Errorf("received %d numbers from empty Generator, expected 0", count)
	}
}

// ============================================================
// 测试 Task 3: PingPong
// ============================================================

func TestPingPong(t *testing.T) {
	result := PingPong(2)
	if result == nil {
		t.Fatal("PingPong returned nil")
	}

	expected := []string{"ping", "pong", "ping", "pong"}
	if len(result) != len(expected) {
		t.Fatalf("len(result) = %d, expected %d", len(result), len(expected))
	}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("result[%d] = %q, expected %q", i, v, expected[i])
		}
	}
}

func TestPingPongZero(t *testing.T) {
	result := PingPong(0)
	if len(result) != 0 {
		t.Errorf("PingPong(0) should return empty slice, got %d items", len(result))
	}
}

// ============================================================
// 测试 Task 4: TaskQueue
// ============================================================

func TestTaskQueue(t *testing.T) {
	queue := CreateTaskQueue(3)
	if queue == nil {
		t.Fatal("CreateTaskQueue returned nil")
	}

	// 检查缓冲区大小
	if cap(queue) != 3 {
		t.Errorf("queue capacity = %d, expected 3", cap(queue))
	}

	counter := &SafeCounter{}

	// 添加任务
	queue <- func() { counter.Inc() }
	queue <- func() { counter.Inc() }
	queue <- func() { counter.Inc() }
	close(queue)

	// 处理任务
	ProcessQueue(queue)

	if counter.Value() != 3 {
		t.Errorf("counter = %d, expected 3", counter.Value())
	}
}

func TestTaskQueueNonBlocking(t *testing.T) {
	queue := CreateTaskQueue(2)

	// 非阻塞发送到缓冲通道
	done := make(chan bool)
	go func() {
		queue <- func() {}
		queue <- func() {}
		done <- true
	}()

	select {
	case <-done:
		// OK, 非阻塞完成
	case <-time.After(100 * time.Millisecond):
		t.Error("sending to buffered channel blocked (buffer should allow 2 items)")
	}

	close(queue)
}
