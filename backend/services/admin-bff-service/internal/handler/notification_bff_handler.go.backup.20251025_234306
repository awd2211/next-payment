package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"payment-platform/admin-service/internal/client"
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
	admin := r.Group("/admin/notifications")
	admin.Use(authMiddleware)
	{
		// 通知列表
		admin.GET("", h.ListNotifications)
		admin.GET("/:id", h.GetNotification)
		admin.POST("", h.CreateNotification)
		admin.DELETE("/:id", h.DeleteNotification)
		admin.POST("/:id/resend", h.ResendNotification)

		// 批量操作
		admin.POST("/batch", h.BatchCreateNotifications)
		admin.POST("/batch-delete", h.BatchDeleteNotifications)

		// 模板管理
		templates := admin.Group("/templates")
		{
			templates.GET("", h.ListTemplates)
			templates.GET("/:id", h.GetTemplate)
			templates.POST("", h.CreateTemplate)
			templates.PUT("/:id", h.UpdateTemplate)
			templates.DELETE("/:id", h.DeleteTemplate)
			templates.POST("/:id/preview", h.PreviewTemplate)
		}

		// Webhook管理
		webhooks := admin.Group("/webhooks")
		{
			webhooks.GET("", h.ListWebhooks)
			webhooks.GET("/:id", h.GetWebhook)
			webhooks.POST("", h.CreateWebhook)
			webhooks.PUT("/:id", h.UpdateWebhook)
			webhooks.DELETE("/:id", h.DeleteWebhook)
			webhooks.POST("/:id/test", h.TestWebhook)
			webhooks.POST("/:id/retry", h.RetryWebhook)
		}

		// Webhook日志
		admin.GET("/webhook-logs", h.ListWebhookLogs)
		admin.GET("/webhook-logs/:id", h.GetWebhookLog)

		// 统计
		admin.GET("/statistics", h.GetStatistics)
	}
}

// ========== 通知列表 ==========

