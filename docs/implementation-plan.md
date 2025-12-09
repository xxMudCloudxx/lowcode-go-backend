# Go 后端项目架构实现计划 (Final v2.2)

> **状态**: 架构评审 A- ✅ | Code Review 已通过  
> **整合来源**: Clean Architecture + Focalboard + json-patch + 架构审查修正  
> **参考文档**: [go-backend-guide.md](./go-backend-guide.md)

---

## 一、设计理念整合

| 来源                   | 提取内容                                                             |
| ---------------------- | -------------------------------------------------------------------- |
| **Clean Architecture** | 分层目录：`api/`, `domain/`, `usecase/`, `repository/`, `bootstrap/` |
| **Focalboard**         | Hub 房间管理、Listeners 订阅、Broadcast 广播                         |
| **json-patch**         | Usecase 层 RFC 6902 Patch 应用                                       |
| **Clerk**              | JWT 验证 + JWKS 公钥缓存 + Webhook 用户同步                          |
| **PostgreSQL**         | JSONB 类型存储 Schema                                                |

---

## 二、已解决的架构问题

| #   | 问题               | 解决方案                                |
| --- | ------------------ | --------------------------------------- |
| 1   | 数据双源陷阱       | PageUseCase 注入 Hub，优先读内存        |
| 2   | 鉴权双模式         | ClerkAuth 支持 Header + Query           |
| 3   | 持久化策略         | SaveTicker (5s) + saveSignal channel    |
| 4   | **竞态条件**       | Channel 串行化写入，flushToDB 加读锁    |
| 5   | **写风暴**         | 非阻塞 select + dirtyCount 重置         |
| 6   | **Goroutine 泄漏** | handleUnregister 调用 close(stopSave)   |
| 7   | **优雅停机**       | Hub.Shutdown + WaitGroup + SIGTERM 监听 |
| 8   | 用户同步           | Clerk Webhook → 本地 users 表           |
| 9   | **Map 并发读写**   | handleRegister/Unregister 加写锁        |
| 10  | **Shutdown 同步**  | Room.Stop() 阻塞等待 flushToDB 完成     |

---

## 三、项目目录结构

```
d:\Desktop\lowercode-go-server\
├── cmd/
│   └── main.go                     # 应用入口 + 优雅停机
├── api/
│   ├── controller/
│   │   ├── page_controller.go      # HTTP 页面接口
│   │   ├── ws_handler.go           # WebSocket 升级处理
│   │   └── webhook_controller.go   # Clerk Webhook
│   ├── route/
│   │   └── route.go                # 路由配置
│   └── middleware/
│       └── clerk_auth.go           # Clerk JWT 双模式鉴权
├── bootstrap/
│   ├── app.go                      # 依赖注入初始化
│   ├── database.go                 # PostgreSQL 配置
│   └── env.go                      # 环境变量加载
├── domain/
│   ├── entity/
│   │   ├── page.go                 # Page 实体
│   │   └── user.go                 # User 实体 (Clerk 同步)
│   └── repository/
│       ├── page_repository.go      # PageRepository 接口
│       └── user_repository.go      # UserRepository 接口
├── usecase/
│   └── page_usecase.go             # 业务逻辑 (注入 Hub, 含 json-patch)
├── repository/
│   ├── page_repository.go          # PageRepository 实现
│   └── user_repository.go          # UserRepository 实现
├── internal/
│   └── ws/                         # WebSocket 服务 (Focalboard 模式)
│       ├── hub.go                  # 房间管理 + Shutdown
│       ├── room.go                 # 实时状态 + saveSignal
│       ├── client.go               # 客户端连接
│       └── message.go              # 消息协议
├── go.mod
├── go.sum
└── .env
```

---

## 四、Domain 层

### 4.1 Page 实体

```go
// domain/entity/page.go

package entity

import (
    "encoding/json"
    "time"
)

// Component 对应前端组件
// ✅ ID 使用 int64 (前端使用时间戳作为 ID，root=1，其他为时间戳如 1765279327172)
// ✅ Props/Styles 使用 json.RawMessage，后端不解析业务字段
type Component struct {
    ID       int64           `json:"id"`           // 时间戳 ID
    Name     string          `json:"name"`
    Desc     string          `json:"desc"`
    ParentID *int64          `json:"parentId,omitempty"`
    Children []int64         `json:"children,omitempty"`
    Props    json.RawMessage `json:"props,omitempty"`
    Styles   json.RawMessage `json:"styles,omitempty"`
}

// Page 数据库模型 (PostgreSQL JSONB)
type Page struct {
    ID        uint      `gorm:"primaryKey"`
    PageID    string    `gorm:"uniqueIndex;size:64"`
    Schema    string    `gorm:"type:jsonb"`
    Version   int64     `gorm:"default:0"`
    CreatorID string    `gorm:"size:64"`  // Clerk user_id
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### 4.2 User 实体 (Clerk 同步)

```go
// domain/entity/user.go

