package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"payment-platform/merchant-bff-service/internal/client"
)

type CashierBFFHandler struct {
	cashierClient *client.ServiceClient
}

func NewCashierBFFHandler(cashierServiceURL string) *CashierBFFHandler {
	return &CashierBFFHandler{
		cashierClient: client.NewServiceClient(cashierServiceURL),
	}
}

func (h *CashierBFFHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	cashier := r.Group("/merchant/cashier")
	cashier.Use(authMiddleware)
	{
		cashier.GET("/templates", h.ListTemplates)
		cashier.PUT("/preference", h.UpdatePreference)
		cashier.GET("/preview", h.PreviewCashier)
	}
}

func (h *CashierBFFHandler) ListTemplates(c *gin.Context) {
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

	result, statusCode, err := h.cashierClient.Get(c.Request.Context(), "/api/v1/templates", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *CashierBFFHandler) UpdatePreference(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req["merchant_id"] = merchantID

	result, statusCode, err := h.cashierClient.Put(c.Request.Context(), "/api/v1/preference", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *CashierBFFHandler) PreviewCashier(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	queryParams := map[string]string{
		"merchant_id": merchantID,
		"template_id": c.DefaultQuery("template_id", ""),
	}

	result, statusCode, err := h.cashierClient.Get(c.Request.Context(), "/api/v1/preview", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}
