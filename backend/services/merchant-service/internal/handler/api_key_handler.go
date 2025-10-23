package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/payment-platform/services/merchant-service/internal/service"
)

// APIKeyHandler API密钥处理器
type APIKeyHandler struct {
	apiKeyService service.APIKeyService
}

// NewAPIKeyHandler 创建API密钥处理器实例
func NewAPIKeyHandler(apiKeyService service.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{
		apiKeyService: apiKeyService,
	}
}

// Create 创建API密钥
// @Summary 创建API密钥
// @Tags APIKey
// @Accept json
// @Produce json
// @Param request body service.CreateAPIKeyInput true "API密钥信息"
// @Success 200 {object} Response
// @Router /api/v1/api-keys [post]
func (h *APIKeyHandler) Create(c *gin.Context) {
	// 从JWT中获取商户ID
	merchantID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req service.CreateAPIKeyInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	apiKey, err := h.apiKeyService.Create(c.Request.Context(), merchantID.(uuid.UUID), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "API密钥创建成功，请妥善保管API Secret，它只会显示这一次",
		"data":    apiKey,
	})
}

// GetByID 根据ID获取API密钥
// @Summary 获取API密钥详情
// @Tags APIKey
// @Produce json
// @Param id path string true "API密钥ID"
// @Success 200 {object} Response
// @Router /api/v1/api-keys/{id} [get]
func (h *APIKeyHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的API密钥ID"})
		return
	}

	apiKey, err := h.apiKeyService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 隐藏Secret
	apiKey.APISecret = "sk_****"

	c.JSON(http.StatusOK, gin.H{
		"data": apiKey,
	})
}

// List 获取API密钥列表
// @Summary 获取API密钥列表
// @Tags APIKey
// @Produce json
// @Param environment query string false "环境筛选（test/production）"
// @Success 200 {object} Response
// @Router /api/v1/api-keys [get]
func (h *APIKeyHandler) List(c *gin.Context) {
	// 从JWT中获取商户ID
	merchantID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	environment := c.Query("environment")

	apiKeys, err := h.apiKeyService.ListByMerchant(c.Request.Context(), merchantID.(uuid.UUID), environment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": apiKeys,
	})
}

// Update 更新API密钥
// @Summary 更新API密钥
// @Tags APIKey
// @Accept json
// @Produce json
// @Param id path string true "API密钥ID"
// @Param request body service.UpdateAPIKeyInput true "更新信息"
// @Success 200 {object} Response
// @Router /api/v1/api-keys/{id} [put]
func (h *APIKeyHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的API密钥ID"})
		return
	}

	var req service.UpdateAPIKeyInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	apiKey, err := h.apiKeyService.Update(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
		"data":    apiKey,
	})
}

// Revoke 撤销API密钥
// @Summary 撤销API密钥
// @Tags APIKey
// @Produce json
// @Param id path string true "API密钥ID"
// @Success 200 {object} Response
// @Router /api/v1/api-keys/{id}/revoke [post]
func (h *APIKeyHandler) Revoke(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的API密钥ID"})
		return
	}

	if err := h.apiKeyService.Revoke(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "API密钥已撤销",
	})
}

// Delete 删除API密钥
// @Summary 删除API密钥
// @Tags APIKey
// @Produce json
// @Param id path string true "API密钥ID"
// @Success 200 {object} Response
// @Router /api/v1/api-keys/{id} [delete]
func (h *APIKeyHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的API密钥ID"})
		return
	}

	if err := h.apiKeyService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "API密钥已删除",
	})
}

// Rotate 轮换API密钥
// @Summary 轮换API密钥（生成新Secret）
// @Tags APIKey
// @Produce json
// @Param id path string true "API密钥ID"
// @Success 200 {object} Response
// @Router /api/v1/api-keys/{id}/rotate [post]
func (h *APIKeyHandler) Rotate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的API密钥ID"})
		return
	}

	apiKey, err := h.apiKeyService.Rotate(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "API密钥已轮换，新的Secret只显示这一次，请妥善保管",
		"data":    apiKey,
	})
}

// RegisterRoutes 注册路由
func (h *APIKeyHandler) RegisterRoutes(r *gin.RouterGroup) {
	apiKeys := r.Group("/api-keys")
	// apiKeys.Use(middleware.AuthMiddleware()) // 需要认证
	{
		apiKeys.POST("", h.Create)
		apiKeys.GET("", h.List)
		apiKeys.GET("/:id", h.GetByID)
		apiKeys.PUT("/:id", h.Update)
		apiKeys.POST("/:id/revoke", h.Revoke)
		apiKeys.POST("/:id/rotate", h.Rotate)
		apiKeys.DELETE("/:id", h.Delete)
	}
}
