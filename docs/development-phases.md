# Go 后端开发阶段指南

> **原则**: 每个阶段独立可测试，逐步叠加功能

---

## 🗺️ 开发路线图

```
Phase 1 ──► Phase 2 ──► Phase 3 ──► Phase 4 ──► Phase 5 ──► Phase 6
  基础       Domain     WebSocket    API层      Bootstrap   集成测试
  骨架        层          服务                    整合
```

---

## Phase 1: 项目骨架 (30 分钟)

### 目标

- 创建目录结构
- 初始化 Go module
- 安装核心依赖

### 任务清单

- [ ] 创建目录结构

```bash
mkdir -p cmd api/controller api/middleware api/route
mkdir -p bootstrap domain/entity domain/repository
mkdir -p usecase repository internal/ws
```

- [ ] 初始化 Go module

```bash
cd d:\Desktop\lowercode-go-server
go mod init lowercode-go-server
```

- [ ] 安装依赖

```bash
go get github.com/gin-gonic/gin
go get github.com/gorilla/websocket
go get github.com/evanphx/json-patch/v5
go get gorm.io/gorm
go get gorm.io/driver/postgres
go get github.com/joho/godotenv
```

- [ ] 创建 `.env` 模板

```env
DATABASE_URL=postgres://user:pass@localhost:5432/lowercode?sslmode=disable
PORT=8080
```

- [ ] 创建最简 `cmd/main.go`

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, LowerCode Server!")
}
```

### 验证

```bash
go run cmd/main.go
# 输出: Hello, LowerCode Server!
```

---

## Phase 2: Domain 层 (1 小时)

### 目标

- 定义核心实体
- 定义 Repository 接口

### 任务清单

- [ ] `domain/entity/page.go`

```go
type Component struct {
    ID       int64           `json:"id"`
    Name     string          `json:"name"`
    // ...完整字段见 implementation-plan.md
}

type Page struct {
    ID        uint      `gorm:"primaryKey"`
    PageID    string    `gorm:"uniqueIndex;size:64"`
    Schema    string    `gorm:"type:jsonb"`
    Version   int64     `gorm:"default:0"`
}
```

- [ ] `domain/entity/user.go`

```go
type User struct {
    ID        string `gorm:"primaryKey;size:64"`
    Email     string `gorm:"size:255"`
    Name      string `gorm:"size:100"`
    AvatarURL string `gorm:"size:500"`
}
```

- [ ] `domain/repository/page_repository.go`

```go
type PageRepository interface {
    GetByPageID(pageID string) (*entity.Page, error)
    Save(page *entity.Page) error
    UpdateSchema(pageID string, schema []byte, version int64) error
}
```

### 验证

```bash
go build ./domain/...
# 无编译错误
```

---

## Phase 3: WebSocket 服务层 (2 小时) ⭐ 核心

### 目标

- 实现 Hub/Room/Client 架构
- 实现 json-patch 应用逻辑
- 实现持久化策略

### 任务清单

**3.1 消息协议**

- [ ] `internal/ws/message.go`
  - 定义 MessageType 枚举
  - 定义 WSMessage 结构

**3.2 Room (最复杂)**

- [ ] `internal/ws/room.go`
  - CurrentState []byte
  - saveSignal channel
  - startSaveLoop() 串行化写入
  - ApplyPatch() 使用 json-patch
  - done channel 同步关闭

**3.3 Client**

- [ ] `internal/ws/client.go`
  - ReadPump / WritePump
  - handleOpPatch 消息处理

**3.4 Hub**

- [ ] `internal/ws/hub.go`
  - rooms map 管理
  - 读写锁保护
  - Shutdown() 优雅停机

### 验证

```bash
go build ./internal/ws/...
# 无编译错误
```

### 单元测试建议

```go
// room_test.go
func TestRoom_ApplyPatch(t *testing.T) {
    // 测试 json-patch 应用
}
```

---

## Phase 4: Repository 实现 (1 小时)

### 目标

- 实现 GORM Repository
- 连接 PostgreSQL

### 任务清单

- [ ] `bootstrap/database.go`

```go
func NewDatabase(dsn string) *gorm.DB {
    db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    db.AutoMigrate(&entity.Page{}, &entity.User{})
    return db
}
```

- [ ] `repository/page_repository.go`

```go
type pageRepository struct {
    db *gorm.DB
}

