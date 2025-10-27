package handler

import (
	"net/http"
	"strconv"
	"time"

	"payment-platform/order-service/internal/repository"
	"payment-platform/order-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/payment-platform/pkg/errors"
	"github.com/payment-platform/pkg/logger"
	"github.com/payment-platform/pkg/middleware"
	"go.uber.org/zap"
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
			orders.POST("/batch", h.BatchGetOrders) // 批量查询订单
			orders.GET("/stats", h.GetOrderStats)   // 添加统计接口
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
//
//	@Summary		创建订单
//	@Description	创建新的订单记录
//	@Tags			Orders
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		service.CreateOrderInput	true	"订单创建请求"
//	@Success		200		{object}	Response
//	@Failure		400		{object}	Response
//	@Failure		500		{object}	Response
//	@Router			/orders [post]
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
//
//	@Summary		获取订单详情
//	@Description	根据订单号获取订单详情
//	@Tags			Orders
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			orderNo	path		string	true	"订单号"
//	@Success		200		{object}	Response
//	@Failure		404		{object}	Response
//	@Failure		500		{object}	Response
//	@Router			/orders/{orderNo} [get]
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
//
//	@Summary		查询订单列表
//	@Description	根据条件查询订单列表，支持分页和多维度筛选
//	@Tags			Orders
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			merchant_id		query		string	false	"商户ID"
//	@Param			customer_id		query		string	false	"客户ID"
//	@Param			status			query		string	false	"订单状态 (pending/confirmed/cancelled/completed)"
//	@Param			pay_status		query		string	false	"支付状态 (unpaid/paid/refunded)"
//	@Param			shipping_status	query		string	false	"发货状态 (unshipped/shipped/delivered)"
//	@Param			currency		query		string	false	"货币类型"
//	@Param			customer_email	query		string	false	"客户邮箱"
//	@Param			keyword			query		string	false	"关键词搜索"
//	@Param			start_time		query		string	false	"开始时间 (RFC3339格式)"
//	@Param			end_time		query		string	false	"结束时间 (RFC3339格式)"
//	@Param			page			query		int		false	"页码"	default(1)
//	@Param			page_size		query		int		false	"每页数量"	default(20)
//	@Success		200				{object}	Response
//	@Failure		400				{object}	Response
//	@Failure		500				{object}	Response
//	@Router			/orders [get]
func (h *OrderHandler) QueryOrders(c *gin.Context) {
	query := &repository.OrderQuery{
		Status:         c.Query("status"),
		PayStatus:      c.Query("pay_status"),
		ShippingStatus: c.Query("shipping_status"),
		Currency:       c.Query("currency"),
		CustomerEmail:  c.Query("customer_email"),
		Keyword:        c.Query("keyword"),
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
	// 验证并限制分页参数（防止DoS攻击）
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100 // 最大限制100条/页
	}

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

// BatchGetOrdersRequest 批量查询订单请求
type BatchGetOrdersRequest struct {
	OrderNos   []string  `json:"order_nos" binding:"required,min=1,max=100"`
	MerchantID uuid.UUID `json:"merchant_id" binding:"required"`
}

// BatchGetOrdersResponse 批量查询订单响应
type BatchGetOrdersResponse struct {
	Results map[string]interface{} `json:"results"` // orderNo -> Order
	Failed  []string               `json:"failed"`  // 查询失败的 orderNo
	Summary struct {
		Total      int `json:"total"`       // 请求的总数
		Found      int `json:"found"`       // 找到的数量
		NotFound   int `json:"not_found"`   // 未找到的数量
	} `json:"summary"`
}

// BatchGetOrders 批量查询订单
//
//	@Summary		批量查询订单
//	@Description	一次性查询多个订单（最多100个）
//	@Tags			Orders
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		BatchGetOrdersRequest	true	"批量查询请求"
//	@Success		200		{object}	BatchGetOrdersResponse
//	@Failure		400		{object}	Response
//	@Failure		500		{object}	Response
//	@Router			/orders/batch [post]
func (h *OrderHandler) BatchGetOrders(c *gin.Context) {
	var req BatchGetOrdersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).
			WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// 调用服务层批量查询
	results, failed, err := h.orderService.BatchGetOrders(c.Request.Context(), req.OrderNos, req.MerchantID)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "批量查询订单失败", err.Error()).
				WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	// 构建响应
	response := BatchGetOrdersResponse{
		Results: make(map[string]interface{}),
		Failed:  failed,
	}
	for orderNo, order := range results {
		response.Results[orderNo] = order
	}
	response.Summary.Total = len(req.OrderNos)
	response.Summary.Found = len(results)
	response.Summary.NotFound = len(failed)

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(response).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// CancelOrder 取消订单
//
//	@Summary		取消订单
//	@Description	取消指定订单
//	@Tags			Orders
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			orderNo	path		string				true	"订单号"
//	@Param			request	body		CancelOrderRequest	true	"取消原因"
//	@Success		200		{object}	Response
//	@Failure		400		{object}	Response
//	@Failure		500		{object}	Response
//	@Router			/orders/{orderNo}/cancel [post]
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
//
//	@Summary		支付订单
//	@Description	标记订单为已支付状态
//	@Tags			Orders
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			orderNo	path		string			true	"订单号"
//	@Param			request	body		PayOrderRequest	true	"支付信息"
//	@Success		200		{object}	Response
//	@Failure		400		{object}	Response
//	@Failure		500		{object}	Response
//	@Router			/orders/{orderNo}/pay [post]
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
//
//	@Summary		退款订单
//	@Description	对已支付订单进行退款
//	@Tags			Orders
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			orderNo	path		string				true	"订单号"
//	@Param			request	body		RefundOrderRequest	true	"退款信息"
//	@Success		200		{object}	Response
//	@Failure		400		{object}	Response
//	@Failure		500		{object}	Response
//	@Router			/orders/{orderNo}/refund [post]
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
//
//	@Summary		订单发货
//	@Description	更新订单发货信息
//	@Tags			Orders
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			orderNo	path		string			true	"订单号"
//	@Param			request	body		ShipOrderRequest	true	"发货信息"
//	@Success		200		{object}	Response
//	@Failure		400		{object}	Response
//	@Failure		500		{object}	Response
//	@Router			/orders/{orderNo}/ship [post]
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
//
//	@Summary		完成订单
//	@Description	将订单标记为已完成状态
//	@Tags			Orders
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			orderNo	path		string	true	"订单号"
//	@Success		200		{object}	Response
//	@Failure		400		{object}	Response
//	@Failure		500		{object}	Response
//	@Router			/orders/{orderNo}/complete [post]
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
//
//	@Summary		更新订单状态
//	@Description	更新订单状态（供支付网关回调使用）
//	@Tags			Orders
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			orderNo	path		string					true	"订单号或支付流水号"
//	@Param			request	body		UpdateOrderStatusRequest	true	"状态更新信息"
//	@Success		200		{object}	Response
//	@Failure		400		{object}	Response
//	@Failure		404		{object}	Response
//	@Failure		500		{object}	Response
//	@Router			/orders/{orderNo}/status [put]
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
//
//	@Summary		获取订单统计
//	@Description	获取指定时间范围内的订单统计信息
//	@Tags			Statistics
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			merchant_id	query		string	true	"商户ID"
//	@Param			start_time	query		string	true	"开始时间 (RFC3339格式)"
//	@Param			end_time	query		string	true	"结束时间 (RFC3339格式)"
//	@Param			currency	query		string	false	"货币类型"	default(USD)
//	@Success		200			{object}	Response
//	@Failure		400			{object}	Response
//	@Failure		500			{object}	Response
//	@Router			/statistics/orders [get]
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
//
//	@Summary		获取每日汇总
//	@Description	获取指定日期的订单每日汇总数据
//	@Tags			Statistics
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			merchant_id	query		string	true	"商户ID"
//	@Param			date		query		string	true	"日期 (YYYY-MM-DD)"
//	@Param			currency	query		string	false	"货币类型"	default(USD)
//	@Success		200			{object}	Response
//	@Failure		400			{object}	Response
//	@Failure		500			{object}	Response
//	@Router			/statistics/daily-summary [get]
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

// GetOrderStats 获取订单统计（实时数据库聚合）
// @Summary 获取订单统计
// @Description 获取全局订单统计数据（总金额、总订单数、已支付/待支付/已取消订单数、今日订单数据）
// @Tags Order
// @Produce json
// @Success 200 {object} map[string]interface{} "订单统计数据"
// @Failure 500 {object} map[string]interface{} "服务器错误"
// @Router /api/v1/orders/stats [get]
func (h *OrderHandler) GetOrderStats(c *gin.Context) {
	ctx := c.Request.Context()
	traceID := middleware.GetRequestID(c)

	// 调用service层获取真实统计数据
	stats, err := h.orderService.GetOrderStats(ctx)
	if err != nil {
		logger.Error("failed to get order stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(
			errors.ErrCodeInternalError,
			"获取订单统计失败",
			err.Error(),
		).WithTraceID(traceID))
		return
	}

	resp := errors.NewSuccessResponse(stats).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}
