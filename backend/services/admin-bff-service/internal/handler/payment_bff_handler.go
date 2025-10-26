package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"payment-platform/admin-service/internal/client"
	localMiddleware "payment-platform/admin-service/internal/middleware"
	"payment-platform/admin-service/internal/service"
	"payment-platform/admin-service/internal/utils"
)

type PaymentBFFHandler struct {
	paymentClient   *client.ServiceClient
	auditLogService service.AuditLogService
	auditHelper     *utils.AuditHelper
}

func NewPaymentBFFHandler(paymentServiceURL string, auditLogService service.AuditLogService) *PaymentBFFHandler {
	return &PaymentBFFHandler{
		paymentClient:   client.NewServiceClient(paymentServiceURL),
		auditLogService: auditLogService,
		auditHelper:     utils.NewAuditHelper(auditLogService),
	}
}

func (h *PaymentBFFHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	admin := r.Group("/admin/payments")
	admin.Use(authMiddleware)
	{
		// 支付查询 (需要权限 + 原因)
		admin.GET("",
			localMiddleware.RequirePermission("payments.view"),
			localMiddleware.RequireReason,
			h.ListPayments,
		)
		admin.GET("/:id",
			localMiddleware.RequirePermission("payments.view"),
			localMiddleware.RequireReason,
			h.GetPayment,
		)
		admin.GET("/payment-no/:payment_no",
			localMiddleware.RequirePermission("payments.view"),
			localMiddleware.RequireReason,
			h.GetPaymentByNo,
		)

		// 退款管理
		refunds := admin.Group("/refunds")
		{
			refunds.GET("",
				localMiddleware.RequirePermission("payments.refund"),
				localMiddleware.RequireReason,
				h.ListRefunds,
			)
			refunds.GET("/:id",
				localMiddleware.RequirePermission("payments.refund"),
				localMiddleware.RequireReason,
				h.GetRefund,
			)
			refunds.POST("",
				localMiddleware.RequirePermission("payments.refund"),
				localMiddleware.RequireReason,
				h.CreateRefund,
			)
			refunds.POST("/:id/approve",
				localMiddleware.RequirePermission("payments.refund"),
				localMiddleware.RequireReason,
				h.ApproveRefund,
			)
			refunds.POST("/:id/reject",
				localMiddleware.RequirePermission("payments.refund"),
				localMiddleware.RequireReason,
				h.RejectRefund,
			)
		}

		// 统计 (不需要reason)
		admin.GET("/statistics",
			localMiddleware.RequirePermission("payments.view"),
			h.GetStatistics,
		)
		admin.GET("/statistics/trend",
			localMiddleware.RequirePermission("payments.view"),
			h.GetTrendStatistics,
		)
		admin.GET("/statistics/channel",
			localMiddleware.RequirePermission("payments.view"),
			h.GetChannelStatistics,
		)

		// 支付渠道统计
		admin.GET("/channels/statistics",
			localMiddleware.RequirePermission("payments.view"),
			h.GetChannelPerformance,
		)

		// 交易监控
		admin.GET("/monitor/realtime",
			localMiddleware.RequirePermission("payments.view"),
			h.GetRealtimeMonitor,
		)
		admin.GET("/monitor/alerts",
			localMiddleware.RequirePermission("payments.view"),
			h.GetAlerts,
		)
	}
}

// ========== 支付查询 ==========