package entity

import "time"

// User Clerk 用户同步表
type User struct {
    ID        string    `gorm:"primaryKey;size:64"` // Clerk user_id
    Email     string    `gorm:"size:255"`
    Name      string    `gorm:"size:100"`
    AvatarURL string    `gorm:"size:500"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### 4.3 Repository 接口

```go
// domain/repository/page_repository.go

package repository

import "lowercode-go-server/domain/entity"

type PageRepository interface {
    GetByPageID(pageID string) (*entity.Page, error)
    Save(page *entity.Page) error
    UpdateSchema(pageID string, schema []byte, version int64) error
}

// domain/repository/user_repository.go

type UserRepository interface {
    Upsert(user *entity.User) error
    GetByID(userID string) (*entity.User, error)
}
```

---

## 五、Usecase 层

### 5.1 PageUseCase (注入 Hub)

```go
// usecase/page_usecase.go

package usecase

import (
    "lowercode-go-server/domain/entity"
    "lowercode-go-server/domain/repository"
    "lowercode-go-server/internal/ws"

    jsonpatch "github.com/evanphx/json-patch/v5"
)

type PageUseCase struct {
    repo repository.PageRepository
    hub  *ws.Hub  // ✅ 注入 Hub，解决数据双源问题
}

func NewPageUseCase(repo repository.PageRepository, hub *ws.Hub) *PageUseCase {
    return &PageUseCase{repo: repo, hub: hub}
}

// GetPage 优先从内存热数据读取
func (uc *PageUseCase) GetPage(pageID string) (*entity.Page, error) {
    // 1. 优先从 Hub 内存读取
    if room := uc.hub.GetRoom(pageID); room != nil {
        snapshot, version := room.GetSnapshot()
        return &entity.Page{
            PageID:  pageID,
            Schema:  string(snapshot),
            Version: version,
        }, nil
    }
    // 2. 内存没有，读数据库
    return uc.repo.GetByPageID(pageID)
}

// ApplyPatch 使用 json-patch 库
func (uc *PageUseCase) ApplyPatch(currentState, patchBytes []byte) ([]byte, error) {
    patch, err := jsonpatch.DecodePatch(patchBytes)
    if err != nil {
        return nil, err
    }
    return patch.Apply(currentState)
}
```

---

## 六、WebSocket 服务层

### 6.1 消息协议

```go
// internal/ws/message.go

package ws

import "encoding/json"

type MessageType string

const (
    TypeOpPatch    MessageType = "op-patch"
    TypeCursorMove MessageType = "cursor-move"
    TypeUserJoin   MessageType = "user-join"
    TypeUserLeave  MessageType = "user-leave"
    TypeSync       MessageType = "sync"
    TypeAck        MessageType = "ack"
    TypeError      MessageType = "error"
)

type WSMessage struct {
    Type      MessageType     `json:"type"`
    SenderID  string          `json:"senderId"`
    Payload   json.RawMessage `json:"payload"`
    Timestamp int64           `json:"ts"`
}

type SyncPayload struct {
    Schema  json.RawMessage `json:"schema"`
    Version int64           `json:"version"`
    Users   []UserInfo      `json:"users"`
}

type UserInfo struct {
    UserID   string `json:"userId"`
    UserName string `json:"userName"`
    Color    string `json:"color,omitempty"`
}
```

### 6.2 Room (无竞态持久化)

```go
// internal/ws/room.go

package ws

import (
    "log"
    "sync"
    "time"

    jsonpatch "github.com/evanphx/json-patch/v5"
)

const (
    SaveInterval  = 5 * time.Second
    SaveThreshold = 20
)

type Room struct {
    ID           string
    CurrentState []byte
    Version      int64
    Clients      map[*Client]bool
    mu           sync.RWMutex
    LastActive   time.Time

    // 持久化策略
    dirtyCount int
    saveSignal chan string    // ✅ 信号通道，避免竞态
    saveTicker *time.Ticker
    stopSave   chan struct{}
    done       chan struct{}  // ✅ 新增: 刷盘完成信号
    repo       PageService
}

func NewRoom(id string, state []byte, repo PageService) *Room {
    r := &Room{
        ID:           id,
        CurrentState: state,
        Version:      1,
        Clients:      make(map[*Client]bool),
        LastActive:   time.Now(),
        saveSignal:   make(chan string, 1),
        saveTicker:   time.NewTicker(SaveInterval),
        stopSave:     make(chan struct{}),
        done:         make(chan struct{}),  // ✅ 新增
        repo:         repo,
    }
    go r.startSaveLoop()
    return r
}

func (r *Room) startSaveLoop() {
    defer close(r.done)  // ✅ 循环结束时通知 Stop()
    for {
        select {
        case <-r.saveTicker.C:
            r.flushToDB("定时")
        case reason := <-r.saveSignal:
            r.flushToDB(reason)
        case <-r.stopSave:
            r.flushToDB("销毁前")
            return  // ✅ return 后 defer 执行, close(done)
        }
    }
}

func (r *Room) ApplyPatch(patchBytes []byte) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    patch, err := jsonpatch.DecodePatch(patchBytes)
    if err != nil {
        return err
    }

    modified, err := patch.Apply(r.CurrentState)
    if err != nil {
        return err
    }

    r.CurrentState = modified
    r.Version++
    r.LastActive = time.Now()
    r.dirtyCount++

    // ✅ 非阻塞发送信号
    if r.dirtyCount >= SaveThreshold {
        select {
        case r.saveSignal <- "阈值触发":
            r.dirtyCount = 0
        default:
            // 正在保存中，跳过
        }
    }
    return nil
}

