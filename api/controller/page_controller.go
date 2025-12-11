package controller

import (
	"errors"
	"net/http"

	"lowercode-go-server/api/middleware"
	domainErrors "lowercode-go-server/domain/errors"
	"lowercode-go-server/usecase"

	"github.com/gin-gonic/gin"
)

// ========== Response DTOs ==========
// 使用强类型结构体代替 gin.H，便于文档化和类型检查

// PageResponse 页面响应结构
type PageResponse struct {
	PageID  string      `json:"pageId"`
	Schema  interface{} `json:"schema"`
	Version int64       `json:"version"`
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// MessageResponse 消息响应结构
type MessageResponse struct {
	Message string `json:"message"`
	PageID  string `json:"pageId,omitempty"`
}

// ========== Controller ==========

// PageController 页面相关的 HTTP 控制器
type PageController struct {
	pageUseCase *usecase.PageUseCase
}

// NewPageController 构造函数
func NewPageController(pageUseCase *usecase.PageUseCase) *PageController {
	return &PageController{pageUseCase: pageUseCase}
}

// GetPage 获取页面
// GET /api/pages/:pageId
// 支持 Hub 内存优先读取，回退到数据库
func (pc *PageController) GetPage(c *gin.Context) {
	pageID := c.Param("pageId")
	if pageID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "pageId 不能为空"})
		return
	}

	page, err := pc.pageUseCase.GetPage(pageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	if page == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "页面不存在"})
		return
	}

	c.JSON(http.StatusOK, PageResponse{
		PageID:  page.PageID,
		Schema:  page.Schema,
		Version: page.Version,
	})
}

// CreatePageRequest 创建页面请求结构
type CreatePageRequest struct {
	PageID string `json:"pageId" binding:"required"`
}

// CreatePage 创建新页面
// POST /api/pages
// 请求体: { "pageId": "xxx" }
func (pc *PageController) CreatePage(c *gin.Context) {
	var req CreatePageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "pageId 不能为空"})
		return
	}

	// 从 Clerk 中间件获取用户 ID（使用常量避免魔法字符串）
	userID, exists := c.Get(middleware.ContextKeyUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "未获取到用户信息"})
		return
	}

	page, err := pc.pageUseCase.CreatePage(req.PageID, userID.(string))
	if err != nil {
		// 使用 errors.Is 判断业务错误类型，而非字符串匹配
		if errors.Is(err, domainErrors.ErrPageAlreadyExists) {
			c.JSON(http.StatusConflict, ErrorResponse{Error: "页面已存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, PageResponse{
		PageID:  page.PageID,
		Schema:  page.Schema,
		Version: page.Version,
	})
}

// DeletePage 删除页面
// DELETE /api/pages/:pageId
// ⚠️ 危险操作：会强制关闭协同编辑房间，踢出所有在线用户
func (pc *PageController) DeletePage(c *gin.Context) {
	pageID := c.Param("pageId")
	if pageID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "pageId 不能为空"})
		return
	}

	// 从 Clerk 中间件获取用户 ID（使用常量避免魔法字符串）
	userID, exists := c.Get(middleware.ContextKeyUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "未获取到用户信息"})
		return
	}

	// 调用 UseCase，权限检查在业务层进行
	if err := pc.pageUseCase.DeletePage(pageID, userID.(string)); err != nil {
		// 使用 errors.Is 判断业务错误类型
		switch {
		case errors.Is(err, domainErrors.ErrPageNotFound):
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "页面不存在"})
		case errors.Is(err, domainErrors.ErrUnauthorized):
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "无权限删除此页面"})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, MessageResponse{
		Message: "页面已删除",
		PageID:  pageID,
	})
}