func (h *NotificationBFFHandler) ListNotifications(c *gin.Context) {
	queryParams := make(map[string]string)
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}
	if notificationType := c.Query("type"); notificationType != "" {
		queryParams["type"] = notificationType
	}
	if status := c.Query("status"); status != "" {
		queryParams["status"] = status
	}
	if startTime := c.Query("start_time"); startTime != "" {
		queryParams["start_time"] = startTime
	}
	if endTime := c.Query("end_time"); endTime != "" {
		queryParams["end_time"] = endTime
	}
	if page := c.Query("page"); page != "" {
		queryParams["page"] = page
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		queryParams["page_size"] = pageSize
	}

	result, statusCode, err := h.notificationClient.Get(c.Request.Context(), "/api/v1/notifications", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Notification Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *NotificationBFFHandler) GetNotification(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.notificationClient.Get(c.Request.Context(), "/api/v1/notifications/"+id, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Notification Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *NotificationBFFHandler) CreateNotification(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	adminID := c.GetString("user_id")
	req["created_by"] = adminID

	result, statusCode, err := h.notificationClient.Post(c.Request.Context(), "/api/v1/notifications", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Notification Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *NotificationBFFHandler) DeleteNotification(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.notificationClient.Delete(c.Request.Context(), "/api/v1/notifications/"+id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Notification Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *NotificationBFFHandler) ResendNotification(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		req = make(map[string]interface{})
	}

	result, statusCode, err := h.notificationClient.Post(c.Request.Context(), "/api/v1/notifications/"+id+"/resend", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Notification Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *NotificationBFFHandler) BatchCreateNotifications(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	adminID := c.GetString("user_id")
	req["created_by"] = adminID

	result, statusCode, err := h.notificationClient.Post(c.Request.Context(), "/api/v1/notifications/batch", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Notification Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *NotificationBFFHandler) BatchDeleteNotifications(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	result, statusCode, err := h.notificationClient.Post(c.Request.Context(), "/api/v1/notifications/batch-delete", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Notification Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== 模板管理 ==========

func (h *NotificationBFFHandler) ListTemplates(c *gin.Context) {
	queryParams := make(map[string]string)
	if templateType := c.Query("type"); templateType != "" {
		queryParams["type"] = templateType
	}
	if status := c.Query("status"); status != "" {
		queryParams["status"] = status
	}
	if page := c.Query("page"); page != "" {
		queryParams["page"] = page
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		queryParams["page_size"] = pageSize
	}

	result, statusCode, err := h.notificationClient.Get(c.Request.Context(), "/api/v1/templates", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Notification Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *NotificationBFFHandler) GetTemplate(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.notificationClient.Get(c.Request.Context(), "/api/v1/templates/"+id, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Notification Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *NotificationBFFHandler) CreateTemplate(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	adminID := c.GetString("user_id")
	req["created_by"] = adminID

	result, statusCode, err := h.notificationClient.Post(c.Request.Context(), "/api/v1/templates", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Notification Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *NotificationBFFHandler) UpdateTemplate(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	adminID := c.GetString("user_id")
	req["updated_by"] = adminID

	result, statusCode, err := h.notificationClient.Put(c.Request.Context(), "/api/v1/templates/"+id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Notification Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *NotificationBFFHandler) DeleteTemplate(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.notificationClient.Delete(c.Request.Context(), "/api/v1/templates/"+id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Notification Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *NotificationBFFHandler) PreviewTemplate(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	result, statusCode, err := h.notificationClient.Post(c.Request.Context(), "/api/v1/templates/"+id+"/preview", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Notification Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== Webhook管理 ==========

func (h *NotificationBFFHandler) ListWebhooks(c *gin.Context) {
	queryParams := make(map[string]string)
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}
	if status := c.Query("status"); status != "" {
		queryParams["status"] = status
	}
	if page := c.Query("page"); page != "" {
		queryParams["page"] = page
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		queryParams["page_size"] = pageSize
	}

	result, statusCode, err := h.notificationClient.Get(c.Request.Context(), "/api/v1/webhooks", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Notification Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *NotificationBFFHandler) GetWebhook(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.notificationClient.Get(c.Request.Context(), "/api/v1/webhooks/"+id, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Notification Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *NotificationBFFHandler) CreateWebhook(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	adminID := c.GetString("user_id")
	req["created_by"] = adminID

	result, statusCode, err := h.notificationClient.Post(c.Request.Context(), "/api/v1/webhooks", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Notification Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *NotificationBFFHandler) UpdateWebhook(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	adminID := c.GetString("user_id")
	req["updated_by"] = adminID

	result, statusCode, err := h.notificationClient.Put(c.Request.Context(), "/api/v1/webhooks/"+id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Notification Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *NotificationBFFHandler) DeleteWebhook(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.notificationClient.Delete(c.Request.Context(), "/api/v1/webhooks/"+id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Notification Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *NotificationBFFHandler) TestWebhook(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		req = make(map[string]interface{})
	}

	result, statusCode, err := h.notificationClient.Post(c.Request.Context(), "/api/v1/webhooks/"+id+"/test", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Notification Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *NotificationBFFHandler) RetryWebhook(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		req = make(map[string]interface{})
	}

	result, statusCode, err := h.notificationClient.Post(c.Request.Context(), "/api/v1/webhooks/"+id+"/retry", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Notification Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== Webhook日志 ==========

func (h *NotificationBFFHandler) ListWebhookLogs(c *gin.Context) {
	queryParams := make(map[string]string)
	if webhookID := c.Query("webhook_id"); webhookID != "" {
		queryParams["webhook_id"] = webhookID
	}
	if status := c.Query("status"); status != "" {
		queryParams["status"] = status
	}
	if startTime := c.Query("start_time"); startTime != "" {
		queryParams["start_time"] = startTime
	}
	if endTime := c.Query("end_time"); endTime != "" {
		queryParams["end_time"] = endTime
	}
	if page := c.Query("page"); page != "" {
		queryParams["page"] = page
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		queryParams["page_size"] = pageSize
	}

	result, statusCode, err := h.notificationClient.Get(c.Request.Context(), "/api/v1/webhook-logs", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Notification Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *NotificationBFFHandler) GetWebhookLog(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.notificationClient.Get(c.Request.Context(), "/api/v1/webhook-logs/"+id, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Notification Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== 统计 ==========

func (h *NotificationBFFHandler) GetStatistics(c *gin.Context) {
	queryParams := make(map[string]string)
	if startTime := c.Query("start_time"); startTime != "" {
		queryParams["start_time"] = startTime
	}
	if endTime := c.Query("end_time"); endTime != "" {
		queryParams["end_time"] = endTime
	}

	result, statusCode, err := h.notificationClient.Get(c.Request.Context(), "/api/v1/statistics", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Notification Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}
