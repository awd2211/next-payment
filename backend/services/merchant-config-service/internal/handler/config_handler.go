package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"payment-platform/merchant-config-service/internal/service"
)

// ConfigHandler 配置处理器（统一管理3个配置服务）
type ConfigHandler struct {
	feeService   service.FeeConfigService
	limitService service.TransactionLimitService
	channelService service.ChannelConfigService
}

// NewConfigHandler 创建配置处理器实例
func NewConfigHandler(
	feeService service.FeeConfigService,
	limitService service.TransactionLimitService,
	channelService service.ChannelConfigService,
) *ConfigHandler {
	return &ConfigHandler{
		feeService:   feeService,
		limitService: limitService,
		channelService: channelService,
	}
}

// Response 统一响应格式
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// RegisterRoutes 注册路由
func (h *ConfigHandler) RegisterRoutes(r *gin.RouterGroup) {
	// 费率配置路由
	feeRoutes := r.Group("/fee-configs")
	{
		feeRoutes.POST("", h.CreateFeeConfig)
		feeRoutes.GET("/:id", h.GetFeeConfig)
		feeRoutes.GET("/merchant/:merchant_id", h.ListMerchantFeeConfigs)
		feeRoutes.PUT("/:id", h.UpdateFeeConfig)
		feeRoutes.DELETE("/:id", h.DeleteFeeConfig)
		feeRoutes.POST("/:id/approve", h.ApproveFeeConfig)
		feeRoutes.POST("/calculate-fee", h.CalculateFee)
	}

	// 交易限额路由
	limitRoutes := r.Group("/transaction-limits")
	{
		limitRoutes.POST("", h.CreateLimit)
		limitRoutes.GET("/:id", h.GetLimit)
		limitRoutes.GET("/merchant/:merchant_id", h.ListMerchantLimits)
		limitRoutes.PUT("/:id", h.UpdateLimit)
		limitRoutes.DELETE("/:id", h.DeleteLimit)
		limitRoutes.POST("/check-limit", h.CheckLimit)
	}

	// 渠道配置路由
	channelRoutes := r.Group("/channel-configs")
	{
		channelRoutes.POST("", h.CreateChannelConfig)
		channelRoutes.GET("/:id", h.GetChannelConfig)
		channelRoutes.GET("/merchant/:merchant_id", h.ListMerchantChannels)
		channelRoutes.GET("/merchant/:merchant_id/channel/:channel", h.GetMerchantChannel)
		channelRoutes.PUT("/:id", h.UpdateChannelConfig)
		channelRoutes.DELETE("/:id", h.DeleteChannelConfig)
		channelRoutes.POST("/:id/enable", h.EnableChannel)
		channelRoutes.POST("/:id/disable", h.DisableChannel)
	}
}

// ========== 费率配置处理器 ==========

// CreateFeeConfig 创建费率配置
func (h *ConfigHandler) CreateFeeConfig(c *gin.Context) {
	var input service.CreateFeeConfigInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 1, Message: err.Error()})
		return
	}

	config, err := h.feeService.CreateFeeConfig(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 1, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, Response{Code: 0, Message: "success", Data: config})
}

// GetFeeConfig 获取费率配置
func (h *ConfigHandler) GetFeeConfig(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 1, Message: "invalid id"})
		return
	}

	config, err := h.feeService.GetFeeConfig(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, Response{Code: 1, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, Response{Code: 0, Message: "success", Data: config})
}

// ListMerchantFeeConfigs 列出商户费率配置
func (h *ConfigHandler) ListMerchantFeeConfigs(c *gin.Context) {
	merchantID, err := uuid.Parse(c.Param("merchant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 1, Message: "invalid merchant_id"})
		return
	}

	configs, err := h.feeService.ListMerchantFeeConfigs(c.Request.Context(), merchantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 1, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, Response{Code: 0, Message: "success", Data: configs})
}

// UpdateFeeConfig 更新费率配置
func (h *ConfigHandler) UpdateFeeConfig(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 1, Message: "invalid id"})
		return
	}

	var input service.UpdateFeeConfigInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 1, Message: err.Error()})
		return
	}

	config, err := h.feeService.UpdateFeeConfig(c.Request.Context(), id, &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 1, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, Response{Code: 0, Message: "success", Data: config})
}

// DeleteFeeConfig 删除费率配置
func (h *ConfigHandler) DeleteFeeConfig(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 1, Message: "invalid id"})
		return
	}

	if err := h.feeService.DeleteFeeConfig(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 1, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, Response{Code: 0, Message: "deleted successfully"})
}

// ApproveFeeConfig 审批费率配置
func (h *ConfigHandler) ApproveFeeConfig(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 1, Message: "invalid id"})
		return
	}

	// TODO: 从 JWT token 获取 approver_id
	approverID := uuid.New()

	if err := h.feeService.ApproveFeeConfig(c.Request.Context(), id, approverID); err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 1, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, Response{Code: 0, Message: "approved successfully"})
}

