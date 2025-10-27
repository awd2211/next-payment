package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/payment-platform/pkg/errors"
	"github.com/payment-platform/pkg/middleware"
	"payment-platform/merchant-auth-service/internal/service"
)

// SecurityHandler 安全处理器
type SecurityHandler struct {
	securityService service.SecurityService
}

// NewSecurityHandler 创建安全处理器实例
func NewSecurityHandler(securityService service.SecurityService) *SecurityHandler {
	return &SecurityHandler{
		securityService: securityService,
	}
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Tags Security
// @Accept json
// @Produce json
// @Param request body ChangePasswordRequest true "修改密码请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/security/password [put]
func (h *SecurityHandler) ChangePassword(c *gin.Context) {
	merchantID, exists := c.Get("user_id")
	if !exists {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeUnauthorized, "未授权", "").WithTraceID(traceID)
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	if err := h.securityService.ChangePassword(c.Request.Context(), merchantID.(uuid.UUID), req.OldPassword, req.NewPassword); err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "修改密码失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{"message": "密码修改成功"}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// Enable2FA 启用双因素认证
// @Summary 启用2FA
// @Tags Security
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/security/2fa/enable [post]
func (h *SecurityHandler) Enable2FA(c *gin.Context) {
	merchantID, exists := c.Get("user_id")
	if !exists {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeUnauthorized, "未授权", "").WithTraceID(traceID)
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	result, err := h.securityService.Enable2FA(c.Request.Context(), merchantID.(uuid.UUID))
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "启用2FA失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{
		"message": "请扫描二维码并验证",
		"data":    result,
	}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// Verify2FA 验证并启用2FA
// @Summary 验证2FA
// @Tags Security
// @Accept json
// @Produce json
// @Param request body Verify2FARequest true "验证请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/security/2fa/verify [post]
func (h *SecurityHandler) Verify2FA(c *gin.Context) {
	merchantID, exists := c.Get("user_id")
	if !exists {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeUnauthorized, "未授权", "").WithTraceID(traceID)
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	var req Verify2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	result, err := h.securityService.Verify2FA(c.Request.Context(), merchantID.(uuid.UUID), req.Code)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "验证2FA失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	if !result.Success {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "验证码错误", "").WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{
		"message": "2FA已启用",
		"data":    result,
	}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// Verify2FARequest 验证2FA请求
type Verify2FARequest struct {
	Code string `json:"code" binding:"required"`
}

// Disable2FA 禁用2FA
// @Summary 禁用2FA
// @Tags Security
// @Accept json
// @Produce json
// @Param request body Disable2FARequest true "禁用请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/security/2fa/disable [post]
func (h *SecurityHandler) Disable2FA(c *gin.Context) {
	merchantID, exists := c.Get("user_id")
	if !exists {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeUnauthorized, "未授权", "").WithTraceID(traceID)
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	var req Disable2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	if err := h.securityService.Disable2FA(c.Request.Context(), merchantID.(uuid.UUID), req.Password); err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "禁用2FA失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{"message": "2FA已禁用"}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// Disable2FARequest 禁用2FA请求
type Disable2FARequest struct {
	Password string `json:"password" binding:"required"`
}

// GetSecuritySettings 获取安全设置
// @Summary 获取安全设置
// @Tags Security
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/security/settings [get]
func (h *SecurityHandler) GetSecuritySettings(c *gin.Context) {
	merchantID, exists := c.Get("user_id")
	if !exists {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeUnauthorized, "未授权", "").WithTraceID(traceID)
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	settings, err := h.securityService.GetSecuritySettings(c.Request.Context(), merchantID.(uuid.UUID))
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "获取安全设置失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(settings).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// UpdateSecuritySettings 更新安全设置
// @Summary 更新安全设置
// @Tags Security
// @Accept json
// @Produce json
// @Param request body service.UpdateSecuritySettingsInput true "安全设置"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/security/settings [put]
func (h *SecurityHandler) UpdateSecuritySettings(c *gin.Context) {
	merchantID, exists := c.Get("user_id")
	if !exists {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeUnauthorized, "未授权", "").WithTraceID(traceID)
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	var req service.UpdateSecuritySettingsInput
	if err := c.ShouldBindJSON(&req); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	settings, err := h.securityService.UpdateSecuritySettings(c.Request.Context(), merchantID.(uuid.UUID), &req)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "更新安全设置失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{
		"message": "安全设置更新成功",
		"data":    settings,
	}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// GetLoginActivities 获取登录活动
// @Summary 获取登录活动
// @Tags Security
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/security/login-activities [get]
func (h *SecurityHandler) GetLoginActivities(c *gin.Context) {
	merchantID, exists := c.Get("user_id")
	if !exists {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeUnauthorized, "未授权", "").WithTraceID(traceID)
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	// 验证并限制分页参数（防止DoS攻击）
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100 // 最大限制100条/页
	}

	activities, total, err := h.securityService.GetLoginActivities(c.Request.Context(), merchantID.(uuid.UUID), page, pageSize)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "获取登录活动失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{
		"list":      activities,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// GetActiveSessions 获取活跃会话
// @Summary 获取活跃会话
// @Tags Security
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/security/sessions [get]
func (h *SecurityHandler) GetActiveSessions(c *gin.Context) {
	merchantID, exists := c.Get("user_id")
	if !exists {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeUnauthorized, "未授权", "").WithTraceID(traceID)
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	sessions, err := h.securityService.GetActiveSessions(c.Request.Context(), merchantID.(uuid.UUID))
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "获取活跃会话失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(sessions).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// RevokeSession 撤销会话
// @Summary 撤销会话
// @Tags Security
// @Produce json
// @Param session_id path string true "会话ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/security/sessions/{session_id} [delete]
func (h *SecurityHandler) RevokeSession(c *gin.Context) {
	sessionID := c.Param("session_id")

	if err := h.securityService.RevokeSession(c.Request.Context(), sessionID); err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "撤销会话失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{"message": "会话已撤销"}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// RevokeAllSessions 撤销所有会话
// @Summary 撤销所有会话
// @Tags Security
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/security/sessions [delete]
func (h *SecurityHandler) RevokeAllSessions(c *gin.Context) {
	merchantID, exists := c.Get("user_id")
	if !exists {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeUnauthorized, "未授权", "").WithTraceID(traceID)
		c.JSON(http.StatusUnauthorized, resp)
		return
	}

	if err := h.securityService.RevokeAllSessions(c.Request.Context(), merchantID.(uuid.UUID)); err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "撤销所有会话失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{"message": "所有会话已撤销"}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

// RegisterRoutes 注册路由
func (h *SecurityHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	security := r.Group("/security")
	security.Use(authMiddleware)
	{
		// 密码管理
		security.PUT("/password", h.ChangePassword)

		// 双因素认证
		security.POST("/2fa/enable", h.Enable2FA)
		security.POST("/2fa/verify", h.Verify2FA)
		security.POST("/2fa/disable", h.Disable2FA)

		// 安全设置
		security.GET("/settings", h.GetSecuritySettings)
		security.PUT("/settings", h.UpdateSecuritySettings)

		// 登录活动
		security.GET("/login-activities", h.GetLoginActivities)

		// 会话管理
		security.GET("/sessions", h.GetActiveSessions)
		security.DELETE("/sessions/:session_id", h.RevokeSession)
		security.DELETE("/sessions", h.RevokeAllSessions)
	}
}
