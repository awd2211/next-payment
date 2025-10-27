package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/payment-platform/pkg/middleware"
	"payment-platform/admin-service/internal/service"
)

// AdminHandler 管理员HTTP处理器
type AdminHandler struct {
	adminService service.AdminService
}

// NewAdminHandler 创建管理员处理器实例
func NewAdminHandler(adminService service.AdminService) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
	}
}

// RegisterRoutes 注册路由
func (h *AdminHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	// 公开路由（不需要认证）
	public := r.Group("/admin")
	{
		public.POST("/login", h.Login)
	}

	// 需要认证的路由
	protected := r.Group("/admin")
	protected.Use(authMiddleware)
	{
		protected.POST("", h.CreateAdmin)
		protected.GET("/:id", h.GetAdmin)
		protected.GET("", h.ListAdmins)
		protected.PUT("/:id", h.UpdateAdmin)
		protected.DELETE("/:id", h.DeleteAdmin)
		protected.POST("/change-password", h.ChangePassword)
		protected.POST("/:id/reset-password", h.ResetPassword)
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 管理员登录
func (h *AdminHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 获取客户端IP
	ip := c.ClientIP()

	resp, err := h.adminService.Login(c.Request.Context(), req.Username, req.Password, ip)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":         resp.Token,
		"refresh_token": resp.RefreshToken,
		"admin":         resp.Admin,
		"expires_in":    resp.ExpiresIn,
	})
}

// CreateAdminRequest 创建管理员请求
type CreateAdminRequest struct {
	Username string   `json:"username" binding:"required"`
	Email    string   `json:"email" binding:"required,email"`
	Password string   `json:"password" binding:"required,min=8"`
	FullName string   `json:"full_name"`
	Phone    string   `json:"phone"`
	RoleIDs  []string `json:"role_ids"`
}

// CreateAdmin 创建管理员
func (h *AdminHandler) CreateAdmin(c *gin.Context) {
	var req CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 转换角色ID
	roleIDs := make([]uuid.UUID, 0, len(req.RoleIDs))
	for _, idStr := range req.RoleIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "角色ID格式错误"})
			return
		}
		roleIDs = append(roleIDs, id)
	}

	admin, err := h.adminService.CreateAdmin(c.Request.Context(), &service.CreateAdminRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
		Phone:    req.Phone,
		RoleIDs:  roleIDs,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": admin})
}

// GetAdmin 获取管理员详情
func (h *AdminHandler) GetAdmin(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID格式错误"})
		return
	}

	admin, err := h.adminService.GetAdmin(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": admin})
}

// ListAdmins 获取管理员列表
func (h *AdminHandler) ListAdmins(c *gin.Context) {
	// 获取分页参数
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
	status := c.Query("status")
	keyword := c.Query("keyword")

	admins, total, err := h.adminService.ListAdmins(c.Request.Context(), page, pageSize, status, keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      admins,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// UpdateAdminRequest 更新管理员请求
type UpdateAdminRequest struct {
	Email    string   `json:"email"`
	FullName string   `json:"full_name"`
	Phone    string   `json:"phone"`
	Status   string   `json:"status"`
	RoleIDs  []string `json:"role_ids"`
}

// UpdateAdmin 更新管理员
func (h *AdminHandler) UpdateAdmin(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID格式错误"})
		return
	}

	var req UpdateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 转换角色ID
	roleIDs := make([]uuid.UUID, 0, len(req.RoleIDs))
	for _, idStr := range req.RoleIDs {
		roleID, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "角色ID格式错误"})
			return
		}
		roleIDs = append(roleIDs, roleID)
	}

	admin, err := h.adminService.UpdateAdmin(c.Request.Context(), &service.UpdateAdminRequest{
		ID:       id,
		Email:    req.Email,
		FullName: req.FullName,
		Phone:    req.Phone,
		Status:   req.Status,
		RoleIDs:  roleIDs,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": admin})
}

// DeleteAdmin 删除管理员
func (h *AdminHandler) DeleteAdmin(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID格式错误"})
		return
	}

	// 检查是否删除自己
	claims, _ := middleware.GetClaims(c)
	if claims.UserID == id {
		c.JSON(http.StatusForbidden, gin.H{"error": "不能删除自己"})
		return
	}

	if err := h.adminService.DeleteAdmin(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// ChangePassword 修改密码
func (h *AdminHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 获取当前登录用户ID
	claims, err := middleware.GetClaims(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	if err := h.adminService.ChangePassword(c.Request.Context(), claims.UserID, req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// ResetPassword 重置密码（管理员为其他用户重置密码）
func (h *AdminHandler) ResetPassword(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID格式错误"})
		return
	}

	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 获取当前登录用户ID
	claims, err := middleware.GetClaims(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 检查是否重置自己的密码（应该使用change-password接口）
	if claims.UserID == id {
		c.JSON(http.StatusForbidden, gin.H{"error": "不能重置自己的密码，请使用修改密码功能"})
		return
	}

	if err := h.adminService.ResetPassword(c.Request.Context(), id, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码重置成功"})
}
