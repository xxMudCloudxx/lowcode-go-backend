package repository

import "lowercode-go-server/domain/entity"

type PageRepository interface {
	// 根据业务 ID 获取页面（不是数据库自增 ID）
	GetByPageID(pageID string) (*entity.Page, error)

	// Create 创建新页面（仅用于首次创建）
	// ⚠️ 禁止使用 GORM Save，它会覆盖 schema 和 version
	Create(page *entity.Page) error

	// UpdateSchema 只更新 Schema（协同编辑的热路径）
	// oldVersion: 上次持久化的版本号，用于乐观锁检查
	// newVersion: 要写入的新版本号（允许跳跃）
	// 如果数据库中的版本与 oldVersion 不匹配，返回 ErrOptimisticLock
	UpdateSchema(pageID string, schema []byte, oldVersion, newVersion int64) error

	// Delete 删除页面
	// ⚠️ 注意：删除前必须先通过 Hub.CloseRoom 关闭内存中的协同房间
	Delete(pageID string) error
}
