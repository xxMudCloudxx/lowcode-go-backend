package usecase

import (
	"lowercode-go-server/domain/entity"
	"lowercode-go-server/domain/repository"
	"lowercode-go-server/internal/ws"

	jsonpatch "github.com/evanphx/json-patch/v5"
)

// PageUseCase 页面业务逻辑层
// ✅ 注入 Hub，解决"数据双源"问题：
// - 有协同编辑时，内存是 source of truth
// - 无协同编辑时，数据库是 source of truth
type PageUseCase struct {
	repo repository.PageRepository
	hub  *ws.Hub
}

// NewPageUseCase 构造函数，依赖注入
func NewPageUseCase(repo repository.PageRepository, hub *ws.Hub) *PageUseCase {
	return &PageUseCase{repo: repo, hub: hub}
}

// GetPage 获取页面
// 优先从 Hub 内存读取（保证读到最新协同状态），否则读数据库
func (uc *PageUseCase) GetPage(pageID string) (*entity.Page, error) {
	// 1. 优先从 Hub 内存读取（协同编辑中的热数据）
	if room := uc.hub.GetOrCreateRoom(pageID); room != nil {
		snapshot, version := room.GetSnapshot()
		return &entity.Page{
			PageID:  pageID,
			Schema:  string(snapshot),
			Version: version,
		}, nil
	}

	// 2. 内存没有，读数据库（非协同编辑状态）
	return uc.repo.GetByPageID(pageID)
}

// CreatePage 创建新页面
func (uc *PageUseCase) CreatePage(pageID, creatorID string) (*entity.Page, error) {
	// 默认空 Schema
	defaultSchema := `{"rootId":1,"components":{"1":{"id":1,"name":"Page","desc":"页面根节点","children":[]}}}`

	page := &entity.Page{
		PageID:    pageID,
		Schema:    defaultSchema,
		Version:   1,
		CreatorID: creatorID,
	}

	if err := uc.repo.Save(page); err != nil {
		return nil, err
	}
	return page, nil
}

// ApplyPatch 应用 JSON Patch 到当前状态
// 使用 RFC 6902 标准的 json-patch 库
func (uc *PageUseCase) ApplyPatch(currentState, patchBytes []byte) ([]byte, error) {
	patch, err := jsonpatch.DecodePatch(patchBytes)
	if err != nil {
		return nil, err
	}
	return patch.Apply(currentState)
}

// SavePage 直接保存页面（非协同编辑场景）
func (uc *PageUseCase) SavePage(page *entity.Page) error {
	return uc.repo.Save(page)
}
