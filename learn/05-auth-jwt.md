# JWT 鉴权 (Day 4-5)

> 目标：实现用户认证，保护 API 接口

## 📚 核心学习资源

| 资源              | 链接                              | 说明             |
| ----------------- | --------------------------------- | ---------------- |
| **golang-jwt 库** | https://github.com/golang-jwt/jwt | ⭐ Go JWT 标准库 |
| Clerk 文档        | https://clerk.com/docs            | 第三方认证服务   |
| JWT 介绍          | https://jwt.io/introduction       | JWT 原理         |

---

## 🚀 快速开始

### 1. 安装依赖

```bash
go get github.com/golang-jwt/jwt/v5
```

### 2. JWT 验证中间件

```go
// internal/api/middleware/auth.go
package middleware

import (
    "strings"
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. 从 Header 获取 Token
        authHeader := c.GetHeader("Authorization")
        if !strings.HasPrefix(authHeader, "Bearer ") {
            c.JSON(401, gin.H{"error": "未授权"})
            c.Abort()
            return
        }

        tokenString := strings.TrimPrefix(authHeader, "Bearer ")

        // 2. 解析验证 Token
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            // 验证签名算法
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("无效的签名算法")
            }
            // 返回密钥（从环境变量获取）
            return []byte(os.Getenv("JWT_SECRET")), nil
        })

        if err != nil || !token.Valid {
            c.JSON(401, gin.H{"error": "Token 无效"})
            c.Abort()
            return
        }

        // 3. 提取用户信息
        claims := token.Claims.(jwt.MapClaims)
        userID := claims["sub"].(string)

        // 4. 存入 Context，后续 Handler 可用
        c.Set("userID", userID)

        c.Next()
    }
}
```

---

## 🎯 必修知识点

### 1. JWT 结构

```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.  // Header
eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4ifQ.  // Payload
SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c  // Signature

Header:  { "alg": "HS256", "typ": "JWT" }
Payload: { "sub": "user-123", "name": "John", "exp": 1234567890 }
```

### 2. 在 Handler 中获取用户

```go
func GetPage(c *gin.Context) {
    // 从 Context 获取用户 ID
    userID := c.GetString("userID")

    // 只返回该用户的页面
    var page model.Page
    data.DB.Where("user_id = ? AND page_id = ?", userID, pageID).First(&page)
}
```

### 3. 使用中间件

```go
// internal/api/router.go
func SetupRouter() *gin.Engine {
    r := gin.Default()

    // 公开接口
    r.GET("/health", healthCheck)

    // 需要认证的接口
    auth := r.Group("/api/v1")
    auth.Use(middleware.AuthMiddleware())
    {
        auth.GET("/pages/:pageId", handler.GetPage)
        auth.POST("/pages/:pageId/save", handler.SavePage)
    }

    return r
}
```

---

## 🔐 Clerk 集成

### 1. 前端获取 Token

```typescript
// React 前端
import { useAuth } from "@clerk/clerk-react";

function App() {
  const { getToken } = useAuth();

  async function callAPI() {
    const token = await getToken();

    fetch("/api/v1/pages/page-001", {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });
  }
}
```

### 2. Clerk JWT 验证

Clerk 使用 RS256 (RSA) 签名，需要使用公钥验证：

```go
import "github.com/golang-jwt/jwt/v5"

func validateClerkToken(tokenString string) (*jwt.Token, error) {
    // Clerk 公钥 (从 Clerk Dashboard 获取)
    publicKey := os.Getenv("CLERK_PUBLIC_KEY")

    key, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(publicKey))

    return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
            return nil, fmt.Errorf("无效的签名算法")
        }
        return key, nil
    })
}
```

---

## 🎫 WebSocket Ticket 机制

WebSocket 不能使用 Header，需要特殊处理：

```go
// 1. HTTP API 生成临时 Ticket
func CreateTicket(c *gin.Context) {
    userID := c.GetString("userID")

    ticket := generateRandomString(32)

    // 存入 Redis，10 秒过期
    redis.Set("ws_ticket:"+ticket, userID, 10*time.Second)

    c.JSON(200, gin.H{"ticket": ticket})
}

// 2. WebSocket 连接时验证 Ticket
func ServeWS(c *gin.Context) {
    ticket := c.Query("ticket")

    userID, err := redis.GetDel("ws_ticket:" + ticket).Result()
    if err != nil {
        c.JSON(401, gin.H{"error": "Ticket 无效"})
        return
    }

    // 升级 WebSocket...
}
```

---

## ✅ Day 4-5 作业（鉴权部分）

1. 实现 `AuthMiddleware`
2. 给 `/api/v1/pages` 路由添加认证
3. 测试：不带 Token 访问返回 401，带 Token 正常访问

---

## 📖 补充阅读

- golang-jwt 文档: https://pkg.go.dev/github.com/golang-jwt/jwt/v5
- Clerk Go SDK: https://clerk.com/docs/references/backend/go
- JWT 最佳实践: https://auth0.com/blog/a-look-at-the-latest-draft-for-jwt-bcp/
