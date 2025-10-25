package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/payment-platform/pkg/middleware"

	"payment-platform/merchant-service/internal/client"
	"payment-platform/merchant-service/internal/errors"
)

// PaymentHandler 支付处理器（代理payment-gateway）
type PaymentHandler struct {
	paymentClient *client.PaymentClient
}

// NewPaymentHandler 创建支付处理器
func NewPaymentHandler(paymentClient *client.PaymentClient) *PaymentHandler {
	return &PaymentHandler{
		paymentClient: paymentClient,
	}
}

// GetPayments 获取支付列表
// @Summary 获取当前商户的支付列表
// @Tags Payment
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param status query string false "支付状态"
// @Param channel query string false "支付渠道"
// @Param start_time query string false "开始时间"
// @Param end_time query string false "结束时间"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/merchant/payments [get]
func (h *PaymentHandler) GetPayments(c *gin.Context) {
	// 从JWT中获取商户ID
	merchantID, exists := c.Get("user_id")
	if !exists {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeUnauthorized, "未授权", "").WithTraceID(traceID)
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	// 构建查询参数
	params := map[string]string{
		"page":       c.DefaultQuery("page", "1"),
		"page_size":  c.DefaultQuery("page_size", "10"),
		"status":     c.Query("status"),
		"channel":    c.Query("channel"),
		"start_time": c.Query("start_time"),
		"end_time":   c.Query("end_time"),
	}

	// 调用payment-gateway获取支付列表
	data, err := h.paymentClient.GetPayments(c.Request.Context(), merchantID.(uuid.UUID), params)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "获取支付列表失败", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(data).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// GetRefunds 获取退款列表
// @Summary 获取当前商户的退款列表
// @Tags Payment
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/merchant/refunds [get]
func (h *PaymentHandler) GetRefunds(c *gin.Context) {
	// 从JWT中获取商户ID
	merchantID, exists := c.Get("user_id")
	if !exists {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeUnauthorized, "未授权", "").WithTraceID(traceID)
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	// 构建查询参数
	params := map[string]string{
		"page":      c.DefaultQuery("page", "1"),
		"page_size": c.DefaultQuery("page_size", "10"),
	}

	// 调用payment-gateway获取退款列表
	data, err := h.paymentClient.GetRefunds(c.Request.Context(), merchantID.(uuid.UUID), params)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "获取退款列表失败", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusInternalServerError, resp)
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(data).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// RegisterRoutes 注册路由
func (h *PaymentHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	payment := r.Group("/merchant")
	payment.Use(authMiddleware)
	{
		payment.GET("/payments", h.GetPayments)
		payment.GET("/refunds", h.GetRefunds)
	}
}
