# Lab 1: 消息与结构体 📦

> 学习目标：掌握结构体定义、方法绑定、常量枚举、Switch 语句

## 📚 背景知识

在聊天室中，消息是最基本的数据单元。每条消息包含：

- **类型**：加入、离开、聊天、系统消息
- **发送者**：用户名
- **内容**：消息文本
- **时间戳**：发送时间

## 🎯 任务清单

### 任务 1：定义消息类型常量

打开 `message.go`，找到 `// TODO: Task 1`，定义 4 个消息类型常量：

```go
const (
    MsgJoin   MessageType = iota  // 0: 用户加入
    MsgLeave                      // 1: 用户离开
    MsgChat                       // 2: 聊天消息
    MsgSystem                     // 3: 系统消息
)
```

**知识点回顾：** [常量](https://gobyexample-cn.github.io/constants) | [iota](https://gobyexample-cn.github.io/constants)

---

### 任务 2：实现 `String()` 方法

为 `MessageType` 实现 `String()` 方法，返回类型的可读名称：

| 类型      | 返回值      |
| --------- | ----------- |
| MsgJoin   | `"JOIN"`    |
| MsgLeave  | `"LEAVE"`   |
| MsgChat   | `"CHAT"`    |
| MsgSystem | `"SYSTEM"`  |
| 其他      | `"UNKNOWN"` |

**提示：** 使用 `switch` 语句

**知识点回顾：** [方法](https://gobyexample-cn.github.io/methods) | [Switch](https://gobyexample-cn.github.io/switch)

---

### 任务 3：完成 `NewMessage()` 函数

实现工厂函数，创建并返回一个 `*Message`：

```go
func NewMessage(msgType MessageType, username, content string) *Message
```

需要设置 `Timestamp` 为当前时间。

**知识点回顾：** [指针](https://gobyexample-cn.github.io/pointers) | [结构体](https://gobyexample-cn.github.io/structs)

---

### 任务 4：实现 `Format()` 方法

为 `Message` 实现格式化方法，返回格式化的字符串：

| 消息类型  | 格式                      | 示例                 |
| --------- | ------------------------- | -------------------- |
| MsgJoin   | `">> {username} joined"`  | `">> Alice joined"`  |
| MsgLeave  | `"<< {username} left"`    | `"<< Alice left"`    |
| MsgChat   | `"{username}: {content}"` | `"Alice: Hello!"`    |
| MsgSystem | `"[System] {content}"`    | `"[System] Welcome"` |

**知识点回顾：** [方法](https://gobyexample-cn.github.io/methods)

---

## 🧪 运行测试

```bash
cd lab1-basics
go test -v
```

预期输出：

```
=== RUN   TestMessageTypeString
--- PASS: TestMessageTypeString
=== RUN   TestNewMessage
--- PASS: TestNewMessage
=== RUN   TestMessageFormat
--- PASS: TestMessageFormat
PASS
```

---

## 💡 提示

如果卡住了，可以查看 [hints.md](./hints.md)