func (r *Room) flushToDB(reason string) {
    r.mu.RLock()  // ✅ 读锁
    stateCopy := make([]byte, len(r.CurrentState))
    copy(stateCopy, r.CurrentState)
    version := r.Version
    r.mu.RUnlock()

    if err := r.repo.SavePageState(r.ID, stateCopy, version); err != nil {
        log.Printf("[Room %s] ⚠️ %s刷盘失败: %v", r.ID, reason, err)
        return
    }
    log.Printf("[Room %s] ✅ %s刷盘, 版本: %d", r.ID, reason, version)
}

func (r *Room) GetSnapshot() ([]byte, int64) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    snapshot := make([]byte, len(r.CurrentState))
    copy(snapshot, r.CurrentState)
    return snapshot, r.Version
}

func (r *Room) Stop() {
    r.saveTicker.Stop()
    close(r.stopSave)
    <-r.done  // ✅ 阻塞等待 flushToDB 完成
}
```

### 6.3 Hub (优雅停机)

```go
// internal/ws/hub.go

package ws

import (
    "log"
    "sync"
)

type PageService interface {
    GetPageState(pageID string) ([]byte, int64, error)
    SavePageState(pageID string, state []byte, version int64) error
}

type Hub struct {
    rooms       map[string]*Room
    listeners   map[*Client]bool
    register    chan *Client
    unregister  chan *Client
    broadcast   chan *BroadcastMessage
    mu          sync.RWMutex
    wg          sync.WaitGroup  // ✅ 优雅停机
    pageService PageService
}

func NewHub(ps PageService) *Hub {
    return &Hub{
        rooms:       make(map[string]*Room),
        listeners:   make(map[*Client]bool),
        register:    make(chan *Client),
        unregister:  make(chan *Client),
        broadcast:   make(chan *BroadcastMessage, 256),
        pageService: ps,
    }
}

func (h *Hub) Run() {
    log.Println("[Hub] 🚀 Hub 事件循环已启动")
    for {
        select {
        case client := <-h.register:
            h.handleRegister(client)
        case client := <-h.unregister:
            h.handleUnregister(client)
        case msg := <-h.broadcast:
            h.handleBroadcast(msg)
        }
    }
}

func (h *Hub) handleRegister(client *Client) {
    roomID := client.RoomID

    // ⚠️ 必须加锁，因为 GetRoom 可能正在另一个协程读取
    h.mu.Lock()
    room, exists := h.rooms[roomID]
    if !exists {
        state, version, err := h.pageService.GetPageState(roomID)
        if err != nil {
            state = []byte(`{"rootId":1,"components":{}}`)
            version = 1
        }
        room = NewRoom(roomID, state, h.pageService)
        room.Version = version
        h.rooms[roomID] = room
        h.wg.Add(1)  // ✅ 计数
        log.Printf("[Hub] 🏠 创建房间: %s", roomID)
    }
    h.mu.Unlock()

    room.Clients[client] = true
    client.Room = room
    h.sendSyncMessage(client, room)
}

