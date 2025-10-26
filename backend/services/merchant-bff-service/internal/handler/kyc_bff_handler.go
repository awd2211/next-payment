package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"payment-platform/merchant-bff-service/internal/client"
)

type KYCBFFHandler struct {
	kycClient *client.ServiceClient
}

func NewKYCBFFHandler(kycServiceURL string) *KYCBFFHandler {
	return &KYCBFFHandler{
		kycClient: client.NewServiceClient(kycServiceURL),
	}
}

func (h *KYCBFFHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	merchant := r.Group("/merchant/kyc")
	merchant.Use(authMiddleware)
	{
		merchant.POST("/documents", h.UploadDocument)
		merchant.GET("/documents", h.ListDocuments)
		merchant.GET("/status", h.GetStatus)
	}
}

func (h *KYCBFFHandler) UploadDocument(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	requestBody["merchant_id"] = merchantID

	result, statusCode, err := h.kycClient.Post(c.Request.Context(), "/api/v1/kyc/documents", requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *KYCBFFHandler) ListDocuments(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	queryParams := map[string]string{
		"merchant_id":   merchantID,
		"page":          c.DefaultQuery("page", "1"),
		"page_size":     c.DefaultQuery("page_size", "10"),
		"document_type": c.Query("document_type"),
		"status":        c.Query("status"),
	}

	result, statusCode, err := h.kycClient.Get(c.Request.Context(), "/api/v1/kyc/documents", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *KYCBFFHandler) GetStatus(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	queryParams := map[string]string{
		"merchant_id": merchantID,
	}

	result, statusCode, err := h.kycClient.Get(c.Request.Context(), "/api/v1/kyc/status", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}
