package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/payment-platform/pkg/errors"
	"github.com/payment-platform/pkg/middleware"
	"payment-platform/config-service/internal/repository"
	"payment-platform/config-service/internal/service"
)

type ConfigHandler struct {
	configService service.ConfigService
}

func NewConfigHandler(configService service.ConfigService) *ConfigHandler {
	return &ConfigHandler{configService: configService}
}

func (h *ConfigHandler) RegisterRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		configs := v1.Group("/configs")
		{
			configs.POST("", h.CreateConfig)
			configs.GET("", h.ListConfigs)
			configs.GET("/:id", h.GetConfigByID)
			configs.PUT("/:id", h.UpdateConfig)
			configs.DELETE("/:id", h.DeleteConfig)
			configs.GET("/:id/history", h.GetConfigHistory)
			configs.POST("/:id/rollback", h.RollbackConfig)
		}

		flags := v1.Group("/feature-flags")
		{
			flags.POST("", h.CreateFeatureFlag)
			flags.GET("", h.ListFeatureFlags)
			flags.GET("/:key", h.GetFeatureFlag)
			flags.GET("/:key/enabled", h.IsFeatureEnabled)
			flags.DELETE("/:id", h.DeleteFeatureFlag)
		}

		services := v1.Group("/services")
		{
			services.POST("/register", h.RegisterService)
			services.GET("", h.ListServices)
			services.GET("/:name", h.GetService)
			services.POST("/:name/heartbeat", h.UpdateServiceHeartbeat)
			services.POST("/:name/deregister", h.DeregisterService)
		}
	}
}

func (h *ConfigHandler) CreateConfig(c *gin.Context) {
	var input service.CreateConfigInput
	if err := c.ShouldBindJSON(&input); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}
	config, err := h.configService.CreateConfig(c.Request.Context(), &input)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "创建配置失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}
	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(config).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

func (h *ConfigHandler) GetConfigByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的ID", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	config, err := h.configService.GetConfigByID(c.Request.Context(), id)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "获取配置失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(config).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

func (h *ConfigHandler) ListConfigs(c *gin.Context) {
	query := &repository.ConfigQuery{
		ServiceName: c.Query("service_name"),
		Environment: c.Query("environment"),
		ConfigKey:   c.Query("config_key"),
	}
	query.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	query.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))

	configs, total, err := h.configService.ListConfigs(c.Request.Context(), query)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "查询配置列表失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}
	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(PageResponse{List: configs, Total: total, Page: query.Page, PageSize: query.PageSize}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

func (h *ConfigHandler) UpdateConfig(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var input service.UpdateConfigInput
	if err := c.ShouldBindJSON(&input); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}
	config, err := h.configService.UpdateConfig(c.Request.Context(), id, &input)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "更新配置失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}
	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(config).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

func (h *ConfigHandler) DeleteConfig(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	if err := h.configService.DeleteConfig(c.Request.Context(), id, "admin"); err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "删除配置失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}
	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(nil).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

func (h *ConfigHandler) GetConfigHistory(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	history, err := h.configService.GetConfigHistory(c.Request.Context(), id, limit)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "获取配置历史失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}
	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(history).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

func (h *ConfigHandler) RollbackConfig(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "无效的ID", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var input struct {
		TargetVersion int    `json:"target_version" binding:"required"`
		RolledBy      string `json:"rolled_by" binding:"required"`
		Reason        string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	config, err := h.configService.RollbackConfig(c.Request.Context(), id, input.TargetVersion, input.RolledBy, input.Reason)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "回滚配置失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(config).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

func (h *ConfigHandler) CreateFeatureFlag(c *gin.Context) {
	var input service.CreateFeatureFlagInput
	if err := c.ShouldBindJSON(&input); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}
	flag, err := h.configService.CreateFeatureFlag(c.Request.Context(), &input)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "创建功能标志失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}
	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(flag).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

func (h *ConfigHandler) GetFeatureFlag(c *gin.Context) {
	flag, err := h.configService.GetFeatureFlag(c.Request.Context(), c.Param("key"))
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeResourceNotFound, "功能标志不存在", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusNotFound, resp)
		}
		return
	}
	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(flag).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

func (h *ConfigHandler) ListFeatureFlags(c *gin.Context) {
	query := &repository.FeatureFlagQuery{Environment: c.Query("environment")}
	query.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	query.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))
	flags, total, err := h.configService.ListFeatureFlags(c.Request.Context(), query)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "查询功能标志列表失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}
	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(PageResponse{List: flags, Total: total, Page: query.Page, PageSize: query.PageSize}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

func (h *ConfigHandler) IsFeatureEnabled(c *gin.Context) {
	enabled, err := h.configService.IsFeatureEnabled(c.Request.Context(), c.Param("key"), nil)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "检查功能标志失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}
	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(gin.H{"enabled": enabled}).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

func (h *ConfigHandler) DeleteFeatureFlag(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	if err := h.configService.DeleteFeatureFlag(c.Request.Context(), id); err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "删除功能标志失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}
	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(nil).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

func (h *ConfigHandler) RegisterService(c *gin.Context) {
	var input service.RegisterServiceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		traceID := middleware.GetRequestID(c)
		resp := errors.NewErrorResponse(errors.ErrCodeInvalidRequest, "请求参数错误", err.Error()).WithTraceID(traceID)
		c.JSON(http.StatusBadRequest, resp)
		return
	}
	svc, err := h.configService.RegisterService(c.Request.Context(), &input)
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "注册服务失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}
	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(svc).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

func (h *ConfigHandler) GetService(c *gin.Context) {
	svc, err := h.configService.GetService(c.Request.Context(), c.Param("name"))
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeResourceNotFound, "服务不存在", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusNotFound, resp)
		}
		return
	}
	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(svc).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

func (h *ConfigHandler) ListServices(c *gin.Context) {
	services, err := h.configService.ListServices(c.Request.Context())
	if err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "查询服务列表失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}
	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(services).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

func (h *ConfigHandler) UpdateServiceHeartbeat(c *gin.Context) {
	if err := h.configService.UpdateServiceHeartbeat(c.Request.Context(), c.Param("name")); err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "更新服务心跳失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}
	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(nil).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

func (h *ConfigHandler) DeregisterService(c *gin.Context) {
	if err := h.configService.DeregisterService(c.Request.Context(), c.Param("name")); err != nil {
		traceID := middleware.GetRequestID(c)
		if bizErr, ok := errors.GetBusinessError(err); ok {
			resp := errors.NewErrorResponseFromBusinessError(bizErr).WithTraceID(traceID)
			c.JSON(errors.GetHTTPStatus(bizErr.Code), resp)
		} else {
			resp := errors.NewErrorResponse(errors.ErrCodeInternalError, "注销服务失败", err.Error()).WithTraceID(traceID)
			c.JSON(http.StatusInternalServerError, resp)
		}
		return
	}
	traceID := middleware.GetRequestID(c)
	resp := errors.NewSuccessResponse(nil).WithTraceID(traceID)
	c.JSON(http.StatusOK, resp)
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type PageResponse struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

func SuccessResponse(data interface{}) Response {
	return Response{Code: 0, Message: "success", Data: data}
}

func ErrorResponse(message string) Response {
	return Response{Code: -1, Message: message}
}
