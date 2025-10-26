package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"payment-platform/merchant-bff-service/internal/client"
)

type MerchantAuthBFFHandler struct {
	authClient *client.ServiceClient
}

func NewMerchantAuthBFFHandler(authServiceURL string) *MerchantAuthBFFHandler {
	return &MerchantAuthBFFHandler{
		authClient: client.NewServiceClient(authServiceURL),
	}
}

func (h *MerchantAuthBFFHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	auth := r.Group("/merchant/auth")
	auth.Use(authMiddleware)
	{
		auth.GET("/api-keys", h.ListAPIKeys)
		auth.POST("/api-keys", h.CreateAPIKey)
		auth.DELETE("/api-keys/:key_id", h.DeleteAPIKey)
		auth.POST("/2fa/enable", h.Enable2FA)
		auth.POST("/2fa/disable", h.Disable2FA)
		auth.GET("/sessions", h.ListSessions)
	}
}

func (h *MerchantAuthBFFHandler) ListAPIKeys(c *gin.Context) {
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

	result, statusCode, err := h.authClient.Get(c.Request.Context(), "/api/v1/api-keys", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *MerchantAuthBFFHandler) CreateAPIKey(c *gin.Context) {
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

	result, statusCode, err := h.authClient.Post(c.Request.Context(), "/api/v1/api-keys", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *MerchantAuthBFFHandler) DeleteAPIKey(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	keyID := c.Param("key_id")
	if keyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "密钥ID不能为空"})
		return
	}

	// Delete方法不需要queryParams，merchant_id应该通过URL路径或body传递
	result, statusCode, err := h.authClient.Delete(c.Request.Context(), "/api/v1/api-keys/"+keyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *MerchantAuthBFFHandler) Enable2FA(c *gin.Context) {
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

	result, statusCode, err := h.authClient.Post(c.Request.Context(), "/api/v1/2fa/enable", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *MerchantAuthBFFHandler) Disable2FA(c *gin.Context) {
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

	result, statusCode, err := h.authClient.Post(c.Request.Context(), "/api/v1/2fa/disable", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *MerchantAuthBFFHandler) ListSessions(c *gin.Context) {
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

	result, statusCode, err := h.authClient.Get(c.Request.Context(), "/api/v1/sessions", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}
