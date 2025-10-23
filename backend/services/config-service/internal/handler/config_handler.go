package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/payment-platform/services/config-service/internal/repository"
	"github.com/payment-platform/services/config-service/internal/service"
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
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}
	config, err := h.configService.CreateConfig(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(config))
}

func (h *ConfigHandler) GetConfigByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("无效的ID"))
		return
	}
	// TODO: Implement GetConfigByID
	c.JSON(http.StatusOK, SuccessResponse(gin.H{"id": id}))
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
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(PageResponse{List: configs, Total: total, Page: query.Page, PageSize: query.PageSize}))
}

func (h *ConfigHandler) UpdateConfig(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	var input service.UpdateConfigInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}
	config, err := h.configService.UpdateConfig(c.Request.Context(), id, &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(config))
}

func (h *ConfigHandler) DeleteConfig(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	if err := h.configService.DeleteConfig(c.Request.Context(), id, "admin"); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(nil))
}

func (h *ConfigHandler) GetConfigHistory(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	history, err := h.configService.GetConfigHistory(c.Request.Context(), id, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(history))
}

func (h *ConfigHandler) CreateFeatureFlag(c *gin.Context) {
	var input service.CreateFeatureFlagInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}
	flag, err := h.configService.CreateFeatureFlag(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(flag))
}

func (h *ConfigHandler) GetFeatureFlag(c *gin.Context) {
	flag, err := h.configService.GetFeatureFlag(c.Request.Context(), c.Param("key"))
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(flag))
}

func (h *ConfigHandler) ListFeatureFlags(c *gin.Context) {
	query := &repository.FeatureFlagQuery{Environment: c.Query("environment")}
	query.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	query.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))
	flags, total, err := h.configService.ListFeatureFlags(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(PageResponse{List: flags, Total: total, Page: query.Page, PageSize: query.PageSize}))
}

func (h *ConfigHandler) IsFeatureEnabled(c *gin.Context) {
	enabled, err := h.configService.IsFeatureEnabled(c.Request.Context(), c.Param("key"), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(gin.H{"enabled": enabled}))
}

func (h *ConfigHandler) DeleteFeatureFlag(c *gin.Context) {
	id, _ := uuid.Parse(c.Param("id"))
	if err := h.configService.DeleteFeatureFlag(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(nil))
}

func (h *ConfigHandler) RegisterService(c *gin.Context) {
	var input service.RegisterServiceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}
	svc, err := h.configService.RegisterService(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(svc))
}

func (h *ConfigHandler) GetService(c *gin.Context) {
	svc, err := h.configService.GetService(c.Request.Context(), c.Param("name"))
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(svc))
}

func (h *ConfigHandler) ListServices(c *gin.Context) {
	services, err := h.configService.ListServices(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(services))
}

func (h *ConfigHandler) UpdateServiceHeartbeat(c *gin.Context) {
	if err := h.configService.UpdateServiceHeartbeat(c.Request.Context(), c.Param("name")); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(nil))
}

func (h *ConfigHandler) DeregisterService(c *gin.Context) {
	if err := h.configService.DeregisterService(c.Request.Context(), c.Param("name")); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}
	c.JSON(http.StatusOK, SuccessResponse(nil))
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
