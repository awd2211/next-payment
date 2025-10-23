package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/payment-platform/services/admin-service/internal/service"
)

// PreferencesHandler 用户偏好设置处理器
type PreferencesHandler struct {
	prefsService service.PreferencesService
}

// NewPreferencesHandler 创建用户偏好设置处理器实例
func NewPreferencesHandler(prefsService service.PreferencesService) *PreferencesHandler {
	return &PreferencesHandler{
		prefsService: prefsService,
	}
}

// GetPreferences 获取用户偏好设置
// @Summary 获取用户偏好设置
// @Tags Preferences
// @Produce json
// @Success 200 {object} Response
// @Router /api/v1/preferences [get]
func (h *PreferencesHandler) GetPreferences(c *gin.Context) {
	userID, _ := c.Get("user_id")
	userType, _ := c.Get("user_type")

	prefs, err := h.prefsService.GetPreferences(
		c.Request.Context(),
		userID.(uuid.UUID),
		userType.(string),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": prefs,
	})
}

// UpdatePreferences 更新用户偏好设置
// @Summary 更新用户偏好设置
// @Tags Preferences
// @Accept json
// @Produce json
// @Param request body service.UpdatePreferencesInput true "更新偏好设置请求"
// @Success 200 {object} Response
// @Router /api/v1/preferences [put]
func (h *PreferencesHandler) UpdatePreferences(c *gin.Context) {
	var req service.UpdatePreferencesInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	userType, _ := c.Get("user_type")

	err := h.prefsService.UpdatePreferences(
		c.Request.Context(),
		userID.(uuid.UUID),
		userType.(string),
		&req,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 返回更新后的偏好设置
	prefs, _ := h.prefsService.GetPreferences(
		c.Request.Context(),
		userID.(uuid.UUID),
		userType.(string),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "偏好设置已更新",
		"data":    prefs,
	})
}

// RegisterRoutes 注册路由
func (h *PreferencesHandler) RegisterRoutes(r *gin.RouterGroup) {
	prefs := r.Group("/preferences")
	{
		prefs.GET("", h.GetPreferences)
		prefs.PUT("", h.UpdatePreferences)
	}
}