func (h *Hub) handleUnregister(client *Client) {
    room := client.Room
    if room == nil {
        return
    }

    delete(room.Clients, client)
    close(client.send)

    // ⚠️ 房间空了，必须善后 + 加写锁
    if len(room.Clients) == 0 {
        room.Stop()               // ✅ 停止 Goroutine

        h.mu.Lock()
        delete(h.rooms, room.ID)
        h.mu.Unlock()

        h.wg.Done()               // ✅ 计数减一
        log.Printf("[Hub] 🗑️ 房间 %s 已销毁", room.ID)
    }
}

// ✅ 优雅停机
func (h *Hub) Shutdown() {
    log.Println("[Hub] 正在关闭所有房间...")
    for _, room := range h.rooms {
        room.Stop()
        h.wg.Done()
    }
    h.wg.Wait()
    log.Println("[Hub] ✅ 所有房间已保存")
}

func (h *Hub) GetRoom(roomID string) *Room {
    h.mu.RLock()
    defer h.mu.RUnlock()
    return h.rooms[roomID]
}

// ... sendSyncMessage, handleBroadcast 等方法
```

### 6.4 Client

```go
// internal/ws/client.go

package ws

import (
    "github.com/gorilla/websocket"
)

type Client struct {
    Hub      *Hub
    Conn     *websocket.Conn
    RoomID   string
    UserInfo UserInfo
    Room     *Room
    send     chan []byte
}

func NewClient(hub *Hub, conn *websocket.Conn, roomID string, userInfo UserInfo) *Client {
    return &Client{
        Hub:      hub,
        Conn:     conn,
        RoomID:   roomID,
        UserInfo: userInfo,
        send:     make(chan []byte, 256),
    }
}

func (c *Client) ReadPump()  { /* 读取消息循环 */ }
func (c *Client) WritePump() { /* 发送消息循环 */ }
```

---

## 七、API 层

### 7.1 Clerk 鉴权中间件 (双模式)

```go
// api/middleware/clerk_auth.go

package middleware

import (
    "strings"

    "github.com/clerk/clerk-sdk-go/v2/jwt"
    "github.com/gin-gonic/gin"
)

func ClerkAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        // ✅ 双模式: Header 优先，Query 兜底 (WebSocket)
        token := c.GetHeader("Authorization")
        if token == "" {
            token = c.Query("token")
        }
        if token == "" {
            c.AbortWithStatusJSON(401, gin.H{"error": "未授权"})
            return
        }

        token = strings.TrimPrefix(token, "Bearer ")

        // Clerk SDK 内置 JWKS 缓存
        claims, err := jwt.Verify(c.Request.Context(), &jwt.VerifyParams{
            Token: token,
        })
        if err != nil {
            c.AbortWithStatusJSON(401, gin.H{"error": "Token 无效"})
            return
        }

        c.Set("userId", claims.Subject)
        c.Next()
    }
}
```

### 7.2 WebSocket Handler

```go
// api/controller/ws_handler.go

package controller

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
    "lowercode-go-server/internal/ws"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true },
}

type WSHandler struct {
    hub *ws.Hub
}

func NewWSHandler(hub *ws.Hub) *WSHandler {
    return &WSHandler{hub: hub}
}

func (h *WSHandler) ServeWS(c *gin.Context) {
    pageID := c.Param("pageId")
    userID := c.GetString("userId")  // 从中间件获取

    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        return
    }

    client := ws.NewClient(h.hub, conn, pageID, ws.UserInfo{
        UserID:   userID,
        UserName: userID,  // 可从 Clerk 获取
    })

    h.hub.Register(client)

    go client.WritePump()
    go client.ReadPump()
}
```

### 7.3 Webhook Controller

```go
// api/controller/webhook_controller.go

package controller

import (
    "github.com/gin-gonic/gin"
    "lowercode-go-server/domain/repository"
)

type WebhookController struct {
    userRepo repository.UserRepository
}

func (c *WebhookController) HandleClerkWebhook(ctx *gin.Context) {
    // 1. 验证 Webhook 签名 (Clerk-Signature header)
    // 2. 解析事件类型 (user.created / user.updated)
    // 3. Upsert 到本地 users 表
}
```

---

## 八、Bootstrap 层

### 8.1 应用初始化

```go
// bootstrap/app.go

package bootstrap

