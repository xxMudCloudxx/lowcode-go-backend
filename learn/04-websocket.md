# WebSocket 协同编辑 (Day 8-10)

> 目标：实现多人实时协同编辑，这是项目的"灵魂"

## 📚 核心学习资源

| 资源                       | 链接                                                           | 说明          |
| -------------------------- | -------------------------------------------------------------- | ------------- |
| **Gorilla WebSocket Chat** | https://github.com/gorilla/websocket/tree/master/examples/chat | ⭐ 必读源码   |
| Gorilla WebSocket 文档     | https://pkg.go.dev/github.com/gorilla/websocket                | API 参考      |
| json-patch 库              | https://github.com/evanphx/json-patch                          | RFC 6902 实现 |

---

## 🚀 快速开始

### 1. 安装依赖

```bash
go get github.com/gorilla/websocket
go get github.com/evanphx/json-patch/v5
```

### 2. 最简 WebSocket 服务

```go
package main

import (
    "log"
    "net/http"
    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true },
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
    conn, _ := upgrader.Upgrade(w, r, nil)
    defer conn.Close()

    for {
        _, msg, err := conn.ReadMessage()
        if err != nil {
            break
        }
        log.Printf("收到: %s", msg)
        conn.WriteMessage(websocket.TextMessage, msg)  // Echo
    }
}

func main() {
    http.HandleFunc("/ws", wsHandler)
    log.Println("WebSocket 服务启动: :8080")
    http.ListenAndServe(":8080", nil)
}
```

---

## 🎯 核心架构：Hub 模式

### 架构图

```
┌─────────────────────────────────────────────────┐
│                     Hub                          │
│  ┌─────────────────────────────────────────────┐ │
│  │  rooms: map[string]*Room                     │ │
│  │                                              │ │
│  │  ┌──────────────────────────────────────┐   │ │
│  │  │  Room: "page-001"                     │   │ │
│  │  │  CurrentState: []byte (最新JSON)      │   │ │
│  │  │  Clients: [Client1, Client2, ...]    │   │ │
│  │  └──────────────────────────────────────┘   │ │
│  │                                              │ │
│  │  ┌──────────────────────────────────────┐   │ │
│  │  │  Room: "page-002"                     │   │ │
│  │  │  ...                                  │   │ │
│  │  └──────────────────────────────────────┘   │ │
│  └─────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────┘
```

---

### 1. Hub 结构 (房间管理器)

```go
// internal/ws/hub.go
type Hub struct {
    rooms      map[string]*Room
    register   chan *Client
    unregister chan *Client
    broadcast  chan *BroadcastMessage
}

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            // 加入房间
        case client := <-h.unregister:
            // 离开房间
        case msg := <-h.broadcast:
            // 广播消息
        }
    }
}
```

---

### 2. Room 结构 (房间状态)

```go
// internal/ws/room.go
type Room struct {
    ID           string
    CurrentState []byte           // 内存中的最新 Schema
    Version      int64
    Clients      map[*Client]bool
    mu           sync.RWMutex
}

// 应用 Patch 并更新状态
func (r *Room) ApplyPatch(patchBytes []byte) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    patch, _ := jsonpatch.DecodePatch(patchBytes)
    r.CurrentState, _ = patch.Apply(r.CurrentState)
    r.Version++
    return nil
}
```

---

### 3. Client 结构 (单个连接)

```go
// internal/ws/client.go
type Client struct {
    Hub      *Hub
    Conn     *websocket.Conn
    RoomID   string
    UserInfo UserInfo
    send     chan []byte
}

// 读取消息
func (c *Client) ReadPump() {
    for {
        _, message, err := c.Conn.ReadMessage()
        if err != nil {
            c.Hub.unregister <- c
            break
        }
        c.handleMessage(message)
    }
}

// 发送消息
func (c *Client) WritePump() {
    for msg := range c.send {
        c.Conn.WriteMessage(websocket.TextMessage, msg)
    }
}
```

---

## 📨 消息协议

```typescript
// 消息格式
interface WSMessage {
  type: 'op-patch' | 'sync' | 'user-join' | 'user-leave' | 'error';
  senderId: string;
  payload: any;
  ts: number;
}

// op-patch payload
{
  patches: [
    { op: 'replace', path: '/components/1/props/title', value: 'Hello' }
  ],
  version: 5
}

// sync payload (新用户加入时收到)
{
  schema: { rootId: 1, components: {...} },
  version: 5,
  users: [{ userId: 'u1', userName: 'Alice' }]
}
```

---

## 🔄 协同流程

```
用户 A 修改组件
    ↓
前端生成 JSON Patch
    ↓
WebSocket 发送 op-patch
    ↓
Go 后端接收
    ↓
Room.ApplyPatch() 更新内存状态
    ↓
广播给房间内其他用户
    ↓
用户 B/C 收到 Patch，应用到本地
```

---

## ✅ Day 8-10 作业

### 阶段验收

1. **Day 8**: 能打开两个浏览器窗口，A 发送消息，B 能收到
2. **Day 9**: 实现 Room 结构，A 发送 Patch，B 能收到并显示
3. **Day 10**: A 修改组件，B 能实时看到变化（UI 同步）

### 测试方法

```javascript
// 浏览器控制台测试
const ws = new WebSocket('ws://localhost:8080/ws/page-001?ticket=xxx');
ws.onmessage = (e) => console.log('收到:', JSON.parse(e.data));
ws.send(JSON.stringify({
  type: 'op-patch',
  payload: { patches: [...], version: 1 }
}));
```

---

## 📖 补充阅读

- Gorilla WebSocket Examples: https://github.com/gorilla/websocket/tree/master/examples
- JSON Patch RFC 6902: https://datatracker.ietf.org/doc/html/rfc6902
- Go sync.Mutex: https://gobyexample-cn.github.io/mutexes
- Go sync.RWMutex: https://pkg.go.dev/sync#RWMutex
