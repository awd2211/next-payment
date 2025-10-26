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

// OrderBFFHandlerSecure Order Service BFF处理器（安全增强版）
type OrderBFFHandlerSecure struct {
	orderClient     *client.ServiceClient
	auditLogService service.AuditLogService
	auditHelper     *utils.AuditHelper
}

// NewOrderBFFHandlerSecure 创建Order BFF处理器（安全版）
func NewOrderBFFHandlerSecure(orderServiceURL string, auditLogService service.AuditLogService) *OrderBFFHandlerSecure {
	return &OrderBFFHandlerSecure{
		orderClient:     client.NewServiceClient(orderServiceURL),
		auditLogService: auditLogService,
		auditHelper:     utils.NewAuditHelper(auditLogService),
	}
}

// RegisterRoutes 注册路由（带RBAC权限控制）
func (h *OrderBFFHandlerSecure) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	admin := r.Group("/admin/orders")
	admin.Use(authMiddleware)
	{
		// 订单查询 (需要权限 + 原因)
		admin.GET("",
			localMiddleware.RequirePermission("orders.view"),
			localMiddleware.RequireReason,
			h.ListOrders,
		)

		admin.GET("/:order_no",
			localMiddleware.RequirePermission("orders.view"),
			localMiddleware.RequireReason,
			h.GetOrder,
		)

		admin.GET("/merchant/:merchant_id",
			localMiddleware.RequirePermission("orders.view"),
			localMiddleware.RequireReason,
			h.GetMerchantOrders,
		)

		// 统计查询 (不需要reason,但需要权限)
		admin.GET("/statistics",
			localMiddleware.RequirePermission("orders.view"),
			h.GetOrderStatistics,
		)

		admin.GET("/status-summary",
			localMiddleware.RequirePermission("orders.view"),
			h.GetOrderStatusSummary,
		)
	}
}

// ListOrders 获取订单列表（安全增强版）
func (h *OrderBFFHandlerSecure) ListOrders(c *gin.Context) {
	// 1. 获取管理员信息
	adminID := c.GetString("user_id")
	adminUsername := c.GetString("username")
	reason := c.GetString("operation_reason") // 来自RequireReason中间件

	// 2. 构建查询参数
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

	// 3. 调用后端服务
	result, statusCode, err := h.orderClient.Get(c.Request.Context(), "/api/v1/orders", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Order Service失败", "details": err.Error()})
		return
	}

	// 4. 数据脱敏（保护用户隐私）
	if data, ok := result["data"].(map[string]interface{}); ok {
		result["data"] = utils.MaskSensitiveData(data)
	}

	// 5. 记录审计日志 (异步)
	go func() {
		adminUUID, _ := uuid.Parse(adminID)
		logReq := &service.CreateAuditLogRequest{
			AdminID:      adminUUID,
			AdminName:    adminUsername,
			Action:       "VIEW_ORDERS",
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

// GetOrder 获取订单详情（安全增强版）
func (h *OrderBFFHandlerSecure) GetOrder(c *gin.Context) {
	// 1. 获取管理员信息
	adminID := c.GetString("user_id")
	adminUsername := c.GetString("username")
	reason := c.GetString("operation_reason")

	// 2. 获取订单号
	orderNo := c.Param("order_no")

	// 3. 调用后端服务
	result, statusCode, err := h.orderClient.Get(c.Request.Context(), "/api/v1/orders/"+orderNo, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Order Service失败", "details": err.Error()})
		return
	}

	// 4. 数据脱敏
	if data, ok := result["data"].(map[string]interface{}); ok {
		result["data"] = utils.MaskSensitiveData(data)
	}

	// 5. 记录审计日志
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

// GetMerchantOrders 获取指定商户的订单列表（安全增强版）
func (h *OrderBFFHandlerSecure) GetMerchantOrders(c *gin.Context) {
	// 1. 获取管理员信息

	// 2. 获取商户ID
	merchantID := c.Param("merchant_id")

	// 3. 构建查询参数
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

	// 4. 调用后端服务
	result, statusCode, err := h.orderClient.Get(c.Request.Context(), "/api/v1/orders", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Order Service失败", "details": err.Error()})
		return
	}

	// 5. 数据脱敏
	if data, ok := result["data"].(map[string]interface{}); ok {
		result["data"] = utils.MaskSensitiveData(data)
	}

	// 6. 记录审计日志
	h.auditHelper.LogCrossTenantAccess(c, "VIEW_MERCHANT_ORDERS", "order", merchantID, merchantID, statusCode)

	// 7. 返回响应
	c.JSON(statusCode, result)
}

// GetOrderStatistics 获取订单统计（不需要reason）
func (h *OrderBFFHandlerSecure) GetOrderStatistics(c *gin.Context) {
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

	// 2. 调用后端服务
	result, statusCode, err := h.orderClient.Get(c.Request.Context(), "/api/v1/orders/statistics", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Order Service失败", "details": err.Error()})
		return
	}

	// 3. 返回响应 (统计数据不脱敏)
	c.JSON(statusCode, result)
}

// GetOrderStatusSummary 获取订单状态汇总
func (h *OrderBFFHandlerSecure) GetOrderStatusSummary(c *gin.Context) {
	// 1. 构建查询参数
	queryParams := make(map[string]string)
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}

	// 2. 调用后端服务
	result, statusCode, err := h.orderClient.Get(c.Request.Context(), "/api/v1/orders/status-summary", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Order Service失败", "details": err.Error()})
		return
	}

	// 3. 返回响应
	c.JSON(statusCode, result)
}
