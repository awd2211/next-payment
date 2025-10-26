package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"payment-platform/merchant-bff-service/internal/client"
)

type MerchantConfigBFFHandler struct {
	configClient *client.ServiceClient
}

func NewMerchantConfigBFFHandler(configServiceURL string) *MerchantConfigBFFHandler {
	return &MerchantConfigBFFHandler{
		configClient: client.NewServiceClient(configServiceURL),
	}
}

func (h *MerchantConfigBFFHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	config := r.Group("/merchant/config")
	config.Use(authMiddleware)
	{
		config.GET("/fee", h.GetFeeConfig)
		config.GET("/limits", h.GetLimits)
		config.PUT("/webhook", h.UpdateWebhook)
		config.GET("/channels", h.GetAvailableChannels)
	}
}

func (h *MerchantConfigBFFHandler) GetFeeConfig(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	queryParams := map[string]string{
		"merchant_id": merchantID,
	}

	result, statusCode, err := h.configClient.Get(c.Request.Context(), "/api/v1/fee", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *MerchantConfigBFFHandler) GetLimits(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	queryParams := map[string]string{
		"merchant_id": merchantID,
	}

	result, statusCode, err := h.configClient.Get(c.Request.Context(), "/api/v1/limits", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *MerchantConfigBFFHandler) UpdateWebhook(c *gin.Context) {
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

	result, statusCode, err := h.configClient.Put(c.Request.Context(), "/api/v1/webhook", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *MerchantConfigBFFHandler) GetAvailableChannels(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到商户ID"})
		return
	}

	queryParams := map[string]string{
		"merchant_id": merchantID,
	}

	result, statusCode, err := h.configClient.Get(c.Request.Context(), "/api/v1/channels", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}
