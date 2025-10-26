package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"payment-platform/merchant-bff-service/internal/client"
)

type SettlementBFFHandler struct {
	settlementClient *client.ServiceClient
}

func NewSettlementBFFHandler(settlementServiceURL string) *SettlementBFFHandler {
	return &SettlementBFFHandler{
		settlementClient: client.NewServiceClient(settlementServiceURL),
	}
}

func (h *SettlementBFFHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	merchant := r.Group("/merchant/settlements")
	merchant.Use(authMiddleware)
	{
		merchant.GET("", h.ListSettlements)
		merchant.GET("/:settlement_no", h.GetSettlement)
		merchant.GET("/statistics", h.GetStatistics)
	}
}

func (h *SettlementBFFHandler) ListSettlements(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	queryParams := map[string]string{
		"merchant_id": merchantID,
		"page":        c.DefaultQuery("page", "1"),
		"page_size":   c.DefaultQuery("page_size", "10"),
		"status":      c.Query("status"),
		"start_time":  c.Query("start_time"),
		"end_time":    c.Query("end_time"),
	}

	result, statusCode, err := h.settlementClient.Get(c.Request.Context(), "/api/v1/settlements", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *SettlementBFFHandler) GetSettlement(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	settlementNo := c.Param("settlement_no")
	if settlementNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "结算单号不能为空"})
		return
	}

	queryParams := map[string]string{
		"merchant_id": merchantID,
	}

	result, statusCode, err := h.settlementClient.Get(c.Request.Context(), "/api/v1/settlements/"+settlementNo, queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *SettlementBFFHandler) GetStatistics(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	queryParams := map[string]string{
		"merchant_id": merchantID,
		"start_time":  c.Query("start_time"),
		"end_time":    c.Query("end_time"),
	}

	result, statusCode, err := h.settlementClient.Get(c.Request.Context(), "/api/v1/settlements/statistics", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}
