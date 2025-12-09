package lab5

import (
	"sync"
	"time"
)

// ============================================================
// Task 1: 通道合并 (Fan-in)
// ============================================================

// Merge 将多个输入通道合并为一个输出通道
//
// TODO: Task 1 - 使用 select 合并通道
//
// 步骤:
//   1. 创建输出通道
//   2. 启动一个协程:
//      - 遍历所有输入通道
//      - 为每个输入通道启动一个协程，读取数据并发送到输出通道
//      - 使用 WaitGroup 等待所有输入协程完成
//      - 关闭输出通道
//   3. 返回输出通道
//
// 提示: 需要用 sync.WaitGroup 等待所有 goroutine
//
// 参考: https://gobyexample-cn.github.io/select
func Merge(channels ...<-chan int) <-chan int {
	outCh := make(chan int)
	var wg sync.WaitGroup
	for _, ch := range channels {
		wg.Add(1)
		go func(c <-chan int) {
			defer wg.Done()
			for v := range c {
				outCh <- v
			}
		}(ch)
	}
	go func() {
		wg.Wait()
		close(outCh)
	}()
	return outCh
}

// ============================================================
// Task 2: 超时处理
// ============================================================

// WithTimeout 带超时的通道读取
//
// TODO: Task 2 - 使用 select + time.After 实现超时
//
// 逻辑:
//   select {
//   case value := <-ch:
//       // 收到数据，返回 (value, true)
//   case <-time.After(timeout):
//       // 超时，返回 (0, false)
//   }
//
// 参考: https://gobyexample-cn.github.io/timeouts
func WithTimeout(ch <-chan int, timeout time.Duration) (int, bool) {
	// 在这里编写你的代码...
	select {
	case value := <- ch:
		return value, true;
	case <-time.After(timeout):
		return 0, false
	}
}

// ============================================================
// Task 3: 非阻塞操作
// ============================================================

// TrySend 非阻塞发送
//
// TODO: Task 3a - 使用 select + default 实现非阻塞发送
//
// 如果通道已满或无缓冲且无接收者，立即返回 false
//
// 参考: https://gobyexample-cn.github.io/non-blocking-channel-operations
func TrySend(ch chan<- int, value int) bool {
	// 在这里编写你的代码...
	// 提示:
	select {
	case ch <- value:
	    return true
	default:
	    return false
	}

}

// TryReceive 非阻塞接收
//
// TODO: Task 3b - 使用 select + default 实现非阻塞接收
//
// 如果通道为空或无缓冲且无发送者，立即返回 (0, false)
func TryReceive(ch <-chan int) (int, bool) {
	// 在这里编写你的代码...
	select {
	case value := <-ch:
		return value, true
	default:
		return 0, false
	}
}

// ============================================================
// Task 4: 广播消息
// ============================================================

// Broadcast 向多个通道广播消息（非阻塞）
//
// TODO: Task 4 - 遍历通道并尝试发送
//
// 规则:
//   - 对每个通道使用非阻塞发送
//   - 如果通道已满，跳过（不阻塞）
//   - 返回成功发送的数量
//
// 提示: 复用 TrySend 函数
func Broadcast(message int, channels ...chan<- int) int {
	// 在这里编写你的代码...
	count := 0
	for _, cn := range channels {
		canSend := TrySend(cn, message)
		if canSend {
			count++
		}
	}
	return count
}

// ============================================================
// Task 5: 通道遍历
// ============================================================

// Drain 排空通道中的所有数据
//
// TODO: Task 5 - 使用 for range 遍历通道
//
// 规则:
//   - 读取所有数据直到通道关闭
//   - 返回所有数据的切片
//
// 注意: 调用者需要保证通道会被关闭，否则会永久阻塞！
//
// 参考: https://gobyexample-cn.github.io/range-over-channels
func Drain(ch <-chan int) []int {
	// 在这里编写你的代码...
	result := []int{}
	for val := range ch {
		result = append(result, val)
	}
	
	return result
}
