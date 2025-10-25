package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/payment-platform/pkg/errors"
	"github.com/payment-platform/pkg/middleware"
	"payment-platform/channel-adapter/internal/service"
)

// ChannelHandler 渠道处理器
type ChannelHandler struct {
	channelService service.ChannelService
}

// NewChannelHandler 创建渠道处理器实例
func NewChannelHandler(channelService service.ChannelService) *ChannelHandler {
	return &ChannelHandler{
		channelService: channelService,
	}
}

// CreatePayment 创建支付
// @Summary 创建支付
// @Tags Channel
// @Accept json
// @Produce json
// @Param request body service.CreatePaymentRequest true "创建支付请求"
// @Success 200 {object} service.CreatePaymentResponse
// @Router /api/v1/channel/payments [post]
func (h *ChannelHandler) CreatePayment(c *gin.Context) {
	var req service.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		traceID := middleware.GetRequestID(c)
		response := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的请求参数", err.Error()).
			WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// 从上下文获取商户ID（通常由认证中间件设置）
	merchantID, exists := c.Get("merchant_id")
	if !exists {
		traceID := middleware.GetRequestID(c)
		response := errors.NewErrorResponse(errors.ErrCodeUnauthorized, "未认证", "").
			WithTraceID(traceID)
		c.JSON(http.StatusUnauthorized, response)
		return
	}
	req.MerchantID = merchantID.(uuid.UUID)

	// 创建支付
	resp, err := h.channelService.CreatePayment(c.Request.Context(), &req)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			response := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), response)
		} else {
			response := errors.NewErrorResponse(errors.ErrCodeInternalError, "创建支付失败", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, response)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	response := errors.NewSuccessResponse(resp).WithTraceID(traceID)
	c.JSON(http.StatusOK, response)
}

// QueryPayment 查询支付
// @Summary 查询支付状态
// @Tags Channel
// @Accept json
// @Produce json
// @Param payment_no path string true "支付流水号"
// @Success 200 {object} service.QueryPaymentResponse
// @Router /api/v1/channel/payments/{payment_no} [get]
func (h *ChannelHandler) QueryPayment(c *gin.Context) {
	paymentNo := c.Param("payment_no")
	if paymentNo == "" {
		traceID := middleware.GetRequestID(c)
		response := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "支付流水号不能为空", "").
			WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// 查询支付
	resp, err := h.channelService.QueryPayment(c.Request.Context(), paymentNo)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			response := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), response)
		} else {
			response := errors.NewErrorResponse(errors.ErrCodeInternalError, "查询支付失败", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, response)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	response := errors.NewSuccessResponse(resp).WithTraceID(traceID)
	c.JSON(http.StatusOK, response)
}

