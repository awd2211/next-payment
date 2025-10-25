package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"payment-platform/merchant-limit-service/internal/service"
)

// LimitHandler 额度HTTP处理器
type LimitHandler struct {
	service service.LimitService
}

// NewLimitHandler 创建处理器实例
func NewLimitHandler(service service.LimitService) *LimitHandler {
	return &LimitHandler{
		service: service,
	}
}

// RegisterRoutes 注册路由
func (h *LimitHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Tier management (Admin only)
	tiers := router.Group("/tiers")
	{
		tiers.POST("", h.CreateTier)
		tiers.GET("", h.ListTiers)
		tiers.GET("/:tier_id", h.GetTier)
		tiers.PUT("/:tier_id", h.UpdateTier)
		tiers.DELETE("/:tier_id", h.DeleteTier)
	}

	// Merchant limit management
	limits := router.Group("/limits")
	{
		limits.POST("/initialize", h.InitializeLimit)
		limits.GET("/:merchant_id", h.GetMerchantLimit)
		limits.PUT("/:merchant_id", h.UpdateMerchantLimit)
		limits.POST("/:merchant_id/change-tier", h.ChangeTier)
		limits.POST("/:merchant_id/suspend", h.SuspendMerchant)
		limits.POST("/:merchant_id/unsuspend", h.UnsuspendMerchant)

		// Limit enforcement (Internal API)
		limits.POST("/check", h.CheckLimit)
		limits.POST("/consume", h.ConsumeLimit)
		limits.POST("/release", h.ReleaseLimit)

		// Usage history
		limits.GET("/:merchant_id/usage-history", h.GetUsageHistory)
		limits.GET("/:merchant_id/statistics", h.GetStatistics)
	}
}

// Tier Management Handlers

