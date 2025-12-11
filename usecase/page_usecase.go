package usecase

import (
	"lowercode-go-server/domain/entity"
	domainErrors "lowercode-go-server/domain/errors"
	"lowercode-go-server/domain/repository"
	"lowercode-go-server/internal/ws"

	"gorm.io/datatypes"
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
// ⚠️ 修复"观察者效应"问题：使用只读的 GetRoom，不创建房间
// 优先从 Hub 内存读取（保证读到最新协同状态），否则读数据库
func (uc *PageUseCase) GetPage(pageID string) (*entity.Page, error) {
	// 1. 优先从 Hub 内存读取（协同编辑中的热数据）
	// ⚠️ 使用 GetRoom 而非 GetOrCreateRoom，避免 HTTP GET 触发房间创建
	if room := uc.hub.GetRoom(pageID); room != nil {
		snapshot, version := room.GetSnapshot()
		return &entity.Page{
			PageID:  pageID,
			Schema:  datatypes.JSON(snapshot), // datatypes.JSON 是 []byte 别名
			Version: version,
		}, nil
	}

	// 2. 内存没有，读数据库（非协同编辑状态）
	return uc.repo.GetByPageID(pageID)
}

// CreatePage 创建新页面
// ⚠️ 使用强类型的 NewDefaultSchema，避免硬编码 JSON 字符串
// ⚠️ 使用 Create 而非 Save，防止意外覆盖已存在的页面
func (uc *PageUseCase) CreatePage(pageID, creatorID string) (*entity.Page, error) {
	// 使用工厂方法创建默认 Schema（类型安全）
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

	// ✅ 使用 Create 而非 Save，只能创建新记录
	if err := uc.repo.Create(page); err != nil {
		return nil, err
	}
	return page, nil
}

// DeletePage 删除页面
// ⚠️ 实现"先杀人，后毁尸"的安全删除流程：
// 1. 检查权限：只有创建者才能删除
// 2. 先斩：强制关闭内存中的协同房间（通知所有在线用户）
// 3. 后奏：删除数据库记录
func (uc *PageUseCase) DeletePage(pageID, operatorID string) error {
	// 1. 先查出页面，检查权限
	page, err := uc.repo.GetByPageID(pageID)
	if err != nil {
		return err
	}
	if page == nil {
		return domainErrors.ErrPageNotFound
	}

	// 2. 权限检查：只有创建者才能删除（防止 IDOR 越权删除）
	if page.CreatorID != operatorID {
		return domainErrors.ErrUnauthorized
	}

	// 3. 先斩：强制关闭内存中的协同房间
	// 这一步必须在删库之前，防止新的写入
	// Hub.CloseRoom 会广播 PAGE_DELETED 消息给所有在线用户
	uc.hub.CloseRoom(pageID)

	// 4. 后奏：删除数据库记录
	return uc.repo.Delete(pageID)
}

// ⚠️ SavePage 方法已删除
// 在协同编辑系统中，禁止使用全量 Save，它会覆盖 schema 和 version
// 如需更新元数据（标题、描述），应使用专门的 UpdateMeta 方法