import (
    "lowercode-go-server/api/controller"
    "lowercode-go-server/domain/repository"
    "lowercode-go-server/internal/ws"
    "lowercode-go-server/usecase"

    "gorm.io/gorm"
)

type Application struct {
    Env       *Env
    DB        *gorm.DB
    Hub       *ws.Hub
    PageCtrl  *controller.PageController
    WSHandler *controller.WSHandler
}

func InitializeApp() *Application {
    // 1. 环境变量
    env := LoadEnv()

    // 2. 数据库
    db := NewDatabase(env.DatabaseURL)

    // 3. Repository
    pageRepo := NewPageRepository(db)

    // 4. Hub (传入 Repo)
    hub := ws.NewHub(pageRepo)
    go hub.Run()

    // 5. Usecase (注入 Repo + Hub)
    pageUC := usecase.NewPageUseCase(pageRepo, hub)

    // 6. Controller
    pageCtrl := controller.NewPageController(pageUC)
    wsHandler := controller.NewWSHandler(hub)

    return &Application{
        Env:       env,
        DB:        db,
        Hub:       hub,
        PageCtrl:  pageCtrl,
        WSHandler: wsHandler,
    }
}
```

### 8.2 主入口 (优雅停机)

```go
// cmd/main.go

package main

import (
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"

    "lowercode-go-server/api/route"
    "lowercode-go-server/bootstrap"
)

func main() {
    app := bootstrap.InitializeApp()

    router := route.Setup(app)

    srv := &http.Server{
        Addr:    ":" + app.Env.Port,
        Handler: router,
    }

    go func() {
        log.Printf("🚀 服务启动: http://localhost:%s", app.Env.Port)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("启动失败: %v", err)
        }
    }()

    // ✅ 优雅停机
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("正在关闭服务...")
    app.Hub.Shutdown()
    log.Println("✅ 服务已安全退出")
}
```

---

## 九、依赖安装

```bash
# 初始化
go mod init lowercode-go-server

# 核心依赖
go get github.com/gin-gonic/gin
go get github.com/gorilla/websocket
go get github.com/evanphx/json-patch/v5
go get gorm.io/gorm
go get gorm.io/driver/postgres
go get github.com/clerk/clerk-sdk-go/v2
go get github.com/joho/godotenv
```

---

## 十、环境变量 (.env)

```env
# 数据库
DATABASE_URL=postgres://user:password@localhost:5432/lowercode?sslmode=disable

# Clerk
CLERK_SECRET_KEY=sk_test_xxxxx

# 服务器
PORT=8080
```

---

## 十一、验证计划

### 1. 结构验证

```bash
cd d:\Desktop\lowercode-go-server
tree /F
```

### 2. 编译验证

```bash
go build ./...
```

### 3. 依赖验证

```bash
go mod tidy
go list -m all | grep -E "(gin|websocket|json-patch|gorm|clerk)"
```

### 4. 功能验证

- [ ] PageUseCase.GetPage 优先读内存
- [ ] Room.ApplyPatch 多并发无 panic
- [ ] SIGTERM 触发所有 Room flushToDB
- [ ] ClerkAuth 双模式工作

---

## 十二、架构图

```mermaid
graph TB
    subgraph "API Layer"
        Route --> Controller
        Controller --> ClerkAuth["ClerkAuth 双模式"]
    end

    subgraph "Usecase Layer"
        PageUseCase --> JP["json-patch"]
        PageUseCase --> Hub
    end

    subgraph "WebSocket Service"
        Hub --> Room
        Room --> SaveSignal["saveSignal channel"]
        Room --> SaveTicker["SaveTicker 5s"]
        Room --> Client
    end

    subgraph "Repository Layer"
        Interface["Repository Interface"]
        Implementation["Repository Impl"]
        Interface -.-> Implementation
    end

    Controller --> PageUseCase
    Hub --> Interface
    PageUseCase --> Interface
    Implementation --> DB[(PostgreSQL)]

    subgraph "Clerk"
        Webhook["Webhook Controller"]
        Webhook --> Users["users 表"]
    end
```

---

## 附录：参考资源

- [go-backend-guide.md](./go-backend-guide.md) - 详细的代码实现参考
- [Clean Architecture](https://github.com/amitshekhariitbhu/go-backend-clean-architecture)
- [Focalboard Server](https://github.com/mattermost/focalboard/tree/main/server)
- [json-patch](https://github.com/evanphx/json-patch)
- [Clerk Go SDK](https://clerk.com/docs/references/go/overview)
