package errors

import "errors"

// ================= 业务领域错误定义 =================
// 所有业务逻辑相关的错误统一在此定义，避免跨包重复定义

// ErrPageNotFound 页面不存在错误
// 当尝试操作一个不存在于数据库中的页面时返回此错误
var ErrPageNotFound = errors.New("page not found in database")

// ErrOptimisticLock 乐观锁冲突错误
// 当数据库中的版本与期望版本不匹配时返回此错误
var ErrOptimisticLock = errors.New("optimistic lock error: version mismatch, please refresh and retry")

// ErrPageAlreadyExists 页面已存在错误
// 当尝试创建一个已存在的页面时返回此错误（pageId 唯一约束冲突）
var ErrPageAlreadyExists = errors.New("page already exists")

// ErrUnauthorized 无权限错误
// 当用户尝试操作没有权限的资源时返回此错误（如删除他人的页面）
var ErrUnauthorized = errors.New("unauthorized: you don't have permission to perform this action")

// ErrRoomClosing 房间正在关闭错误
// 当 WebSocket 尝试连接一个正在关闭的房间时返回此错误，客户端应重试
var ErrRoomClosing = errors.New("room is closing, please retry")
