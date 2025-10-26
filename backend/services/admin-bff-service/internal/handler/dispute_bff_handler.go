package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"payment-platform/admin-service/internal/client"
)

// DisputeBFFHandler Dispute Service BFF处理器
type DisputeBFFHandler struct {
	disputeClient *client.ServiceClient
}

// NewDisputeBFFHandler 创建Dispute BFF处理器
func NewDisputeBFFHandler(disputeServiceURL string) *DisputeBFFHandler {
	return &DisputeBFFHandler{
		disputeClient: client.NewServiceClient(disputeServiceURL),
	}
}

// RegisterRoutes 注册路由
func (h *DisputeBFFHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	admin := r.Group("/admin/disputes")
	admin.Use(authMiddleware)
	{
		admin.POST("", h.CreateDispute)
		admin.GET("", h.ListDisputes)
		admin.GET("/:dispute_id", h.GetDisputeDetails)
		admin.PUT("/:dispute_id/status", h.UpdateStatus)
		admin.POST("/:dispute_id/assign", h.AssignDispute)

		// 证据管理
		admin.POST("/:dispute_id/evidence", h.UploadEvidence)
		admin.GET("/:dispute_id/evidence", h.ListEvidence)
		admin.DELETE("/evidence/:evidence_id", h.DeleteEvidence)

		// Stripe集成
		admin.POST("/:dispute_id/submit", h.SubmitToStripe)
		admin.POST("/sync/:channel_dispute_id", h.SyncFromStripe)

		// 统计
		admin.GET("/statistics", h.GetStatistics)
	}
}

func (h *DisputeBFFHandler) CreateDispute(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, statusCode, err := h.disputeClient.Post(c.Request.Context(), "/api/v1/disputes", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *DisputeBFFHandler) ListDisputes(c *gin.Context) {
	queryParams := map[string]string{
		"page":      c.DefaultQuery("page", "1"),
		"page_size": c.DefaultQuery("page_size", "10"),
	}
	if status := c.Query("status"); status != "" {
		queryParams["status"] = status
	}
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}
	result, statusCode, err := h.disputeClient.Get(c.Request.Context(), "/api/v1/disputes", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *DisputeBFFHandler) GetDisputeDetails(c *gin.Context) {
	disputeID := c.Param("dispute_id")
	result, statusCode, err := h.disputeClient.Get(c.Request.Context(), "/api/v1/disputes/"+disputeID, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *DisputeBFFHandler) UpdateStatus(c *gin.Context) {
	disputeID := c.Param("dispute_id")
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, statusCode, err := h.disputeClient.Put(c.Request.Context(), "/api/v1/disputes/"+disputeID+"/status", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *DisputeBFFHandler) AssignDispute(c *gin.Context) {
	disputeID := c.Param("dispute_id")
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, statusCode, err := h.disputeClient.Post(c.Request.Context(), "/api/v1/disputes/"+disputeID+"/assign", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *DisputeBFFHandler) UploadEvidence(c *gin.Context) {
	disputeID := c.Param("dispute_id")
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, statusCode, err := h.disputeClient.Post(c.Request.Context(), "/api/v1/disputes/"+disputeID+"/evidence", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *DisputeBFFHandler) ListEvidence(c *gin.Context) {
	disputeID := c.Param("dispute_id")
	result, statusCode, err := h.disputeClient.Get(c.Request.Context(), "/api/v1/disputes/"+disputeID+"/evidence", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *DisputeBFFHandler) DeleteEvidence(c *gin.Context) {
	evidenceID := c.Param("evidence_id")
	result, statusCode, err := h.disputeClient.Delete(c.Request.Context(), "/api/v1/disputes/evidence/"+evidenceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *DisputeBFFHandler) SubmitToStripe(c *gin.Context) {
	disputeID := c.Param("dispute_id")
	var req map[string]interface{}
	c.ShouldBindJSON(&req)
	result, statusCode, err := h.disputeClient.Post(c.Request.Context(), "/api/v1/disputes/"+disputeID+"/submit", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *DisputeBFFHandler) SyncFromStripe(c *gin.Context) {
	channelDisputeID := c.Param("channel_dispute_id")
	var req map[string]interface{}
	c.ShouldBindJSON(&req)
	result, statusCode, err := h.disputeClient.Post(c.Request.Context(), "/api/v1/disputes/sync/"+channelDisputeID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *DisputeBFFHandler) GetStatistics(c *gin.Context) {
	queryParams := map[string]string{}
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}
	result, statusCode, err := h.disputeClient.Get(c.Request.Context(), "/api/v1/disputes/statistics", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}
