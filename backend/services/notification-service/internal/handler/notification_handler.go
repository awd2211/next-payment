package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/payment-platform/pkg/errors"
	"github.com/payment-platform/pkg/middleware"
	"payment-platform/notification-service/internal/model"
	"payment-platform/notification-service/internal/repository"
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
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的请求参数", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// 从上下文获取商户ID
	merchantID, exists := c.Get("merchant_id")
	if !exists {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeUnauthorized, "未认证", "").WithTraceID(traceID)
		c.JSON(http.StatusUnauthorized, resp)
		return
	}
	req.MerchantID = merchantID.(uuid.UUID)

	// 发送邮件
	if err := h.notificationService.SendEmail(c.Request.Context(), &req); err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "邮件发送失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{"message": "邮件发送成功"}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
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
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的请求参数", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// 从上下文获取商户ID
	merchantID, exists := c.Get("merchant_id")
	if !exists {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeUnauthorized, "未认证", "").WithTraceID(traceID)
		c.JSON(http.StatusUnauthorized, resp)
		return
	}
	req.MerchantID = merchantID.(uuid.UUID)

	// 发送短信
	if err := h.notificationService.SendSMS(c.Request.Context(), &req); err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "短信发送失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{"message": "短信发送成功"}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
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
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的请求参数", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// 从上下文获取商户ID
	merchantID, exists := c.Get("merchant_id")
	if !exists {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeUnauthorized, "未认证", "").WithTraceID(traceID)
		c.JSON(http.StatusUnauthorized, resp)
		return
	}
	req.MerchantID = merchantID.(uuid.UUID)

	// 发送 Webhook
	if err := h.notificationService.SendWebhook(c.Request.Context(), &req); err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "Webhook 发送失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{"message": "Webhook 发送成功"}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
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
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的请求参数", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// 从上下文获取商户ID
	merchantID, exists := c.Get("merchant_id")
	if !exists {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeUnauthorized, "未认证", "").WithTraceID(traceID)
		c.JSON(http.StatusUnauthorized, resp)
		return
	}
	req.MerchantID = merchantID.(uuid.UUID)

	// 发送邮件
	if err := h.notificationService.SendEmailByTemplate(c.Request.Context(), &req); err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "模板邮件发送失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{"message": "邮件发送成功"}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// GetNotification 获取通知详情
// @Summary 获取通知详情
// @Tags Notification
// @Accept json
// @Produce json
// @Param id path string true "通知ID"
// @Success 200 {object} model.Notification
// @Router /api/v1/notifications/{id} [get]
func (h *NotificationHandler) GetNotification(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的通知ID", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	notification, err := h.notificationService.GetNotification(c.Request.Context(), id)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "获取通知失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	if notification == nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "通知不存在", "").WithTraceID(traceID)
		c.JSON(http.StatusNotFound, resp)
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(notification).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// ListNotifications 列出通知
// @Summary 列出通知
// @Tags Notification
// @Accept json
// @Produce json
// @Param merchant_id query string false "商户ID"
// @Param type query string false "通知类型"
// @Param channel query string false "通知渠道"
// @Param status query string false "状态"
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/notifications [get]
func (h *NotificationHandler) ListNotifications(c *gin.Context) {
	var query repository.NotificationQuery

	// 解析查询参数
	if merchantIDStr := c.Query("merchant_id"); merchantIDStr != "" {
		merchantID, err := uuid.Parse(merchantIDStr)
		if err != nil {
			traceID := middleware.GetRequestID(c)
			resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的商户ID", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		query.MerchantID = &merchantID
	}

	query.Type = c.Query("type")
	query.Channel = c.Query("channel")
	query.Status = c.Query("status")

	// 解析分页参数
	query.Page = 1
	query.PageSize = 20
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			query.Page = page
		}
	}
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 && pageSize <= 100 {
			query.PageSize = pageSize
		}
	}

	// 查询通知列表
	notifications, total, err := h.notificationService.ListNotifications(c.Request.Context(), &query)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "查询通知列表失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{
		"data":      notifications,
		"total":     total,
		"page":      query.Page,
		"page_size": query.PageSize,
	}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// CreateTemplate 创建通知模板
