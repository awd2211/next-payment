package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/payment-platform/services/payment-gateway/internal/repository"
	"github.com/payment-platform/services/payment-gateway/internal/service"
)

// PaymentHandler 支付处理器
type PaymentHandler struct {
	paymentService service.PaymentService
}

// NewPaymentHandler 创建支付处理器实例
func NewPaymentHandler(paymentService service.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

// RegisterRoutes 注册路由
func (h *PaymentHandler) RegisterRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		// 支付管理
		payments := v1.Group("/payments")
		{
			payments.POST("", h.CreatePayment)
			payments.GET("/:paymentNo", h.GetPayment)
			payments.GET("", h.QueryPayments)
			payments.POST("/:paymentNo/cancel", h.CancelPayment)
		}

		// 退款管理
		refunds := v1.Group("/refunds")
		{
			refunds.POST("", h.CreateRefund)
			refunds.GET("/:refundNo", h.GetRefund)
			refunds.GET("", h.QueryRefunds)
		}

		// 回调处理
		webhooks := v1.Group("/webhooks")
		{
			webhooks.POST("/stripe", h.HandleStripeWebhook)
			webhooks.POST("/paypal", h.HandlePayPalWebhook)
		}
	}
}

// CreatePayment 创建支付
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var input service.CreatePaymentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	payment, err := h.paymentService.CreatePayment(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(payment))
}

// GetPayment 获取支付详情
func (h *PaymentHandler) GetPayment(c *gin.Context) {
	paymentNo := c.Param("paymentNo")

	payment, err := h.paymentService.GetPayment(c.Request.Context(), paymentNo)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(payment))
}

// QueryPayments 查询支付列表
func (h *PaymentHandler) QueryPayments(c *gin.Context) {
	query := &repository.PaymentQuery{
		Channel:       c.Query("channel"),
		Status:        c.Query("status"),
		Currency:      c.Query("currency"),
		CustomerEmail: c.Query("customer_email"),
		Keyword:       c.Query("keyword"),
	}

	if merchantIDStr := c.Query("merchant_id"); merchantIDStr != "" {
		merchantID, err := uuid.Parse(merchantIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse("无效的商户ID"))
			return
		}
		query.MerchantID = &merchantID
	}

	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err == nil {
			query.StartTime = &startTime
		}
	}
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		endTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err == nil {
			query.EndTime = &endTime
		}
	}

	if minAmountStr := c.Query("min_amount"); minAmountStr != "" {
		minAmount, err := strconv.ParseInt(minAmountStr, 10, 64)
		if err == nil {
			query.MinAmount = &minAmount
		}
	}
	if maxAmountStr := c.Query("max_amount"); maxAmountStr != "" {
		maxAmount, err := strconv.ParseInt(maxAmountStr, 10, 64)
		if err == nil {
			query.MaxAmount = &maxAmount
		}
	}

	query.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	query.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))

	payments, total, err := h.paymentService.QueryPayment(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(PageResponse{
		List:     payments,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}))
}

// CancelPayment 取消支付
func (h *PaymentHandler) CancelPayment(c *gin.Context) {
	paymentNo := c.Param("paymentNo")

	var req CancelPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	if err := h.paymentService.CancelPayment(c.Request.Context(), paymentNo, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(nil))
}

// CreateRefund 创建退款
func (h *PaymentHandler) CreateRefund(c *gin.Context) {
	var input service.CreateRefundInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	refund, err := h.paymentService.CreateRefund(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(refund))
}

// GetRefund 获取退款详情
func (h *PaymentHandler) GetRefund(c *gin.Context) {
	refundNo := c.Param("refundNo")

	refund, err := h.paymentService.GetRefund(c.Request.Context(), refundNo)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(refund))
}

// QueryRefunds 查询退款列表
func (h *PaymentHandler) QueryRefunds(c *gin.Context) {
	query := &repository.RefundQuery{
		Status: c.Query("status"),
	}

	if paymentIDStr := c.Query("payment_id"); paymentIDStr != "" {
		paymentID, err := uuid.Parse(paymentIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse("无效的支付ID"))
			return
		}
		query.PaymentID = &paymentID
	}

	if merchantIDStr := c.Query("merchant_id"); merchantIDStr != "" {
		merchantID, err := uuid.Parse(merchantIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse("无效的商户ID"))
			return
		}
		query.MerchantID = &merchantID
	}

	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err == nil {
			query.StartTime = &startTime
		}
	}
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		endTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err == nil {
			query.EndTime = &endTime
		}
	}

	query.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	query.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))

	refunds, total, err := h.paymentService.QueryRefunds(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(PageResponse{
		List:     refunds,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}))
}

// HandleStripeWebhook 处理Stripe回调
func (h *PaymentHandler) HandleStripeWebhook(c *gin.Context) {
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	if err := h.paymentService.HandleCallback(c.Request.Context(), "stripe", data); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(nil))
}

// HandlePayPalWebhook 处理PayPal回调
func (h *PaymentHandler) HandlePayPalWebhook(c *gin.Context) {
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	if err := h.paymentService.HandleCallback(c.Request.Context(), "paypal", data); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(nil))
}

// Request/Response structures

type CancelPaymentRequest struct {
	Reason string `json:"reason" binding:"required"`
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type PageResponse struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

func SuccessResponse(data interface{}) Response {
	return Response{
		Code:    0,
		Message: "success",
		Data:    data,
	}
}

func ErrorResponse(message string) Response {
	return Response{
		Code:    -1,
		Message: message,
	}
}
