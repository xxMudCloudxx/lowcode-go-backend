# Go 语法基础 (Day 1-2)

> 目标：掌握 Go 核心语法，能够处理 JSON 数据和基础并发

## 📚 核心学习资源

### 主要教程

| 资源                     | 链接                              | 说明                  |
| ------------------------ | --------------------------------- | --------------------- |
| **Go by Example 中文版** | https://gobyexample-cn.github.io/ | ⭐ 推荐首选，简洁实用 |
| Go 官方教程              | https://go.dev/tour/welcome/1     | 交互式学习            |
| Go 语言圣经              | https://gopl-zh.github.io/        | 深入学习参考          |

---

## 🎯 必修知识点

### 1. 变量与基本类型

```go
// := 短变量声明（函数内使用）
name := "LowCode"
count := 42
isReady := true

// var 显式声明
var port int = 8080
```

**学习链接**: https://gobyexample-cn.github.io/variables

---

### 2. Structs 结构体 ⭐ 重点

```go
// 定义结构体（对应 JSON 数据）
type Component struct {
    ID       int    `json:"id"`
    Name     string `json:"name"`
    ParentID *int   `json:"parentId,omitempty"`
}

// 创建实例
btn := Component{ID: 1, Name: "Button"}
```

**学习链接**: https://gobyexample-cn.github.io/structs

---

### 3. Methods 方法

```go
type Component struct {
    ID   int
    Name string
}

// 给结构体绑定方法
func (c *Component) Rename(newName string) {
    c.Name = newName
}

// 使用
btn := &Component{ID: 1, Name: "Button"}
btn.Rename("SubmitBtn")
```

**学习链接**: https://gobyexample-cn.github.io/methods

---

### 4. Interfaces 接口

```go
// 定义接口
type PageService interface {
    GetPage(id string) ([]byte, error)
    SavePage(id string, data []byte) error
}

// 任何实现这两个方法的类型都满足该接口
```

**学习链接**: https://gobyexample-cn.github.io/interfaces

---

### 5. Goroutines & Channels ⭐ 并发核心

```go
// Goroutine：轻量级线程
go func() {
    fmt.Println("异步执行")
}()

// Channel：Goroutine 间通信
messages := make(chan string)

go func() {
    messages <- "ping"  // 发送
}()

msg := <-messages  // 接收
fmt.Println(msg)
```

**学习链接**:

- Goroutines: https://gobyexample-cn.github.io/goroutines
- Channels: https://gobyexample-cn.github.io/channels

---

### 6. JSON 处理 ⭐ 必须掌握

```go
import "encoding/json"

type Component struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

// 序列化：Go -> JSON
btn := Component{ID: 1, Name: "Button"}
data, _ := json.Marshal(btn)
// data = []byte(`{"id":1,"name":"Button"}`)

// 反序列化：JSON -> Go
var c Component
json.Unmarshal([]byte(`{"id":2,"name":"Input"}`), &c)
```

**学习链接**: https://gobyexample-cn.github.io/json

---

### 7. 错误处理

```go
// Go 使用显式错误返回，没有 try-catch
result, err := doSomething()
if err != nil {
    log.Printf("出错了: %v", err)
    return err
}
```

**学习链接**: https://gobyexample-cn.github.io/errors

---

## ✅ Day 2 作业

写一个 CLI 工具，完成以下功能：

1. 读取一个 JSON 文件
2. 修改里面的某个字段
3. 保存回文件

```go
// 参考框架
package main

import (
    "encoding/json"
    "os"
)

type Page struct {
    Title string `json:"title"`
    // ...
}

func main() {
    // 1. 读取文件
    data, _ := os.ReadFile("page.json")

    // 2. 解析 JSON
    var page Page
    json.Unmarshal(data, &page)

    // 3. 修改字段
    page.Title = "新标题"

    // 4. 保存回文件
    newData, _ := json.MarshalIndent(page, "", "  ")
    os.WriteFile("page.json", newData, 0644)
}
```

---

## 📖 补充阅读

- Maps: https://gobyexample-cn.github.io/maps
- Slices: https://gobyexample-cn.github.io/slices
- Pointers: https://gobyexample-cn.github.io/pointers
- Mutexes: https://gobyexample-cn.github.io/mutexes (Week 2 用到)