// @Summary 创建通知模板
// @Tags Template
// @Accept json
// @Produce json
// @Param request body model.NotificationTemplate true "模板信息"
// @Success 200 {object} model.NotificationTemplate
// @Router /api/v1/templates [post]
func (h *NotificationHandler) CreateTemplate(c *gin.Context) {
	var template model.NotificationTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的请求参数", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// 从上下文获取商户ID（如果有）
	if merchantID, exists := c.Get("merchant_id"); exists {
		template.MerchantID = merchantID.(uuid.UUID)
		template.IsSystem = false
	}

	if err := h.notificationService.CreateTemplate(c.Request.Context(), &template); err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "创建模板失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(template).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// GetTemplate 获取通知模板
// @Summary 获取通知模板
// @Tags Template
// @Accept json
// @Produce json
// @Param code query string true "模板编码"
// @Success 200 {object} model.NotificationTemplate
// @Router /api/v1/templates/{code} [get]
func (h *NotificationHandler) GetTemplate(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "模板编码不能为空", "").WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var merchantID *uuid.UUID
	if mid, exists := c.Get("merchant_id"); exists {
		id := mid.(uuid.UUID)
		merchantID = &id
	}

	template, err := h.notificationService.GetTemplate(c.Request.Context(), code, merchantID)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "获取模板失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	if template == nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "模板不存在", "").WithTraceID(traceID)
		c.JSON(http.StatusNotFound, resp)
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(template).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// ListTemplates 列出通知模板
// @Summary 列出通知模板
// @Tags Template
// @Accept json
// @Produce json
// @Success 200 {array} model.NotificationTemplate
// @Router /api/v1/templates [get]
func (h *NotificationHandler) ListTemplates(c *gin.Context) {
	var merchantID *uuid.UUID
	if mid, exists := c.Get("merchant_id"); exists {
		id := mid.(uuid.UUID)
		merchantID = &id
	}

	templates, err := h.notificationService.ListTemplates(c.Request.Context(), merchantID)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "查询模板列表失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(templates).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// UpdateTemplate 更新通知模板
// @Summary 更新通知模板
// @Tags Template
// @Accept json
// @Produce json
// @Param id path string true "模板ID"
// @Param request body model.NotificationTemplate true "模板信息"
// @Success 200 {object} model.NotificationTemplate
// @Router /api/v1/templates/{id} [put]
func (h *NotificationHandler) UpdateTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的模板ID", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var template model.NotificationTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的请求参数", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	template.ID = id

	if err := h.notificationService.UpdateTemplate(c.Request.Context(), &template); err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "更新模板失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(template).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// DeleteTemplate 删除通知模板
// @Summary 删除通知模板
// @Tags Template
// @Accept json
// @Produce json
// @Param id path string true "模板ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/templates/{id} [delete]
func (h *NotificationHandler) DeleteTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的模板ID", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	if err := h.notificationService.DeleteTemplate(c.Request.Context(), id); err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "删除模板失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{"message": "模板删除成功"}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// CreateWebhookEndpoint 创建 Webhook 端点
// @Summary 创建 Webhook 端点
// @Tags Webhook
// @Accept json
// @Produce json
// @Param request body model.WebhookEndpoint true "端点信息"
// @Success 200 {object} model.WebhookEndpoint
// @Router /api/v1/webhooks/endpoints [post]
func (h *NotificationHandler) CreateWebhookEndpoint(c *gin.Context) {
	var endpoint model.WebhookEndpoint
	if err := c.ShouldBindJSON(&endpoint); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的请求参数", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// 从上下文获取商户ID
	merchantID, exists := c.Get("merchant_id")
	if !exists {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeUnauthorized, "未认证", "").WithTraceID(traceID)
		c.JSON(http.StatusUnauthorized, resp)
		return
	}
	endpoint.MerchantID = merchantID.(uuid.UUID)

	if err := h.notificationService.CreateWebhookEndpoint(c.Request.Context(), &endpoint); err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "创建 Webhook 端点失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(endpoint).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// ListWebhookEndpoints 列出 Webhook 端点
// @Summary 列出 Webhook 端点
// @Tags Webhook
// @Accept json
// @Produce json
// @Success 200 {array} model.WebhookEndpoint
// @Router /api/v1/webhooks/endpoints [get]
func (h *NotificationHandler) ListWebhookEndpoints(c *gin.Context) {
	// 从上下文获取商户ID
	merchantID, exists := c.Get("merchant_id")
	if !exists {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeUnauthorized, "未认证", "").WithTraceID(traceID)
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	// 列出端点
	endpoints, err := h.notificationService.ListWebhookEndpoints(c.Request.Context(), merchantID.(uuid.UUID))
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "查询 Webhook 端点列表失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(endpoints).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// UpdateWebhookEndpoint 更新 Webhook 端点
// @Summary 更新 Webhook 端点
// @Tags Webhook
// @Accept json
// @Produce json
// @Param id path string true "端点ID"
// @Param request body model.WebhookEndpoint true "端点信息"
// @Success 200 {object} model.WebhookEndpoint
// @Router /api/v1/webhooks/endpoints/{id} [put]
func (h *NotificationHandler) UpdateWebhookEndpoint(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的端点ID", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var endpoint model.WebhookEndpoint
	if err := c.ShouldBindJSON(&endpoint); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的请求参数", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	endpoint.ID = id

	// 从上下文获取商户ID
	merchantID, exists := c.Get("merchant_id")
	if !exists {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeUnauthorized, "未认证", "").WithTraceID(traceID)
		c.JSON(http.StatusUnauthorized, resp)
		return
	}
	endpoint.MerchantID = merchantID.(uuid.UUID)

	if err := h.notificationService.UpdateWebhookEndpoint(c.Request.Context(), &endpoint); err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "更新 Webhook 端点失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(endpoint).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// DeleteWebhookEndpoint 删除 Webhook 端点
// @Summary 删除 Webhook 端点
// @Tags Webhook
// @Accept json
// @Produce json
// @Param id path string true "端点ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/webhooks/endpoints/{id} [delete]
func (h *NotificationHandler) DeleteWebhookEndpoint(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的端点ID", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	if err := h.notificationService.DeleteWebhookEndpoint(c.Request.Context(), id); err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "删除 Webhook 端点失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{"message": "端点删除成功"}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// ListWebhookDeliveries 列出 Webhook 投递记录
// @Summary 列出 Webhook 投递记录
// @Tags Webhook
// @Accept json
// @Produce json
// @Param endpoint_id query string false "端点ID"
// @Param merchant_id query string false "商户ID"
// @Param event_type query string false "事件类型"
// @Param status query string false "状态"
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/webhooks/deliveries [get]
func (h *NotificationHandler) ListWebhookDeliveries(c *gin.Context) {
	var query repository.DeliveryQuery

	// 解析查询参数
	if endpointIDStr := c.Query("endpoint_id"); endpointIDStr != "" {
		endpointID, err := uuid.Parse(endpointIDStr)
		if err != nil {
			traceID := middleware.GetRequestID(c)
			resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的端点ID", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		query.EndpointID = &endpointID
	}

	if merchantIDStr := c.Query("merchant_id"); merchantIDStr != "" {
		merchantID, err := uuid.Parse(merchantIDStr)
		if err != nil {
			traceID := middleware.GetRequestID(c)
			resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的商户ID", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		query.MerchantID = &merchantID
	} else {
		// 如果没有指定商户ID，从上下文获取
		if mid, exists := c.Get("merchant_id"); exists {
			merchantID := mid.(uuid.UUID)
			query.MerchantID = &merchantID
		}
	}

	query.EventType = c.Query("event_type")
	query.Status = c.Query("status")

	// 解析分页参数
	query.Page = 1
	query.PageSize = 20
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			query.Page = page
		}
	}
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 && pageSize <= 100 {
			query.PageSize = pageSize
		}
	}

	// 查询投递记录列表
	deliveries, total, err := h.notificationService.ListWebhookDeliveries(c.Request.Context(), &query)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "查询 Webhook 投递记录失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{
		"data":      deliveries,
		"total":     total,
		"page":      query.Page,
		"page_size": query.PageSize,
	}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// CreatePreference 创建通知偏好
// @Summary 创建通知偏好
// @Tags Preference
// @Accept json
// @Produce json
// @Param request body model.NotificationPreference true "偏好设置"
// @Success 200 {object} model.NotificationPreference
// @Router /api/v1/preferences [post]
func (h *NotificationHandler) CreatePreference(c *gin.Context) {
	var preference model.NotificationPreference
	if err := c.ShouldBindJSON(&preference); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的请求参数", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// 从上下文获取商户ID
	merchantID, exists := c.Get("merchant_id")
	if !exists {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeUnauthorized, "未认证", "").WithTraceID(traceID)
		c.JSON(http.StatusUnauthorized, resp)
		return
	}
	preference.MerchantID = merchantID.(uuid.UUID)

	if err := h.notificationService.CreatePreference(c.Request.Context(), &preference); err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "创建通知偏好失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(preference).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// GetPreference 获取通知偏好详情
// @Summary 获取通知偏好详情
// @Tags Preference
// @Accept json
// @Produce json
// @Param id path string true "偏好ID"
// @Success 200 {object} model.NotificationPreference
// @Router /api/v1/preferences/{id} [get]
func (h *NotificationHandler) GetPreference(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的偏好ID", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	preference, err := h.notificationService.GetPreference(c.Request.Context(), id)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "获取通知偏好失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	if preference == nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "偏好设置不存在", "").WithTraceID(traceID)
		c.JSON(http.StatusNotFound, resp)
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(preference).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// ListPreferences 列出通知偏好
// @Summary 列出通知偏好
// @Tags Preference
// @Accept json
// @Produce json
// @Param user_id query string false "用户ID"
// @Success 200 {array} model.NotificationPreference
// @Router /api/v1/preferences [get]
func (h *NotificationHandler) ListPreferences(c *gin.Context) {
	// 从上下文获取商户ID
	merchantID, exists := c.Get("merchant_id")
	if !exists {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeUnauthorized, "未认证", "").WithTraceID(traceID)
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	var userID *uuid.UUID
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		uid, err := uuid.Parse(userIDStr)
		if err != nil {
			traceID := middleware.GetRequestID(c)
			resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的用户ID", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		userID = &uid
	}

	preferences, err := h.notificationService.ListPreferences(c.Request.Context(), merchantID.(uuid.UUID), userID)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "查询通知偏好列表失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(preferences).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// UpdatePreference 更新通知偏好
// @Summary 更新通知偏好
// @Tags Preference
// @Accept json
// @Produce json
// @Param id path string true "偏好ID"
// @Param request body model.NotificationPreference true "偏好设置"
// @Success 200 {object} model.NotificationPreference
// @Router /api/v1/preferences/{id} [put]
func (h *NotificationHandler) UpdatePreference(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的偏好ID", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var preference model.NotificationPreference
	if err := c.ShouldBindJSON(&preference); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的请求参数", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	preference.ID = id

	if err := h.notificationService.UpdatePreference(c.Request.Context(), &preference); err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "更新通知偏好失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(preference).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// DeletePreference 删除通知偏好
// @Summary 删除通知偏好
// @Tags Preference
// @Accept json
// @Produce json
// @Param id path string true "偏好ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/preferences/{id} [delete]
func (h *NotificationHandler) DeletePreference(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的偏好ID", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	if err := h.notificationService.DeletePreference(c.Request.Context(), id); err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "删除通知偏好失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{"message": "偏好设置删除成功"}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// RegisterRoutes 注册路由
func (h *NotificationHandler) RegisterRoutes(router *gin.Engine, authMiddleware ...gin.HandlerFunc) {
	api := router.Group("/api/v1")

	// 需要认证的路由
	if len(authMiddleware) > 0 {
		api.Use(authMiddleware...)
	}

	{
		// 通知发送
		api.POST("/notifications/email", h.SendEmail)
		api.POST("/notifications/sms", h.SendSMS)
		api.POST("/notifications/webhook", h.SendWebhook)
		api.POST("/notifications/email/template", h.SendEmailByTemplate)

		// 通知查询
		api.GET("/notifications", h.ListNotifications)
		api.GET("/notifications/:id", h.GetNotification)

		// 模板管理
		api.POST("/templates", h.CreateTemplate)
		api.GET("/templates/:code", h.GetTemplate)
		api.GET("/templates", h.ListTemplates)
		api.PUT("/templates/:id", h.UpdateTemplate)
		api.DELETE("/templates/:id", h.DeleteTemplate)

		// Webhook 端点管理
		api.POST("/webhooks/endpoints", h.CreateWebhookEndpoint)
		api.GET("/webhooks/endpoints", h.ListWebhookEndpoints)
		api.PUT("/webhooks/endpoints/:id", h.UpdateWebhookEndpoint)
		api.DELETE("/webhooks/endpoints/:id", h.DeleteWebhookEndpoint)

		// Webhook 投递记录
		api.GET("/webhooks/deliveries", h.ListWebhookDeliveries)

		// 通知偏好设置
		api.POST("/preferences", h.CreatePreference)
		api.GET("/preferences/:id", h.GetPreference)
		api.GET("/preferences", h.ListPreferences)
		api.PUT("/preferences/:id", h.UpdatePreference)
		api.DELETE("/preferences/:id", h.DeletePreference)
	}
}