func (h *PaymentBFFHandler) ListPayments(c *gin.Context) {
	adminID := c.GetString("user_id")
	adminUsername := c.GetString("username")
	reason := c.GetString("operation_reason")

	queryParams := make(map[string]string)
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}
	if orderNo := c.Query("order_no"); orderNo != "" {
		queryParams["order_no"] = orderNo
	}
	if paymentNo := c.Query("payment_no"); paymentNo != "" {
		queryParams["payment_no"] = paymentNo
	}
	if channel := c.Query("channel"); channel != "" {
		queryParams["channel"] = channel
	}
	if status := c.Query("status"); status != "" {
		queryParams["status"] = status
	}
	if currency := c.Query("currency"); currency != "" {
		queryParams["currency"] = currency
	}
	if startTime := c.Query("start_time"); startTime != "" {
		queryParams["start_time"] = startTime
	}
	if endTime := c.Query("end_time"); endTime != "" {
		queryParams["end_time"] = endTime
	}
	if minAmount := c.Query("min_amount"); minAmount != "" {
		queryParams["min_amount"] = minAmount
	}
	if maxAmount := c.Query("max_amount"); maxAmount != "" {
		queryParams["max_amount"] = maxAmount
	}
	if page := c.Query("page"); page != "" {
		queryParams["page"] = page
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		queryParams["page_size"] = pageSize
	}

	result, statusCode, err := h.paymentClient.Get(c.Request.Context(), "/api/v1/payments", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Payment Gateway失败", "details": err.Error()})
		return
	}

	// 数据脱敏
	if data, ok := result["data"].(map[string]interface{}); ok {
		result["data"] = utils.MaskSensitiveData(data)
	}

	// 记录审计日志
	go func() {
		adminUUID, _ := uuid.Parse(adminID)
		logReq := &service.CreateAuditLogRequest{
			AdminID:      adminUUID,
			AdminName:    adminUsername,
			Action:       "VIEW_PAYMENTS",
			Resource:     "payment",
			ResourceID:   queryParams["merchant_id"],
			Method:       "GET",
			Path:         "/api/v1/admin/payments",
			IP:           c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			Description:  reason,
			ResponseCode: statusCode,
		}
		_ = h.auditLogService.CreateLog(c.Request.Context(), logReq)
	}()

	c.JSON(statusCode, result)
}

func (h *PaymentBFFHandler) GetPayment(c *gin.Context) {
	adminID := c.GetString("user_id")
	adminUsername := c.GetString("username")
	reason := c.GetString("operation_reason")
	id := c.Param("id")

	result, statusCode, err := h.paymentClient.Get(c.Request.Context(), "/api/v1/payments/"+id, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Payment Gateway失败", "details": err.Error()})
		return
	}

	// 数据脱敏
	if data, ok := result["data"].(map[string]interface{}); ok {
		result["data"] = utils.MaskSensitiveData(data)
	}

	// 记录审计日志
	go func() {
		adminUUID, _ := uuid.Parse(adminID)
		logReq := &service.CreateAuditLogRequest{
			AdminID:      adminUUID,
			AdminName:    adminUsername,
			Action:       "VIEW_PAYMENT_DETAIL",
			Resource:     "payment",
			ResourceID:   id,
			Method:       "GET",
			Path:         "/api/v1/admin/payments/" + id,
			IP:           c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			Description:  reason,
			ResponseCode: statusCode,
		}
		_ = h.auditLogService.CreateLog(c.Request.Context(), logReq)
	}()

	c.JSON(statusCode, result)
}

func (h *PaymentBFFHandler) GetPaymentByNo(c *gin.Context) {
	paymentNo := c.Param("payment_no")

	result, statusCode, err := h.paymentClient.Get(c.Request.Context(), "/api/v1/payments/payment-no/"+paymentNo, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Payment Gateway失败", "details": err.Error()})
		return
	}

	// 数据脱敏
	if data, ok := result["data"].(map[string]interface{}); ok {
		result["data"] = utils.MaskSensitiveData(data)
	}

	// 记录审计日志
	h.auditHelper.LogCrossTenantAccess(c, "VIEW_PAYMENT_BY_NO", "payment", paymentNo, "", statusCode)

	c.JSON(statusCode, result)
}

// ========== 退款管理 ==========

