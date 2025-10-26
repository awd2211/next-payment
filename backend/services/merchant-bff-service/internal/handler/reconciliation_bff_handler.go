package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"payment-platform/merchant-bff-service/internal/client"
)

type ReconciliationBFFHandler struct {
	reconciliationClient *client.ServiceClient
}

func NewReconciliationBFFHandler(reconciliationServiceURL string) *ReconciliationBFFHandler {
	return &ReconciliationBFFHandler{
		reconciliationClient: client.NewServiceClient(reconciliationServiceURL),
	}
}

func (h *ReconciliationBFFHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	reconciliation := r.Group("/merchant/reconciliation")
	reconciliation.Use(authMiddleware)
	{
		reconciliation.GET("/reports", h.ListReports)
		reconciliation.GET("/reports/:report_id", h.GetReport)
		reconciliation.GET("/discrepancies", h.ListDiscrepancies)
	}
}

func (h *ReconciliationBFFHandler) ListReports(c *gin.Context) {
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
		"status":      c.DefaultQuery("status", ""),
	}

	result, statusCode, err := h.reconciliationClient.Get(c.Request.Context(), "/api/v1/reports", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *ReconciliationBFFHandler) GetReport(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	reportID := c.Param("report_id")
	if reportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "报告ID不能为空"})
		return
	}

	queryParams := map[string]string{
		"merchant_id": merchantID,
	}

	result, statusCode, err := h.reconciliationClient.Get(c.Request.Context(), "/api/v1/reports/"+reportID, queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *ReconciliationBFFHandler) ListDiscrepancies(c *gin.Context) {
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
		"status":      c.DefaultQuery("status", ""),
		"type":        c.DefaultQuery("type", ""),
	}

	result, statusCode, err := h.reconciliationClient.Get(c.Request.Context(), "/api/v1/discrepancies", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}
