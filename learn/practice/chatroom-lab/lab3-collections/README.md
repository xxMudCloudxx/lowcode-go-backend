# Lab 3: 数据集合 📊

> 学习目标：掌握切片、Map、Range 遍历

## 📚 背景知识

Go 有三种主要的集合类型：

- **数组**: 固定长度，很少直接使用
- **切片 (Slice)**: 动态数组，最常用
- **Map**: 键值对，类似其他语言的 HashMap/Dictionary

## 🎯 任务清单

### 任务 1：实现 `MessageHistory` 结构体方法

一个消息历史记录管理器：

```go
type MessageHistory struct {
    messages []*Message
    maxSize  int
}
```

实现以下方法：

#### `Add(msg *Message)`

- 添加消息到历史记录
- 如果超过 maxSize，删除最早的消息

#### `GetRecent(n int) []*Message`

- 返回最近的 n 条消息
- 如果 n > 实际数量，返回全部

#### `Clear()`

- 清空所有消息

**知识点回顾：** [切片](https://gobyexample-cn.github.io/slices)

---

### 任务 2：实现 `UserRegistry` 结构体方法

一个用户注册表：

```go
type UserRegistry struct {
    users map[string]*User
}

type User struct {
    ID       string
    Username string
    JoinTime time.Time
}
```

实现以下方法：

#### `Register(id, username string) error`

- 注册新用户
- 如果 ID 已存在，返回 `ErrUserExists`

#### `Unregister(id string) error`

- 注销用户
- 如果 ID 不存在，返回 `ErrUserNotFound`

#### `GetUser(id string) (*User, bool)`

- 获取用户，返回用户和是否存在

#### `Count() int`

- 返回用户总数

**知识点回顾：** [Map](https://gobyexample-cn.github.io/maps)

---

### 任务 3：实现 `GetUsernames()` 方法

遍历所有用户，返回用户名列表：

```go
func (r *UserRegistry) GetUsernames() []string
```

**知识点回顾：** [Range 遍历](https://gobyexample-cn.github.io/range)

---

### 任务 4：实现 `FilterMessages()` 函数

按类型过滤消息：

```go
func FilterMessages(messages []*Message, msgType MessageType) []*Message
```

返回所有匹配类型的消息切片。

**知识点回顾：** [Range 遍历](https://gobyexample-cn.github.io/range) | [切片](https://gobyexample-cn.github.io/slices)

---

## 🧪 运行测试

```bash
cd lab3-collections
go test -v
```
