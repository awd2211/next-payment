package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"payment-platform/admin-service/internal/client"
)

type MerchantAuthBFFHandler struct {
	authClient *client.ServiceClient
}

func NewMerchantAuthBFFHandler(merchantAuthServiceURL string) *MerchantAuthBFFHandler {
	return &MerchantAuthBFFHandler{
		authClient: client.NewServiceClient(merchantAuthServiceURL),
	}
}

func (h *MerchantAuthBFFHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	admin := r.Group("/admin/merchant-auth")
	admin.Use(authMiddleware)
	{
		// API Key管理
		apiKeys := admin.Group("/api-keys")
		{
			apiKeys.GET("", h.ListAPIKeys)
			apiKeys.GET("/:id", h.GetAPIKey)
			apiKeys.POST("", h.CreateAPIKey)
			apiKeys.PUT("/:id", h.UpdateAPIKey)
			apiKeys.DELETE("/:id", h.DeleteAPIKey)
			apiKeys.POST("/:id/revoke", h.RevokeAPIKey)
			apiKeys.POST("/:id/regenerate", h.RegenerateAPIKey)
		}

		// 2FA管理
		twoFA := admin.Group("/2fa")
		{
			twoFA.GET("/:merchant_id/status", h.Get2FAStatus)
			twoFA.POST("/:merchant_id/enable", h.Enable2FA)
			twoFA.POST("/:merchant_id/disable", h.Disable2FA)
			twoFA.POST("/:merchant_id/reset", h.Reset2FA)
		}

		// 会话管理
		sessions := admin.Group("/sessions")
		{
			sessions.GET("", h.ListSessions)
			sessions.GET("/:id", h.GetSession)
			sessions.DELETE("/:id", h.TerminateSession)
			sessions.POST("/:merchant_id/terminate-all", h.TerminateAllSessions)
		}

		// 登录日志
		admin.GET("/login-logs", h.ListLoginLogs)
		admin.GET("/login-logs/:id", h.GetLoginLog)
	}
}

// ========== API Key管理 ==========

func (h *MerchantAuthBFFHandler) ListAPIKeys(c *gin.Context) {
	queryParams := make(map[string]string)
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}
	if status := c.Query("status"); status != "" {
		queryParams["status"] = status
	}
	if page := c.Query("page"); page != "" {
		queryParams["page"] = page
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		queryParams["page_size"] = pageSize
	}

	result, statusCode, err := h.authClient.Get(c.Request.Context(), "/api/v1/api-keys", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Auth Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *MerchantAuthBFFHandler) GetAPIKey(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.authClient.Get(c.Request.Context(), "/api/v1/api-keys/"+id, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Auth Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *MerchantAuthBFFHandler) CreateAPIKey(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	adminID := c.GetString("user_id")
	req["created_by"] = adminID

	result, statusCode, err := h.authClient.Post(c.Request.Context(), "/api/v1/api-keys", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Auth Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *MerchantAuthBFFHandler) UpdateAPIKey(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	result, statusCode, err := h.authClient.Put(c.Request.Context(), "/api/v1/api-keys/"+id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Auth Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *MerchantAuthBFFHandler) DeleteAPIKey(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.authClient.Delete(c.Request.Context(), "/api/v1/api-keys/"+id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Auth Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *MerchantAuthBFFHandler) RevokeAPIKey(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		req = make(map[string]interface{})
	}

	adminID := c.GetString("user_id")
	req["revoked_by"] = adminID

	result, statusCode, err := h.authClient.Post(c.Request.Context(), "/api/v1/api-keys/"+id+"/revoke", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Auth Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *MerchantAuthBFFHandler) RegenerateAPIKey(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		req = make(map[string]interface{})
	}

	result, statusCode, err := h.authClient.Post(c.Request.Context(), "/api/v1/api-keys/"+id+"/regenerate", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Auth Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== 2FA管理 ==========

func (h *MerchantAuthBFFHandler) Get2FAStatus(c *gin.Context) {
	merchantID := c.Param("merchant_id")

	result, statusCode, err := h.authClient.Get(c.Request.Context(), "/api/v1/2fa/"+merchantID+"/status", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Auth Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *MerchantAuthBFFHandler) Enable2FA(c *gin.Context) {
	merchantID := c.Param("merchant_id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		req = make(map[string]interface{})
	}

	adminID := c.GetString("user_id")
	req["enabled_by"] = adminID

	result, statusCode, err := h.authClient.Post(c.Request.Context(), "/api/v1/2fa/"+merchantID+"/enable", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Auth Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *MerchantAuthBFFHandler) Disable2FA(c *gin.Context) {
	merchantID := c.Param("merchant_id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		req = make(map[string]interface{})
	}

	adminID := c.GetString("user_id")
	req["disabled_by"] = adminID

	result, statusCode, err := h.authClient.Post(c.Request.Context(), "/api/v1/2fa/"+merchantID+"/disable", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Auth Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *MerchantAuthBFFHandler) Reset2FA(c *gin.Context) {
	merchantID := c.Param("merchant_id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		req = make(map[string]interface{})
	}

	adminID := c.GetString("user_id")
	req["reset_by"] = adminID

	result, statusCode, err := h.authClient.Post(c.Request.Context(), "/api/v1/2fa/"+merchantID+"/reset", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Auth Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== 会话管理 ==========

func (h *MerchantAuthBFFHandler) ListSessions(c *gin.Context) {
	queryParams := make(map[string]string)
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}
	if status := c.Query("status"); status != "" {
		queryParams["status"] = status
	}
	if page := c.Query("page"); page != "" {
		queryParams["page"] = page
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		queryParams["page_size"] = pageSize
	}

	result, statusCode, err := h.authClient.Get(c.Request.Context(), "/api/v1/sessions", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Auth Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *MerchantAuthBFFHandler) GetSession(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.authClient.Get(c.Request.Context(), "/api/v1/sessions/"+id, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Auth Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *MerchantAuthBFFHandler) TerminateSession(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.authClient.Delete(c.Request.Context(), "/api/v1/sessions/"+id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Auth Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *MerchantAuthBFFHandler) TerminateAllSessions(c *gin.Context) {
	merchantID := c.Param("merchant_id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		req = make(map[string]interface{})
	}

	adminID := c.GetString("user_id")
	req["terminated_by"] = adminID

	result, statusCode, err := h.authClient.Post(c.Request.Context(), "/api/v1/sessions/"+merchantID+"/terminate-all", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Auth Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== 登录日志 ==========

func (h *MerchantAuthBFFHandler) ListLoginLogs(c *gin.Context) {
	queryParams := make(map[string]string)
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}
	if status := c.Query("status"); status != "" {
		queryParams["status"] = status
	}
	if startTime := c.Query("start_time"); startTime != "" {
		queryParams["start_time"] = startTime
	}
	if endTime := c.Query("end_time"); endTime != "" {
		queryParams["end_time"] = endTime
	}
	if page := c.Query("page"); page != "" {
		queryParams["page"] = page
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		queryParams["page_size"] = pageSize
	}

	result, statusCode, err := h.authClient.Get(c.Request.Context(), "/api/v1/login-logs", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Auth Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *MerchantAuthBFFHandler) GetLoginLog(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.authClient.Get(c.Request.Context(), "/api/v1/login-logs/"+id, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Merchant Auth Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}
