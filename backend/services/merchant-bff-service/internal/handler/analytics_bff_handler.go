package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"payment-platform/merchant-bff-service/internal/client"
)

type AnalyticsBFFHandler struct {
	analyticsClient *client.ServiceClient
}

func NewAnalyticsBFFHandler(analyticsServiceURL string) *AnalyticsBFFHandler {
	return &AnalyticsBFFHandler{
		analyticsClient: client.NewServiceClient(analyticsServiceURL),
	}
}

func (h *AnalyticsBFFHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	merchant := r.Group("/merchant/analytics")
	merchant.Use(authMiddleware)
	{
		merchant.GET("/dashboard", h.GetDashboard)
		merchant.GET("/payments", h.GetPaymentTrends)
		merchant.GET("/revenue", h.GetRevenueAnalysis)
	}
}

func (h *AnalyticsBFFHandler) GetDashboard(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	queryParams := map[string]string{
		"merchant_id": merchantID,
		"time_range":  c.DefaultQuery("time_range", "30d"),
	}

	result, statusCode, err := h.analyticsClient.Get(c.Request.Context(), "/api/v1/analytics/dashboard", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *AnalyticsBFFHandler) GetPaymentTrends(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	queryParams := map[string]string{
		"merchant_id": merchantID,
		"start_time":  c.Query("start_time"),
		"end_time":    c.Query("end_time"),
		"granularity": c.DefaultQuery("granularity", "day"),
	}

	result, statusCode, err := h.analyticsClient.Get(c.Request.Context(), "/api/v1/analytics/payments/trends", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *AnalyticsBFFHandler) GetRevenueAnalysis(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	queryParams := map[string]string{
		"merchant_id": merchantID,
		"start_time":  c.Query("start_time"),
		"end_time":    c.Query("end_time"),
		"group_by":    c.DefaultQuery("group_by", "channel"),
	}

	result, statusCode, err := h.analyticsClient.Get(c.Request.Context(), "/api/v1/analytics/revenue", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}
