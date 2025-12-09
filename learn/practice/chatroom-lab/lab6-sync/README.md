# Lab 6: 同步原语 🔒

> 学习目标：掌握互斥锁、WaitGroup、原子计数器

## 📚 背景知识

并发程序中，多个协程访问共享数据时需要同步：

- **互斥锁 (Mutex)**: 保护临界区，一次只允许一个协程访问
- **读写锁 (RWMutex)**: 允许多个读取者，但写入者独占
- **WaitGroup**: 等待一组协程完成
- **原子操作**: 无锁的线程安全计数

> ⚠️ **重要警告**：Go 中并发读写 Map 会直接 Panic！必须用锁保护。

## 🎯 任务清单

### 任务 1：实现 `Counter` 结构体

线程安全的计数器：

```go
type Counter struct {
    value int64
}

func (c *Counter) Inc()         // 原子加 1
func (c *Counter) Add(n int64)  // 原子加 n
func (c *Counter) Value() int64 // 原子读取
```

**知识点回顾：** [原子计数器](https://gobyexample-cn.github.io/atomic-counters)

---

### 任务 2：实现 `SafeMap` 结构体

线程安全的 Map（解决并发读写问题）：

```go
type SafeMap struct {
    mu   sync.RWMutex
    data map[string]interface{}
}

func (m *SafeMap) Set(key string, value interface{})
func (m *SafeMap) Get(key string) (interface{}, bool)
func (m *SafeMap) Delete(key string)
func (m *SafeMap) Len() int
```

**规则：**

- `Set` 和 `Delete` 使用写锁 (`mu.Lock()`)
- `Get` 和 `Len` 使用读锁 (`mu.RLock()`)

**知识点回顾：** [互斥锁](https://gobyexample-cn.github.io/mutexes)

---

### 任务 3：实现 `WorkerPool`

工作池模式：

```go
type WorkerPool struct {
    jobs    chan func()
    wg      sync.WaitGroup
    workers int
}

func (p *WorkerPool) Start()              // 启动 workers 个协程
func (p *WorkerPool) Submit(job func())   // 提交任务
func (p *WorkerPool) Stop()               // 关闭并等待完成
```

**知识点回顾：** [WaitGroup](https://gobyexample-cn.github.io/waitgroups) | [工作池](https://gobyexample-cn.github.io/worker-pools)

---

### 任务 4：实现 `RateLimiter`

简单的速率限制器：

```go
type RateLimiter struct {
    tokens  int64        // 当前令牌数（原子操作）
    maxTokens int64
}

func (r *RateLimiter) Allow() bool  // 尝试获取一个令牌
func (r *RateLimiter) Refill()      // 补充一个令牌
```

**知识点回顾：** [原子计数器](https://gobyexample-cn.github.io/atomic-counters)

---

## 🧪 运行测试

```bash
cd lab6-sync
go test -v -race  # -race 很重要，检测数据竞争！
```

> 💡 **提示**：如果测试报 `DATA RACE`，说明你的锁没有正确使用。

---

## ⚠️ 常见错误

### 1. Map 并发读写 Panic

```go
// ❌ 错误：没有加锁
func (m *SafeMap) Set(key string, value interface{}) {
    m.data[key] = value  // 并发调用会 Panic!
}

// ✅ 正确：使用互斥锁
func (m *SafeMap) Set(key string, value interface{}) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.data[key] = value
}
```

### 2. 忘记 Unlock

```go
// ❌ 错误：忘记 Unlock
func (m *SafeMap) Get(key string) interface{} {
    m.mu.RLock()
    return m.data[key]  // 没有 RUnlock，死锁！
}

// ✅ 正确：使用 defer
func (m *SafeMap) Get(key string) interface{} {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.data[key]
}
```
