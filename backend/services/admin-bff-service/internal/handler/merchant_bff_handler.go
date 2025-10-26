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

// MerchantBFFHandler Merchant Service BFF处理器
type MerchantBFFHandler struct {
	merchantClient  *client.ServiceClient
	auditLogService service.AuditLogService
	auditHelper     *utils.AuditHelper
}

// NewMerchantBFFHandler 创建Merchant BFF处理器
func NewMerchantBFFHandler(merchantServiceURL string, auditLogService service.AuditLogService) *MerchantBFFHandler {
	return &MerchantBFFHandler{
		merchantClient:  client.NewServiceClient(merchantServiceURL),
		auditLogService: auditLogService,
		auditHelper:     utils.NewAuditHelper(auditLogService),
	}
}

// RegisterRoutes 注册路由
func (h *MerchantBFFHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	admin := r.Group("/admin/merchants")
	admin.Use(authMiddleware)
	{
		// 商户管理
		admin.POST("",
			localMiddleware.RequirePermission("merchants.create"),
			h.CreateMerchant,
		)
		admin.GET("/:id",
			localMiddleware.RequirePermission("merchants.view"),
			localMiddleware.RequireReason,
			h.GetMerchant,
		)
		admin.GET("",
			localMiddleware.RequirePermission("merchants.view"),
			localMiddleware.RequireReason,
			h.ListMerchants,
		)
		admin.PUT("/:id",
			localMiddleware.RequirePermission("merchants.update"),
			h.UpdateMerchant,
		)
		admin.DELETE("/:id",
			localMiddleware.RequirePermission("merchants.delete"),
			localMiddleware.RequireReason,
			h.DeleteMerchant,
		)

		// 商户状态管理 (敏感操作需要reason)
		admin.POST("/:id/approve",
			localMiddleware.RequirePermission("merchants.approve"),
			localMiddleware.RequireReason,
			h.ApproveMerchant,
		)
		admin.POST("/:id/reject",
			localMiddleware.RequirePermission("merchants.approve"),
			localMiddleware.RequireReason,
			h.RejectMerchant,
		)
		admin.POST("/:id/freeze",
			localMiddleware.RequirePermission("merchants.freeze"),
			localMiddleware.RequireReason,
			h.FreezeMerchant,
		)
		admin.POST("/:id/unfreeze",
			localMiddleware.RequirePermission("merchants.freeze"),
			localMiddleware.RequireReason,
			h.UnfreezeMerchant,
		)

		// 商户统计 (不需要reason)
		admin.GET("/statistics",
			localMiddleware.RequirePermission("merchants.view"),
			h.GetStatistics,
		)
	}
}

// ========== 商户管理 ==========

// CreateMerchant 创建商户
func (h *MerchantBFFHandler) CreateMerchant(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	result, statusCode, err := h.merchantClient.Post(c.Request.Context(), "/api/v1/merchant", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// GetMerchant 获取商户详情
func (h *MerchantBFFHandler) GetMerchant(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.merchantClient.Get(c.Request.Context(), "/api/v1/merchant/"+id, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Service失败", "details": err.Error()})
		return
	}

	// 数据脱敏
	if data, ok := result["data"].(map[string]interface{}); ok {
		result["data"] = utils.MaskSensitiveData(data)
	}

	// 记录审计日志
	h.auditHelper.LogCrossTenantAccess(c, "VIEW_MERCHANT_DETAIL", "merchant", id, id, statusCode)

	c.JSON(statusCode, result)
}

// ListMerchants 获取商户列表
func (h *MerchantBFFHandler) ListMerchants(c *gin.Context) {
	adminID := c.GetString("user_id")
	adminUsername := c.GetString("username")
	reason := c.GetString("operation_reason")

	queryParams := make(map[string]string)
	if page := c.Query("page"); page != "" {
		queryParams["page"] = page
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		queryParams["page_size"] = pageSize
	}
	if status := c.Query("status"); status != "" {
		queryParams["status"] = status
	}
	if kycStatus := c.Query("kyc_status"); kycStatus != "" {
		queryParams["kyc_status"] = kycStatus
	}
	if businessType := c.Query("business_type"); businessType != "" {
		queryParams["business_type"] = businessType
	}
	if search := c.Query("search"); search != "" {
		queryParams["search"] = search
	}

	result, statusCode, err := h.merchantClient.Get(c.Request.Context(), "/api/v1/merchant", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Service失败", "details": err.Error()})
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
			Action:       "VIEW_MERCHANTS",
			Resource:     "merchant",
			Method:       "GET",
			Path:         "/api/v1/admin/merchants",
			IP:           c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			Description:  reason,
			ResponseCode: statusCode,
		}
		_ = h.auditLogService.CreateLog(c.Request.Context(), logReq)
	}()

	c.JSON(statusCode, result)
}

// UpdateMerchant 更新商户信息
func (h *MerchantBFFHandler) UpdateMerchant(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	result, statusCode, err := h.merchantClient.Put(c.Request.Context(), "/api/v1/merchant/"+id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// DeleteMerchant 删除商户
func (h *MerchantBFFHandler) DeleteMerchant(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.merchantClient.Delete(c.Request.Context(), "/api/v1/merchant/"+id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== 商户状态管理 ==========

// ApproveMerchant 批准商户
func (h *MerchantBFFHandler) ApproveMerchant(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		req = make(map[string]interface{})
	}

	// 添加审核人信息
	adminID := c.GetString("user_id")
	req["approved_by"] = adminID

	result, statusCode, err := h.merchantClient.Post(c.Request.Context(), "/api/v1/merchant/"+id+"/approve", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// RejectMerchant 拒绝商户
func (h *MerchantBFFHandler) RejectMerchant(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 添加审核人信息
	adminID := c.GetString("user_id")
	req["rejected_by"] = adminID

	result, statusCode, err := h.merchantClient.Post(c.Request.Context(), "/api/v1/merchant/"+id+"/reject", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// FreezeMerchant 冻结商户
func (h *MerchantBFFHandler) FreezeMerchant(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 添加操作人信息
	adminID := c.GetString("user_id")
	req["operator_id"] = adminID

	result, statusCode, err := h.merchantClient.Post(c.Request.Context(), "/api/v1/merchant/"+id+"/freeze", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// UnfreezeMerchant 解冻商户
func (h *MerchantBFFHandler) UnfreezeMerchant(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		req = make(map[string]interface{})
	}

	// 添加操作人信息
	adminID := c.GetString("user_id")
	req["operator_id"] = adminID

	result, statusCode, err := h.merchantClient.Post(c.Request.Context(), "/api/v1/merchant/"+id+"/unfreeze", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// GetStatistics 获取商户统计
func (h *MerchantBFFHandler) GetStatistics(c *gin.Context) {
	queryParams := make(map[string]string)
	if period := c.Query("period"); period != "" {
		queryParams["period"] = period
	}

	result, statusCode, err := h.merchantClient.Get(c.Request.Context(), "/api/v1/merchant/statistics", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}
