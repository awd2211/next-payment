package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/payment-platform/pkg/errors"
	"github.com/payment-platform/pkg/middleware"
	"payment-platform/order-service/internal/repository"
	"payment-platform/order-service/internal/service"
)

// OrderHandler 订单处理器
type OrderHandler struct {
	orderService service.OrderService
}

// NewOrderHandler 创建订单处理器实例
func NewOrderHandler(orderService service.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// RegisterRoutes 注册路由
func (h *OrderHandler) RegisterRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		// 订单管理
		orders := v1.Group("/orders")
		{
			orders.POST("", h.CreateOrder)
			orders.GET("/:orderNo", h.GetOrder)
			orders.GET("", h.QueryOrders)
			orders.POST("/:orderNo/cancel", h.CancelOrder)
			orders.POST("/:orderNo/pay", h.PayOrder)
			orders.POST("/:orderNo/refund", h.RefundOrder)
			orders.POST("/:orderNo/ship", h.ShipOrder)
			orders.POST("/:orderNo/complete", h.CompleteOrder)
			orders.PUT("/:orderNo/status", h.UpdateOrderStatus)
		}

		// 统计分析
		statistics := v1.Group("/statistics")
		{
			statistics.GET("/orders", h.GetOrderStatistics)
			statistics.GET("/daily-summary", h.GetDailySummary)
		}
	}
}

// CreateOrder 创建订单
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var input service.CreateOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).
			WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	order, err := h.orderService.CreateOrder(c.Request.Context(), &input)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "创建订单失败", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(order).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// GetOrder 获取订单详情
func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderNo := c.Param("orderNo")

	order, err := h.orderService.GetOrder(c.Request.Context(), orderNo)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeResourceNotFound, "订单不存在", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusNotFound, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(order).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// QueryOrders 查询订单列表
func (h *OrderHandler) QueryOrders(c *gin.Context) {
	query := &repository.OrderQuery{
		Status:        c.Query("status"),
		PayStatus:     c.Query("pay_status"),
		ShippingStatus: c.Query("shipping_status"),
		Currency:      c.Query("currency"),
		CustomerEmail: c.Query("customer_email"),
		Keyword:       c.Query("keyword"),
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

	if customerIDStr := c.Query("customer_id"); customerIDStr != "" {
		customerID, err := uuid.Parse(customerIDStr)
		if err != nil {
			traceID := middleware.GetRequestID(c)
			resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的客户ID", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		query.CustomerID = &customerID
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

	orders, total, err := h.orderService.QueryOrders(c.Request.Context(), query)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "查询订单列表失败", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(PageResponse{
		List:     orders,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// CancelOrder 取消订单
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	orderNo := c.Param("orderNo")

	var req CancelOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).
			WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	operatorID := uuid.Nil
	if req.OperatorID != "" {
		id, err := uuid.Parse(req.OperatorID)
		if err == nil {
			operatorID = id
		}
	}

	err := h.orderService.CancelOrder(c.Request.Context(), orderNo, req.Reason, operatorID, req.OperatorType)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "取消订单失败", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(nil).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// PayOrder 支付订单
func (h *OrderHandler) PayOrder(c *gin.Context) {
	orderNo := c.Param("orderNo")

	var req PayOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).
			WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	err := h.orderService.PayOrder(c.Request.Context(), orderNo, req.PaymentNo)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "支付订单失败", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(nil).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// RefundOrder 退款订单
func (h *OrderHandler) RefundOrder(c *gin.Context) {
	orderNo := c.Param("orderNo")

	var req RefundOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).
			WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	err := h.orderService.RefundOrder(c.Request.Context(), orderNo, req.Amount, req.Reason)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "退款订单失败", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(nil).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// ShipOrder 发货
func (h *OrderHandler) ShipOrder(c *gin.Context) {
	orderNo := c.Param("orderNo")

	var req ShipOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).
			WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	err := h.orderService.ShipOrder(c.Request.Context(), orderNo, req.ShippingInfo)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "发货失败", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(nil).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// CompleteOrder 完成订单
func (h *OrderHandler) CompleteOrder(c *gin.Context) {
	orderNo := c.Param("orderNo")

	err := h.orderService.CompleteOrder(c.Request.Context(), orderNo)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "完成订单失败", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(nil).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// UpdateOrderStatus 更新订单状态（支付网关回调使用）
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	orderNoOrPaymentNo := c.Param("orderNo") // 可以是订单号或支付流水号

	var req UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).
			WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// 尝试先按支付流水号查询
	order, err := h.orderService.GetOrderByPaymentNo(c.Request.Context(), orderNoOrPaymentNo)
	if err != nil || order == nil {
		// 如果找不到，再尝试按订单号查询
		order, err = h.orderService.GetOrder(c.Request.Context(), orderNoOrPaymentNo)
		if err != nil {
			traceID := middleware.GetRequestID(c)
			resp := errors.NewErrorResponse(errors.ErrCodeResourceNotFound, "订单不存在", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusNotFound, resp)
			return
		}
	}

	// 更新订单状态
	operatorID := uuid.Nil
	err = h.orderService.UpdateOrderStatus(c.Request.Context(), order.OrderNo, req.Status, operatorID, "system")
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "更新订单状态失败", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(nil).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// GetOrderStatistics 获取订单统计
func (h *OrderHandler) GetOrderStatistics(c *gin.Context) {
	merchantIDStr := c.Query("merchant_id")
	if merchantIDStr == "" {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "缺少商户ID", "").
			WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的商户ID", err.Error()).
			WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的开始时间", err.Error()).
			WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的结束时间", err.Error()).
			WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	currency := c.DefaultQuery("currency", "USD")

	statistics, err := h.orderService.GetOrderStatistics(c.Request.Context(), merchantID, startTime, endTime, currency)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "获取订单统计失败", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(statistics).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// GetDailySummary 获取每日汇总
func (h *OrderHandler) GetDailySummary(c *gin.Context) {
	merchantIDStr := c.Query("merchant_id")
	if merchantIDStr == "" {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "缺少商户ID", "").
			WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的商户ID", err.Error()).
			WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	dateStr := c.Query("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		date = time.Now()
	}

	summary, err := h.orderService.GetDailySummary(c.Request.Context(), merchantID, date)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "获取每日汇总失败", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(summary).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// Request structures

type CancelOrderRequest struct {
	Reason       string `json:"reason" binding:"required"`
	OperatorID   string `json:"operator_id"`
	OperatorType string `json:"operator_type"`
}

type PayOrderRequest struct {
	PaymentNo string `json:"payment_no" binding:"required"`
}

type RefundOrderRequest struct {
	Amount int64  `json:"amount" binding:"required,gt=0"`
	Reason string `json:"reason" binding:"required"`
}

type ShipOrderRequest struct {
	ShippingInfo map[string]interface{} `json:"shipping_info" binding:"required"`
}

type UpdateOrderStatusRequest struct {
	Status         string `json:"status" binding:"required"`
	ChannelOrderNo string `json:"channel_order_no,omitempty"`
	PaidAt         string `json:"paid_at,omitempty"`
	ErrorCode      string `json:"error_code,omitempty"`
	ErrorMsg       string `json:"error_msg,omitempty"`
}

// Response structures

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