// CalculateFee 计算手续费
func (h *ConfigHandler) CalculateFee(c *gin.Context) {
	var req struct {
		MerchantID    uuid.UUID `json:"merchant_id"`
		Channel       string    `json:"channel"`
		PaymentMethod string    `json:"payment_method"`
		Amount        int64     `json:"amount"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 1, Message: err.Error()})
		return
	}

	fee, err := h.feeService.CalculateFee(c.Request.Context(), req.MerchantID, req.Channel, req.PaymentMethod, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 1, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, Response{Code: 0, Message: "success", Data: gin.H{"fee": fee}})
}

// ========== 交易限额处理器 ==========

// CreateLimit 创建交易限额
func (h *ConfigHandler) CreateLimit(c *gin.Context) {
	var input service.CreateLimitInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 1, Message: err.Error()})
		return
	}

	// 设置默认生效时间
	if input.EffectiveDate.IsZero() {
		input.EffectiveDate = time.Now()
	}

	limit, err := h.limitService.CreateLimit(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 1, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, Response{Code: 0, Message: "success", Data: limit})
}

// GetLimit 获取交易限额
func (h *ConfigHandler) GetLimit(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 1, Message: "invalid id"})
		return
	}

	limit, err := h.limitService.GetLimit(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, Response{Code: 1, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, Response{Code: 0, Message: "success", Data: limit})
}

// ListMerchantLimits 列出商户交易限额
func (h *ConfigHandler) ListMerchantLimits(c *gin.Context) {
	merchantID, err := uuid.Parse(c.Param("merchant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 1, Message: "invalid merchant_id"})
		return
	}

	limits, err := h.limitService.ListMerchantLimits(c.Request.Context(), merchantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 1, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, Response{Code: 0, Message: "success", Data: limits})
}

// UpdateLimit 更新交易限额
func (h *ConfigHandler) UpdateLimit(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 1, Message: "invalid id"})
		return
	}

	var input service.UpdateLimitInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 1, Message: err.Error()})
		return
	}

	limit, err := h.limitService.UpdateLimit(c.Request.Context(), id, &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 1, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, Response{Code: 0, Message: "success", Data: limit})
}

// DeleteLimit 删除交易限额
func (h *ConfigHandler) DeleteLimit(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 1, Message: "invalid id"})
		return
	}

	if err := h.limitService.DeleteLimit(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 1, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, Response{Code: 0, Message: "deleted successfully"})
}

// CheckLimit 检查交易限额
func (h *ConfigHandler) CheckLimit(c *gin.Context) {
	var req struct {
		MerchantID    uuid.UUID `json:"merchant_id"`
		Channel       string    `json:"channel"`
		PaymentMethod string    `json:"payment_method"`
		Amount        int64     `json:"amount"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 1, Message: err.Error()})
		return
	}

	if err := h.limitService.CheckLimit(c.Request.Context(), req.MerchantID, req.Channel, req.PaymentMethod, req.Amount); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 1, Message: err.Error(), Data: gin.H{"passed": false}})
		return
	}

	c.JSON(http.StatusOK, Response{Code: 0, Message: "success", Data: gin.H{"passed": true}})
}

// ========== 渠道配置处理器 ==========

// CreateChannelConfig 创建渠道配置
func (h *ConfigHandler) CreateChannelConfig(c *gin.Context) {
	var input service.CreateChannelConfigInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 1, Message: err.Error()})
		return
	}

	config, err := h.channelService.CreateChannelConfig(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 1, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, Response{Code: 0, Message: "success", Data: config})
}

// GetChannelConfig 获取渠道配置
func (h *ConfigHandler) GetChannelConfig(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 1, Message: "invalid id"})
		return
	}

	config, err := h.channelService.GetChannelConfig(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, Response{Code: 1, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, Response{Code: 0, Message: "success", Data: config})
}

// ListMerchantChannels 列出商户渠道配置
func (h *ConfigHandler) ListMerchantChannels(c *gin.Context) {
	merchantID, err := uuid.Parse(c.Param("merchant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 1, Message: "invalid merchant_id"})
		return
	}

	configs, err := h.channelService.ListMerchantChannels(c.Request.Context(), merchantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 1, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, Response{Code: 0, Message: "success", Data: configs})
}

// GetMerchantChannel 获取商户指定渠道配置
func (h *ConfigHandler) GetMerchantChannel(c *gin.Context) {
	merchantID, err := uuid.Parse(c.Param("merchant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 1, Message: "invalid merchant_id"})
		return
	}

	channel := c.Param("channel")
	if channel == "" {
		c.JSON(http.StatusBadRequest, Response{Code: 1, Message: "channel is required"})
		return
	}

	config, err := h.channelService.GetMerchantChannel(c.Request.Context(), merchantID, channel)
	if err != nil {
		c.JSON(http.StatusNotFound, Response{Code: 1, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, Response{Code: 0, Message: "success", Data: config})
}

// UpdateChannelConfig 更新渠道配置
func (h *ConfigHandler) UpdateChannelConfig(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 1, Message: "invalid id"})
		return
	}

	var input service.UpdateChannelConfigInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 1, Message: err.Error()})
		return
	}

	config, err := h.channelService.UpdateChannelConfig(c.Request.Context(), id, &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 1, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, Response{Code: 0, Message: "success", Data: config})
}

// DeleteChannelConfig 删除渠道配置
func (h *ConfigHandler) DeleteChannelConfig(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 1, Message: "invalid id"})
		return
	}

	if err := h.channelService.DeleteChannelConfig(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 1, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, Response{Code: 0, Message: "deleted successfully"})
}

// EnableChannel 启用渠道
func (h *ConfigHandler) EnableChannel(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 1, Message: "invalid id"})
		return
	}

	if err := h.channelService.EnableChannel(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 1, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, Response{Code: 0, Message: "enabled successfully"})
}

// DisableChannel 停用渠道
func (h *ConfigHandler) DisableChannel(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{Code: 1, Message: "invalid id"})
		return
	}

	if err := h.channelService.DisableChannel(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, Response{Code: 1, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, Response{Code: 0, Message: "disabled successfully"})
}
