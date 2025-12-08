# Gin Web 框架 (Day 3)

> 目标：搭建标准项目结构，跑通 Hello World Web Server

## 📚 核心学习资源

| 资源                    | 链接                                  | 说明     |
| ----------------------- | ------------------------------------- | -------- |
| **Gin 官方文档 (中文)** | https://gin-gonic.com/zh-cn/docs/     | ⭐ 首选  |
| Gin GitHub 示例         | https://github.com/gin-gonic/examples | 实战参考 |

---

## 🚀 快速开始

### 1. 安装 Gin

```bash
go get -u github.com/gin-gonic/gin
```

### 2. Hello World

```go
// cmd/server/main.go
package main

import "github.com/gin-gonic/gin"

func main() {
    r := gin.Default()

    r.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "pong",
        })
    })

    r.Run(":8080")  // 监听 localhost:8080
}
```

运行：`go run cmd/server/main.go`

访问：http://localhost:8080/ping

---

## 🎯 必修知识点

### 1. 路由定义

```go
// RESTful API
r.GET("/pages/:id", getPage)      // 获取
r.POST("/pages", createPage)      // 创建
r.PUT("/pages/:id", updatePage)   // 更新
r.DELETE("/pages/:id", deletePage) // 删除

// 路由分组
v1 := r.Group("/api/v1")
{
    v1.GET("/pages", listPages)
    v1.GET("/pages/:id", getPage)
}
```

**文档**: https://gin-gonic.com/zh-cn/docs/examples/param-in-path/

---

### 2. 获取请求参数

```go
func getPage(c *gin.Context) {
    // 路径参数 /pages/:id
    id := c.Param("id")

    // 查询参数 ?name=xxx
    name := c.Query("name")

    // POST Body
    var body struct {
        Title string `json:"title"`
    }
    c.ShouldBindJSON(&body)

    c.JSON(200, gin.H{"id": id, "title": body.Title})
}
```

---

### 3. 中间件

```go
// 自定义中间件
func Logger() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()

        c.Next()  // 处理请求

        latency := time.Since(start)
        log.Printf("[%s] %s - %v", c.Request.Method, c.Request.URL.Path, latency)
    }
}

// 使用中间件
r.Use(Logger())
```

**文档**: https://gin-gonic.com/zh-cn/docs/examples/custom-middleware/

---

### 4. CORS 跨域配置

```go
import "github.com/gin-contrib/cors"

func main() {
    r := gin.Default()

    // 配置 CORS
    r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"http://localhost:3000"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
        AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
        AllowCredentials: true,
    }))

    // ...
}
```

安装：`go get github.com/gin-contrib/cors`

---

## 📁 标准项目结构

```
/lowCode-go-backend
├── cmd
│   └── server
│       └── main.go           # 入口
├── internal
│   ├── api
│   │   ├── handler           # 控制器
│   │   │   └── page.go
│   │   ├── middleware        # 中间件
│   │   │   └── auth.go
│   │   └── router.go         # 路由定义
│   ├── core                  # 业务逻辑
│   └── data                  # 数据层
├── config
│   └── config.yaml
├── go.mod
└── go.sum
```

---

## ✅ Day 3 作业

1. 初始化项目：`go mod init lowcode-backend`
2. 按上面的目录结构创建文件
3. 实现以下 API：

```
GET  /api/v1/ping              → {"message": "pong"}
GET  /api/v1/pages/:id         → {"id": ":id", "title": "Demo"}
POST /api/v1/pages             → 返回请求体
```

---

## 📖 补充阅读

- 路由参数: https://gin-gonic.com/zh-cn/docs/examples/param-in-path/
- JSON 绑定: https://gin-gonic.com/zh-cn/docs/examples/binding-and-validation/
- 错误处理: https://gin-gonic.com/zh-cn/docs/examples/custom-http-config/
