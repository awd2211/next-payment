package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"payment-platform/merchant-auth-service/internal/service"
)

// APIKeyHandler API密钥处理器
type APIKeyHandler struct {
	service service.APIKeyService
}

// NewAPIKeyHandler 创建API密钥处理器
func NewAPIKeyHandler(service service.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{service: service}
}

// ValidateSignature 验证 API 签名（供 payment-gateway 调用）
// @Summary		验证API签名
// @Description	验证商户API Key和请求签名
// @Tags		Auth
// @Accept		json
// @Produce		json
// @Param		request	body	ValidateSignatureRequest	true	"签名验证请求"
// @Success		200		{object}	ValidateSignatureResponse
// @Failure		400		{object}	ErrorResponse
// @Failure		401		{object}	ErrorResponse
// @Router		/auth/validate-signature [post]
func (h *APIKeyHandler) ValidateSignature(c *gin.Context) {
	var req ValidateSignatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	key, err := h.service.ValidateAPIKey(c.Request.Context(), req.APIKey, req.Signature, req.Payload)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, ValidateSignatureResponse{
		Valid:       true,
		MerchantID:  key.MerchantID.String(),
		Environment: key.Environment,
	})
}

// CreateAPIKey 创建新的API Key
// @Summary		创建API Key
// @Description	为商户创建新的API Key
// @Tags		API Keys
// @Accept		json
// @Produce		json
// @Security	BearerAuth
// @Param		request	body	CreateAPIKeyRequest	true	"创建请求"
// @Success		201		{object}	CreateAPIKeyResponse
// @Failure		400		{object}	ErrorResponse
// @Failure		401		{object}	ErrorResponse
// @Router		/api-keys [post]
func (h *APIKeyHandler) CreateAPIKey(c *gin.Context) {
	// 从JWT中获取merchant_id
	merchantIDStr, exists := c.Get("merchant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	merchantID, err := uuid.Parse(merchantIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid merchant_id"})
		return
	}

	var req CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	key, secret, err := h.service.CreateAPIKey(c.Request.Context(), merchantID, req.Name, req.Environment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, CreateAPIKeyResponse{
		ID:          key.ID.String(),
		APIKey:      key.APIKey,
		APISecret:   secret, // 仅此一次返回
		Name:        key.Name,
		Environment: key.Environment,
		CreatedAt:   key.CreatedAt,
	})
}

// ListAPIKeys 获取API Key列表
// @Summary		获取API Key列表
// @Description	获取商户的所有API Key
// @Tags		API Keys
// @Produce		json
// @Security	BearerAuth
// @Success		200	{array}		APIKeyInfo
// @Failure		401	{object}	ErrorResponse
// @Router		/api-keys [get]
func (h *APIKeyHandler) ListAPIKeys(c *gin.Context) {
	merchantIDStr, exists := c.Get("merchant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	merchantID, err := uuid.Parse(merchantIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid merchant_id"})
		return
	}

	keys, err := h.service.ListAPIKeys(c.Request.Context(), merchantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, keys)
}

// DeleteAPIKey 删除API Key
// @Summary		删除API Key
// @Description	删除指定的API Key
// @Tags		API Keys
// @Produce		json
// @Security	BearerAuth
// @Param		id	path	string	true	"API Key ID"
// @Success		200	{object}	SuccessResponse
// @Failure		401	{object}	ErrorResponse
// @Failure		404	{object}	ErrorResponse
// @Router		/api-keys/{id} [delete]
func (h *APIKeyHandler) DeleteAPIKey(c *gin.Context) {
	merchantIDStr, exists := c.Get("merchant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	merchantID, err := uuid.Parse(merchantIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid merchant_id"})
		return
	}

	keyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid key id"})
		return
	}

	if err := h.service.DeleteAPIKey(c.Request.Context(), merchantID, keyID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Message: "API Key deleted successfully"})
}

// RegisterAPIKeyRoutes 注册API Key路由
func RegisterAPIKeyRoutes(r *gin.RouterGroup, handler *APIKeyHandler, authMiddleware gin.HandlerFunc) {
	// 公开路由（供payment-gateway调用）
	auth := r.Group("/auth")
	{
		auth.POST("/validate-signature", handler.ValidateSignature)
	}

	// 需要认证的路由（供merchant-portal调用）
	apiKeys := r.Group("/api-keys")
	apiKeys.Use(authMiddleware)
	{
		apiKeys.POST("", handler.CreateAPIKey)
		apiKeys.GET("", handler.ListAPIKeys)
		apiKeys.DELETE("/:id", handler.DeleteAPIKey)
	}
}

// DTO定义

type ValidateSignatureRequest struct {
	APIKey    string `json:"api_key" binding:"required"`
	Signature string `json:"signature" binding:"required"`
	Payload   string `json:"payload" binding:"required"`
}

type ValidateSignatureResponse struct {
	Valid       bool   `json:"valid"`
	MerchantID  string `json:"merchant_id"`
	Environment string `json:"environment"`
}

type CreateAPIKeyRequest struct {
	Name        string `json:"name" binding:"required"`
	Environment string `json:"environment" binding:"required,oneof=test production"`
}

type CreateAPIKeyResponse struct {
	ID          string    `json:"id"`
	APIKey      string    `json:"api_key"`
	APISecret   string    `json:"api_secret"` // 仅此一次返回
	Name        string    `json:"name"`
	Environment string    `json:"environment"`
	CreatedAt   time.Time `json:"created_at"`
}

type APIKeyInfo struct {
	ID          string     `json:"id"`
	APIKey      string     `json:"api_key"`
	Name        string     `json:"name"`
	Environment string     `json:"environment"`
	IsActive    bool       `json:"is_active"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}
