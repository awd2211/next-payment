package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"payment-platform/admin-service/internal/client"
	"payment-platform/admin-service/internal/service"
)

// OrderBFFHandler Order Service BFF处理器
type OrderBFFHandler struct {
	orderClient     *client.ServiceClient
	auditLogService service.AuditLogService
}

// NewOrderBFFHandler 创建Order BFF处理器
func NewOrderBFFHandler(orderServiceURL string, auditLogService service.AuditLogService) *OrderBFFHandler {
	return &OrderBFFHandler{
		orderClient:     client.NewServiceClient(orderServiceURL),
		auditLogService: auditLogService,
	}
}

// RegisterRoutes 注册路由
func (h *OrderBFFHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	admin := r.Group("/admin/orders")
	admin.Use(authMiddleware)
	{
		// 订单查询 (只读)
		admin.GET("", h.ListOrders)
		admin.GET("/:order_no", h.GetOrder)
		admin.GET("/merchant/:merchant_id", h.GetMerchantOrders)

		// 订单统计 (只读)
		admin.GET("/statistics", h.GetOrderStatistics)
		admin.GET("/status-summary", h.GetOrderStatusSummary)
	}
}

// ListOrders 获取订单列表
func (h *OrderBFFHandler) ListOrders(c *gin.Context) {
	// 1. 获取管理员信息
	adminID := c.GetString("user_id")
	adminUsername := c.GetString("username")

	// 2. 必须提供查看原因
	reason := c.Query("reason")
	if reason == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "必须提供查看原因 (reason参数)",
			"details": map[string]string{
				"example": "GET /api/v1/admin/orders?reason=客户投诉调查&merchant_id=xxx",
			},
		})
		return
	}

	// 3. 构建查询参数
	queryParams := make(map[string]string)
	if page := c.Query("page"); page != "" {
		queryParams["page"] = page
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		queryParams["page_size"] = pageSize
	}
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
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

	// 4. 调用 Order Service
	result, statusCode, err := h.orderClient.Get(c.Request.Context(), "/api/v1/orders", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Order Service失败", "details": err.Error()})
		return
	}

	// 5. 记录审计日志 (异步)
	go func() {
		adminUUID, _ := uuid.Parse(adminID)
		logReq := &service.CreateAuditLogRequest{
			AdminID:      adminUUID,
			AdminName:    adminUsername,
			Action:       "VIEW_MERCHANT_ORDERS",
			Resource:     "order",
			ResourceID:   queryParams["merchant_id"],
			Method:       "GET",
			Path:         "/api/v1/admin/orders",
			IP:           c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			Description:  reason,
			ResponseCode: statusCode,
		}
		_ = h.auditLogService.CreateLog(c.Request.Context(), logReq)
	}()

	// 6. 返回响应
	c.JSON(statusCode, result)
}

// GetOrder 获取订单详情
func (h *OrderBFFHandler) GetOrder(c *gin.Context) {
	// 1. 获取管理员信息
	adminID := c.GetString("user_id")
	adminUsername := c.GetString("username")

	// 2. 必须提供查看原因
	reason := c.Query("reason")
	if reason == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "必须提供查看原因 (reason参数)",
		})
		return
	}

	// 3. 获取订单号
	orderNo := c.Param("order_no")

	// 4. 调用 Order Service
	result, statusCode, err := h.orderClient.Get(c.Request.Context(), "/api/v1/orders/"+orderNo, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Order Service失败", "details": err.Error()})
		return
	}

	// 5. 记录审计日志 (异步)
	go func() {
		adminUUID, _ := uuid.Parse(adminID)
		logReq := &service.CreateAuditLogRequest{
			AdminID:      adminUUID,
			AdminName:    adminUsername,
			Action:       "VIEW_ORDER_DETAIL",
			Resource:     "order",
			ResourceID:   orderNo,
			Method:       "GET",
			Path:         "/api/v1/admin/orders/" + orderNo,
			IP:           c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			Description:  reason,
			ResponseCode: statusCode,
		}
		_ = h.auditLogService.CreateLog(c.Request.Context(), logReq)
	}()

	// 6. 返回响应
	c.JSON(statusCode, result)
}

// GetMerchantOrders 获取指定商户的订单列表
func (h *OrderBFFHandler) GetMerchantOrders(c *gin.Context) {
	// 1. 获取管理员信息
	adminID := c.GetString("user_id")
	adminUsername := c.GetString("username")

	// 2. 必须提供查看原因
	reason := c.Query("reason")
	if reason == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "必须提供查看原因 (reason参数)",
		})
		return
	}

	// 3. 获取商户ID
	merchantID := c.Param("merchant_id")

	// 4. 构建查询参数
	queryParams := make(map[string]string)
	queryParams["merchant_id"] = merchantID
	if page := c.Query("page"); page != "" {
		queryParams["page"] = page
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		queryParams["page_size"] = pageSize
	}
	if status := c.Query("status"); status != "" {
		queryParams["status"] = status
	}

	// 5. 调用 Order Service
	result, statusCode, err := h.orderClient.Get(c.Request.Context(), "/api/v1/orders", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Order Service失败", "details": err.Error()})
		return
	}

	// 6. 记录审计日志 (异步)
	go func() {
		adminUUID, _ := uuid.Parse(adminID)
		logReq := &service.CreateAuditLogRequest{
			AdminID:      adminUUID,
			AdminName:    adminUsername,
			Action:       "VIEW_MERCHANT_ORDERS",
			Resource:     "order",
			ResourceID:   merchantID,
			Method:       "GET",
			Path:         "/api/v1/admin/orders/merchant/" + merchantID,
			IP:           c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			Description:  reason,
			ResponseCode: statusCode,
		}
		_ = h.auditLogService.CreateLog(c.Request.Context(), logReq)
	}()

	// 7. 返回响应
	c.JSON(statusCode, result)
}

// GetOrderStatistics 获取订单统计
func (h *OrderBFFHandler) GetOrderStatistics(c *gin.Context) {
	// 1. 构建查询参数
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

	// 2. 调用 Order Service
	result, statusCode, err := h.orderClient.Get(c.Request.Context(), "/api/v1/orders/statistics", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Order Service失败", "details": err.Error()})
		return
	}

	// 3. 返回响应 (统计数据不需要审计日志)
	c.JSON(statusCode, result)
}

// GetOrderStatusSummary 获取订单状态汇总
func (h *OrderBFFHandler) GetOrderStatusSummary(c *gin.Context) {
	// 1. 构建查询参数
	queryParams := make(map[string]string)
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}

	// 2. 调用 Order Service
	result, statusCode, err := h.orderClient.Get(c.Request.Context(), "/api/v1/orders/status-summary", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Order Service失败", "details": err.Error()})
		return
	}

	// 3. 返回响应 (统计数据不需要审计日志)
	c.JSON(statusCode, result)
}
