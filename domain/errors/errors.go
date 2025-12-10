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
