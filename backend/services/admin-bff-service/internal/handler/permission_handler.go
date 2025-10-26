package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"payment-platform/admin-service/internal/service"
)

// PermissionHandler 权限HTTP处理器
type PermissionHandler struct {
	permissionService service.PermissionService
}

// NewPermissionHandler 创建权限处理器实例
func NewPermissionHandler(permissionService service.PermissionService) *PermissionHandler {
	return &PermissionHandler{
		permissionService: permissionService,
	}
}

// RegisterRoutes 注册路由
func (h *PermissionHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	permissions := r.Group("/permissions")
	permissions.Use(authMiddleware)
	{
		permissions.GET("/:id", h.GetPermission)
		permissions.GET("", h.ListPermissions)
		permissions.GET("/grouped", h.ListPermissionsByResource)
	}
}

// GetPermission 获取权限详情
func (h *PermissionHandler) GetPermission(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "权限ID格式错误"})
		return
	}

	permission, err := h.permissionService.GetPermission(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取权限失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": permission,
	})
}

// ListPermissions 获取权限列表
func (h *PermissionHandler) ListPermissions(c *gin.Context) {
	resource := c.Query("resource")

	permissions, err := h.permissionService.ListPermissions(c.Request.Context(), resource)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取权限列表失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  permissions,
		"total": len(permissions),
	})
}

// ListPermissionsByResource 按资源分组获取权限列表
func (h *PermissionHandler) ListPermissionsByResource(c *gin.Context) {
	grouped, err := h.permissionService.ListPermissionsByResource(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取权限列表失败", "details": err.Error()})
		return
	}

	// 计算总数
	total := 0
	for _, perms := range grouped {
		total += len(perms)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  grouped,
		"total": total,
	})
}
