package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"payment-platform/admin-service/internal/response"
	"payment-platform/admin-service/internal/service"
)

// SystemConfigHandler 系统配置HTTP处理器
type SystemConfigHandler struct {
	configService service.SystemConfigService
}

// NewSystemConfigHandler 创建系统配置处理器实例
func NewSystemConfigHandler(configService service.SystemConfigService) *SystemConfigHandler {
	return &SystemConfigHandler{
		configService: configService,
	}
}

// RegisterRoutes 注册路由
func (h *SystemConfigHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	configs := r.Group("/system-configs")
	configs.Use(authMiddleware)
	{
		configs.POST("", h.CreateConfig)
		configs.GET("/:id", h.GetConfig)
		configs.GET("/key/:key", h.GetConfigByKey)
		configs.GET("", h.ListConfigs)
		configs.GET("/grouped", h.ListConfigsByCategory)
		configs.PUT("/:id", h.UpdateConfig)
		configs.DELETE("/:id", h.DeleteConfig)
		configs.POST("/batch", h.BatchUpdateConfigs)
	}
}

// CreateConfigRequest 创建配置请求
type CreateConfigRequest struct {
	Key         string `json:"key" binding:"required"`
	Value       string `json:"value" binding:"required"`
	Type        string `json:"type" binding:"required"`
	Category    string `json:"category" binding:"required"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
}

// CreateConfig 创建系统配置
//
//	@Summary		创建系统配置
//	@Description	创建新的系统配置项
//	@Tags			系统配置
//	@Accept			json
//	@Produce		json
//	@Param			config	body		CreateConfigRequest	true	"配置信息"
//	@Success		201		{object}	response.Response
//	@Failure		400		{object}	response.Response
//	@Failure		401		{object}	response.Response
//	@Failure		409		{object}	response.Response
//	@Failure		500		{object}	response.Response
//	@Security		BearerAuth
//	@Router			/system-configs [post]
func (h *SystemConfigHandler) CreateConfig(c *gin.Context) {
	var req CreateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "请求参数错误", err.Error())
		return
	}

	// 获取当前管理员ID
	adminID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未授权")
		return
	}

	config, err := h.configService.CreateConfig(c.Request.Context(), &service.CreateConfigRequest{
		Key:         req.Key,
		Value:       req.Value,
		Type:        req.Type,
		Category:    req.Category,
		Description: req.Description,
		IsPublic:    req.IsPublic,
		UpdatedBy:   adminID.(uuid.UUID),
	})
	if err != nil {
		response.HandleServiceError(c, err, "创建配置失败")
		return
	}

	response.Created(c, "配置创建成功", config)
}

// GetConfig 获取配置详情
func (h *SystemConfigHandler) GetConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "配置ID格式错误", err.Error())
		return
	}

	config, err := h.configService.GetConfig(c.Request.Context(), id)
	if err != nil {
		response.HandleServiceError(c, err, "获取配置失败")
		return
	}

	response.SuccessWithData(c, config)
}

// GetConfigByKey 根据键获取配置
func (h *SystemConfigHandler) GetConfigByKey(c *gin.Context) {
	key := c.Param("key")

	config, err := h.configService.GetConfigByKey(c.Request.Context(), key)
	if err != nil {
		response.HandleServiceError(c, err, "获取配置失败")
		return
	}

	response.SuccessWithData(c, config)
}

// ListConfigs 获取配置列表
//
//	@Summary		获取系统配置列表
//	@Description	分页查询系统配置列表，支持按类别过滤
//	@Tags			系统配置
//	@Accept			json
//	@Produce		json
//	@Param			page		query		int		false	"页码"			default(1)
//	@Param			page_size	query		int		false	"每页数量"			default(20)
//	@Param			category	query		string	false	"配置类别"
//	@Success		200			{object}	response.ListResponse
//	@Failure		401			{object}	response.Response
//	@Failure		500			{object}	response.Response
//	@Security		BearerAuth
//	@Router			/system-configs [get]
func (h *SystemConfigHandler) ListConfigs(c *gin.Context) {
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
	category := c.Query("category")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	configs, total, err := h.configService.ListConfigs(c.Request.Context(), category, page, pageSize)
	if err != nil {
		response.InternalServerError(c, "获取配置列表失败", err.Error())
		return
	}

	response.SuccessList(c, configs, response.NewPagination(page, pageSize, total))
}

// ListConfigsByCategory 按类别分组获取配置
func (h *SystemConfigHandler) ListConfigsByCategory(c *gin.Context) {
	grouped, err := h.configService.ListConfigsByCategory(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "获取配置列表失败", err.Error())
		return
	}

	// 计算总数
	total := 0
	for _, configs := range grouped {
		total += len(configs)
	}

	response.SuccessWithData(c, gin.H{
		"configs": grouped,
		"total":   total,
	})
}

// UpdateConfigRequest 更新配置请求
type UpdateConfigRequest struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Type        string `json:"type"`
	Category    string `json:"category"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
}

// UpdateConfig 更新配置
func (h *SystemConfigHandler) UpdateConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "配置ID格式错误", err.Error())
		return
	}

	var req UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "请求参数错误", err.Error())
		return
	}

	// 获取当前管理员ID
	adminID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未授权")
		return
	}

	config, err := h.configService.UpdateConfig(c.Request.Context(), &service.UpdateConfigRequest{
		ID:          id,
		Key:         req.Key,
		Value:       req.Value,
		Type:        req.Type,
		Category:    req.Category,
		Description: req.Description,
		IsPublic:    req.IsPublic,
		UpdatedBy:   adminID.(uuid.UUID),
	})
	if err != nil {
		response.HandleServiceError(c, err, "更新配置失败")
		return
	}

	response.Success(c, "配置更新成功", config)
}

// DeleteConfig 删除配置
func (h *SystemConfigHandler) DeleteConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "配置ID格式错误", err.Error())
		return
	}

	err = h.configService.DeleteConfig(c.Request.Context(), id)
	if err != nil {
		response.HandleServiceError(c, err, "删除配置失败")
		return
	}

	response.Success(c, "配置删除成功", nil)
}

// BatchUpdateConfigsRequest 批量更新配置请求
type BatchUpdateConfigsRequest struct {
	Configs []struct {
		ID          string `json:"id" binding:"required"`
		Value       string `json:"value" binding:"required"`
		Description string `json:"description"`
	} `json:"configs" binding:"required"`
}

// BatchUpdateConfigs 批量更新配置
func (h *SystemConfigHandler) BatchUpdateConfigs(c *gin.Context) {
	var req BatchUpdateConfigsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "请求参数错误", err.Error())
		return
	}

	// 获取当前管理员ID
	adminID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "未授权")
		return
	}

	// 转换请求
	configs := make([]service.UpdateConfigRequest, 0, len(req.Configs))
	for _, cfg := range req.Configs {
		id, err := uuid.Parse(cfg.ID)
		if err != nil {
			response.BadRequest(c, "配置ID格式错误", err.Error())
			return
		}
		configs = append(configs, service.UpdateConfigRequest{
			ID:          id,
			Value:       cfg.Value,
			Description: cfg.Description,
			UpdatedBy:   adminID.(uuid.UUID),
		})
	}

	if err := h.configService.BatchUpdateConfigs(c.Request.Context(), configs); err != nil {
		response.InternalServerError(c, "批量更新配置失败", err.Error())
		return
	}

	response.Success(c, "配置批量更新成功", nil)
}
