package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"payment-platform/merchant-bff-service/internal/client"
)

type DisputeBFFHandler struct {
	disputeClient *client.ServiceClient
}

func NewDisputeBFFHandler(disputeServiceURL string) *DisputeBFFHandler {
	return &DisputeBFFHandler{
		disputeClient: client.NewServiceClient(disputeServiceURL),
	}
}

func (h *DisputeBFFHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	disputes := r.Group("/merchant/disputes")
	disputes.Use(authMiddleware)
	{
		disputes.GET("", h.ListDisputes)
		disputes.GET("/:dispute_id", h.GetDispute)
		disputes.POST("/:dispute_id/evidence", h.UploadEvidence)
		disputes.GET("/:dispute_id/evidence", h.ListEvidence)
	}
}

func (h *DisputeBFFHandler) ListDisputes(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	queryParams := map[string]string{
		"merchant_id": merchantID,
		"page":        c.DefaultQuery("page", "1"),
		"page_size":   c.DefaultQuery("page_size", "10"),
		"status":      c.DefaultQuery("status", ""),
		"start_date":  c.DefaultQuery("start_date", ""),
		"end_date":    c.DefaultQuery("end_date", ""),
	}

	result, statusCode, err := h.disputeClient.Get(c.Request.Context(), "/api/v1/disputes", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *DisputeBFFHandler) GetDispute(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	disputeID := c.Param("dispute_id")
	if disputeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "争议ID不能为空"})
		return
	}

	queryParams := map[string]string{
		"merchant_id": merchantID,
	}

	result, statusCode, err := h.disputeClient.Get(c.Request.Context(), "/api/v1/disputes/"+disputeID, queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *DisputeBFFHandler) UploadEvidence(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	disputeID := c.Param("dispute_id")
	if disputeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "争议ID不能为空"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req["merchant_id"] = merchantID

	result, statusCode, err := h.disputeClient.Post(c.Request.Context(), "/api/v1/disputes/"+disputeID+"/evidence", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *DisputeBFFHandler) ListEvidence(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	disputeID := c.Param("dispute_id")
	if disputeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "争议ID不能为空"})
		return
	}

	queryParams := map[string]string{
		"merchant_id": merchantID,
		"page":        c.DefaultQuery("page", "1"),
		"page_size":   c.DefaultQuery("page_size", "10"),
	}

	result, statusCode, err := h.disputeClient.Get(c.Request.Context(), "/api/v1/disputes/"+disputeID+"/evidence", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}
