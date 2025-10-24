package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/payment-platform/pkg/auth"
	"payment-platform/cashier-service/internal/service"
)

// CashierHandler 收银台处理器
type CashierHandler struct {
	service service.CashierService
}

// NewCashierHandler 创建处理器实例
func NewCashierHandler(service service.CashierService) *CashierHandler {
	return &CashierHandler{service: service}
}

// RegisterRoutes 注册路由
func (h *CashierHandler) RegisterRoutes(router *gin.RouterGroup) {
	cashier := router.Group("/cashier")
	{
		// 配置管理 (需要商户认证)
		cashier.POST("/configs", h.CreateOrUpdateConfig)
		cashier.GET("/configs", h.GetConfig)
		cashier.DELETE("/configs", h.DeleteConfig)

		// 会话管理 (服务端API,需要商户认证)
		cashier.POST("/sessions", h.CreateSession)
		cashier.GET("/sessions/:token", h.GetSession)
		cashier.POST("/sessions/:token/complete", h.CompleteSession)
		cashier.DELETE("/sessions/:token", h.CancelSession)

		// 日志记录 (公开API,收银台前端调用)
		cashier.POST("/logs", h.RecordLog)

		// 统计分析
		cashier.GET("/analytics", h.GetAnalytics)
	}

	// 管理员API
	admin := router.Group("/admin/cashier")
	{
		admin.GET("/templates", h.ListTemplates)
		admin.POST("/templates", h.CreateTemplate)
		admin.PUT("/templates/:id", h.UpdateTemplate)
		admin.DELETE("/templates/:id", h.DeleteTemplate)
		admin.GET("/stats", h.GetPlatformStats)
	}
}

// CreateOrUpdateConfig 创建或更新配置
func (h *CashierHandler) CreateOrUpdateConfig(c *gin.Context) {
	var input service.ConfigInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input", "message": err.Error()})
		return
	}

	// 从认证中间件获取商户ID
	merchantID, err := getMerchantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	config, err := h.service.CreateOrUpdateConfig(c.Request.Context(), merchantID, &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save config", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": config,
		"message": "success",
	})
}

// GetConfig 获取配置
func (h *CashierHandler) GetConfig(c *gin.Context) {
	merchantID, err := getMerchantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	config, err := h.service.GetConfig(c.Request.Context(), merchantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get config", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": config,
		"message": "success",
	})
}

// DeleteConfig 删除配置
func (h *CashierHandler) DeleteConfig(c *gin.Context) {
	merchantID, err := getMerchantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	err = h.service.DeleteConfig(c.Request.Context(), merchantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete config", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"message": "success",
	})
}

// CreateSession 创建会话
func (h *CashierHandler) CreateSession(c *gin.Context) {
	var input service.SessionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input", "message": err.Error()})
		return
	}

	// 从认证中间件获取商户ID
	merchantID, err := getMerchantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	input.MerchantID = merchantID

	// 获取客户IP
	if input.CustomerIP == "" {
		input.CustomerIP = c.ClientIP()
	}

	session, token, err := h.service.CreateSession(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"session_token": token,
			"session":       session,
			"cashier_url":   "/cashier/checkout/" + token, // 收银台URL
		},
		"message": "success",
	})
}

// GetSession 获取会话
func (h *CashierHandler) GetSession(c *gin.Context) {
	token := c.Param("token")

	session, err := h.service.GetSession(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": session,
		"message": "success",
	})
}

// CompleteSession 完成会话
func (h *CashierHandler) CompleteSession(c *gin.Context) {
	token := c.Param("token")

	var input struct {
		PaymentNo string `json:"payment_no" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	err := h.service.CompleteSession(c.Request.Context(), token, input.PaymentNo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to complete session", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"message": "success",
	})
}

// CancelSession 取消会话
func (h *CashierHandler) CancelSession(c *gin.Context) {
	token := c.Param("token")

	err := h.service.CancelSession(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel session", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"message": "success",
	})
}

// RecordLog 记录日志
func (h *CashierHandler) RecordLog(c *gin.Context) {
	var input service.LogInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	err := h.service.RecordLog(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record log", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"message": "success",
	})
}

// GetAnalytics 获取统计数据
func (h *CashierHandler) GetAnalytics(c *gin.Context) {
	merchantID, err := getMerchantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// 解析时间范围
	startTime, _ := time.Parse(time.RFC3339, c.DefaultQuery("start_time", time.Now().AddDate(0, 0, -7).Format(time.RFC3339)))
	endTime, _ := time.Parse(time.RFC3339, c.DefaultQuery("end_time", time.Now().Format(time.RFC3339)))

	analytics, err := h.service.GetAnalytics(c.Request.Context(), merchantID, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get analytics", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": analytics,
		"message": "success",
	})
}

// getMerchantID 从上下文获取商户ID
func getMerchantID(c *gin.Context) (uuid.UUID, error) {
	claims, exists := c.Get("claims")
	if !exists {
		return uuid.Nil, nil
	}

	jwtClaims, ok := claims.(*auth.Claims)
	if !ok {
		return uuid.Nil, nil
	}

	// 尝试使用 TenantID 或 UserID 作为 merchant_id
	merchantID := jwtClaims.TenantID
	if merchantID == uuid.Nil {
		// 如果 TenantID 为空,使用 UserID
		merchantID = jwtClaims.UserID
	}

	return merchantID, nil
}

// ListTemplates 列出所有模板 (管理员)
func (h *CashierHandler) ListTemplates(c *gin.Context) {
	templates, err := h.service.ListTemplates(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list templates", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"data":    templates,
		"message": "success",
	})
}

// CreateTemplate 创建模板 (管理员)
func (h *CashierHandler) CreateTemplate(c *gin.Context) {
	var input service.TemplateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input", "message": err.Error()})
		return
	}

	template, err := h.service.CreateTemplate(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create template", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"data":    template,
		"message": "success",
	})
}

// UpdateTemplate 更新模板 (管理员)
func (h *CashierHandler) UpdateTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template id"})
		return
	}

	var input service.TemplateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input", "message": err.Error()})
		return
	}

	template, err := h.service.UpdateTemplate(c.Request.Context(), id, &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update template", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"data":    template,
		"message": "success",
	})
}

// DeleteTemplate 删除模板 (管理员)
func (h *CashierHandler) DeleteTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template id"})
		return
	}

	if err := h.service.DeleteTemplate(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete template", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// GetPlatformStats 获取平台统计 (管理员)
func (h *CashierHandler) GetPlatformStats(c *gin.Context) {
	stats, err := h.service.GetPlatformStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get platform stats", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"data":    stats,
		"message": "success",
	})
}
