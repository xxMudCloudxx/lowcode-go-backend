# Lab 4: 并发基础 🚀

> 学习目标：掌握协程、通道、通道缓冲

## 📚 背景知识

Go 的并发模型基于 CSP (Communicating Sequential Processes)：

- **Goroutine**: 轻量级线程，用 `go` 关键字启动
- **Channel**: 协程间通信的管道

## 🎯 任务清单

### 任务 1：实现 `RunAsync()` 函数

并行执行多个任务：

```go
func RunAsync(tasks ...func())
```

- 为每个 task 启动一个协程
- 等待所有任务完成后返回

**提示：** 使用 `sync.WaitGroup`（可以先简单用 `time.Sleep`，后面 Lab 6 会学 WaitGroup）

**知识点回顾：** [协程](https://gobyexample-cn.github.io/goroutines)

---

### 任务 2：实现 `Generator()` 函数

创建一个生成器，返回发送数字的通道：

```go
func Generator(nums ...int) <-chan int
```

- 创建一个通道
- 启动协程依次发送每个数字
- 发送完成后关闭通道
- 返回只读通道

**知识点回顾：** [通道](https://gobyexample-cn.github.io/channels) | [通道方向](https://gobyexample-cn.github.io/channel-directions)

---

### 任务 3：实现 `Ping Pong` 通信

两个协程互相传递消息：

```go
func PingPong(count int) []string
```

- 创建两个通道
- Ping 协程发送 "ping"
- Pong 协程收到后发送 "pong"
- 重复 count 次
- 返回所有消息记录

**知识点回顾：** [通道同步](https://gobyexample-cn.github.io/channel-synchronization)

---

### 任务 4：实现缓冲通道任务队列

```go
func CreateTaskQueue(bufferSize int) chan func()
```

- 创建一个缓冲通道
- 返回该通道

```go
func ProcessQueue(queue chan func())
```

- 从队列读取任务并执行
- 队列关闭后退出

**知识点回顾：** [通道缓冲](https://gobyexample-cn.github.io/channel-buffering)

---

## 🧪 运行测试

```bash
cd lab4-concurrency
go test -v -race  # -race 检测数据竞争
```
