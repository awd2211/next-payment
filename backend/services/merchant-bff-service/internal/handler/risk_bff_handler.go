package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"payment-platform/merchant-bff-service/internal/client"
)

type RiskBFFHandler struct {
	riskClient *client.ServiceClient
}

func NewRiskBFFHandler(riskServiceURL string) *RiskBFFHandler {
	return &RiskBFFHandler{
		riskClient: client.NewServiceClient(riskServiceURL),
	}
}

func (h *RiskBFFHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	risk := r.Group("/merchant/risk")
	risk.Use(authMiddleware)
	{
		risk.GET("/rules", h.GetRules)
		risk.GET("/alerts", h.GetAlerts)
		risk.GET("/statistics", h.GetStatistics)
	}
}

func (h *RiskBFFHandler) GetRules(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	queryParams := map[string]string{
		"merchant_id": merchantID,
		"page":        c.DefaultQuery("page", "1"),
		"page_size":   c.DefaultQuery("page_size", "10"),
	}

	result, statusCode, err := h.riskClient.Get(c.Request.Context(), "/api/v1/rules", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *RiskBFFHandler) GetAlerts(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	queryParams := map[string]string{
		"merchant_id": merchantID,
		"page":        c.DefaultQuery("page", "1"),
		"page_size":   c.DefaultQuery("page_size", "10"),
		"start_date":  c.DefaultQuery("start_date", ""),
		"end_date":    c.DefaultQuery("end_date", ""),
		"level":       c.DefaultQuery("level", ""),
	}

	result, statusCode, err := h.riskClient.Get(c.Request.Context(), "/api/v1/alerts", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *RiskBFFHandler) GetStatistics(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	queryParams := map[string]string{
		"merchant_id": merchantID,
		"start_date":  c.DefaultQuery("start_date", ""),
		"end_date":    c.DefaultQuery("end_date", ""),
		"granularity": c.DefaultQuery("granularity", "day"),
	}

	result, statusCode, err := h.riskClient.Get(c.Request.Context(), "/api/v1/statistics", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}
