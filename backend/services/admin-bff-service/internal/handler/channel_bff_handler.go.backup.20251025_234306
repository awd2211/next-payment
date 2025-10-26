package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"payment-platform/admin-service/internal/client"
)

// ChannelBFFHandler Channel Adapter BFF处理器
type ChannelBFFHandler struct {
	channelClient *client.ServiceClient
}

// NewChannelBFFHandler 创建Channel BFF处理器
func NewChannelBFFHandler(channelServiceURL string) *ChannelBFFHandler {
	return &ChannelBFFHandler{
		channelClient: client.NewServiceClient(channelServiceURL),
	}
}

// RegisterRoutes 注册路由
func (h *ChannelBFFHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	admin := r.Group("/admin/channels")
	admin.Use(authMiddleware)
	{
		// 支付通道管理
		admin.POST("", h.CreateChannel)
		admin.GET("/:code", h.GetChannel)
		admin.GET("", h.ListChannels)
		admin.PUT("/:code", h.UpdateChannel)
		admin.DELETE("/:code", h.DeleteChannel)

		// 通道配置
		admin.GET("/:code/config", h.GetChannelConfig)
		admin.PUT("/:code/config", h.UpdateChannelConfig)

		// 通道状态管理
		admin.POST("/:code/enable", h.EnableChannel)
		admin.POST("/:code/disable", h.DisableChannel)

		// 汇率管理
		admin.GET("/exchange-rates", h.GetExchangeRates)
		admin.POST("/exchange-rates/update", h.UpdateExchangeRates)
	}
}

// ========== 支付通道管理 ==========

// CreateChannel 创建支付通道
func (h *ChannelBFFHandler) CreateChannel(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	result, statusCode, err := h.channelClient.Post(c.Request.Context(), "/api/v1/admin/channels", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Channel Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// GetChannel 获取支付通道详情
func (h *ChannelBFFHandler) GetChannel(c *gin.Context) {
	code := c.Param("code")

	result, statusCode, err := h.channelClient.Get(c.Request.Context(), "/api/v1/admin/channels/"+code, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Channel Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ListChannels 获取支付通道列表
func (h *ChannelBFFHandler) ListChannels(c *gin.Context) {
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
	if channelType := c.Query("type"); channelType != "" {
		queryParams["type"] = channelType
	}

	result, statusCode, err := h.channelClient.Get(c.Request.Context(), "/api/v1/admin/channels", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Channel Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// UpdateChannel 更新支付通道
func (h *ChannelBFFHandler) UpdateChannel(c *gin.Context) {
	code := c.Param("code")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	result, statusCode, err := h.channelClient.Put(c.Request.Context(), "/api/v1/admin/channels/"+code, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Channel Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// DeleteChannel 删除支付通道
func (h *ChannelBFFHandler) DeleteChannel(c *gin.Context) {
	code := c.Param("code")

	result, statusCode, err := h.channelClient.Delete(c.Request.Context(), "/api/v1/admin/channels/"+code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Channel Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== 通道配置管理 ==========

// GetChannelConfig 获取通道配置
func (h *ChannelBFFHandler) GetChannelConfig(c *gin.Context) {
	code := c.Param("code")

	result, statusCode, err := h.channelClient.Get(c.Request.Context(), "/api/v1/channel/config/"+code, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Channel Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// UpdateChannelConfig 更新通道配置
func (h *ChannelBFFHandler) UpdateChannelConfig(c *gin.Context) {
	code := c.Param("code")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	result, statusCode, err := h.channelClient.Put(c.Request.Context(), "/api/v1/channel/config/"+code, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Channel Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== 通道状态管理 ==========

// EnableChannel 启用支付通道
func (h *ChannelBFFHandler) EnableChannel(c *gin.Context) {
	code := c.Param("code")

	result, statusCode, err := h.channelClient.Post(c.Request.Context(), "/api/v1/admin/channels/"+code+"/enable", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Channel Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// DisableChannel 禁用支付通道
func (h *ChannelBFFHandler) DisableChannel(c *gin.Context) {
	code := c.Param("code")

	result, statusCode, err := h.channelClient.Post(c.Request.Context(), "/api/v1/admin/channels/"+code+"/disable", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Channel Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== 汇率管理 ==========

// GetExchangeRates 获取汇率列表
func (h *ChannelBFFHandler) GetExchangeRates(c *gin.Context) {
	queryParams := make(map[string]string)
	if baseCurrency := c.Query("base_currency"); baseCurrency != "" {
		queryParams["base_currency"] = baseCurrency
	}

	result, statusCode, err := h.channelClient.Get(c.Request.Context(), "/api/v1/channel/exchange-rates", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Channel Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// UpdateExchangeRates 更新汇率
func (h *ChannelBFFHandler) UpdateExchangeRates(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	result, statusCode, err := h.channelClient.Post(c.Request.Context(), "/api/v1/channel/exchange-rates/update", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Channel Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}