func (h *PaymentBFFHandler) ListRefunds(c *gin.Context) {
	queryParams := make(map[string]string)
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}
	if paymentNo := c.Query("payment_no"); paymentNo != "" {
		queryParams["payment_no"] = paymentNo
	}
	if refundNo := c.Query("refund_no"); refundNo != "" {
		queryParams["refund_no"] = refundNo
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

	result, statusCode, err := h.paymentClient.Get(c.Request.Context(), "/api/v1/refunds", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Payment Gateway失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *PaymentBFFHandler) GetRefund(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.paymentClient.Get(c.Request.Context(), "/api/v1/refunds/"+id, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Payment Gateway失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *PaymentBFFHandler) CreateRefund(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	adminID := c.GetString("user_id")
	req["operator_id"] = adminID

	result, statusCode, err := h.paymentClient.Post(c.Request.Context(), "/api/v1/refunds", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Payment Gateway失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *PaymentBFFHandler) ApproveRefund(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		req = make(map[string]interface{})
	}

	adminID := c.GetString("user_id")
	req["approved_by"] = adminID

	result, statusCode, err := h.paymentClient.Post(c.Request.Context(), "/api/v1/refunds/"+id+"/approve", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Payment Gateway失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *PaymentBFFHandler) RejectRefund(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	adminID := c.GetString("user_id")
	req["rejected_by"] = adminID

	result, statusCode, err := h.paymentClient.Post(c.Request.Context(), "/api/v1/refunds/"+id+"/reject", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Payment Gateway失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== 统计 ==========

func (h *PaymentBFFHandler) GetStatistics(c *gin.Context) {
	queryParams := make(map[string]string)
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}
	if startTime := c.Query("start_time"); startTime != "" {
		queryParams["start_time"] = startTime
	}
	if endTime := c.Query("end_time"); endTime != "" {
		queryParams["end_time"] = endTime
	}

	result, statusCode, err := h.paymentClient.Get(c.Request.Context(), "/api/v1/statistics", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Payment Gateway失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *PaymentBFFHandler) GetTrendStatistics(c *gin.Context) {
	queryParams := make(map[string]string)
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}
	if period := c.Query("period"); period != "" {
		queryParams["period"] = period
	}
	if startTime := c.Query("start_time"); startTime != "" {
		queryParams["start_time"] = startTime
	}
	if endTime := c.Query("end_time"); endTime != "" {
		queryParams["end_time"] = endTime
	}

	result, statusCode, err := h.paymentClient.Get(c.Request.Context(), "/api/v1/statistics/trend", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Payment Gateway失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *PaymentBFFHandler) GetChannelStatistics(c *gin.Context) {
	queryParams := make(map[string]string)
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}
	if startTime := c.Query("start_time"); startTime != "" {
		queryParams["start_time"] = startTime
	}
	if endTime := c.Query("end_time"); endTime != "" {
		queryParams["end_time"] = endTime
	}

	result, statusCode, err := h.paymentClient.Get(c.Request.Context(), "/api/v1/statistics/channel", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Payment Gateway失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *PaymentBFFHandler) GetChannelPerformance(c *gin.Context) {
	queryParams := make(map[string]string)
	if startTime := c.Query("start_time"); startTime != "" {
		queryParams["start_time"] = startTime
	}
	if endTime := c.Query("end_time"); endTime != "" {
		queryParams["end_time"] = endTime
	}

	result, statusCode, err := h.paymentClient.Get(c.Request.Context(), "/api/v1/channels/statistics", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Payment Gateway失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== 交易监控 ==========

func (h *PaymentBFFHandler) GetRealtimeMonitor(c *gin.Context) {
	queryParams := make(map[string]string)
	if interval := c.Query("interval"); interval != "" {
		queryParams["interval"] = interval
	}

	result, statusCode, err := h.paymentClient.Get(c.Request.Context(), "/api/v1/monitor/realtime", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Payment Gateway失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *PaymentBFFHandler) GetAlerts(c *gin.Context) {
	queryParams := make(map[string]string)
	if severity := c.Query("severity"); severity != "" {
		queryParams["severity"] = severity
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

	result, statusCode, err := h.paymentClient.Get(c.Request.Context(), "/api/v1/monitor/alerts", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Payment Gateway失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}
