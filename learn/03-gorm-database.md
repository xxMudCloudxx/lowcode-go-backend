# GORM 数据库操作 (Day 4-5)

> 目标：使用 GORM 操作 PostgreSQL，完成数据持久化

## 📚 核心学习资源

| 资源                     | 链接                        | 说明       |
| ------------------------ | --------------------------- | ---------- |
| **GORM 官方文档 (中文)** | https://gorm.io/zh_CN/docs/ | ⭐ 首选    |
| Supabase 文档            | https://supabase.com/docs   | 数据库托管 |

---

## 🚀 快速开始

### 1. 安装依赖

```bash
go get -u gorm.io/gorm
go get -u gorm.io/driver/postgres
```

### 2. 连接 Supabase

```go
// internal/data/db.go
package data

import (
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
    // Supabase 连接字符串
    dsn := "host=xxx.supabase.co user=postgres password=xxx dbname=postgres port=5432"

    var err error
    DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        panic("数据库连接失败: " + err.Error())
    }
}
```

---

## 🎯 必修知识点

### 1. 定义模型

```go
// internal/model/page.go
package model

import "time"

type Page struct {
    ID        uint      `gorm:"primaryKey"`
    PageID    string    `gorm:"uniqueIndex;size:64"`
    UserID    string    `gorm:"index;size:64"`  // Clerk User ID
    Schema    string    `gorm:"type:jsonb"`     // PostgreSQL JSONB
    Version   int64     `gorm:"default:0"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

**文档**: https://gorm.io/zh_CN/docs/models.html

---

### 2. 自动迁移

```go
// 自动创建/更新表结构
DB.AutoMigrate(&model.Page{})
```

---

### 3. CRUD 操作

```go
// 创建
page := model.Page{PageID: "page-001", UserID: "user-123", Schema: "{}"}
DB.Create(&page)

// 查询
var page model.Page
DB.Where("page_id = ?", "page-001").First(&page)

// 更新
DB.Model(&page).Update("schema", newSchema)

// 删除
DB.Delete(&page)
```

**文档**: https://gorm.io/zh_CN/docs/create.html

---

### 4. 在 Handler 中使用

```go
// internal/api/handler/page.go
package handler

import (
    "github.com/gin-gonic/gin"
    "lowcode-backend/internal/data"
    "lowcode-backend/internal/model"
)

func GetPage(c *gin.Context) {
    pageID := c.Param("pageId")

    var page model.Page
    result := data.DB.Where("page_id = ?", pageID).First(&page)

    if result.Error != nil {
        c.JSON(404, gin.H{"error": "页面不存在"})
        return
    }

    c.JSON(200, gin.H{
        "pageId": page.PageID,
        "schema": page.Schema,
    })
}

func SavePage(c *gin.Context) {
    pageID := c.Param("pageId")

    var body struct {
        Schema string `json:"schema"`
    }
    c.ShouldBindJSON(&body)

    // Upsert (存在则更新，不存在则创建)
    page := model.Page{PageID: pageID, Schema: body.Schema}
    data.DB.Where("page_id = ?", pageID).
        Assign(model.Page{Schema: body.Schema}).
        FirstOrCreate(&page)

    c.JSON(200, gin.H{"success": true})
}
```

---

## 🔐 配置 Supabase

### 1. 创建 Supabase 项目

1. 访问 https://supabase.com/
2. 创建新项目
3. 进入 Settings → Database → Connection string
4. 复制 URI 格式连接字符串

### 2. 环境变量配置

```go
// 不要硬编码！使用环境变量
import "os"

dsn := os.Getenv("DATABASE_URL")
```

---

## ✅ Day 4-5 作业

实现以下 API（带数据库操作）：

```
POST /api/v1/pages/:pageId/save  → 保存页面 Schema
GET  /api/v1/pages/:pageId       → 读取页面 Schema
```

验证：

1. 调用 POST 保存一个 JSON
2. 调用 GET 能读取回来
3. 再次 POST 更新，GET 能获取最新值

---

## 📖 补充阅读

- 查询: https://gorm.io/zh_CN/docs/query.html
- 更新: https://gorm.io/zh_CN/docs/update.html
- 事务: https://gorm.io/zh_CN/docs/transactions.html
- JSONB 查询: https://gorm.io/zh_CN/docs/data_types.html
