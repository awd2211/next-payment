package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"payment-platform/admin-service/internal/client"
)

// AnalyticsBFFHandler Analytics Service BFF处理器
type AnalyticsBFFHandler struct {
	analyticsClient *client.ServiceClient
}

// NewAnalyticsBFFHandler 创建Analytics BFF处理器
func NewAnalyticsBFFHandler(analyticsServiceURL string) *AnalyticsBFFHandler {
	return &AnalyticsBFFHandler{
		analyticsClient: client.NewServiceClient(analyticsServiceURL),
	}
}

// RegisterRoutes 注册路由
func (h *AnalyticsBFFHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	admin := r.Group("/admin/analytics")
	admin.Use(authMiddleware)
	{
		// 平台总览
		platform := admin.Group("/platform")
		{
			platform.GET("/overview", h.GetPlatformOverview)
			platform.GET("/trends", h.GetPlatformTrends)
			platform.GET("/statistics", h.GetPlatformStatistics)
		}

		// Dashboard 数据聚合
		admin.GET("/dashboard", h.GetAdminDashboard)

		// 支付分析
		payments := admin.Group("/payments")
		{
			payments.GET("/statistics", h.GetPaymentStatistics)
			payments.GET("/trends", h.GetPaymentTrends)
			payments.GET("/channels", h.GetChannelAnalysis)
		}

		// 商户分析
		merchants := admin.Group("/merchants")
		{
			merchants.GET("/statistics", h.GetMerchantStatistics)
			merchants.GET("/trends", h.GetMerchantTrends)
			merchants.GET("/top", h.GetTopMerchants)
		}
	}
}

// ========== 平台总览 ==========

// GetPlatformOverview 获取平台总览
func (h *AnalyticsBFFHandler) GetPlatformOverview(c *gin.Context) {
	queryParams := make(map[string]string)
	if period := c.Query("period"); period != "" {
		queryParams["period"] = period
	}

	result, statusCode, err := h.analyticsClient.Get(c.Request.Context(), "/api/v1/analytics/platform/overview", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Analytics Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// GetPlatformTrends 获取平台趋势
func (h *AnalyticsBFFHandler) GetPlatformTrends(c *gin.Context) {
	queryParams := make(map[string]string)
	if period := c.Query("period"); period != "" {
		queryParams["period"] = period
	}
	if startTime := c.Query("start_time"); startTime != "" {
		queryParams["start_time"] = startTime
	}
	if endTime := c.Query("end_time"); endTime != "" {
		queryParams["end_time"] = endTime
	}

	result, statusCode, err := h.analyticsClient.Get(c.Request.Context(), "/api/v1/analytics/platform/trends", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Analytics Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// GetPlatformStatistics 获取平台统计
func (h *AnalyticsBFFHandler) GetPlatformStatistics(c *gin.Context) {
	queryParams := make(map[string]string)
	if period := c.Query("period"); period != "" {
		queryParams["period"] = period
	}

	result, statusCode, err := h.analyticsClient.Get(c.Request.Context(), "/api/v1/analytics/platform/statistics", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Analytics Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== Dashboard ==========

// GetAdminDashboard 获取管理员Dashboard数据
func (h *AnalyticsBFFHandler) GetAdminDashboard(c *gin.Context) {
	queryParams := make(map[string]string)
	period := c.DefaultQuery("period", "today")
	queryParams["period"] = period

	result, statusCode, err := h.analyticsClient.Get(c.Request.Context(), "/api/v1/analytics/dashboard", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Analytics Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== 支付分析 ==========

// GetPaymentStatistics 获取支付统计
func (h *AnalyticsBFFHandler) GetPaymentStatistics(c *gin.Context) {
	queryParams := make(map[string]string)
	if period := c.Query("period"); period != "" {
		queryParams["period"] = period
	}
	if startTime := c.Query("start_time"); startTime != "" {
		queryParams["start_time"] = startTime
	}
	if endTime := c.Query("end_time"); endTime != "" {
		queryParams["end_time"] = endTime
	}
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}

	result, statusCode, err := h.analyticsClient.Get(c.Request.Context(), "/api/v1/analytics/payments/statistics", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Analytics Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// GetPaymentTrends 获取支付趋势
func (h *AnalyticsBFFHandler) GetPaymentTrends(c *gin.Context) {
	queryParams := make(map[string]string)
	if period := c.Query("period"); period != "" {
		queryParams["period"] = period
	}
	if startTime := c.Query("start_time"); startTime != "" {
		queryParams["start_time"] = startTime
	}
	if endTime := c.Query("end_time"); endTime != "" {
		queryParams["end_time"] = endTime
	}
	if groupBy := c.Query("group_by"); groupBy != "" {
		queryParams["group_by"] = groupBy
	}

	result, statusCode, err := h.analyticsClient.Get(c.Request.Context(), "/api/v1/analytics/payments/trends", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Analytics Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// GetChannelAnalysis 获取渠道分析
func (h *AnalyticsBFFHandler) GetChannelAnalysis(c *gin.Context) {
	queryParams := make(map[string]string)
	if period := c.Query("period"); period != "" {
		queryParams["period"] = period
	}
	if startTime := c.Query("start_time"); startTime != "" {
		queryParams["start_time"] = startTime
	}
	if endTime := c.Query("end_time"); endTime != "" {
		queryParams["end_time"] = endTime
	}

	result, statusCode, err := h.analyticsClient.Get(c.Request.Context(), "/api/v1/analytics/payments/channels", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Analytics Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== 商户分析 ==========

// GetMerchantStatistics 获取商户统计
func (h *AnalyticsBFFHandler) GetMerchantStatistics(c *gin.Context) {
	queryParams := make(map[string]string)
	if period := c.Query("period"); period != "" {
		queryParams["period"] = period
	}
	if startTime := c.Query("start_time"); startTime != "" {
		queryParams["start_time"] = startTime
	}
	if endTime := c.Query("end_time"); endTime != "" {
		queryParams["end_time"] = endTime
	}

	result, statusCode, err := h.analyticsClient.Get(c.Request.Context(), "/api/v1/analytics/merchants/statistics", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Analytics Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// GetMerchantTrends 获取商户趋势
func (h *AnalyticsBFFHandler) GetMerchantTrends(c *gin.Context) {
	queryParams := make(map[string]string)
	if period := c.Query("period"); period != "" {
		queryParams["period"] = period
	}
	if startTime := c.Query("start_time"); startTime != "" {
		queryParams["start_time"] = startTime
	}
	if endTime := c.Query("end_time"); endTime != "" {
		queryParams["end_time"] = endTime
	}

	result, statusCode, err := h.analyticsClient.Get(c.Request.Context(), "/api/v1/analytics/merchants/trends", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Analytics Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// GetTopMerchants 获取Top商户
func (h *AnalyticsBFFHandler) GetTopMerchants(c *gin.Context) {
	queryParams := make(map[string]string)
	if period := c.Query("period"); period != "" {
		queryParams["period"] = period
	}
	if limit := c.Query("limit"); limit != "" {
		queryParams["limit"] = limit
	}
	if orderBy := c.Query("order_by"); orderBy != "" {
		queryParams["order_by"] = orderBy
	}

	result, statusCode, err := h.analyticsClient.Get(c.Request.Context(), "/api/v1/analytics/merchants/top", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Analytics Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}
