package repository

import (
	"errors"

	"lowercode-go-server/domain/entity"
	domainRepo "lowercode-go-server/domain/repository"

	"gorm.io/gorm"
)

// ErrOptimisticLock 乐观锁冲突错误
// 当数据库中的版本与期望版本不匹配时返回此错误
var ErrOptimisticLock = errors.New("optimistic lock error: version mismatch, please refresh and retry")

// pageRepository GORM 实现 PageRepository 接口
// 同时实现 ws.PageService 接口供 Hub 使用
type pageRepository struct {
	db *gorm.DB
}

// NewPageRepository 构造函数
func NewPageRepository(db *gorm.DB) domainRepo.PageRepository {
	return &pageRepository{db: db}
}

// ================= domain.PageRepository 接口实现 =================

// GetByPageID 根据业务 ID 查询页面
func (r *pageRepository) GetByPageID(pageID string) (*entity.Page, error) {
	var page entity.Page
	err := r.db.Where("page_id = ?", pageID).First(&page).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil // 返回 nil 表示不存在，调用方需处理
	}
	return &page, err
}

// Save 创建或更新整个页面实体
func (r *pageRepository) Save(page *entity.Page) error {
	return r.db.Save(page).Error
}

// UpdateSchema 只更新 Schema 字段（协同编辑热路径）
// 实现乐观锁：只有当数据库中的 version == oldVersion 时才更新
func (r *pageRepository) UpdateSchema(pageID string, schema []byte, oldVersion int64) error {
	result := r.db.Model(&entity.Page{}).
		// ⚠️ 关键：WHERE 必须包含版本检查，实现乐观锁
		Where("page_id = ? AND version = ?", pageID, oldVersion).
		Updates(map[string]interface{}{
			"schema":  string(schema),
			"version": oldVersion + 1, // 版本号递增
		})

	if result.Error != nil {
		return result.Error
	}

	// ⚠️ 关键：检查是否真的更新了记录
	// 如果 RowsAffected == 0，说明版本冲突或页面不存在
	if result.RowsAffected == 0 {
		return ErrOptimisticLock
	}

	return nil
}

// ================= ws.PageService 接口实现 =================
// 这些方法供 Hub 直接调用，无需额外适配器

// GetPageState 获取页面状态（供 Hub 使用）
func (r *pageRepository) GetPageState(pageID string) ([]byte, int64, error) {
	page, err := r.GetByPageID(pageID)
	if err != nil {
		return nil, 0, err
	}
	if page == nil {
		// 页面不存在，返回默认空 Schema
		emptySchema := `{"rootId":1,"components":{"1":{"id":1,"name":"Page","props":{},"desc":"页面","parentId":null}}}`
		return []byte(emptySchema), 1, nil
	}
	return []byte(page.Schema), page.Version, nil
}

// SavePageState 保存页面状态（供 Hub 使用）
func (r *pageRepository) SavePageState(pageID string, state []byte, version int64) error {
	return r.UpdateSchema(pageID, state, version)
}
