# Lab 1 提示 💡

> ⚠️ 尝试自己解决后再看提示！

---

## Task 1 提示

<details>
<summary>点击展开</summary>

使用 `iota` 自动递增：

```go
const (
    MsgJoin   MessageType = iota  // iota = 0
    MsgLeave                      // iota = 1 (自动递增)
    MsgChat                       // iota = 2
    MsgSystem                     // iota = 3
)
```

</details>

---

## Task 2 提示

<details>
<summary>点击展开</summary>

```go
func (t MessageType) String() string {
    switch t {
    case MsgJoin:
        return "JOIN"
    case MsgLeave:
        return "LEAVE"
    // ... 继续其他 case
    default:
        return "UNKNOWN"
    }
}
```

</details>

---

## Task 3 提示

<details>
<summary>点击展开</summary>

创建结构体指针的两种方式：

```go
// 方式 1: 使用 & 取地址
msg := &Message{
    Type:      msgType,
    Username:  username,
    Content:   content,
    Timestamp: time.Now(),
}
return msg

// 方式 2: 使用 new() 然后赋值
msg := new(Message)
msg.Type = msgType
// ... 其他字段
return msg
```

</details>

---

## Task 4 提示

<details>
<summary>点击展开</summary>

使用 `fmt.Sprintf` 格式化：

```go
import "fmt"

func (m *Message) Format() string {
    switch m.Type {
    case MsgJoin:
        return fmt.Sprintf(">> %s joined", m.Username)
    case MsgLeave:
        return fmt.Sprintf("<< %s left", m.Username)
    // ... 继续其他 case
    default:
        return ""
    }
}
```

</details>