func (r *pageRepository) GetByPageID(pageID string) (*entity.Page, error) {
    var page entity.Page
    err := r.db.Where("page_id = ?", pageID).First(&page).Error
    return &page, err
}
```

### 验证

```bash
# 启动 PostgreSQL
# 运行迁移测试
go test ./repository/...
```

---

## Phase 5: API 层 (1.5 小时)

### 目标

- 实现 HTTP Controller
- 实现 WebSocket Handler
- 实现路由配置

### 任务清单

**5.1 中间件**

- [ ] `api/middleware/clerk_auth.go` (双模式鉴权)

**5.2 Controller**

- [ ] `api/controller/page_controller.go`
  - GetPage / SavePage
- [ ] `api/controller/ws_handler.go`
  - ServeWS 升级 WebSocket

**5.3 路由**

- [ ] `api/route/route.go`

```go
func Setup(app *bootstrap.Application) *gin.Engine {
    r := gin.Default()

    api := r.Group("/api")
    api.GET("/pages/:pageId", app.PageCtrl.GetPage)

    ws := r.Group("/ws")
    ws.GET("/:pageId", app.WSHandler.ServeWS)

    return r
}
```

### 验证

```bash
go build ./api/...
```

---

## Phase 6: Bootstrap 整合 (1 小时)

### 目标

- 依赖注入初始化
- 优雅停机
- 完整可运行

### 任务清单

- [ ] `bootstrap/env.go`
- [ ] `bootstrap/app.go` (依赖注入)
- [ ] `cmd/main.go` (优雅停机)

### 验证

```bash
go run cmd/main.go
# 服务启动，可接受连接
```

---

## Phase 7: 集成测试 (1 小时)

### 测试场景

1. **HTTP 测试**

```bash
curl http://localhost:8080/api/pages/test-page
```

2. **WebSocket 测试**
   使用 wscat 或浏览器 DevTools:

```bash
wscat -c ws://localhost:8080/ws/test-page
```

3. **Patch 应用测试**
   发送真实的 Patch 消息:

```json
{
  "type": "op-patch",
  "payload": {
    "patches": [
      { "op": "replace", "path": "/components/1/desc", "value": "测试" }
    ]
  }
}
```

---

## 📋 进度追踪

| Phase             | 状态 | 预计时间   | 实际时间 |
| ----------------- | ---- | ---------- | -------- |
| 1. 项目骨架       | ⬜   | 30 分钟    |          |
| 2. Domain 层      | ⬜   | 1 小时     |          |
| 3. WebSocket 服务 | ⬜   | 2 小时     |          |
| 4. Repository     | ⬜   | 1 小时     |          |
| 5. API 层         | ⬜   | 1.5 小时   |          |
| 6. Bootstrap      | ⬜   | 1 小时     |          |
| 7. 集成测试       | ⬜   | 1 小时     |          |
| **总计**          |      | **8 小时** |          |

---

## 💡 开发建议

1. **每个 Phase 完成后提交 Git**
2. **优先跑通 Phase 1-3**，这是核心
3. **Phase 4-6 可以用 Mock 替代**
4. **遇到问题先写 TODO 注释，不要卡住**

---

## 📚 参考文档

- [implementation-plan.md](./implementation-plan.md) - 详细代码
- [go-backend-guide.md](./go-backend-guide.md) - 设计理念
- [frontend-patch-format.md](./frontend-patch-format.md) - Patch 格式
