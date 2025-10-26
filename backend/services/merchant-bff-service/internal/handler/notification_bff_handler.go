package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"payment-platform/merchant-bff-service/internal/client"
)

type NotificationBFFHandler struct {
	notificationClient *client.ServiceClient
}

func NewNotificationBFFHandler(notificationServiceURL string) *NotificationBFFHandler {
	return &NotificationBFFHandler{
		notificationClient: client.NewServiceClient(notificationServiceURL),
	}
}

func (h *NotificationBFFHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	notifications := r.Group("/merchant/notifications")
	notifications.Use(authMiddleware)
	{
		notifications.GET("", h.ListNotifications)
		notifications.PUT("/:id/read", h.MarkAsRead)
		notifications.GET("/webhooks", h.GetWebhookConfig)
		notifications.PUT("/webhooks", h.UpdateWebhookConfig)
		notifications.POST("/webhooks/test", h.TestWebhook)
	}
}

func (h *NotificationBFFHandler) ListNotifications(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	queryParams := map[string]string{
		"merchant_id": merchantID,
		"page":        c.DefaultQuery("page", "1"),
		"page_size":   c.DefaultQuery("page_size", "10"),
		"status":      c.DefaultQuery("status", ""),
		"type":        c.DefaultQuery("type", ""),
	}

	result, statusCode, err := h.notificationClient.Get(c.Request.Context(), "/api/v1/notifications", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *NotificationBFFHandler) MarkAsRead(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	notificationID := c.Param("id")
	if notificationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "通知ID不能为空"})
		return
	}

	req := map[string]interface{}{
		"merchant_id": merchantID,
	}

	result, statusCode, err := h.notificationClient.Put(c.Request.Context(), "/api/v1/notifications/"+notificationID+"/read", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *NotificationBFFHandler) GetWebhookConfig(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	queryParams := map[string]string{
		"merchant_id": merchantID,
	}

	result, statusCode, err := h.notificationClient.Get(c.Request.Context(), "/api/v1/webhooks", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *NotificationBFFHandler) UpdateWebhookConfig(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req["merchant_id"] = merchantID

	result, statusCode, err := h.notificationClient.Put(c.Request.Context(), "/api/v1/webhooks", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *NotificationBFFHandler) TestWebhook(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req["merchant_id"] = merchantID

	result, statusCode, err := h.notificationClient.Post(c.Request.Context(), "/api/v1/webhooks/test", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}
