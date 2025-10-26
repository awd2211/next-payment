package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"payment-platform/merchant-bff-service/internal/client"
)

type AccountingBFFHandler struct {
	accountingClient *client.ServiceClient
}

func NewAccountingBFFHandler(accountingServiceURL string) *AccountingBFFHandler {
	return &AccountingBFFHandler{
		accountingClient: client.NewServiceClient(accountingServiceURL),
	}
}

func (h *AccountingBFFHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	merchant := r.Group("/merchant/accounting")
	merchant.Use(authMiddleware)
	{
		merchant.GET("/balance", h.GetBalance)
		merchant.GET("/transactions", h.ListTransactions)
		merchant.GET("/invoices", h.ListInvoices)
	}
}

func (h *AccountingBFFHandler) GetBalance(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	queryParams := map[string]string{
		"merchant_id": merchantID,
		"currency":    c.Query("currency"),
	}

	result, statusCode, err := h.accountingClient.Get(c.Request.Context(), "/api/v1/accounting/balance", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *AccountingBFFHandler) ListTransactions(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	queryParams := map[string]string{
		"merchant_id":  merchantID,
		"page":         c.DefaultQuery("page", "1"),
		"page_size":    c.DefaultQuery("page_size", "10"),
		"account_type": c.Query("account_type"),
		"start_time":   c.Query("start_time"),
		"end_time":     c.Query("end_time"),
	}

	result, statusCode, err := h.accountingClient.Get(c.Request.Context(), "/api/v1/accounting/transactions", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *AccountingBFFHandler) ListInvoices(c *gin.Context) {
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

	result, statusCode, err := h.accountingClient.Get(c.Request.Context(), "/api/v1/accounting/invoices", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}
