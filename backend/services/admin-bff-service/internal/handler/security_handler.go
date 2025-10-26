package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"payment-platform/admin-service/internal/service"
)

// SecurityHandler 安全功能处理器
type SecurityHandler struct {
	securityService service.SecurityService
}

// NewSecurityHandler 创建安全功能处理器实例
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
// @Router /api/v1/security/change-password [post]
func (h *SecurityHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 从上下文获取用户信息
	userID, _ := c.Get("user_id")
	userType, _ := c.Get("user_type")

	err := h.securityService.ChangePassword(
		c.Request.Context(),
		userID.(uuid.UUID),
		userType.(string),
		req.OldPassword,
		req.NewPassword,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "密码修改成功",
	})
}

// Setup2FARequest 设置2FA请求
type Setup2FARequest struct {
	AccountName string `json:"account_name" binding:"required"` // 账户名称（如邮箱）
}

// Setup2FA 设置2FA
// @Summary 设置双因素认证
// @Tags Security
// @Accept json
// @Produce json
// @Param request body Setup2FARequest true "设置2FA请求"
// @Success 200 {object} service.Setup2FAResponse
// @Router /api/v1/security/2fa/setup [post]
func (h *SecurityHandler) Setup2FA(c *gin.Context) {
	var req Setup2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	userType, _ := c.Get("user_type")

	result, err := h.securityService.Setup2FA(
		c.Request.Context(),
		userID.(uuid.UUID),
		userType.(string),
		req.AccountName,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "2FA设置成功，请使用Google Authenticator或其他TOTP应用扫描二维码",
		"data":    result,
	})
}

// Verify2FARequest 验证2FA请求
type Verify2FARequest struct {
	Code string `json:"code" binding:"required"` // 6位验证码
}

// Verify2FA 验证并启用2FA
// @Summary 验证并启用2FA
// @Tags Security
// @Accept json
// @Produce json
// @Param request body Verify2FARequest true "验证2FA请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/security/2fa/verify [post]
func (h *SecurityHandler) Verify2FA(c *gin.Context) {
	var req Verify2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	userType, _ := c.Get("user_type")

	err := h.securityService.Verify2FA(
		c.Request.Context(),
		userID.(uuid.UUID),
		userType.(string),
		req.Code,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "2FA已成功启用",
	})
}

// Disable2FARequest 禁用2FA请求
type Disable2FARequest struct {
	Password string `json:"password" binding:"required"` // 密码验证
}

// Disable2FA 禁用2FA
// @Summary 禁用双因素认证
// @Tags Security
// @Accept json
// @Produce json
// @Param request body Disable2FARequest true "禁用2FA请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/security/2fa/disable [post]
func (h *SecurityHandler) Disable2FA(c *gin.Context) {
	var req Disable2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	userType, _ := c.Get("user_type")

	err := h.securityService.Disable2FA(
		c.Request.Context(),
		userID.(uuid.UUID),
		userType.(string),
		req.Password,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "2FA已禁用",
	})
}

// RegenerateBackupCodes 重新生成备用代码
// @Summary 重新生成备用恢复代码
// @Tags Security
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/security/2fa/backup-codes [post]
func (h *SecurityHandler) RegenerateBackupCodes(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userType, _ := c.Get("user_type")

	codes, err := h.securityService.RegenerateBackupCodes(
		c.Request.Context(),
		userID.(uuid.UUID),
		userType.(string),
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "备用代码已重新生成",
		"data": gin.H{
			"backup_codes": codes,
		},
	})
}

// GetLoginActivities 获取登录活动记录
// @Summary 获取登录活动记录
// @Tags Security
// @Produce json
// @Param limit query int false "记录数量" default(50)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/security/login-activities [get]
func (h *SecurityHandler) GetLoginActivities(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userType, _ := c.Get("user_type")

	limit := 50
	if l, ok := c.GetQuery("limit"); ok {
		if parsedLimit, err := strconv.Atoi(l); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	activities, err := h.securityService.GetLoginActivities(
		c.Request.Context(),
		userID.(uuid.UUID),
		userType.(string),
		limit,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": activities,
	})
}

