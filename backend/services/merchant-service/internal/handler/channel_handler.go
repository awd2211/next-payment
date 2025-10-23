package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/payment-platform/pkg/errors"
	"github.com/payment-platform/pkg/middleware"
	"payment-platform/merchant-service/internal/service"
)

// ChannelHandler 渠道处理器
type ChannelHandler struct {
	channelService service.ChannelService
}

// NewChannelHandler 创建渠道处理器实例
func NewChannelHandler(channelService service.ChannelService) *ChannelHandler {
	return &ChannelHandler{
		channelService: channelService,
	}
}

// RegisterRoutes 注册渠道路由
func (h *ChannelHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	channels := r.Group("/channels")
	channels.Use(authMiddleware)
	{
		channels.POST("", h.CreateChannel)                     // 创建渠道配置
		channels.GET("", h.ListChannels)                       // 获取所有渠道配置
		channels.GET("/:id", h.GetChannel)                     // 获取单个渠道配置
		channels.PUT("/:id", h.UpdateChannel)                  // 更新渠道配置
		channels.DELETE("/:id", h.DeleteChannel)               // 删除渠道配置
		channels.POST("/:id/toggle", h.ToggleChannel)          // 启用/禁用渠道
	}
}

// CreateChannel 创建渠道配置
// @Summary 创建渠道配置
// @Tags Channel
// @Accept json
// @Produce json
// @Param request body service.CreateChannelInput true "渠道配置"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/channels [post]
func (h *ChannelHandler) CreateChannel(c *gin.Context) {
	merchantID, exists := c.Get("user_id")
	if !exists {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeUnauthorized, "未授权", "").WithTraceID(traceID)
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	var req service.CreateChannelInput
	if err := c.ShouldBindJSON(&req); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// 设置商户ID
	req.MerchantID = merchantID.(uuid.UUID)

	channel, err := h.channelService.CreateChannel(c.Request.Context(), &req)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "创建渠道配置失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{
		"message": "渠道配置创建成功",
		"data":    channel,
	}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// GetChannel 获取渠道配置
// @Summary 获取渠道配置
// @Tags Channel
// @Produce json
// @Param id path string true "渠道ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/channels/{id} [get]
func (h *ChannelHandler) GetChannel(c *gin.Context) {
	merchantID, exists := c.Get("user_id")
	if !exists {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeUnauthorized, "未授权", "").WithTraceID(traceID)
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的渠道ID", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	channel, err := h.channelService.GetChannel(c.Request.Context(), channelID, merchantID.(uuid.UUID))
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "获取渠道配置失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(channel).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// ListChannels 获取所有渠道配置
// @Summary 获取所有渠道配置
// @Tags Channel
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/channels [get]
func (h *ChannelHandler) ListChannels(c *gin.Context) {
	merchantID, exists := c.Get("user_id")
	if !exists {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeUnauthorized, "未授权", "").WithTraceID(traceID)
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	channels, err := h.channelService.ListChannels(c.Request.Context(), merchantID.(uuid.UUID))
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "获取渠道列表失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(channels).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// UpdateChannel 更新渠道配置
// @Summary 更新渠道配置
// @Tags Channel
// @Accept json
// @Produce json
// @Param id path string true "渠道ID"
// @Param request body service.UpdateChannelInput true "渠道配置"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/channels/{id} [put]
func (h *ChannelHandler) UpdateChannel(c *gin.Context) {
	merchantID, exists := c.Get("user_id")
	if !exists {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeUnauthorized, "未授权", "").WithTraceID(traceID)
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的渠道ID", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var req service.UpdateChannelInput
	if err := c.ShouldBindJSON(&req); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	channel, err := h.channelService.UpdateChannel(c.Request.Context(), channelID, merchantID.(uuid.UUID), &req)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "更新渠道配置失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{
		"message": "渠道配置更新成功",
		"data":    channel,
	}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// DeleteChannel 删除渠道配置
// @Summary 删除渠道配置
// @Tags Channel
// @Produce json
// @Param id path string true "渠道ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/channels/{id} [delete]
func (h *ChannelHandler) DeleteChannel(c *gin.Context) {
	merchantID, exists := c.Get("user_id")
	if !exists {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeUnauthorized, "未授权", "").WithTraceID(traceID)
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的渠道ID", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	err = h.channelService.DeleteChannel(c.Request.Context(), channelID, merchantID.(uuid.UUID))
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "删除渠道配置失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{"message": "渠道配置删除成功"}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// ToggleChannel 启用/禁用渠道
// @Summary 启用/禁用渠道
// @Tags Channel
// @Accept json
// @Produce json
// @Param id path string true "渠道ID"
// @Param request body map[string]bool true "enabled"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/channels/{id}/toggle [post]
func (h *ChannelHandler) ToggleChannel(c *gin.Context) {
	merchantID, exists := c.Get("user_id")
	if !exists {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeUnauthorized, "未授权", "").WithTraceID(traceID)
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的渠道ID", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var req struct {
		Enabled bool `json:"enabled" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	err = h.channelService.ToggleChannel(c.Request.Context(), channelID, merchantID.(uuid.UUID), req.Enabled)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "切换渠道状态失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	status := "禁用"
	if req.Enabled {
		status = "启用"
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{"message": "渠道已" + status}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}
