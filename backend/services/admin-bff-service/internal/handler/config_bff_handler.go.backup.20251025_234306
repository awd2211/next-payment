package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"payment-platform/admin-service/internal/client"
)

// ConfigBFFHandler Config Service BFF处理器
type ConfigBFFHandler struct {
	configClient *client.ServiceClient
}

// NewConfigBFFHandler 创建Config BFF处理器
func NewConfigBFFHandler(configServiceURL string) *ConfigBFFHandler {
	return &ConfigBFFHandler{
		configClient: client.NewServiceClient(configServiceURL),
	}
}

// RegisterRoutes 注册路由
func (h *ConfigBFFHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	admin := r.Group("/admin")
	admin.Use(authMiddleware)
	{
		// 系统配置管理
		configs := admin.Group("/configs")
		{
			configs.POST("", h.CreateConfig)
			configs.GET("/:key", h.GetConfig)
			configs.GET("", h.ListConfigs)
			configs.PUT("/:key", h.UpdateConfig)
			configs.DELETE("/:key", h.DeleteConfig)
		}

		// 功能开关管理
		flags := admin.Group("/feature-flags")
		{
			flags.POST("", h.CreateFeatureFlag)
			flags.GET("/:name", h.GetFeatureFlag)
			flags.GET("", h.ListFeatureFlags)
			flags.PUT("/:name", h.UpdateFeatureFlag)
			flags.DELETE("/:name", h.DeleteFeatureFlag)
			flags.POST("/:name/toggle", h.ToggleFeatureFlag)
		}

		// 服务注册管理
		services := admin.Group("/services")
		{
			services.POST("", h.RegisterService)
			services.GET("/:name", h.GetService)
			services.GET("", h.ListServices)
			services.PUT("/:name", h.UpdateService)
			services.DELETE("/:name", h.UnregisterService)
		}
	}
}

// ========== 系统配置管理 ==========

// CreateConfig 创建配置
func (h *ConfigBFFHandler) CreateConfig(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	result, statusCode, err := h.configClient.Post(c.Request.Context(), "/api/v1/configs", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Config Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// GetConfig 获取配置
func (h *ConfigBFFHandler) GetConfig(c *gin.Context) {
	key := c.Param("key")

	result, statusCode, err := h.configClient.Get(c.Request.Context(), "/api/v1/configs/"+key, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Config Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ListConfigs 获取配置列表
func (h *ConfigBFFHandler) ListConfigs(c *gin.Context) {
	// 获取查询参数
	queryParams := make(map[string]string)
	if page := c.Query("page"); page != "" {
		queryParams["page"] = page
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		queryParams["page_size"] = pageSize
	}
	if category := c.Query("category"); category != "" {
		queryParams["category"] = category
	}

	result, statusCode, err := h.configClient.Get(c.Request.Context(), "/api/v1/configs", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Config Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// UpdateConfig 更新配置
func (h *ConfigBFFHandler) UpdateConfig(c *gin.Context) {
	key := c.Param("key")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	result, statusCode, err := h.configClient.Put(c.Request.Context(), "/api/v1/configs/"+key, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Config Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// DeleteConfig 删除配置
func (h *ConfigBFFHandler) DeleteConfig(c *gin.Context) {
	key := c.Param("key")

	result, statusCode, err := h.configClient.Delete(c.Request.Context(), "/api/v1/configs/"+key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Config Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== 功能开关管理 ==========

// CreateFeatureFlag 创建功能开关
func (h *ConfigBFFHandler) CreateFeatureFlag(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	result, statusCode, err := h.configClient.Post(c.Request.Context(), "/api/v1/feature-flags", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Config Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// GetFeatureFlag 获取功能开关
func (h *ConfigBFFHandler) GetFeatureFlag(c *gin.Context) {
	name := c.Param("name")

	result, statusCode, err := h.configClient.Get(c.Request.Context(), "/api/v1/feature-flags/"+name, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Config Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ListFeatureFlags 获取功能开关列表
func (h *ConfigBFFHandler) ListFeatureFlags(c *gin.Context) {
	// 获取查询参数
	queryParams := make(map[string]string)
	if page := c.Query("page"); page != "" {
		queryParams["page"] = page
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		queryParams["page_size"] = pageSize
	}

	result, statusCode, err := h.configClient.Get(c.Request.Context(), "/api/v1/feature-flags", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Config Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// UpdateFeatureFlag 更新功能开关
func (h *ConfigBFFHandler) UpdateFeatureFlag(c *gin.Context) {
	name := c.Param("name")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	result, statusCode, err := h.configClient.Put(c.Request.Context(), "/api/v1/feature-flags/"+name, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Config Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// DeleteFeatureFlag 删除功能开关
func (h *ConfigBFFHandler) DeleteFeatureFlag(c *gin.Context) {
	name := c.Param("name")

	result, statusCode, err := h.configClient.Delete(c.Request.Context(), "/api/v1/feature-flags/"+name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Config Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ToggleFeatureFlag 切换功能开关
func (h *ConfigBFFHandler) ToggleFeatureFlag(c *gin.Context) {
	name := c.Param("name")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	result, statusCode, err := h.configClient.Post(c.Request.Context(), "/api/v1/feature-flags/"+name+"/toggle", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Config Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== 服务注册管理 ==========

// RegisterService 注册服务
func (h *ConfigBFFHandler) RegisterService(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	result, statusCode, err := h.configClient.Post(c.Request.Context(), "/api/v1/services", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Config Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// GetService 获取服务信息
func (h *ConfigBFFHandler) GetService(c *gin.Context) {
	name := c.Param("name")

	result, statusCode, err := h.configClient.Get(c.Request.Context(), "/api/v1/services/"+name, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Config Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ListServices 获取服务列表
func (h *ConfigBFFHandler) ListServices(c *gin.Context) {
	// 获取查询参数
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

	result, statusCode, err := h.configClient.Get(c.Request.Context(), "/api/v1/services", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Config Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// UpdateService 更新服务信息
func (h *ConfigBFFHandler) UpdateService(c *gin.Context) {
	name := c.Param("name")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	result, statusCode, err := h.configClient.Put(c.Request.Context(), "/api/v1/services/"+name, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Config Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// UnregisterService 注销服务
func (h *ConfigBFFHandler) UnregisterService(c *gin.Context) {
	name := c.Param("name")

	result, statusCode, err := h.configClient.Delete(c.Request.Context(), "/api/v1/services/"+name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Config Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}
