package controller

import (
	"net/http"

	"lowercode-go-server/usecase"

	"github.com/gin-gonic/gin"
)

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
		c.JSON(http.StatusBadRequest, gin.H{"error": "pageId 不能为空"})
		return
	}

	page, err := pc.pageUseCase.GetPage(pageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if page == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "页面不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pageId":  page.PageID,
		"schema":  page.Schema,
		"version": page.Version,
	})
}

// CreatePage 创建新页面
// POST /api/pages
// 请求体: { "pageId": "xxx" }
func (pc *PageController) CreatePage(c *gin.Context) {
	var req struct {
		PageID string `json:"pageId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pageId 不能为空"})
		return
	}

	// 从 Clerk 中间件获取用户 ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未获取到用户信息"})
		return
	}

	page, err := pc.pageUseCase.CreatePage(req.PageID, userID.(string))
	if err != nil {
		// 判断是否是重复创建
		if err.Error() == "UNIQUE constraint failed" ||
			(len(err.Error()) > 0 && err.Error()[:9] == "duplicate") {
			c.JSON(http.StatusConflict, gin.H{"error": "页面已存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"pageId":  page.PageID,
		"schema":  page.Schema,
		"version": page.Version,
	})
}

// DeletePage 删除页面
// DELETE /api/pages/:pageId
// ⚠️ 危险操作：会强制关闭协同编辑房间，踢出所有在线用户
func (pc *PageController) DeletePage(c *gin.Context) {
	pageID := c.Param("pageId")
	if pageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pageId 不能为空"})
		return
	}

	// 从 Clerk 中间件获取用户 ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未获取到用户信息"})
		return
	}

	// TODO: 检查用户是否有权限删除（是否是创建者）
	// page, _ := pc.pageUseCase.GetPage(pageID)
	// if page.CreatorID != userID.(string) { ... }
	_ = userID // 暂时忽略，后续添加权限检查

	if err := pc.pageUseCase.DeletePage(pageID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "页面已删除",
		"pageId":  pageID,
	})
}
