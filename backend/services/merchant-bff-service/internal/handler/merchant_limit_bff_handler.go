package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"payment-platform/merchant-bff-service/internal/client"
)

type MerchantLimitBFFHandler struct {
	limitClient *client.ServiceClient
}

func NewMerchantLimitBFFHandler(limitServiceURL string) *MerchantLimitBFFHandler {
	return &MerchantLimitBFFHandler{
		limitClient: client.NewServiceClient(limitServiceURL),
	}
}

func (h *MerchantLimitBFFHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	limits := r.Group("/merchant/limits")
	limits.Use(authMiddleware)
	{
		limits.GET("/tier", h.GetCurrentTier)
		limits.GET("/usage", h.GetUsage)
		limits.GET("/history", h.GetHistory)
	}
}

func (h *MerchantLimitBFFHandler) GetCurrentTier(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	queryParams := map[string]string{
		"merchant_id": merchantID,
	}

	result, statusCode, err := h.limitClient.Get(c.Request.Context(), "/api/v1/tier", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *MerchantLimitBFFHandler) GetUsage(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	queryParams := map[string]string{
		"merchant_id": merchantID,
		"period":      c.DefaultQuery("period", "current"),
	}

	result, statusCode, err := h.limitClient.Get(c.Request.Context(), "/api/v1/usage", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *MerchantLimitBFFHandler) GetHistory(c *gin.Context) {
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
	}

	result, statusCode, err := h.limitClient.Get(c.Request.Context(), "/api/v1/history", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}
