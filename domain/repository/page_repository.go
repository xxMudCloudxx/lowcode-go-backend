package repository

import "lowercode-go-server/domain/entity"

type PageRepository interface {
	// 根据业务 ID 获取页面（不是数据库自增 ID）
	GetByPageID(pageID string) (*entity.Page, error)

	// 完整保存（创建或更新整个实体）
	Save(page *entity.Page) error

	// UpdateSchema 只更新 Schema（协同编辑的热路径）
	// oldVersion: 当前已知的版本号，用于乐观锁检查
	// 成功时版本号自动递增为 oldVersion + 1
	// 如果数据库中的版本与 oldVersion 不匹配，返回错误
	UpdateSchema(pageID string, schema []byte, oldVersion int64) error
}
