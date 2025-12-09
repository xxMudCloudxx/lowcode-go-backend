package lab6

import (
	"sync"
	"sync/atomic"
)

// ============================================================
// Task 1: 原子计数器
// ============================================================

// Counter 线程安全的计数器
type Counter struct {
	value int64
}

// Inc 原子加 1
//
// TODO: Task 1a - 使用 atomic.AddInt64
//
// 参考: https://gobyexample-cn.github.io/atomic-counters
func (c *Counter) Inc() {
	atomic.AddInt64(&c.value, 1)
}

// Add 原子加 n
//
// TODO: Task 1b - 使用 atomic.AddInt64
func (c *Counter) Add(n int64) {
	atomic.AddInt64(&c.value, n)
}

// Value 原子读取当前值
//
// TODO: Task 1c - 使用 atomic.LoadInt64
func (c *Counter) Value() int64 {
	return atomic.LoadInt64(&c.value)
}

// ============================================================
// Task 2: 线程安全 Map
// ============================================================

// SafeMap 线程安全的 Map
//
// ⚠️ Go 中并发读写普通 Map 会 Panic！
// 必须使用 sync.RWMutex 保护
type SafeMap struct {
	mu   sync.RWMutex
	data map[string]interface{}
}

// NewSafeMap 创建 SafeMap
func NewSafeMap() *SafeMap {
	return &SafeMap{
		data: make(map[string]interface{}),
	}
}

// Set 设置键值对（需要写锁）
//
// TODO: Task 2a - 使用 mu.Lock() 保护写操作
//
// 步骤:
//   1. 获取写锁: m.mu.Lock()
//   2. 使用 defer 确保解锁: defer m.mu.Unlock()
//   3. 写入数据: m.data[key] = value
//
// 参考: https://gobyexample-cn.github.io/mutexes
func (m *SafeMap) Set(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
}

// Get 获取值（需要读锁）
//
// TODO: Task 2b - 使用 mu.RLock() 保护读操作
//
// 读锁允许多个协程同时读取
func (m *SafeMap) Get(key string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.data[key]
	return val, ok
}

// Delete 删除键（需要写锁）
//
// TODO: Task 2c - 使用写锁保护删除操作
func (m *SafeMap) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
}

// Len 返回长度（需要读锁）
//
// TODO: Task 2d - 使用读锁保护
func (m *SafeMap) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.data)
}

// ============================================================
// Task 3: 工作池
// ============================================================

// WorkerPool 工作池
type WorkerPool struct {
	jobs    chan func()
	wg      sync.WaitGroup
	workers int
}

// NewWorkerPool 创建工作池
func NewWorkerPool(workers, queueSize int) *WorkerPool {
	return &WorkerPool{
		jobs:    make(chan func(), queueSize),
		workers: workers,
	}
}

// Start 启动工作协程
//
// TODO: Task 3a - 启动 workers 个协程
//
// 每个工作协程:
//   1. 从 jobs 通道读取任务
//   2. 执行任务
//   3. 通道关闭后退出
//   4. 退出前调用 wg.Done()
//
// 参考: https://gobyexample-cn.github.io/worker-pools
func (p *WorkerPool) Start() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for job := range p.jobs {
				if job != nil {
					job()
				}
			}
		}()
	}
}

// Submit 提交任务到工作池
//
// TODO: Task 3b - 将任务发送到 jobs 通道
func (p *WorkerPool) Submit(job func()) {
	p.jobs <- job
}

// Stop 关闭工作池并等待完成
//
// TODO: Task 3c - 关闭通道并等待
//
// 步骤:
//   1. 关闭 jobs 通道: close(p.jobs)
//   2. 等待所有工人完成: p.wg.Wait()
func (p *WorkerPool) Stop() {
	close(p.jobs)
	p.wg.Wait()
}

// ============================================================
// Task 4: 速率限制器
// ============================================================

// RateLimiter 简单的令牌桶速率限制器
type RateLimiter struct {
	tokens    int64
	maxTokens int64
}

// NewRateLimiter 创建速率限制器
func NewRateLimiter(maxTokens int64) *RateLimiter {
	return &RateLimiter{
		tokens:    maxTokens,
		maxTokens: maxTokens,
	}
}

// Allow 尝试获取一个令牌
//
// TODO: Task 4a - 原子减少令牌
//
// 逻辑:
//   1. 读取当前令牌数
//   2. 如果 > 0，减少 1 并返回 true
//   3. 否则返回 false
//
// 提示: 使用 atomic.AddInt64 返回值判断
// 或使用 atomic.CompareAndSwapInt64 实现
func (r *RateLimiter) Allow() bool {
	for {
		current := atomic.LoadInt64(&r.tokens)
		if current <= 0 {
			return false
		}
		if atomic.CompareAndSwapInt64(&r.tokens, current, current-1) {
			return true
		}
	}
}

// Refill 补充一个令牌
//
// TODO: Task 4b - 原子增加令牌（不超过 maxTokens）
//
// 提示: 使用循环 + CompareAndSwap 实现原子操作
func (r *RateLimiter) Refill() {
	for {
		current := atomic.LoadInt64(&r.tokens)
		if current >= r.maxTokens {
			return
		}
		if atomic.CompareAndSwapInt64(&r.tokens, current, current+1) {
			return
		}
	}
}

// Tokens 返回当前令牌数
func (r *RateLimiter) Tokens() int64 {
	return atomic.LoadInt64(&r.tokens)
}