// CreateTier 创建等级
func (h *LimitHandler) CreateTier(c *gin.Context) {
	var req service.CreateTierInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_REQUEST", err.Error()))
		return
	}

	tier, err := h.service.CreateTier(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("CREATE_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(tier))
}

// GetTier 获取等级
func (h *LimitHandler) GetTier(c *gin.Context) {
	tierID, err := uuid.Parse(c.Param("tier_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_TIER_ID", "Invalid tier ID format"))
		return
	}

	tier, err := h.service.GetTier(c.Request.Context(), tierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("GET_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(tier))
}

// ListTiers 查询所有等级
func (h *LimitHandler) ListTiers(c *gin.Context) {
	tiers, err := h.service.ListTiers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("LIST_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(tiers))
}

// UpdateTier 更新等级
func (h *LimitHandler) UpdateTier(c *gin.Context) {
	tierID, err := uuid.Parse(c.Param("tier_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_TIER_ID", "Invalid tier ID format"))
		return
	}

	var req service.UpdateTierInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_REQUEST", err.Error()))
		return
	}

	tier, err := h.service.UpdateTier(c.Request.Context(), tierID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("UPDATE_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(tier))
}

// DeleteTier 删除等级
func (h *LimitHandler) DeleteTier(c *gin.Context) {
	tierID, err := uuid.Parse(c.Param("tier_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_TIER_ID", "Invalid tier ID format"))
		return
	}

	if err := h.service.DeleteTier(c.Request.Context(), tierID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("DELETE_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "Tier deleted successfully"}))
}

// Limit Management Handlers

// InitializeLimit 初始化商户额度
func (h *LimitHandler) InitializeLimit(c *gin.Context) {
	var req struct {
		MerchantID string `json:"merchant_id" binding:"required"`
		TierID     string `json:"tier_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_REQUEST", err.Error()))
		return
	}

	merchantID, err := uuid.Parse(req.MerchantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_MERCHANT_ID", "Invalid merchant_id format"))
		return
	}

	tierID, err := uuid.Parse(req.TierID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_TIER_ID", "Invalid tier_id format"))
		return
	}

	limit, err := h.service.InitializeMerchantLimit(c.Request.Context(), merchantID, tierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("INITIALIZE_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(limit))
}

// GetMerchantLimit 获取商户额度
func (h *LimitHandler) GetMerchantLimit(c *gin.Context) {
	merchantID, err := uuid.Parse(c.Param("merchant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_MERCHANT_ID", "Invalid merchant ID format"))
		return
	}

	details, err := h.service.GetMerchantLimit(c.Request.Context(), merchantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("GET_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(details))
}

// UpdateMerchantLimit 更新商户额度
func (h *LimitHandler) UpdateMerchantLimit(c *gin.Context) {
	merchantID, err := uuid.Parse(c.Param("merchant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_MERCHANT_ID", "Invalid merchant ID format"))
		return
	}

	var req service.UpdateLimitInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_REQUEST", err.Error()))
		return
	}

	limit, err := h.service.UpdateMerchantLimit(c.Request.Context(), merchantID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("UPDATE_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(limit))
}

// ChangeTier 更改商户等级
func (h *LimitHandler) ChangeTier(c *gin.Context) {
	merchantID, err := uuid.Parse(c.Param("merchant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_MERCHANT_ID", "Invalid merchant ID format"))
		return
	}

	var req struct {
		NewTierID string `json:"new_tier_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_REQUEST", err.Error()))
		return
	}

	newTierID, err := uuid.Parse(req.NewTierID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_TIER_ID", "Invalid tier_id format"))
		return
	}

	if err := h.service.ChangeMerchantTier(c.Request.Context(), merchantID, newTierID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("CHANGE_TIER_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "Tier changed successfully"}))
}

// SuspendMerchant 暂停商户
func (h *LimitHandler) SuspendMerchant(c *gin.Context) {
	merchantID, err := uuid.Parse(c.Param("merchant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_MERCHANT_ID", "Invalid merchant ID format"))
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_REQUEST", err.Error()))
		return
	}

	if err := h.service.SuspendMerchant(c.Request.Context(), merchantID, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("SUSPEND_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "Merchant suspended successfully"}))
}

// UnsuspendMerchant 恢复商户
func (h *LimitHandler) UnsuspendMerchant(c *gin.Context) {
	merchantID, err := uuid.Parse(c.Param("merchant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_MERCHANT_ID", "Invalid merchant ID format"))
		return
	}

	if err := h.service.UnsuspendMerchant(c.Request.Context(), merchantID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("UNSUSPEND_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "Merchant unsuspended successfully"}))
}

// Limit Enforcement Handlers (Internal API)

// CheckLimit 检查额度
func (h *LimitHandler) CheckLimit(c *gin.Context) {
	var req struct {
		MerchantID string `json:"merchant_id" binding:"required"`
		Amount     int64  `json:"amount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_REQUEST", err.Error()))
		return
	}

	merchantID, err := uuid.Parse(req.MerchantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_MERCHANT_ID", "Invalid merchant_id format"))
		return
	}

	result, err := h.service.CheckLimit(c.Request.Context(), merchantID, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("CHECK_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(result))
}

// ConsumeLimit 消费额度
func (h *LimitHandler) ConsumeLimit(c *gin.Context) {
	var req service.ConsumeLimitInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_REQUEST", err.Error()))
		return
	}

	if err := h.service.ConsumeLimit(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("CONSUME_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "Limit consumed successfully"}))
}

// ReleaseLimit 释放额度
func (h *LimitHandler) ReleaseLimit(c *gin.Context) {
	var req service.ReleaseLimitInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_REQUEST", err.Error()))
		return
	}

	if err := h.service.ReleaseLimit(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("RELEASE_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "Limit released successfully"}))
}

// Usage History Handlers

// GetUsageHistory 获取使用历史
func (h *LimitHandler) GetUsageHistory(c *gin.Context) {
	merchantID, err := uuid.Parse(c.Param("merchant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_MERCHANT_ID", "Invalid merchant ID format"))
		return
	}

	var startTime, endTime *time.Time
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		st, err := time.Parse("2006-01-02", startTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_DATE", "Invalid start_time format"))
			return
		}
		startTime = &st
	}

	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		et, err := time.Parse("2006-01-02", endTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_DATE", "Invalid end_time format"))
			return
		}
		endTime = &et
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	result, err := h.service.GetUsageHistory(c.Request.Context(), merchantID, startTime, endTime, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("GET_HISTORY_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(result))
}

// GetStatistics 获取统计信息
func (h *LimitHandler) GetStatistics(c *gin.Context) {
	merchantID, err := uuid.Parse(c.Param("merchant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_MERCHANT_ID", "Invalid merchant ID format"))
		return
	}

	stats, err := h.service.GetStatistics(c.Request.Context(), merchantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("GET_STATS_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(stats))
}

// Response helpers

type Response struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	TraceID string      `json:"trace_id,omitempty"`
}

func SuccessResponse(data interface{}) Response {
	return Response{
		Code:    "SUCCESS",
		Message: "操作成功",
		Data:    data,
	}
}

func ErrorResponse(code, message string) Response {
	return Response{
		Code:    code,
		Message: message,
	}
}
