package lab4

import (
	"sync"
	"time"
)

// ============================================================
// Task 1: 并行执行任务
// ============================================================

// RunAsync 并行执行多个任务，等待全部完成
//
// TODO: Task 1 - 使用协程并行执行所有任务
//
// 步骤:
//   1. 创建 sync.WaitGroup
//   2. 对每个 task，调用 wg.Add(1)
//   3. 启动协程执行 task，完成后调用 wg.Done()
//   4. 调用 wg.Wait() 等待所有任务完成
//
// 参考: https://gobyexample-cn.github.io/goroutines
// 参考: https://gobyexample-cn.github.io/waitgroups
func RunAsync(tasks ...func()) {
	// 在这里编写你的代码...
	
	// 提示: 使用 var wg sync.WaitGroup
}

// ============================================================
// Task 2: 通道生成器
// ============================================================

// Generator 创建一个发送数字的通道
//
// TODO: Task 2 - 返回只读通道
//
// 步骤:
//   1. 创建 int 类型通道: ch := make(chan int)
//   2. 启动协程:
//      - 遍历 nums，将每个数字发送到通道
//      - 发送完成后关闭通道: close(ch)
//   3. 返回通道（类型为 <-chan int 表示只读）
//
// 参考: https://gobyexample-cn.github.io/channels
// 参考: https://gobyexample-cn.github.io/channel-directions
func Generator(nums ...int) <-chan int {
	// 在这里编写你的代码...
	return nil
}

// ============================================================
// Task 3: Ping Pong 通信
// ============================================================

// PingPong 两个协程互相传递消息
//
// TODO: Task 3 - 实现乒乓通信
//
// 逻辑:
//   1. 创建两个通道: pingCh, pongCh
//   2. 启动 Ping 协程:
//      - 重复 count 次:
//        - 发送 "ping" 到 pingCh
//        - 从 pongCh 接收
//   3. 启动 Pong 协程:
//      - 重复 count 次:
//        - 从 pingCh 接收
//        - 发送 "pong" 到 pongCh
//   4. 收集所有消息到 []string 并返回
//
// 示例: PingPong(2) 返回 ["ping", "pong", "ping", "pong"]
//
// 参考: https://gobyexample-cn.github.io/channel-synchronization
func PingPong(count int) []string {
	// 在这里编写你的代码...
	return nil
}

// ============================================================
// Task 4: 缓冲通道任务队列
// ============================================================

// CreateTaskQueue 创建任务队列（缓冲通道）
//
// TODO: Task 4a - 创建并返回缓冲通道
//
// 提示: make(chan Type, bufferSize)
//
// 参考: https://gobyexample-cn.github.io/channel-buffering
func CreateTaskQueue(bufferSize int) chan func() {
	// 在这里编写你的代码...
	return nil
}

// ProcessQueue 处理任务队列
//
// TODO: Task 4b - 从队列读取并执行任务
//
// 步骤:
//   1. 使用 for range 遍历通道
//   2. 执行每个任务
//   3. 通道关闭后自动退出循环
//
// 参考: https://gobyexample-cn.github.io/range-over-channels
func ProcessQueue(queue chan func()) {
	// 在这里编写你的代码...
}

// --- 辅助类型（不要修改）---

// SafeCounter 线程安全计数器（用于测试）
type SafeCounter struct {
	mu    sync.Mutex
	count int
}

func (c *SafeCounter) Inc() {
	c.mu.Lock()
	c.count++
	c.mu.Unlock()
}

func (c *SafeCounter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}

// 防止未使用的导入错误
var _ = time.Sleep
