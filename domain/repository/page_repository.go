package repository

import "lowercode-go-server/domain/entity"

type PageRepository interface {
	// 根据业务 ID 获取页面（不是数据库自增 ID）
	GetByPageID(pageID string) (*entity.Page, error)

	 // 完整保存（创建或更新整个实体）
	Save(page *entity.Page) error

	// 只更新 Schema（协同编辑的热路径，性能优化）
	UpdateSchema(pageID string, schema []byte, version int64) error
}