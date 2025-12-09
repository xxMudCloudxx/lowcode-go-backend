package lab6

import (
	"sync"
	"testing"
	"time"
)

// ============================================================
// 测试 Task 1: Counter
// ============================================================

func TestCounterInc(t *testing.T) {
	c := &Counter{}
	c.Inc()
	c.Inc()
	c.Inc()

	if c.Value() != 3 {
		t.Errorf("Value() = %d, expected 3", c.Value())
	}
}

func TestCounterAdd(t *testing.T) {
	c := &Counter{}
	c.Add(10)
	c.Add(5)

	if c.Value() != 15 {
		t.Errorf("Value() = %d, expected 15", c.Value())
	}
}

func TestCounterConcurrent(t *testing.T) {
	c := &Counter{}
	var wg sync.WaitGroup

	// 100 个协程，每个加 100 次
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				c.Inc()
			}
		}()
	}

	wg.Wait()

	if c.Value() != 10000 {
		t.Errorf("Value() = %d, expected 10000 (data race?)", c.Value())
	}
}

// ============================================================
// 测试 Task 2: SafeMap
// ============================================================

func TestSafeMapBasic(t *testing.T) {
	m := NewSafeMap()

	m.Set("key1", "value1")
	m.Set("key2", 42)

	v1, ok := m.Get("key1")
	if !ok || v1 != "value1" {
		t.Errorf("Get(key1) = %v, %v, expected value1, true", v1, ok)
	}

	v2, ok := m.Get("key2")
	if !ok || v2 != 42 {
		t.Errorf("Get(key2) = %v, %v, expected 42, true", v2, ok)
	}

	_, ok = m.Get("nonexistent")
	if ok {
		t.Error("Get(nonexistent) should return false")
	}
}

func TestSafeMapDelete(t *testing.T) {
	m := NewSafeMap()
	m.Set("key", "value")

	m.Delete("key")

	_, ok := m.Get("key")
	if ok {
		t.Error("key should be deleted")
	}
}

func TestSafeMapLen(t *testing.T) {
	m := NewSafeMap()
	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)

	if m.Len() != 3 {
		t.Errorf("Len() = %d, expected 3", m.Len())
	}
}

func TestSafeMapConcurrent(t *testing.T) {
	m := NewSafeMap()
	var wg sync.WaitGroup

	// 并发写入
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := string(rune('a' + i%26))
			m.Set(key, i)
		}(i)
	}

	// 并发读取
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.Get("a")
			m.Len()
		}()
	}

	wg.Wait()
	// 如果没有 panic，测试通过
}

// ============================================================
// 测试 Task 3: WorkerPool
// ============================================================

func TestWorkerPool(t *testing.T) {
	pool := NewWorkerPool(3, 10)
	pool.Start()

	counter := &Counter{}

	for i := 0; i < 10; i++ {
		pool.Submit(func() {
			counter.Inc()
		})
	}

	pool.Stop()

	if counter.Value() != 10 {
		t.Errorf("counter = %d, expected 10", counter.Value())
	}
}

func TestWorkerPoolParallel(t *testing.T) {
	pool := NewWorkerPool(3, 10)
	pool.Start()

	counter := &Counter{}

	// 提交 3 个需要 50ms 的任务
	start := time.Now()
	for i := 0; i < 3; i++ {
		pool.Submit(func() {
			time.Sleep(50 * time.Millisecond)
			counter.Inc()
		})
	}

	pool.Stop()
	elapsed := time.Since(start)

	if counter.Value() != 3 {
		t.Errorf("counter = %d, expected 3", counter.Value())
	}

	// 3 个工人并行执行，应该约 50ms 而不是 150ms
	if elapsed > 100*time.Millisecond {
		t.Errorf("elapsed = %v, should be ~50ms (not parallel?)", elapsed)
	}
}

// ============================================================
// 测试 Task 4: RateLimiter
// ============================================================

func TestRateLimiterAllow(t *testing.T) {
	r := NewRateLimiter(3)

	// 前 3 次应该成功
	if !r.Allow() {
		t.Error("Allow() should return true (1/3)")
	}
	if !r.Allow() {
		t.Error("Allow() should return true (2/3)")
	}
	if !r.Allow() {
		t.Error("Allow() should return true (3/3)")
	}

	// 第 4 次应该失败
	if r.Allow() {
		t.Error("Allow() should return false (no tokens left)")
	}

	if r.Tokens() != 0 {
		t.Errorf("Tokens() = %d, expected 0", r.Tokens())
	}
}

func TestRateLimiterRefill(t *testing.T) {
	r := NewRateLimiter(2)

	r.Allow() // tokens: 1
	r.Allow() // tokens: 0

	r.Refill() // tokens: 1

	if r.Tokens() != 1 {
		t.Errorf("Tokens() = %d, expected 1", r.Tokens())
	}

	if !r.Allow() {
		t.Error("Allow() should return true after Refill")
	}
}

func TestRateLimiterRefillMax(t *testing.T) {
	r := NewRateLimiter(2)

	// 尝试超过最大值
	r.Refill()
	r.Refill()
	r.Refill()

	if r.Tokens() > 2 {
		t.Errorf("Tokens() = %d, should not exceed max (2)", r.Tokens())
	}
}

func TestRateLimiterConcurrent(t *testing.T) {
	r := NewRateLimiter(100)
	var wg sync.WaitGroup

	allowed := &Counter{}

	// 200 个协程同时请求
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if r.Allow() {
				allowed.Inc()
			}
		}()
	}

	wg.Wait()

	if allowed.Value() > 100 {
		t.Errorf("allowed %d requests, should be <= 100", allowed.Value())
	}
}
