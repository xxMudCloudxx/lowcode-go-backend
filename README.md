# LowerCode Go Server

低代码编辑器的 Go 后端服务，支持实时协同编辑。

## 项目结构

```
lowercode-go-server/
├── cmd/                    # 应用入口
│   └── main.go             # 主函数 + 优雅停机
│
├── api/                    # API 层 (接收请求)
│   ├── controller/         # 控制器 (处理 HTTP/WS 请求)
│   ├── route/              # 路由配置
│   └── middleware/         # 中间件 (Clerk 鉴权)
│
├── bootstrap/              # 启动配置
│   ├── app.go              # 依赖注入
│   ├── database.go         # PostgreSQL 连接
│   └── env.go              # 环境变量
│
├── domain/                 # 领域层 (核心定义)
│   ├── entity/             # 实体 (Page, User)
│   └── repository/         # Repository 接口
│
├── usecase/                # 用例层 (业务逻辑)
│   └── page_usecase.go     # 页面业务 + json-patch
│
├── repository/             # Repository 实现
│   └── page_repository.go  # GORM 实现
│
├── internal/ws/            # WebSocket 服务 (Focalboard 模式)
│   ├── hub.go              # 房间管理器
│   ├── room.go             # 单个房间 (状态 + 持久化)
│   ├── client.go           # 客户端连接
│   └── message.go          # 消息协议
│
├── docs/                   # 开发文档
└── learn/                  # 学习笔记
```

## 分层职责

| 层              | 职责                      | 依赖方向     |
| --------------- | ------------------------- | ------------ |
| **api**         | 接收请求，调用 usecase    | → usecase    |
| **usecase**     | 业务逻辑，调用 repository | → domain     |
| **repository**  | 数据库操作                | → domain     |
| **domain**      | 核心实体定义              | 无依赖       |
| **internal/ws** | WebSocket 实时服务        | → repository |

## 快速开始

```bash
# 安装依赖
go mod tidy

# 启动服务
go run cmd/main.go
```

## 相关文档

- [开发阶段指南](docs/development-phases.md)
- [实现计划](docs/implementation-plan.md)
- [后端设计指南](docs/go-backend-guide.md)
