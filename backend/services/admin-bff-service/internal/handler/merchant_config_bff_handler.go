package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"payment-platform/admin-service/internal/client"
)

type MerchantConfigBFFHandler struct {
	configClient *client.ServiceClient
}

func NewMerchantConfigBFFHandler(merchantConfigServiceURL string) *MerchantConfigBFFHandler {
	return &MerchantConfigBFFHandler{
		configClient: client.NewServiceClient(merchantConfigServiceURL),
	}
}

func (h *MerchantConfigBFFHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	admin := r.Group("/admin/merchant-config")
	admin.Use(authMiddleware)
	{
		// 费率配置
		feeConfigs := admin.Group("/fee-configs")
		{
			feeConfigs.GET("", h.ListFeeConfigs)
			feeConfigs.GET("/:id", h.GetFeeConfig)
			feeConfigs.POST("", h.CreateFeeConfig)
			feeConfigs.PUT("/:id", h.UpdateFeeConfig)
			feeConfigs.DELETE("/:id", h.DeleteFeeConfig)
			feeConfigs.POST("/:id/activate", h.ActivateFeeConfig)
			feeConfigs.POST("/:id/deactivate", h.DeactivateFeeConfig)
		}

		// 限额配置
		limitConfigs := admin.Group("/limit-configs")
		{
			limitConfigs.GET("", h.ListLimitConfigs)
			limitConfigs.GET("/:id", h.GetLimitConfig)
			limitConfigs.POST("", h.CreateLimitConfig)
			limitConfigs.PUT("/:id", h.UpdateLimitConfig)
			limitConfigs.DELETE("/:id", h.DeleteLimitConfig)
			limitConfigs.POST("/:id/activate", h.ActivateLimitConfig)
			limitConfigs.POST("/:id/deactivate", h.DeactivateLimitConfig)
		}

		// 批量操作
		admin.POST("/fee-configs/batch", h.BatchCreateFeeConfigs)
		admin.POST("/limit-configs/batch", h.BatchCreateLimitConfigs)

		// 配置模板
		admin.GET("/templates/fee", h.ListFeeTemplates)
		admin.GET("/templates/limit", h.ListLimitTemplates)
	}
}

// ========== 费率配置 ==========

func (h *MerchantConfigBFFHandler) ListFeeConfigs(c *gin.Context) {
	queryParams := make(map[string]string)
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}
	if channel := c.Query("channel"); channel != "" {
		queryParams["channel"] = channel
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

	result, statusCode, err := h.configClient.Get(c.Request.Context(), "/api/v1/fee-configs", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Config Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *MerchantConfigBFFHandler) GetFeeConfig(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.configClient.Get(c.Request.Context(), "/api/v1/fee-configs/"+id, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Config Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *MerchantConfigBFFHandler) CreateFeeConfig(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	adminID := c.GetString("user_id")
	req["created_by"] = adminID

	result, statusCode, err := h.configClient.Post(c.Request.Context(), "/api/v1/fee-configs", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Config Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *MerchantConfigBFFHandler) UpdateFeeConfig(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	adminID := c.GetString("user_id")
	req["updated_by"] = adminID

	result, statusCode, err := h.configClient.Put(c.Request.Context(), "/api/v1/fee-configs/"+id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Config Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *MerchantConfigBFFHandler) DeleteFeeConfig(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.configClient.Delete(c.Request.Context(), "/api/v1/fee-configs/"+id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Config Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *MerchantConfigBFFHandler) ActivateFeeConfig(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		req = make(map[string]interface{})
	}

	adminID := c.GetString("user_id")
	req["activated_by"] = adminID

	result, statusCode, err := h.configClient.Post(c.Request.Context(), "/api/v1/fee-configs/"+id+"/activate", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Config Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *MerchantConfigBFFHandler) DeactivateFeeConfig(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		req = make(map[string]interface{})
	}

	adminID := c.GetString("user_id")
	req["deactivated_by"] = adminID

	result, statusCode, err := h.configClient.Post(c.Request.Context(), "/api/v1/fee-configs/"+id+"/deactivate", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Config Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== 限额配置 ==========

func (h *MerchantConfigBFFHandler) ListLimitConfigs(c *gin.Context) {
	queryParams := make(map[string]string)
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}
	if channel := c.Query("channel"); channel != "" {
		queryParams["channel"] = channel
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

	result, statusCode, err := h.configClient.Get(c.Request.Context(), "/api/v1/limit-configs", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Config Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *MerchantConfigBFFHandler) GetLimitConfig(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.configClient.Get(c.Request.Context(), "/api/v1/limit-configs/"+id, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Config Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *MerchantConfigBFFHandler) CreateLimitConfig(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	adminID := c.GetString("user_id")
	req["created_by"] = adminID

	result, statusCode, err := h.configClient.Post(c.Request.Context(), "/api/v1/limit-configs", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Config Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *MerchantConfigBFFHandler) UpdateLimitConfig(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	adminID := c.GetString("user_id")
	req["updated_by"] = adminID

	result, statusCode, err := h.configClient.Put(c.Request.Context(), "/api/v1/limit-configs/"+id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Config Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *MerchantConfigBFFHandler) DeleteLimitConfig(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.configClient.Delete(c.Request.Context(), "/api/v1/limit-configs/"+id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Config Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *MerchantConfigBFFHandler) ActivateLimitConfig(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		req = make(map[string]interface{})
	}

	adminID := c.GetString("user_id")
	req["activated_by"] = adminID

	result, statusCode, err := h.configClient.Post(c.Request.Context(), "/api/v1/limit-configs/"+id+"/activate", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Config Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *MerchantConfigBFFHandler) DeactivateLimitConfig(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		req = make(map[string]interface{})
	}

	adminID := c.GetString("user_id")
	req["deactivated_by"] = adminID

	result, statusCode, err := h.configClient.Post(c.Request.Context(), "/api/v1/limit-configs/"+id+"/deactivate", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Config Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== 批量操作 ==========

func (h *MerchantConfigBFFHandler) BatchCreateFeeConfigs(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	adminID := c.GetString("user_id")
	req["created_by"] = adminID

	result, statusCode, err := h.configClient.Post(c.Request.Context(), "/api/v1/fee-configs/batch", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Config Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *MerchantConfigBFFHandler) BatchCreateLimitConfigs(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	adminID := c.GetString("user_id")
	req["created_by"] = adminID

	result, statusCode, err := h.configClient.Post(c.Request.Context(), "/api/v1/limit-configs/batch", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Config Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== 配置模板 ==========

func (h *MerchantConfigBFFHandler) ListFeeTemplates(c *gin.Context) {
	queryParams := make(map[string]string)
	if industry := c.Query("industry"); industry != "" {
		queryParams["industry"] = industry
	}

	result, statusCode, err := h.configClient.Get(c.Request.Context(), "/api/v1/templates/fee", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Config Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *MerchantConfigBFFHandler) ListLimitTemplates(c *gin.Context) {
	queryParams := make(map[string]string)
	if tier := c.Query("tier"); tier != "" {
		queryParams["tier"] = tier
	}

	result, statusCode, err := h.configClient.Get(c.Request.Context(), "/api/v1/templates/limit", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Config Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}
