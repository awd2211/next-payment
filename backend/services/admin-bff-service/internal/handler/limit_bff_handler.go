package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"payment-platform/admin-service/internal/client"
)

// LimitBFFHandler Merchant Limit Service BFF处理器
type LimitBFFHandler struct {
	limitClient *client.ServiceClient
}

// NewLimitBFFHandler 创建Limit BFF处理器
func NewLimitBFFHandler(limitServiceURL string) *LimitBFFHandler {
	return &LimitBFFHandler{
		limitClient: client.NewServiceClient(limitServiceURL),
	}
}

// RegisterRoutes 注册路由
func (h *LimitBFFHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	admin := r.Group("/admin")
	admin.Use(authMiddleware)
	{
		// Tier 管理
		tiers := admin.Group("/merchant-tiers")
		{
			tiers.POST("", h.CreateTier)
			tiers.GET("/:id", h.GetTier)
			tiers.GET("", h.ListTiers)
			tiers.PUT("/:id", h.UpdateTier)
			tiers.DELETE("/:id", h.DeleteTier)
		}

		// Limit 管理
		limits := admin.Group("/merchant-limits")
		{
			limits.GET("", h.ListAllMerchantLimits)
			limits.GET("/:merchant_id", h.GetMerchantLimit)
			limits.POST("/:merchant_id/adjust", h.AdjustLimit)
			limits.POST("/:merchant_id/change-tier", h.ChangeTier)
			limits.GET("/:merchant_id/usage-history", h.GetUsageHistory)
		}
	}
}

// ========== Tier 管理 ==========

// CreateTier 创建层级
func (h *LimitBFFHandler) CreateTier(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	result, statusCode, err := h.limitClient.Post(c.Request.Context(), "/api/v1/tiers", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Limit Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// GetTier 获取层级详情
func (h *LimitBFFHandler) GetTier(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.limitClient.Get(c.Request.Context(), "/api/v1/tiers/"+id, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Limit Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ListTiers 获取层级列表
func (h *LimitBFFHandler) ListTiers(c *gin.Context) {
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

	result, statusCode, err := h.limitClient.Get(c.Request.Context(), "/api/v1/tiers", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Limit Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// UpdateTier 更新层级
func (h *LimitBFFHandler) UpdateTier(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	result, statusCode, err := h.limitClient.Put(c.Request.Context(), "/api/v1/tiers/"+id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Limit Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// DeleteTier 删除层级
func (h *LimitBFFHandler) DeleteTier(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.limitClient.Delete(c.Request.Context(), "/api/v1/tiers/"+id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Limit Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== Limit 管理 ==========

// ListAllMerchantLimits 获取所有商户限额
func (h *LimitBFFHandler) ListAllMerchantLimits(c *gin.Context) {
	queryParams := make(map[string]string)
	if page := c.Query("page"); page != "" {
		queryParams["page"] = page
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		queryParams["page_size"] = pageSize
	}
	if tierID := c.Query("tier_id"); tierID != "" {
		queryParams["tier_id"] = tierID
	}

	result, statusCode, err := h.limitClient.Get(c.Request.Context(), "/api/v1/limits", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Limit Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// GetMerchantLimit 获取商户限额
func (h *LimitBFFHandler) GetMerchantLimit(c *gin.Context) {
	merchantID := c.Param("merchant_id")

	result, statusCode, err := h.limitClient.Get(c.Request.Context(), "/api/v1/limits/"+merchantID, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Limit Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// AdjustLimit 调整商户限额
func (h *LimitBFFHandler) AdjustLimit(c *gin.Context) {
	merchantID := c.Param("merchant_id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 添加操作人信息
	adminID := c.GetString("user_id")
	req["operator_id"] = adminID

	result, statusCode, err := h.limitClient.Post(c.Request.Context(), "/api/v1/limits/"+merchantID+"/adjust", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Limit Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ChangeTier 更改商户层级
func (h *LimitBFFHandler) ChangeTier(c *gin.Context) {
	merchantID := c.Param("merchant_id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 添加操作人信息
	adminID := c.GetString("user_id")
	req["operator_id"] = adminID

	result, statusCode, err := h.limitClient.Post(c.Request.Context(), "/api/v1/limits/"+merchantID+"/change-tier", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Limit Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// GetUsageHistory 获取使用历史
func (h *LimitBFFHandler) GetUsageHistory(c *gin.Context) {
	merchantID := c.Param("merchant_id")

	queryParams := make(map[string]string)
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

	result, statusCode, err := h.limitClient.Get(c.Request.Context(), "/api/v1/limits/"+merchantID+"/usage-history", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Limit Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}