// GetAbnormalActivities 获取异常登录活动
// @Summary 获取异常登录活动
// @Tags Security
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/security/abnormal-activities [get]
func (h *SecurityHandler) GetAbnormalActivities(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userType, _ := c.Get("user_type")

	activities, err := h.securityService.GetAbnormalActivities(
		c.Request.Context(),
		userID.(uuid.UUID),
		userType.(string),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": activities,
	})
}

// GetSecuritySettings 获取安全设置
// @Summary 获取安全设置
// @Tags Security
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/security/settings [get]
func (h *SecurityHandler) GetSecuritySettings(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userType, _ := c.Get("user_type")

	settings, err := h.securityService.GetSecuritySettings(
		c.Request.Context(),
		userID.(uuid.UUID),
		userType.(string),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": settings,
	})
}

// UpdateSecuritySettings 更新安全设置
// @Summary 更新安全设置
// @Tags Security
// @Accept json
// @Produce json
// @Param request body service.UpdateSecuritySettingsInput true "更新安全设置请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/security/settings [put]
func (h *SecurityHandler) UpdateSecuritySettings(c *gin.Context) {
	var req service.UpdateSecuritySettingsInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	userType, _ := c.Get("user_type")

	err := h.securityService.UpdateSecuritySettings(
		c.Request.Context(),
		userID.(uuid.UUID),
		userType.(string),
		&req,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "安全设置已更新",
	})
}

// GetActiveSessions 获取活跃会话
// @Summary 获取活跃会话列表
// @Tags Security
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/security/sessions [get]
func (h *SecurityHandler) GetActiveSessions(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userType, _ := c.Get("user_type")

	sessions, err := h.securityService.GetActiveSessions(
		c.Request.Context(),
		userID.(uuid.UUID),
		userType.(string),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": sessions,
	})
}

// DeactivateSessionRequest 停用会话请求
type DeactivateSessionRequest struct {
	SessionID string `json:"session_id" binding:"required"`
}

// DeactivateSession 停用指定会话
// @Summary 停用指定会话
// @Tags Security
// @Accept json
// @Produce json
// @Param request body DeactivateSessionRequest true "停用会话请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/security/sessions/deactivate [post]
func (h *SecurityHandler) DeactivateSession(c *gin.Context) {
	var req DeactivateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.securityService.DeactivateSession(c.Request.Context(), req.SessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "会话已停用",
	})
}

// DeactivateAllOtherSessions 停用其他所有会话
// @Summary 停用其他所有会话
// @Tags Security
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/security/sessions/deactivate-others [post]
func (h *SecurityHandler) DeactivateAllOtherSessions(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userType, _ := c.Get("user_type")
	sessionID, _ := c.Get("session_id")

	err := h.securityService.DeactivateAllOtherSessions(
		c.Request.Context(),
		userID.(uuid.UUID),
		userType.(string),
		sessionID.(string),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "其他所有会话已停用",
	})
}

// RegisterRoutes 注册路由
func (h *SecurityHandler) RegisterRoutes(r *gin.RouterGroup) {
	security := r.Group("/security")
	{
		// 密码管理
		security.POST("/change-password", h.ChangePassword)

		// 2FA管理
		twoFA := security.Group("/2fa")
		{
			twoFA.POST("/setup", h.Setup2FA)
			twoFA.POST("/verify", h.Verify2FA)
			twoFA.POST("/disable", h.Disable2FA)
			twoFA.POST("/backup-codes", h.RegenerateBackupCodes)
		}

		// 登录活动
		security.GET("/login-activities", h.GetLoginActivities)
		security.GET("/abnormal-activities", h.GetAbnormalActivities)

		// 安全设置
		security.GET("/settings", h.GetSecuritySettings)
		security.PUT("/settings", h.UpdateSecuritySettings)

		// 会话管理
		sessions := security.Group("/sessions")
		{
			sessions.GET("", h.GetActiveSessions)
			sessions.POST("/deactivate", h.DeactivateSession)
			sessions.POST("/deactivate-others", h.DeactivateAllOtherSessions)
		}
	}
}
