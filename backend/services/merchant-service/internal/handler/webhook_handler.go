package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"payment-platform/merchant-service/internal/service"
)

// WebhookHandler Webhook处理器
type WebhookHandler struct {
	webhookService service.WebhookService
}

// NewWebhookHandler 创建Webhook处理器实例
func NewWebhookHandler(webhookService service.WebhookService) *WebhookHandler {
	return &WebhookHandler{
		webhookService: webhookService,
	}
}

// RegisterRoutes 注册Webhook路由
func (h *WebhookHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	webhook := r.Group("/webhooks")
	webhook.Use(authMiddleware)
	{
		webhook.POST("", h.CreateWebhook)            // 创建Webhook配置
		webhook.GET("", h.GetWebhook)                // 获取Webhook配置
		webhook.PUT("", h.UpdateWebhook)             // 更新Webhook配置
		webhook.DELETE("", h.DeleteWebhook)          // 删除Webhook配置
		webhook.POST("/regenerate-secret", h.RegenerateSecret) // 重新生成密钥
	}
}

// CreateWebhook 创建Webhook配置
// @Summary 创建Webhook配置
// @Tags Webhook
// @Accept json
// @Produce json
// @Param request body service.CreateWebhookInput true "Webhook配置"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/webhooks [post]
func (h *WebhookHandler) CreateWebhook(c *gin.Context) {
	// 从JWT中获取商户ID
	merchantID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req service.CreateWebhookInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置商户ID
	req.MerchantID = merchantID.(uuid.UUID)

	webhook, err := h.webhookService.CreateWebhook(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Webhook配置创建成功",
		"data":    webhook,
	})
}

// GetWebhook 获取Webhook配置
// @Summary 获取Webhook配置
// @Tags Webhook
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/webhooks [get]
func (h *WebhookHandler) GetWebhook(c *gin.Context) {
	merchantID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	webhook, err := h.webhookService.GetWebhook(c.Request.Context(), merchantID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": webhook,
	})
}

// UpdateWebhook 更新Webhook配置
// @Summary 更新Webhook配置
// @Tags Webhook
// @Accept json
// @Produce json
// @Param request body service.UpdateWebhookInput true "Webhook配置"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/webhooks [put]
func (h *WebhookHandler) UpdateWebhook(c *gin.Context) {
	merchantID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req service.UpdateWebhookInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	webhook, err := h.webhookService.UpdateWebhook(c.Request.Context(), merchantID.(uuid.UUID), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Webhook配置更新成功",
		"data":    webhook,
	})
}

// DeleteWebhook 删除Webhook配置
// @Summary 删除Webhook配置
// @Tags Webhook
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/webhooks [delete]
func (h *WebhookHandler) DeleteWebhook(c *gin.Context) {
	merchantID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	err := h.webhookService.DeleteWebhook(c.Request.Context(), merchantID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Webhook配置删除成功",
	})
}

// RegenerateSecret 重新生成签名密钥
// @Summary 重新生成Webhook签名密钥
// @Tags Webhook
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/webhooks/regenerate-secret [post]
func (h *WebhookHandler) RegenerateSecret(c *gin.Context) {
	merchantID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	webhook, err := h.webhookService.RegenerateSecret(c.Request.Context(), merchantID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "密钥重新生成成功",
		"data":    webhook,
	})
}
