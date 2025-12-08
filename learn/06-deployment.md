# Docker 部署 (Day 13)

> 目标：将 Go 后端容器化并部署到云平台

## 📚 核心学习资源

| 资源            | 链接                     | 说明            |
| --------------- | ------------------------ | --------------- |
| **Railway**     | https://railway.app/     | ⭐ 推荐部署平台 |
| Render          | https://render.com/      | 备选平台        |
| Docker 官方文档 | https://docs.docker.com/ | 容器化参考      |

---

## 🐳 Dockerfile

### 基础版本

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server

# 运行阶段
FROM alpine:latest

WORKDIR /root/

# 安装证书（HTTPS 请求需要）
RUN apk --no-cache add ca-certificates

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]
```

### 本地测试

```bash
# 构建镜像
docker build -t lowcode-backend .

# 运行容器
docker run -p 8080:8080 \
  -e DATABASE_URL="your-supabase-url" \
  -e JWT_SECRET="your-secret" \
  lowcode-backend
```

---

## 🚀 Railway 部署

### 1. 准备工作

1. 将代码推送到 GitHub
2. 注册 [Railway](https://railway.app/)
3. 连接你的 GitHub 账号

### 2. 创建项目

1. New Project → Deploy from GitHub repo
2. 选择你的仓库
3. Railway 会自动检测 Dockerfile 或 go.mod

### 3. 配置环境变量

在 Railway Dashboard → Variables 中添加：

```
DATABASE_URL=postgres://user:pass@host:5432/db
JWT_SECRET=your-super-secret-key
CLERK_PUBLIC_KEY=your-clerk-public-key
PORT=8080
```

### 4. 配置 Procfile (可选)

```procfile
# Procfile
web: ./main
```

### 5. 获取域名

部署成功后，Railway 会提供：

- `your-app.railway.app` 公网域名
- 支持自定义域名

---

## 🔧 环境变量管理

### Go 代码中读取

```go
import "os"

func main() {
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    dbURL := os.Getenv("DATABASE_URL")
    if dbURL == "" {
        panic("DATABASE_URL 未设置")
    }

    // 启动服务
    r := setupRouter()
    r.Run(":" + port)
}
```

### 本地开发：.env 文件

```bash
# .env (不要提交到 Git!)
DATABASE_URL=postgres://localhost:5432/lowcode
JWT_SECRET=dev-secret
```

使用 godotenv 加载：

```go
import "github.com/joho/godotenv"

func init() {
    godotenv.Load()  // 加载 .env 文件
}
```

安装：`go get github.com/joho/godotenv`

---

## 📋 部署检查清单

### 部署前

- [ ] 所有敏感信息使用环境变量
- [ ] .env 文件已添加到 .gitignore
- [ ] Dockerfile 能本地构建成功
- [ ] 所有 API 已测试通过

### 部署后

- [ ] 健康检查接口可访问 `GET /health`
- [ ] 数据库连接正常
- [ ] WebSocket 连接正常
- [ ] 前端能正常调用 API

---

## 🔒 生产环境配置

### 1. Gin 生产模式

```go
import "github.com/gin-gonic/gin"

func main() {
    // 生产模式（减少日志输出）
    gin.SetMode(gin.ReleaseMode)

    r := gin.New()
    r.Use(gin.Recovery())  // 只保留 panic 恢复
    // ...
}
```

### 2. 优雅关闭

```go
import (
    "context"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
)

func main() {
    r := setupRouter()

    srv := &http.Server{
        Addr:    ":8080",
        Handler: r,
    }

    // 异步启动服务器
    go func() {
        if err := srv.ListenAndServe(); err != nil {
            log.Printf("服务停止: %v", err)
        }
    }()

    // 等待中断信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("正在关闭服务...")

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("强制关闭:", err)
    }

    log.Println("服务已关闭")
}
```

---

## ✅ Day 13 作业

1. 编写 Dockerfile
2. 本地 `docker run` 测试通过
3. 部署到 Railway
4. 获取公网 URL，测试 API

---

## 📖 补充阅读

- Docker 最佳实践: https://docs.docker.com/develop/develop-images/dockerfile_best-practices/
- Railway 文档: https://docs.railway.app/
- Go 生产部署: https://blog.golang.org/deploying-go-servers-with-docker
