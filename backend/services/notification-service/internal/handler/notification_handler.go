package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"payment-platform/notification-service/internal/service"
)

// NotificationHandler 通知处理器
type NotificationHandler struct {
	notificationService service.NotificationService
}

// NewNotificationHandler 创建通知处理器实例
func NewNotificationHandler(notificationService service.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// SendEmail 发送邮件
// @Summary 发送邮件
// @Tags Notification
// @Accept json
// @Produce json
// @Param request body service.SendEmailRequest true "发送邮件请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/notifications/email [post]
func (h *NotificationHandler) SendEmail(c *gin.Context) {
	var req service.SendEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 从上下文获取商户ID
	merchantID, exists := c.Get("merchant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}
	req.MerchantID = merchantID.(uuid.UUID)

	// 发送邮件
	if err := h.notificationService.SendEmail(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "邮件发送成功"})
}

// SendSMS 发送短信
// @Summary 发送短信
// @Tags Notification
// @Accept json
// @Produce json
// @Param request body service.SendSMSRequest true "发送短信请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/notifications/sms [post]
func (h *NotificationHandler) SendSMS(c *gin.Context) {
	var req service.SendSMSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 从上下文获取商户ID
	merchantID, exists := c.Get("merchant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}
	req.MerchantID = merchantID.(uuid.UUID)

	// 发送短信
	if err := h.notificationService.SendSMS(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "短信发送成功"})
}

// SendWebhook 发送 Webhook
// @Summary 发送 Webhook
// @Tags Notification
// @Accept json
// @Produce json
// @Param request body service.SendWebhookRequest true "发送 Webhook 请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/notifications/webhook [post]
func (h *NotificationHandler) SendWebhook(c *gin.Context) {
	var req service.SendWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 从上下文获取商户ID
	merchantID, exists := c.Get("merchant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}
	req.MerchantID = merchantID.(uuid.UUID)

	// 发送 Webhook
	if err := h.notificationService.SendWebhook(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Webhook 发送成功"})
}

// SendEmailByTemplate 使用模板发送邮件
// @Summary 使用模板发送邮件
// @Tags Notification
// @Accept json
// @Produce json
// @Param request body service.SendEmailByTemplateRequest true "使用模板发送邮件请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/notifications/email/template [post]
func (h *NotificationHandler) SendEmailByTemplate(c *gin.Context) {
	var req service.SendEmailByTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 从上下文获取商户ID
	merchantID, exists := c.Get("merchant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}
	req.MerchantID = merchantID.(uuid.UUID)

	// 发送邮件
	if err := h.notificationService.SendEmailByTemplate(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "邮件发送成功"})
}

// ListNotifications 列出通知
// @Summary 列出通知
// @Tags Notification
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/notifications [get]
func (h *NotificationHandler) ListNotifications(c *gin.Context) {
	// TODO: 解析查询参数
	c.JSON(http.StatusOK, gin.H{"message": "功能开发中"})
}

// CreateWebhookEndpoint 创建 Webhook 端点
// @Summary 创建 Webhook 端点
// @Tags Webhook
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/webhooks/endpoints [post]
func (h *NotificationHandler) CreateWebhookEndpoint(c *gin.Context) {
	// TODO: 实现创建 Webhook 端点
	c.JSON(http.StatusOK, gin.H{"message": "功能开发中"})
}

// ListWebhookEndpoints 列出 Webhook 端点
// @Summary 列出 Webhook 端点
// @Tags Webhook
// @Accept json
// @Produce json
// @Success 200 {array} map[string]interface{}
// @Router /api/v1/webhooks/endpoints [get]
func (h *NotificationHandler) ListWebhookEndpoints(c *gin.Context) {
	// 从上下文获取商户ID
	merchantID, exists := c.Get("merchant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 列出端点
	endpoints, err := h.notificationService.ListWebhookEndpoints(c.Request.Context(), merchantID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, endpoints)
}

// RegisterRoutes 注册路由
func (h *NotificationHandler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	{
		// 通知发送
		api.POST("/notifications/email", h.SendEmail)
		api.POST("/notifications/sms", h.SendSMS)
		api.POST("/notifications/webhook", h.SendWebhook)
		api.POST("/notifications/email/template", h.SendEmailByTemplate)
		api.GET("/notifications", h.ListNotifications)

		// Webhook 端点管理
		api.POST("/webhooks/endpoints", h.CreateWebhookEndpoint)
		api.GET("/webhooks/endpoints", h.ListWebhookEndpoints)
	}
}
