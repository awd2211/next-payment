package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"payment-platform/admin-service/internal/service"
)

// RoleHandler 角色HTTP处理器
type RoleHandler struct {
	roleService service.RoleService
}

// NewRoleHandler 创建角色处理器实例
func NewRoleHandler(roleService service.RoleService) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
	}
}

// RegisterRoutes 注册路由
func (h *RoleHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	roles := r.Group("/roles")
	roles.Use(authMiddleware)
	{
		roles.POST("", h.CreateRole)
		roles.GET("/:id", h.GetRole)
		roles.GET("", h.ListRoles)
		roles.PUT("/:id", h.UpdateRole)
		roles.DELETE("/:id", h.DeleteRole)
		roles.POST("/:id/permissions", h.AssignPermissions)
		roles.POST("/assign", h.AssignRoleToAdmin)
	}
}

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name          string   `json:"name" binding:"required"`
	DisplayName   string   `json:"display_name" binding:"required"`
	Description   string   `json:"description"`
	PermissionIDs []string `json:"permission_ids"`
}

// CreateRole 创建角色
func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 转换权限ID
	permissionIDs := make([]uuid.UUID, 0, len(req.PermissionIDs))
	for _, idStr := range req.PermissionIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "权限ID格式错误"})
			return
		}
		permissionIDs = append(permissionIDs, id)
	}

	role, err := h.roleService.CreateRole(c.Request.Context(), &service.CreateRoleRequest{
		Name:          req.Name,
		DisplayName:   req.DisplayName,
		Description:   req.Description,
		PermissionIDs: permissionIDs,
	})
	if err != nil {
		if err == service.ErrRoleExists {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建角色失败", "details": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "角色创建成功",
		"data":    role,
	})
}

// GetRole 获取角色详情
func (h *RoleHandler) GetRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "角色ID格式错误"})
		return
	}

	role, err := h.roleService.GetRole(c.Request.Context(), id)
	if err != nil {
		if err == service.ErrRoleNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取角色失败", "details": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": role,
	})
}

// ListRoles 获取角色列表
func (h *RoleHandler) ListRoles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	roles, total, err := h.roleService.ListRoles(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取角色列表失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": roles,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	DisplayName   string   `json:"display_name"`
	Description   string   `json:"description"`
	PermissionIDs []string `json:"permission_ids"`
}

// UpdateRole 更新角色
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "角色ID格式错误"})
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 转换权限ID
	permissionIDs := make([]uuid.UUID, 0, len(req.PermissionIDs))
	for _, idStr := range req.PermissionIDs {
		permID, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "权限ID格式错误"})
			return
		}
		permissionIDs = append(permissionIDs, permID)
	}

	role, err := h.roleService.UpdateRole(c.Request.Context(), &service.UpdateRoleRequest{
		ID:            id,
		DisplayName:   req.DisplayName,
		Description:   req.Description,
		PermissionIDs: permissionIDs,
	})
	if err != nil {
		if err == service.ErrRoleNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新角色失败", "details": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "角色更新成功",
		"data":    role,
	})
}

// DeleteRole 删除角色
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "角色ID格式错误"})
		return
	}

	err = h.roleService.DeleteRole(c.Request.Context(), id)
	if err != nil {
		if err == service.ErrRoleNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if err == service.ErrSystemRoleProtect {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除角色失败", "details": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "角色删除成功",
	})
}

// AssignPermissionsRequest 分配权限请求
type AssignPermissionsRequest struct {
	PermissionIDs []string `json:"permission_ids" binding:"required"`
}

// AssignPermissions 为角色分配权限
func (h *RoleHandler) AssignPermissions(c *gin.Context) {
	idStr := c.Param("id")
	roleID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "角色ID格式错误"})
		return
	}

	var req AssignPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 转换权限ID
	permissionIDs := make([]uuid.UUID, 0, len(req.PermissionIDs))
	for _, idStr := range req.PermissionIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "权限ID格式错误"})
			return
		}
		permissionIDs = append(permissionIDs, id)
	}

	err = h.roleService.AssignPermissions(c.Request.Context(), roleID, permissionIDs)
	if err != nil {
		if err == service.ErrRoleNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "分配权限失败", "details": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "权限分配成功",
	})
}

// AssignRoleToAdminRequest 为管理员分配角色请求
type AssignRoleToAdminRequest struct {
	AdminID string `json:"admin_id" binding:"required"`
	RoleID  string `json:"role_id" binding:"required"`
}

// AssignRoleToAdmin 为管理员分配角色
func (h *RoleHandler) AssignRoleToAdmin(c *gin.Context) {
	var req AssignRoleToAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	adminID, err := uuid.Parse(req.AdminID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "管理员ID格式错误"})
		return
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "角色ID格式错误"})
		return
	}

	err = h.roleService.AssignRoleToAdmin(c.Request.Context(), adminID, roleID)
	if err != nil {
		if err == service.ErrAdminNotFound || err == service.ErrRoleNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "分配角色失败", "details": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "角色分配成功",
	})
}
