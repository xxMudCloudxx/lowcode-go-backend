# Lab 2: 函数与错误处理 ⚙️

> 学习目标：掌握函数定义、多返回值、错误处理、闭包

## 📚 背景知识

Go 的函数是一等公民，可以：

- 返回多个值（常用于返回结果+错误）
- 作为参数传递
- 作为返回值（闭包）

## 🎯 任务清单

### 任务 1：实现 `ParseCommand()` 函数

解析用户输入的命令，返回命令名和参数：

```go
func ParseCommand(input string) (cmd string, args string, err error)
```

**规则：**

- 输入 `"/join room1"` → 返回 `"join"`, `"room1"`, `nil`
- 输入 `"/name Alice"` → 返回 `"name"`, `"Alice"`, `nil`
- 输入 `"/quit"` → 返回 `"quit"`, `""`, `nil`
- 输入 `"hello"` (不以 `/` 开头) → 返回 `"", "", ErrNotCommand`
- 输入 `""` (空字符串) → 返回 `"", "", ErrEmptyInput`

**知识点回顾：** [多返回值](https://gobyexample-cn.github.io/multiple-return-values) | [错误处理](https://gobyexample-cn.github.io/errors)

---

### 任务 2：实现 `ValidateUsername()` 函数

验证用户名是否合法：

```go
func ValidateUsername(name string) error
```

**规则：**

- 长度必须在 2-20 个字符之间
- 不能包含空格
- 通过验证返回 `nil`，否则返回对应错误

**知识点回顾：** [错误处理](https://gobyexample-cn.github.io/errors)

---

### 任务 3：实现 `FormatMessages()` 变参函数

格式化多条消息：

```go
func FormatMessages(prefix string, messages ...*Message) []string
```

- 对每条消息调用 `Format()` 方法
- 在结果前加上 prefix
- 返回格式化后的字符串切片

**知识点回顾：** [变参函数](https://gobyexample-cn.github.io/variadic-functions)

---

### 任务 4：实现 `CreateCounter()` 闭包

创建一个计数器函数：

```go
func CreateCounter(start int) func() int
```

返回一个闭包，每次调用返回递增的值：

```go
counter := CreateCounter(0)
counter() // 返回 1
counter() // 返回 2
counter() // 返回 3
```

**知识点回顾：** [闭包](https://gobyexample-cn.github.io/closures)

---

### 任务 5：实现 `Fibonacci()` 递归函数

计算第 n 个斐波那契数：

```go
func Fibonacci(n int) int
```

- `Fibonacci(0)` = 0
- `Fibonacci(1)` = 1
- `Fibonacci(n)` = `Fibonacci(n-1)` + `Fibonacci(n-2)`

**知识点回顾：** [递归](https://gobyexample-cn.github.io/recursion)

---

## 🧪 运行测试

```bash
cd lab2-functions
go test -v
```
