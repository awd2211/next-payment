package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/payment-platform/pkg/errors"
	"github.com/payment-platform/pkg/logger"
	"github.com/payment-platform/pkg/middleware"
	"go.uber.org/zap"
	"payment-platform/payment-gateway/internal/repository"
	"payment-platform/payment-gateway/internal/service"
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
		// 支付管理（外部API，需要API Key）
		payments := v1.Group("/payments")
		{
			payments.POST("", h.CreatePayment)
			payments.GET("/:paymentNo", h.GetPayment)
			payments.GET("", h.QueryPayments)
			payments.POST("/batch", h.BatchGetPayments) // 批量查询支付
			payments.POST("/:paymentNo/cancel", h.CancelPayment)
		}

		// 退款管理（外部API，需要API Key）
		refunds := v1.Group("/refunds")
		{
			refunds.POST("", h.CreateRefund)
			refunds.GET("/:refundNo", h.GetRefund)
			refunds.GET("", h.QueryRefunds)
			refunds.POST("/batch", h.BatchGetRefunds) // 批量查询退款
		}

		// 商户后台查询路由（使用JWT认证）
		merchantPayments := v1.Group("/merchant/payments")
		{
			merchantPayments.GET("", h.QueryPayments)
			merchantPayments.GET("/:paymentNo", h.GetPayment)
			merchantPayments.POST("/batch", h.BatchGetPayments) // 批量查询支付
		}

		merchantRefunds := v1.Group("/merchant/refunds")
		{
			merchantRefunds.GET("", h.QueryRefunds)
			merchantRefunds.GET("/:refundNo", h.GetRefund)
			merchantRefunds.POST("/batch", h.BatchGetRefunds) // 批量查询退款
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
//
//	@Summary		创建支付
//	@Description	创建支付订单，支持Stripe、PayPal等多种支付渠道
//	@Tags			Payments
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		service.CreatePaymentInput	true	"支付创建请求"
//	@Success		200		{object}	Response
//	@Failure		400		{object}	Response
//	@Failure		500		{object}	Response
//	@Router			/payments [post]
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var input service.CreatePaymentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).
			WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	payment, err := h.paymentService.CreatePayment(c.Request.Context(), &input)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		// 检查是否为业务错误
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "内部服务错误", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(payment).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// GetPayment 获取支付详情
//
//	@Summary		获取支付详情
//	@Description	根据支付流水号获取支付详情
//	@Tags			Payments
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			paymentNo	path		string	true	"支付流水号"
//	@Success		200			{object}	Response
//	@Failure		404			{object}	Response
//	@Failure		500			{object}	Response
//	@Router			/payments/{paymentNo} [get]
func (h *PaymentHandler) GetPayment(c *gin.Context) {
	paymentNo := c.Param("paymentNo")

	payment, err := h.paymentService.GetPayment(c.Request.Context(), paymentNo)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeResourceNotFound, "支付记录不存在", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusNotFound, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(payment).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// QueryPayments 查询支付列表
//
//	@Summary		查询支付列表
//	@Description	根据条件查询支付列表，支持分页和多维度筛选
//	@Tags			Payments
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			merchant_id		query		string	false	"商户ID"
//	@Param			channel			query		string	false	"支付渠道 (stripe/paypal)"
//	@Param			status			query		string	false	"支付状态 (pending/success/failed/cancelled)"
//	@Param			currency		query		string	false	"货币类型 (USD/EUR/CNY)"
//	@Param			customer_email	query		string	false	"客户邮箱"
//	@Param			keyword			query		string	false	"关键词搜索（订单号/支付号）"
//	@Param			start_time		query		string	false	"开始时间 (RFC3339格式)"
//	@Param			end_time		query		string	false	"结束时间 (RFC3339格式)"
//	@Param			min_amount		query		int		false	"最小金额（分）"
//	@Param			max_amount		query		int		false	"最大金额（分）"
//	@Param			page			query		int		false	"页码"	default(1)
//	@Param			page_size		query		int		false	"每页数量"	default(20)
//	@Success		200				{object}	Response
//	@Failure		400				{object}	Response
//	@Failure		500				{object}	Response
//	@Router			/payments [get]
func (h *PaymentHandler) QueryPayments(c *gin.Context) {
	query := &repository.PaymentQuery{
		Channel:       c.Query("channel"),
		Status:        c.Query("status"),
		Currency:      c.Query("currency"),
		CustomerEmail: c.Query("customer_email"),
		Keyword:       c.Query("keyword"),
	}

	// 从query参数获取merchant_id（外部API调用方式）
	if merchantIDStr := c.Query("merchant_id"); merchantIDStr != "" {
		logger.Info("Parsing merchant_id from query parameter",
			zap.String("merchant_id_str", merchantIDStr))
		merchantID, err := uuid.Parse(merchantIDStr)
		if err != nil {
			traceID := middleware.GetRequestID(c)
			logger.Error("Failed to parse merchant_id",
				zap.String("merchant_id_str", merchantIDStr),
				zap.Error(err))
			resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的商户ID", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		query.MerchantID = &merchantID
	}

	// 从JWT token获取merchant_id（商户后台调用方式）
	if query.MerchantID == nil {
		if userID, exists := c.Get("user_id"); exists {
			logger.Info("Got user_id from context",
				zap.Any("user_id", userID),
				zap.String("user_id_type", fmt.Sprintf("%T", userID)))
			if merchantID, ok := userID.(uuid.UUID); ok {
				query.MerchantID = &merchantID
				logger.Info("Set merchant_id from JWT token",
					zap.String("merchant_id", merchantID.String()))
			} else {
				logger.Warn("Failed to cast user_id to UUID",
					zap.Any("user_id", userID),
					zap.String("user_id_type", fmt.Sprintf("%T", userID)))
			}
		} else {
			logger.Warn("user_id not found in context")
		}
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
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "查询支付列表失败", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(PageResponse{
		List:     payments,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// CancelPayment 取消支付
//
//	@Summary		取消支付
//	@Description	取消待支付的支付订单
//	@Tags			Payments
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			paymentNo	path		string					true	"支付流水号"
//	@Param			request		body		CancelPaymentRequest	true	"取消原因"
//	@Success		200			{object}	Response
//	@Failure		400			{object}	Response
//	@Failure		500			{object}	Response
//	@Router			/payments/{paymentNo}/cancel [post]
func (h *PaymentHandler) CancelPayment(c *gin.Context) {
	paymentNo := c.Param("paymentNo")

	var req CancelPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).
			WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	if err := h.paymentService.CancelPayment(c.Request.Context(), paymentNo, req.Reason); err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "取消支付失败", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(nil).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// CreateRefund 创建退款
//
//	@Summary		创建退款
//	@Description	为已支付的订单创建退款申请
//	@Tags			Refunds
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		service.CreateRefundInput	true	"退款创建请求"
//	@Success		200		{object}	Response
//	@Failure		400		{object}	Response
//	@Failure		500		{object}	Response
//	@Router			/refunds [post]
func (h *PaymentHandler) CreateRefund(c *gin.Context) {
	var input service.CreateRefundInput
	if err := c.ShouldBindJSON(&input); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).
			WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	refund, err := h.paymentService.CreateRefund(c.Request.Context(), &input)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "创建退款失败", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(refund).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// GetRefund 获取退款详情
//
//	@Summary		获取退款详情
//	@Description	根据退款流水号获取退款详情
//	@Tags			Refunds
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			refundNo	path		string	true	"退款流水号"
//	@Success		200			{object}	Response
//	@Failure		404			{object}	Response
//	@Failure		500			{object}	Response
//	@Router			/refunds/{refundNo} [get]
func (h *PaymentHandler) GetRefund(c *gin.Context) {
	refundNo := c.Param("refundNo")

	refund, err := h.paymentService.GetRefund(c.Request.Context(), refundNo)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeResourceNotFound, "退款记录不存在", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusNotFound, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(refund).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// QueryRefunds 查询退款列表
//
//	@Summary		查询退款列表
//	@Description	根据条件查询退款列表，支持分页和多维度筛选
//	@Tags			Refunds
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			merchant_id	query		string	false	"商户ID"
//	@Param			payment_id	query		string	false	"支付ID"
//	@Param			status		query		string	false	"退款状态 (pending/success/failed)"
//	@Param			start_time	query		string	false	"开始时间 (RFC3339格式)"
//	@Param			end_time	query		string	false	"结束时间 (RFC3339格式)"
//	@Param			page		query		int		false	"页码"	default(1)
//	@Param			page_size	query		int		false	"每页数量"	default(20)
//	@Success		200			{object}	Response
//	@Failure		400			{object}	Response
//	@Failure		500			{object}	Response
//	@Router			/refunds [get]
func (h *PaymentHandler) QueryRefunds(c *gin.Context) {
	query := &repository.RefundQuery{
		Status: c.Query("status"),
	}

	if paymentIDStr := c.Query("payment_id"); paymentIDStr != "" {
		paymentID, err := uuid.Parse(paymentIDStr)
		if err != nil {
			traceID := middleware.GetRequestID(c)
			resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的支付ID", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		query.PaymentID = &paymentID
	}

	if merchantIDStr := c.Query("merchant_id"); merchantIDStr != "" {
		merchantID, err := uuid.Parse(merchantIDStr)
		if err != nil {
			traceID := middleware.GetRequestID(c)
			resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的商户ID", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusBadRequest, resp)
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
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "查询退款列表失败", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(PageResponse{
		List:     refunds,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// HandleStripeWebhook 处理Stripe回调
//
//	@Summary		处理Stripe Webhook
//	@Description	接收并处理Stripe支付网关的异步通知
//	@Tags			Webhooks
//	@Accept			json
//	@Produce		json
//	@Param			request	body		map[string]interface{}	true	"Stripe webhook payload"
//	@Success		200		{object}	Response
//	@Failure		400		{object}	Response
//	@Failure		500		{object}	Response
//	@Router			/webhooks/stripe [post]
func (h *PaymentHandler) HandleStripeWebhook(c *gin.Context) {
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).
			WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	if err := h.paymentService.HandleCallback(c.Request.Context(), "stripe", data); err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "处理Stripe回调失败", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(nil).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// HandlePayPalWebhook 处理PayPal回调
//
//	@Summary		处理PayPal Webhook
//	@Description	接收并处理PayPal支付网关的异步通知
//	@Tags			Webhooks
//	@Accept			json
//	@Produce		json
//	@Param			request	body		map[string]interface{}	true	"PayPal webhook payload"
//	@Success		200		{object}	Response
//	@Failure		400		{object}	Response
//	@Failure		500		{object}	Response
//	@Router			/webhooks/paypal [post]
func (h *PaymentHandler) HandlePayPalWebhook(c *gin.Context) {
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).
			WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	if err := h.paymentService.HandleCallback(c.Request.Context(), "paypal", data); err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "处理PayPal回调失败", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(nil).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
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

// BatchGetPaymentsRequest 批量查询支付请求
type BatchGetPaymentsRequest struct {
	PaymentNos []string  `json:"payment_nos" binding:"required,min=1,max=100"`
	MerchantID uuid.UUID `json:"merchant_id" binding:"required"`
}

// BatchGetPaymentsResponse 批量查询支付响应
type BatchGetPaymentsResponse struct {
	Results map[string]interface{} `json:"results"` // paymentNo -> Payment
	Failed  []string               `json:"failed"`  // 查询失败的 paymentNo
	Summary struct {
		Total    int `json:"total"`     // 请求的总数
		Found    int `json:"found"`     // 找到的数量
		NotFound int `json:"not_found"` // 未找到的数量
	} `json:"summary"`
}

// BatchGetPayments 批量查询支付
//
//	@Summary		批量查询支付
//	@Description	一次性查询多个支付记录（最多100个）
//	@Tags			Payments
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		BatchGetPaymentsRequest	true	"批量查询请求"
//	@Success		200		{object}	BatchGetPaymentsResponse
//	@Failure		400		{object}	Response
//	@Failure		500		{object}	Response
//	@Router			/payments/batch [post]
func (h *PaymentHandler) BatchGetPayments(c *gin.Context) {
	var req BatchGetPaymentsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).
			WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// 调用服务层批量查询
	results, failed, err := h.paymentService.BatchGetPayments(c.Request.Context(), req.PaymentNos, req.MerchantID)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "批量查询支付失败", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	// 构建响应
	response := BatchGetPaymentsResponse{
		Results: make(map[string]interface{}),
		Failed:  failed,
	}
	for paymentNo, payment := range results {
		response.Results[paymentNo] = payment
	}
	response.Summary.Total = len(req.PaymentNos)
	response.Summary.Found = len(results)
	response.Summary.NotFound = len(failed)

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(response).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// BatchGetRefundsRequest 批量查询退款请求
type BatchGetRefundsRequest struct {
	RefundNos  []string  `json:"refund_nos" binding:"required,min=1,max=100"`
	MerchantID uuid.UUID `json:"merchant_id" binding:"required"`
}

// BatchGetRefundsResponse 批量查询退款响应
type BatchGetRefundsResponse struct {
	Results map[string]interface{} `json:"results"` // refundNo -> Refund
	Failed  []string               `json:"failed"`  // 查询失败的 refundNo
	Summary struct {
		Total    int `json:"total"`     // 请求的总数
		Found    int `json:"found"`     // 找到的数量
		NotFound int `json:"not_found"` // 未找到的数量
	} `json:"summary"`
}

// BatchGetRefunds 批量查询退款
//
//	@Summary		批量查询退款
//	@Description	一次性查询多个退款记录（最多100个）
//	@Tags			Refunds
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		BatchGetRefundsRequest	true	"批量查询请求"
//	@Success		200		{object}	BatchGetRefundsResponse
//	@Failure		400		{object}	Response
//	@Failure		500		{object}	Response
//	@Router			/refunds/batch [post]
func (h *PaymentHandler) BatchGetRefunds(c *gin.Context) {
	var req BatchGetRefundsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).
			WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// 调用服务层批量查询
	results, failed, err := h.paymentService.BatchGetRefunds(c.Request.Context(), req.RefundNos, req.MerchantID)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "批量查询退款失败", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	// 构建响应
	response := BatchGetRefundsResponse{
		Results: make(map[string]interface{}),
		Failed:  failed,
	}
	for refundNo, refund := range results {
		response.Results[refundNo] = refund
	}
	response.Summary.Total = len(req.RefundNos)
	response.Summary.Found = len(results)
	response.Summary.NotFound = len(failed)

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(response).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}
