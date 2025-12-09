# Lab 5: 通道进阶 📡

> 学习目标：掌握 Select、超时处理、非阻塞操作、通道关闭

## 📚 背景知识

`select` 是 Go 并发的核心控制结构，类似 switch，但用于通道操作。

## 🎯 任务清单

### 任务 1：实现 `Merge()` 函数

合并多个通道为一个（Fan-in 模式）：

```go
func Merge(channels ...<-chan int) <-chan int
```

- 从所有输入通道读取数据
- 输出到一个合并通道
- 所有输入关闭后，关闭输出通道

**知识点回顾：** [通道选择器](https://gobyexample-cn.github.io/select)

---

### 任务 2：实现 `WithTimeout()` 函数

带超时的通道读取：

```go
func WithTimeout(ch <-chan int, timeout time.Duration) (int, bool)
```

- 在 timeout 内收到数据，返回 `(值, true)`
- 超时，返回 `(0, false)`

**知识点回顾：** [超时处理](https://gobyexample-cn.github.io/timeouts)

---

### 任务 3：实现 `TrySend()` 和 `TryReceive()` 函数

非阻塞通道操作：

```go
func TrySend(ch chan<- int, value int) bool  // 成功返回 true
func TryReceive(ch <-chan int) (int, bool)   // 成功返回 (值, true)
```

**知识点回顾：** [非阻塞通道操作](https://gobyexample-cn.github.io/non-blocking-channel-operations)

---

### 任务 4：实现 `Broadcast()` 函数

向多个通道广播消息：

```go
func Broadcast(message int, channels ...chan<- int) int
```

- 尝试向所有通道发送 message
- 使用非阻塞发送，跳过已满的通道
- 返回成功发送的数量

**知识点回顾：** [非阻塞通道操作](https://gobyexample-cn.github.io/non-blocking-channel-operations)

---

### 任务 5：实现 `Drain()` 函数

排空通道中的所有数据：

```go
func Drain(ch <-chan int) []int
```

- 读取通道中所有数据直到关闭
- 返回所有数据的切片

**知识点回顾：** [通道遍历](https://gobyexample-cn.github.io/range-over-channels)

---

## 🧪 运行测试

```bash
cd lab5-channels
go test -v -race
```
