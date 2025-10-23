package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/payment-platform/services/order-service/internal/repository"
	"github.com/payment-platform/services/order-service/internal/service"
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
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	order, err := h.orderService.CreateOrder(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(order))
}

// GetOrder 获取订单详情
func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderNo := c.Param("orderNo")

	order, err := h.orderService.GetOrder(c.Request.Context(), orderNo)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(order))
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
			c.JSON(http.StatusBadRequest, ErrorResponse("无效的商户ID"))
			return
		}
		query.MerchantID = &merchantID
	}

	if customerIDStr := c.Query("customer_id"); customerIDStr != "" {
		customerID, err := uuid.Parse(customerIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse("无效的客户ID"))
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
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(PageResponse{
		List:     orders,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}))
}

// CancelOrder 取消订单
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	orderNo := c.Param("orderNo")

	var req CancelOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
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
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(nil))
}

// PayOrder 支付订单
func (h *OrderHandler) PayOrder(c *gin.Context) {
	orderNo := c.Param("orderNo")

	var req PayOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	err := h.orderService.PayOrder(c.Request.Context(), orderNo, req.PaymentNo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(nil))
}

// RefundOrder 退款订单
func (h *OrderHandler) RefundOrder(c *gin.Context) {
	orderNo := c.Param("orderNo")

	var req RefundOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	err := h.orderService.RefundOrder(c.Request.Context(), orderNo, req.Amount, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(nil))
}

// ShipOrder 发货
func (h *OrderHandler) ShipOrder(c *gin.Context) {
	orderNo := c.Param("orderNo")

	var req ShipOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	err := h.orderService.ShipOrder(c.Request.Context(), orderNo, req.ShippingInfo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(nil))
}

// CompleteOrder 完成订单
func (h *OrderHandler) CompleteOrder(c *gin.Context) {
	orderNo := c.Param("orderNo")

	err := h.orderService.CompleteOrder(c.Request.Context(), orderNo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(nil))
}

// GetOrderStatistics 获取订单统计
func (h *OrderHandler) GetOrderStatistics(c *gin.Context) {
	merchantIDStr := c.Query("merchant_id")
	if merchantIDStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse("缺少商户ID"))
		return
	}

	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("无效的商户ID"))
		return
	}

	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("无效的开始时间"))
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("无效的结束时间"))
		return
	}

	currency := c.DefaultQuery("currency", "USD")

	statistics, err := h.orderService.GetOrderStatistics(c.Request.Context(), merchantID, startTime, endTime, currency)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(statistics))
}

// GetDailySummary 获取每日汇总
func (h *OrderHandler) GetDailySummary(c *gin.Context) {
	merchantIDStr := c.Query("merchant_id")
	if merchantIDStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse("缺少商户ID"))
		return
	}

	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("无效的商户ID"))
		return
	}

	dateStr := c.Query("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		date = time.Now()
	}

	summary, err := h.orderService.GetDailySummary(c.Request.Context(), merchantID, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(summary))
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
