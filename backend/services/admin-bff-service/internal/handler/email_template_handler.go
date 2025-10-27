package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/payment-platform/pkg/middleware"
	"payment-platform/admin-service/internal/model"
	"payment-platform/admin-service/internal/service"
)

// EmailTemplateHandler 邮件模板HTTP处理器
type EmailTemplateHandler struct {
	templateService service.EmailTemplateService
}

// NewEmailTemplateHandler 创建邮件模板处理器实例
func NewEmailTemplateHandler(templateService service.EmailTemplateService) *EmailTemplateHandler {
	return &EmailTemplateHandler{
		templateService: templateService,
	}
}

// RegisterRoutes 注册路由
func (h *EmailTemplateHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	// 需要认证的路由
	protected := r.Group("/email-templates")
	protected.Use(authMiddleware)
	{
		// 模板管理
		protected.POST("", h.CreateTemplate)
		protected.GET("/:id", h.GetTemplate)
		protected.GET("", h.ListTemplates)
		protected.PUT("/:id", h.UpdateTemplate)
		protected.DELETE("/:id", h.DeleteTemplate)

		// 测试模板
		protected.POST("/:id/test", h.TestTemplate)

		// 发送邮件
		protected.POST("/send", h.SendEmail)
		protected.POST("/send-template", h.SendTemplateEmail)

		// 邮件日志
		protected.GET("/logs", h.ListEmailLogs)

		// 初始化默认模板
		protected.POST("/init-defaults", h.InitDefaultTemplates)
	}
}

// CreateTemplateRequest 创建模板请求
type CreateTemplateRequest struct {
	Code        string                           `json:"code" binding:"required"`
	Name        string                           `json:"name" binding:"required"`
	Subject     string                           `json:"subject" binding:"required"`
	HTMLContent string                           `json:"html_content" binding:"required"`
	TextContent string                           `json:"text_content"`
	Description string                           `json:"description"`
	Category    string                           `json:"category" binding:"required"`
	Variables   []model.EmailTemplateVariable    `json:"variables"`
}

// CreateTemplate 创建邮件模板
func (h *EmailTemplateHandler) CreateTemplate(c *gin.Context) {
	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	template, err := h.templateService.CreateTemplate(c.Request.Context(), &service.CreateTemplateRequest{
		Code:        req.Code,
		Name:        req.Name,
		Subject:     req.Subject,
		HTMLContent: req.HTMLContent,
		TextContent: req.TextContent,
		Description: req.Description,
		Category:    req.Category,
		Variables:   req.Variables,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": template})
}

// GetTemplate 获取模板详情
func (h *EmailTemplateHandler) GetTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID格式错误"})
		return
	}

	template, err := h.templateService.GetTemplate(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": template})
}

// ListTemplates 获取模板列表
func (h *EmailTemplateHandler) ListTemplates(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	// 验证并限制分页参数（防止DoS攻击）
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100 // 最大限制100条/页
	}
	category := c.Query("category")

	// 状态筛选
	var isActive *bool
	if activeStr := c.Query("is_active"); activeStr != "" {
		active := activeStr == "true"
		isActive = &active
	}

	templates, total, err := h.templateService.ListTemplates(c.Request.Context(), page, pageSize, category, isActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      templates,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// UpdateTemplateRequest 更新模板请求
type UpdateTemplateRequest struct {
	Name        string                           `json:"name"`
	Subject     string                           `json:"subject"`
	HTMLContent string                           `json:"html_content"`
	TextContent string                           `json:"text_content"`
	Description string                           `json:"description"`
	IsActive    bool                             `json:"is_active"`
	Variables   []model.EmailTemplateVariable    `json:"variables"`
}

// UpdateTemplate 更新模板
func (h *EmailTemplateHandler) UpdateTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID格式错误"})
		return
	}

	var req UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 获取当前登录用户ID
	claims, _ := middleware.GetClaims(c)

	template, err := h.templateService.UpdateTemplate(c.Request.Context(), &service.UpdateTemplateRequest{
		ID:          id,
		Name:        req.Name,
		Subject:     req.Subject,
		HTMLContent: req.HTMLContent,
		TextContent: req.TextContent,
		Description: req.Description,
		IsActive:    req.IsActive,
		Variables:   req.Variables,
		UpdatedBy:   claims.UserID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": template})
}

// DeleteTemplate 删除模板
func (h *EmailTemplateHandler) DeleteTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID格式错误"})
		return
	}

	if err := h.templateService.DeleteTemplate(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// SendEmailRequest 发送邮件请求
type SendEmailRequest struct {
	To          []string `json:"to" binding:"required"`
	Subject     string   `json:"subject" binding:"required"`
	HTMLContent string   `json:"html_content"`
	TextContent string   `json:"text_content"`
}

// SendEmail 直接发送邮件
func (h *EmailTemplateHandler) SendEmail(c *gin.Context) {
	var req SendEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	if err := h.templateService.SendEmail(c.Request.Context(), &service.SendEmailRequest{
		To:          req.To,
		Subject:     req.Subject,
		HTMLContent: req.HTMLContent,
		TextContent: req.TextContent,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "邮件发送成功"})
}

// SendTemplateEmailRequest 使用模板发送邮件请求
type SendTemplateEmailRequest struct {
	To           []string               `json:"to" binding:"required"`
	TemplateCode string                 `json:"template_code" binding:"required"`
	Data         map[string]interface{} `json:"data" binding:"required"`
}

// SendTemplateEmail 使用模板发送邮件
func (h *EmailTemplateHandler) SendTemplateEmail(c *gin.Context) {
	var req SendTemplateEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	if err := h.templateService.SendTemplateEmail(c.Request.Context(), &service.SendTemplateEmailRequest{
		To:           req.To,
		TemplateCode: req.TemplateCode,
		Data:         req.Data,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "邮件发送成功"})
}

// TestTemplateRequest 测试模板请求
type TestTemplateRequest struct {
	Data map[string]interface{} `json:"data" binding:"required"`
}

// TestTemplate 测试模板渲染
func (h *EmailTemplateHandler) TestTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID格式错误"})
		return
	}

	var req TestTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	html, err := h.templateService.TestTemplate(c.Request.Context(), &service.TestTemplateRequest{
		TemplateID: id,
		Data:       req.Data,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"html": html})
}

// ListEmailLogs 获取邮件日志列表
func (h *EmailTemplateHandler) ListEmailLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	// 验证并限制分页参数（防止DoS攻击）
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100 // 最大限制100条/页
	}
	status := c.Query("status")
	to := c.Query("to")

	logs, total, err := h.templateService.ListEmailLogs(c.Request.Context(), page, pageSize, status, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// InitDefaultTemplates 初始化默认模板
func (h *EmailTemplateHandler) InitDefaultTemplates(c *gin.Context) {
	if err := h.templateService.InitDefaultTemplates(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "默认模板初始化成功"})
}
