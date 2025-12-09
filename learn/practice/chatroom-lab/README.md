# Go 聊天室 Lab 🧪

> 一个 CS61A/CS61B 风格的 Go 语言练习项目

## 📋 项目介绍

你将实现一个 **实时聊天室服务器**，通过逐步完成 6 个 Lab，掌握 Go 语言核心概念。

**学习方式：**

1. 阅读每个 Lab 的 `README.md` 了解任务要求
2. 在带有 `// TODO:` 注释的地方编写代码
3. 运行测试检验你的实现是否正确
4. 通过所有测试后进入下一个 Lab

---

## 🗂️ Lab 列表

| Lab                                   | 主题           | 知识点                         | 难度   |
| ------------------------------------- | -------------- | ------------------------------ | ------ |
| [Lab 1](./lab1-basics/README.md)      | 消息与结构体   | 结构体、方法、常量、Switch     | ⭐     |
| [Lab 2](./lab2-functions/README.md)   | 函数与错误处理 | 函数、多返回值、错误处理、闭包 | ⭐     |
| [Lab 3](./lab3-collections/README.md) | 数据集合       | 切片、Map、Range               | ⭐⭐   |
| [Lab 4](./lab4-concurrency/README.md) | 并发基础       | 协程、通道、通道缓冲           | ⭐⭐   |
| [Lab 5](./lab5-channels/README.md)    | 通道进阶       | Select、超时、非阻塞、关闭     | ⭐⭐⭐ |
| [Lab 6](./lab6-sync/README.md)        | 同步原语       | 互斥锁、WaitGroup、原子计数    | ⭐⭐⭐ |

---

## 🚀 快速开始

```bash
# 进入项目目录
cd learn/practice/chatroom-lab

# 初始化模块（只需运行一次）
go mod init chatroom-lab

# 进入第一个 Lab
cd lab1-basics

# 运行测试（检查你的实现）
go test -v
```

---

## 📝 代码规范

1. **只修改 `// TODO:` 标记的地方**
2. **不要修改测试文件** (`*_test.go`)
3. **不要修改函数签名**（参数和返回类型）
4. 如果卡住了，可以查看 `hints.md` 获取提示

---

## ✅ 完成标准

每个 Lab 需要通过所有测试：

```bash
go test -v
# 看到 PASS 表示通过
```

全部 Lab 完成后，运行完整聊天室：

```bash
cd final
go run .
```