// QueryPaymentCompat 查询支付（兼容 Payment-Gateway 的请求格式）
// @Summary 查询支付状态（兼容接口）
// @Tags Channel
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "查询请求"
// @Success 200 {object} service.QueryPaymentResponse
// @Router /api/v1/channel/query [post]
func (h *ChannelHandler) QueryPaymentCompat(c *gin.Context) {
	var req struct {
		PaymentNo      string `json:"payment_no"`
		ChannelOrderNo string `json:"channel_order_no"`
		Channel        string `json:"channel"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		traceID := middleware.GetRequestID(c)
		response := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的请求参数", err.Error()).
			WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// 优先使用 payment_no
	paymentNo := req.PaymentNo
	if paymentNo == "" {
		traceID := middleware.GetRequestID(c)
		response := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "支付流水号不能为空", "").
			WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// 查询支付
	resp, err := h.channelService.QueryPayment(c.Request.Context(), paymentNo)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			response := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), response)
		} else {
			response := errors.NewErrorResponse(errors.ErrCodeInternalError, "查询支付失败", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, response)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	response := errors.NewSuccessResponse(resp).WithTraceID(traceID)
	c.JSON(http.StatusOK, response)
}

// CancelPayment 取消支付
// @Summary 取消支付
// @Tags Channel
// @Accept json
// @Produce json
// @Param payment_no path string true "支付流水号"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/channel/payments/{payment_no}/cancel [post]
func (h *ChannelHandler) CancelPayment(c *gin.Context) {
	paymentNo := c.Param("payment_no")
	if paymentNo == "" {
		traceID := middleware.GetRequestID(c)
		response := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "支付流水号不能为空", "").
			WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// 取消支付
	if err := h.channelService.CancelPayment(c.Request.Context(), paymentNo); err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			response := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), response)
		} else {
			response := errors.NewErrorResponse(errors.ErrCodeInternalError, "取消支付失败", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, response)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	response := errors.NewSuccessResponse(gin.H{"message": "取消成功"}).WithTraceID(traceID)
	c.JSON(http.StatusOK, response)
}

// CreateRefund 创建退款
// @Summary 创建退款
// @Tags Channel
// @Accept json
// @Produce json
// @Param request body service.CreateRefundRequest true "创建退款请求"
// @Success 200 {object} service.CreateRefundResponse
// @Router /api/v1/channel/refunds [post]
func (h *ChannelHandler) CreateRefund(c *gin.Context) {
	var req service.CreateRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		traceID := middleware.GetRequestID(c)
		response := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的请求参数", err.Error()).
			WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// 从上下文获取商户ID
	merchantID, exists := c.Get("merchant_id")
	if !exists {
		traceID := middleware.GetRequestID(c)
		response := errors.NewErrorResponse(errors.ErrCodeUnauthorized, "未认证", "").
			WithTraceID(traceID)
		c.JSON(http.StatusUnauthorized, response)
		return
	}
	req.MerchantID = merchantID.(uuid.UUID)

	// 创建退款
	resp, err := h.channelService.CreateRefund(c.Request.Context(), &req)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			response := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), response)
		} else {
			response := errors.NewErrorResponse(errors.ErrCodeInternalError, "创建退款失败", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, response)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	response := errors.NewSuccessResponse(resp).WithTraceID(traceID)
	c.JSON(http.StatusOK, response)
}

// QueryRefund 查询退款
// @Summary 查询退款状态
// @Tags Channel
// @Accept json
// @Produce json
// @Param refund_no path string true "退款流水号"
// @Success 200 {object} service.QueryRefundResponse
// @Router /api/v1/channel/refunds/{refund_no} [get]
func (h *ChannelHandler) QueryRefund(c *gin.Context) {
	refundNo := c.Param("refund_no")
	if refundNo == "" {
		traceID := middleware.GetRequestID(c)
		response := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "退款流水号不能为空", "").
			WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// 查询退款
	resp, err := h.channelService.QueryRefund(c.Request.Context(), refundNo)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			response := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), response)
		} else {
			response := errors.NewErrorResponse(errors.ErrCodeInternalError, "查询退款失败", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, response)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	response := errors.NewSuccessResponse(resp).WithTraceID(traceID)
	c.JSON(http.StatusOK, response)
}

// HandleStripeWebhook 处理 Stripe Webhook
// @Summary 处理 Stripe Webhook 回调
// @Tags Webhook
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/webhooks/stripe [post]
func (h *ChannelHandler) HandleStripeWebhook(c *gin.Context) {
	// 读取请求体
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		response := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "读取请求体失败", err.Error()).
			WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// 获取签名
	signature := c.GetHeader("Stripe-Signature")
	if signature == "" {
		traceID := middleware.GetRequestID(c)
		response := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "缺少签名", "").
			WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// 获取所有请求头
	headers := make(map[string]string)
	for key, values := range c.Request.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// 处理 Webhook
	if err := h.channelService.HandleWebhook(c.Request.Context(), "stripe", signature, body, headers); err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			response := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), response)
		} else {
			response := errors.NewErrorResponse(errors.ErrCodeInternalError, "处理Stripe回调失败", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, response)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	response := errors.NewSuccessResponse(gin.H{"received": true}).WithTraceID(traceID)
	c.JSON(http.StatusOK, response)
}

// HandlePayPalWebhook 处理 PayPal Webhook
// @Summary 处理 PayPal Webhook 回调
// @Tags Webhook
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/webhooks/paypal [post]
func (h *ChannelHandler) HandlePayPalWebhook(c *gin.Context) {
	// 读取请求体
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		response := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "读取请求体失败", err.Error()).
			WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// 获取签名（PayPal 使用不同的签名头）
	signature := c.GetHeader("PAYPAL-TRANSMISSION-SIG")
	if signature == "" {
		traceID := middleware.GetRequestID(c)
		response := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "缺少签名", "").
			WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// 获取所有请求头
	headers := make(map[string]string)
	for key, values := range c.Request.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// 处理 Webhook
	if err := h.channelService.HandleWebhook(c.Request.Context(), "paypal", signature, body, headers); err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			response := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), response)
		} else {
			response := errors.NewErrorResponse(errors.ErrCodeInternalError, "处理PayPal回调失败", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, response)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	response := errors.NewSuccessResponse(gin.H{"received": true}).WithTraceID(traceID)
	c.JSON(http.StatusOK, response)
}

// GetChannelConfig 获取渠道配置
// @Summary 获取渠道配置
// @Tags Channel
// @Accept json
// @Produce json
// @Param channel path string true "渠道名称"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/channel/config/{channel} [get]
func (h *ChannelHandler) GetChannelConfig(c *gin.Context) {
	channel := c.Param("channel")
	if channel == "" {
		traceID := middleware.GetRequestID(c)
		response := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "渠道名称不能为空", "").
			WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// 从上下文获取商户ID
	merchantID, exists := c.Get("merchant_id")
	if !exists {
		traceID := middleware.GetRequestID(c)
		response := errors.NewErrorResponse(errors.ErrCodeUnauthorized, "未认证", "").
			WithTraceID(traceID)
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	// 获取配置
	config, err := h.channelService.GetChannelConfig(c.Request.Context(), merchantID.(uuid.UUID), channel)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			response := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), response)
		} else {
			response := errors.NewErrorResponse(errors.ErrCodeInternalError, "获取渠道配置失败", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, response)
		}
		return
	}
	if config == nil {
		traceID := middleware.GetRequestID(c)
		response := errors.NewErrorResponse(errors.ErrCodeResourceNotFound, "配置不存在", "").
			WithTraceID(traceID)
		c.JSON(http.StatusNotFound, response)
		return
	}

	traceID := middleware.GetRequestID(c)
	response := errors.NewSuccessResponse(config).WithTraceID(traceID)
	c.JSON(http.StatusOK, response)
}

// ListChannelConfigs 列出所有渠道配置
// @Summary 列出所有渠道配置
// @Tags Channel
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/channel/config [get]
func (h *ChannelHandler) ListChannelConfigs(c *gin.Context) {
	// 从上下文获取商户ID
	merchantID, exists := c.Get("merchant_id")
	if !exists {
		traceID := middleware.GetRequestID(c)
		response := errors.NewErrorResponse(errors.ErrCodeUnauthorized, "未认证", "").
			WithTraceID(traceID)
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	// 列出配置
	configs, err := h.channelService.ListChannelConfigs(c.Request.Context(), merchantID.(uuid.UUID))
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			response := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), response)
		} else {
			response := errors.NewErrorResponse(errors.ErrCodeInternalError, "列出渠道配置失败", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, response)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	response := errors.NewSuccessResponse(configs).WithTraceID(traceID)
	c.JSON(http.StatusOK, response)
}

// CreatePreAuth 创建预授权
// @Summary 创建预授权
// @Tags Channel
// @Accept json
// @Produce json
// @Param request body service.CreatePreAuthRequest true "创建预授权请求"
// @Success 200 {object} service.CreatePreAuthResponse
// @Router /api/v1/channel/pre-auth [post]
func (h *ChannelHandler) CreatePreAuth(c *gin.Context) {
	var req service.CreatePreAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		traceID := middleware.GetRequestID(c)
		response := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的请求参数", err.Error()).
			WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	resp, err := h.channelService.CreatePreAuth(c.Request.Context(), &req)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		response := errors.NewErrorResponse(errors.ErrCodeInternalError, "创建预授权失败", err.Error()).
			WithTraceID(traceID)
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "创建预授权成功",
		"data":    resp,
	})
}

// CapturePreAuth 确认预授权（扣款）
// @Summary 确认预授权
// @Tags Channel
// @Accept json
// @Produce json
// @Param request body service.CapturePreAuthRequest true "确认预授权请求"
// @Success 200 {object} service.CapturePreAuthResponse
// @Router /api/v1/channel/pre-auth/capture [post]
func (h *ChannelHandler) CapturePreAuth(c *gin.Context) {
	var req service.CapturePreAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		traceID := middleware.GetRequestID(c)
		response := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的请求参数", err.Error()).
			WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	resp, err := h.channelService.CapturePreAuth(c.Request.Context(), &req)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		response := errors.NewErrorResponse(errors.ErrCodeInternalError, "确认预授权失败", err.Error()).
			WithTraceID(traceID)
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "确认预授权成功",
		"data":    resp,
	})
}

// CancelPreAuth 取消预授权（释放资金）
// @Summary 取消预授权
// @Tags Channel
// @Accept json
// @Produce json
// @Param request body service.CancelPreAuthRequest true "取消预授权请求"
// @Success 200 {object} service.CancelPreAuthResponse
// @Router /api/v1/channel/pre-auth/cancel [post]
func (h *ChannelHandler) CancelPreAuth(c *gin.Context) {
	var req service.CancelPreAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		traceID := middleware.GetRequestID(c)
		response := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的请求参数", err.Error()).
			WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	resp, err := h.channelService.CancelPreAuth(c.Request.Context(), &req)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		response := errors.NewErrorResponse(errors.ErrCodeInternalError, "取消预授权失败", err.Error()).
			WithTraceID(traceID)
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "取消预授权成功",
		"data":    resp,
	})
}

// QueryPreAuth 查询预授权状态
// @Summary 查询预授权
// @Tags Channel
// @Produce json
// @Param channel_pre_auth_no path string true "渠道预授权号"
// @Success 200 {object} service.QueryPreAuthResponse
// @Router /api/v1/channel/pre-auth/{channel_pre_auth_no} [get]
func (h *ChannelHandler) QueryPreAuth(c *gin.Context) {
	channelPreAuthNo := c.Param("channel_pre_auth_no")
	if channelPreAuthNo == "" {
		traceID := middleware.GetRequestID(c)
		response := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "缺少预授权号", "").
			WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	resp, err := h.channelService.QueryPreAuth(c.Request.Context(), channelPreAuthNo)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		response := errors.NewErrorResponse(errors.ErrCodeInternalError, "查询预授权失败", err.Error()).
			WithTraceID(traceID)
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    "SUCCESS",
		"message": "查询预授权成功",
		"data":    resp,
	})
}

// RegisterRoutes 注册路由
func (h *ChannelHandler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	{
		// 支付相关（完整RESTful路由）
		api.POST("/channel/payments", h.CreatePayment)
		api.GET("/channel/payments/:payment_no", h.QueryPayment)
		api.POST("/channel/payments/:payment_no/cancel", h.CancelPayment)

		// 退款相关（完整RESTful路由）
		api.POST("/channel/refunds", h.CreateRefund)
		api.GET("/channel/refunds/:refund_no", h.QueryRefund)

		// 预授权相关（完整RESTful路由）
		api.POST("/channel/pre-auth", h.CreatePreAuth)
		api.POST("/channel/pre-auth/capture", h.CapturePreAuth)
		api.POST("/channel/pre-auth/cancel", h.CancelPreAuth)
		api.GET("/channel/pre-auth/:channel_pre_auth_no", h.QueryPreAuth)

		// Payment-Gateway 兼容路由（简化版）
		api.POST("/channel/payment", h.CreatePayment)      // 别名路由
		api.POST("/channel/refund", h.CreateRefund)        // 别名路由
		api.POST("/channel/query", h.QueryPaymentCompat)   // 查询接口

		// Webhook 回调（不需要认证）
		api.POST("/webhooks/stripe", h.HandleStripeWebhook)
		api.POST("/webhooks/paypal", h.HandlePayPalWebhook)

		// 渠道配置
		api.GET("/channel/config", h.ListChannelConfigs)
		api.GET("/channel/config/:channel", h.GetChannelConfig)
	}
}
