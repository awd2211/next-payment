package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"payment-platform/merchant-quota-service/internal/service"
)

// QuotaHandler 配额处理器
type QuotaHandler struct {
	quotaService service.QuotaService
}

// NewQuotaHandler 创建配额处理器实例
func NewQuotaHandler(quotaService service.QuotaService) *QuotaHandler {
	return &QuotaHandler{
		quotaService: quotaService,
	}
}

// InitializeQuota godoc
// @Summary 初始化商户配额
// @Description 为商户创建配额记录（新商户注册时调用）
// @Tags Quota
// @Accept json
// @Produce json
// @Param body body service.InitializeQuotaInput true "初始化配额请求"
// @Success 200 {object} SuccessResponse "成功返回配额信息"
// @Failure 400 {object} ErrorResponse "参数错误"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Security BearerAuth
// @Router /quotas/initialize [post]
func (h *QuotaHandler) InitializeQuota(c *gin.Context) {
	var input service.InitializeQuotaInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	quota, err := h.quotaService.InitializeQuota(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    0,
		Message: "quota initialized successfully",
		Data:    quota,
	})
}

// ConsumeQuota godoc
// @Summary 消耗配额
// @Description 交易时调用，扣减商户配额（使用乐观锁）
// @Tags Quota
// @Accept json
// @Produce json
// @Param body body service.ConsumeQuotaInput true "消耗配额请求"
// @Success 200 {object} SuccessResponse "成功返回操作结果"
// @Failure 400 {object} ErrorResponse "参数错误"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Security BearerAuth
// @Router /quotas/consume [post]
func (h *QuotaHandler) ConsumeQuota(c *gin.Context) {
	var input service.ConsumeQuotaInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	result, err := h.quotaService.ConsumeQuota(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    0,
		Message: result.Message,
		Data:    result,
	})
}

// ReleaseQuota godoc
// @Summary 释放配额
// @Description 退款时调用，释放已占用的配额
// @Tags Quota
// @Accept json
// @Produce json
// @Param body body service.ReleaseQuotaInput true "释放配额请求"
// @Success 200 {object} SuccessResponse "成功返回操作结果"
// @Failure 400 {object} ErrorResponse "参数错误"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Security BearerAuth
// @Router /quotas/release [post]
func (h *QuotaHandler) ReleaseQuota(c *gin.Context) {
	var input service.ReleaseQuotaInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	result, err := h.quotaService.ReleaseQuota(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    0,
		Message: result.Message,
		Data:    result,
	})
}

// AdjustQuota godoc
// @Summary 调整配额
// @Description 管理员手动调整商户配额使用量
// @Tags Quota
// @Accept json
// @Produce json
// @Param body body service.AdjustQuotaInput true "调整配额请求"
// @Success 200 {object} SuccessResponse "成功返回调整后的配额"
// @Failure 400 {object} ErrorResponse "参数错误"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Security BearerAuth
// @Router /quotas/adjust [post]
func (h *QuotaHandler) AdjustQuota(c *gin.Context) {
	var input service.AdjustQuotaInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	quota, err := h.quotaService.AdjustQuota(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    0,
		Message: "quota adjusted successfully",
		Data:    quota,
	})
}

// SuspendQuota godoc
// @Summary 暂停配额
// @Description 暂停商户配额，暂停后无法消耗配额
// @Tags Quota
// @Accept json
// @Produce json
// @Param merchant_id query string true "商户ID"
// @Param currency query string true "币种"
// @Success 200 {object} SuccessResponse "成功暂停"
// @Failure 400 {object} ErrorResponse "参数错误"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Security BearerAuth
// @Router /quotas/suspend [post]
func (h *QuotaHandler) SuspendQuota(c *gin.Context) {
	merchantIDStr := c.Query("merchant_id")
	if merchantIDStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "merchant_id is required"})
		return
	}

	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid merchant_id"})
		return
	}

	currency := c.Query("currency")
	if currency == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "currency is required"})
		return
	}

	if err := h.quotaService.SuspendQuota(c.Request.Context(), merchantID, currency); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    0,
		Message: "quota suspended successfully",
		Data:    nil,
	})
}

// ResumeQuota godoc
// @Summary 恢复配额
// @Description 恢复已暂停的商户配额
// @Tags Quota
// @Accept json
// @Produce json
// @Param merchant_id query string true "商户ID"
// @Param currency query string true "币种"
// @Success 200 {object} SuccessResponse "成功恢复"
// @Failure 400 {object} ErrorResponse "参数错误"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Security BearerAuth
// @Router /quotas/resume [post]
func (h *QuotaHandler) ResumeQuota(c *gin.Context) {
	merchantIDStr := c.Query("merchant_id")
	if merchantIDStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "merchant_id is required"})
		return
	}

	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid merchant_id"})
		return
	}

	currency := c.Query("currency")
	if currency == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "currency is required"})
		return
	}

	if err := h.quotaService.ResumeQuota(c.Request.Context(), merchantID, currency); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    0,
		Message: "quota resumed successfully",
		Data:    nil,
	})
}

// GetQuota godoc
// @Summary 查询配额
// @Description 查询商户指定币种的配额信息
// @Tags Quota
// @Accept json
// @Produce json
// @Param merchant_id query string true "商户ID"
// @Param currency query string true "币种"
// @Success 200 {object} SuccessResponse "成功返回配额信息"
// @Failure 400 {object} ErrorResponse "参数错误"
// @Failure 404 {object} ErrorResponse "配额不存在"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Security BearerAuth
// @Router /quotas [get]
func (h *QuotaHandler) GetQuota(c *gin.Context) {
	merchantIDStr := c.Query("merchant_id")
	if merchantIDStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "merchant_id is required"})
		return
	}

	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid merchant_id"})
		return
	}

	currency := c.Query("currency")
	if currency == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "currency is required"})
		return
	}

	quota, err := h.quotaService.GetQuota(c.Request.Context(), merchantID, currency)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    0,
		Message: "success",
		Data:    quota,
	})
}

// ListQuotas godoc
// @Summary 查询配额列表
// @Description 查询配额列表，支持按商户、币种、状态筛选和分页
// @Tags Quota
// @Accept json
// @Produce json
// @Param merchant_id query string false "商户ID"
// @Param currency query string false "币种"
// @Param is_suspended query boolean false "是否暂停"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} ListResponse "成功返回配额列表"
// @Failure 400 {object} ErrorResponse "参数错误"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Security BearerAuth
// @Router /quotas/list [get]
func (h *QuotaHandler) ListQuotas(c *gin.Context) {
	var merchantID *uuid.UUID
	if merchantIDStr := c.Query("merchant_id"); merchantIDStr != "" {
		id, err := uuid.Parse(merchantIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid merchant_id"})
			return
		}
		merchantID = &id
	}

	currency := c.Query("currency")

	var isSuspended *bool
	if isSuspendedStr := c.Query("is_suspended"); isSuspendedStr != "" {
		val, err := strconv.ParseBool(isSuspendedStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid is_suspended value"})
			return
		}
		isSuspended = &val
	}

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

	result, err := h.quotaService.ListQuotas(c.Request.Context(), merchantID, currency, isSuspended, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, ListResponse{
		Code:       0,
		Message:    "success",
		Data:       result.Quotas,
		Total:      result.Total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
	})
}
