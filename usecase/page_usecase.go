package usecase

import (
	"lowercode-go-server/domain/entity"
	domainErrors "lowercode-go-server/domain/errors"
	"lowercode-go-server/domain/repository"
	"lowercode-go-server/internal/ws"

	"gorm.io/datatypes"
)

// PageUseCase 页面业务逻辑层
// 注入 Hub 解决"数据双源"问题：
//   - 有协同编辑时，内存是 source of truth
//   - 无协同编辑时，数据库是 source of truth
type PageUseCase struct {
	repo repository.PageRepository
	hub  *ws.Hub
}

// NewPageUseCase 创建 PageUseCase 实例
func NewPageUseCase(repo repository.PageRepository, hub *ws.Hub) *PageUseCase {
	return &PageUseCase{repo: repo, hub: hub}
}

// GetPage 获取页面
// 优先从 Hub 内存读取（保证读到最新协同状态），否则读数据库。
// 使用只读的 GetRoom 不会创建房间，避免"观察者效应"。
func (uc *PageUseCase) GetPage(pageID string) (*entity.Page, error) {
	// 优先从 Hub 内存读取
	if room := uc.hub.GetRoom(pageID); room != nil {
		snapshot, version := room.GetSnapshot()
		return &entity.Page{
			PageID:  pageID,
			Schema:  datatypes.JSON(snapshot),
			Version: version,
		}, nil
	}

	// 内存没有，读数据库
	return uc.repo.GetByPageID(pageID)
}

// CreatePage 创建新页面
func (uc *PageUseCase) CreatePage(pageID, creatorID string) (*entity.Page, error) {
	defaultSchema := entity.NewDefaultSchema()
	schemaBytes, err := defaultSchema.ToBytes()
	if err != nil {
		return nil, err
	}

	page := &entity.Page{
		PageID:    pageID,
		Schema:    datatypes.JSON(schemaBytes),
		Version:   1,
		CreatorID: creatorID,
	}

	if err := uc.repo.Create(page); err != nil {
		return nil, err
	}
	return page, nil
}

// DeletePage 删除页面
// 执行"先关房间后删数据"的安全删除流程：
//  1. 检查权限：只有创建者才能删除
//  2. 强制关闭内存中的协同房间
//  3. 删除数据库记录
func (uc *PageUseCase) DeletePage(pageID, operatorID string) error {
	// 查出页面，检查权限
	page, err := uc.repo.GetByPageID(pageID)
	if err != nil {
		return err
	}
	if page == nil {
		return domainErrors.ErrPageNotFound
	}

	// 权限检查：只有创建者才能删除
	if page.CreatorID != operatorID {
		return domainErrors.ErrUnauthorized
	}

	// 先关闭内存中的协同房间
	uc.hub.CloseRoom(pageID)

	// 删除数据库记录
	return uc.repo.Delete(pageID)
}

// 注意：SavePage 方法已删除
// 在协同编辑系统中，禁止使用全量 Save，它会覆盖 schema 和 version。
// 如需更新元数据，应使用专门的 UpdateMeta 方法。
