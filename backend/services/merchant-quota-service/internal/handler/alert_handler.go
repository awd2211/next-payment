package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"payment-platform/merchant-quota-service/internal/service"
)

// AlertHandler 配额预警处理器
type AlertHandler struct {
	alertService service.AlertService
}

// NewAlertHandler 创建预警处理器实例
func NewAlertHandler(alertService service.AlertService) *AlertHandler {
	return &AlertHandler{
		alertService: alertService,
	}
}

// CheckMerchantQuotaAlert godoc
// @Summary 检查商户配额预警
// @Description 手动触发检查指定商户的配额预警
// @Tags Alert
// @Accept json
// @Produce json
// @Param merchant_id query string true "商户ID"
// @Param currency query string true "币种"
// @Success 200 {object} SuccessResponse "成功检查预警"
// @Failure 400 {object} ErrorResponse "参数错误"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Security BearerAuth
// @Router /alerts/check [post]
func (h *AlertHandler) CheckMerchantQuotaAlert(c *gin.Context) {
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

	if err := h.alertService.CheckMerchantQuotaAlert(c.Request.Context(), merchantID, currency); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    0,
		Message: "quota alert check completed",
		Data:    nil,
	})
}

// ResolveAlert godoc
// @Summary 标记预警为已处理
// @Description 管理员标记预警为已处理
// @Tags Alert
// @Accept json
// @Produce json
// @Param alert_id path string true "预警ID"
// @Param resolved_by query string true "处理人ID"
// @Success 200 {object} SuccessResponse "成功标记为已处理"
// @Failure 400 {object} ErrorResponse "参数错误"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Security BearerAuth
// @Router /alerts/{alert_id}/resolve [post]
func (h *AlertHandler) ResolveAlert(c *gin.Context) {
	alertIDStr := c.Param("alert_id")
	alertID, err := uuid.Parse(alertIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid alert_id"})
		return
	}

	resolvedByStr := c.Query("resolved_by")
	if resolvedByStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "resolved_by is required"})
		return
	}

	resolvedBy, err := uuid.Parse(resolvedByStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid resolved_by"})
		return
	}

	if err := h.alertService.ResolveAlert(c.Request.Context(), alertID, resolvedBy); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    0,
		Message: "alert resolved successfully",
		Data:    nil,
	})
}

// GetActiveAlerts godoc
// @Summary 获取商户活跃预警
// @Description 查询商户未处理的预警列表
// @Tags Alert
// @Accept json
// @Produce json
// @Param merchant_id query string true "商户ID"
// @Param alert_level query string false "预警级别 (warning, critical)"
// @Success 200 {object} SuccessResponse "成功返回预警列表"
// @Failure 400 {object} ErrorResponse "参数错误"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Security BearerAuth
// @Router /alerts/active [get]
func (h *AlertHandler) GetActiveAlerts(c *gin.Context) {
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

	alertLevel := c.Query("alert_level")

	alerts, err := h.alertService.GetActiveAlerts(c.Request.Context(), merchantID, alertLevel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Code:    0,
		Message: "success",
		Data:    alerts,
	})
}

// ListAlerts godoc
// @Summary 查询预警列表
// @Description 查询预警列表，支持按商户、级别、类型、状态筛选和分页
// @Tags Alert
// @Accept json
// @Produce json
// @Param merchant_id query string false "商户ID"
// @Param alert_level query string false "预警级别"
// @Param alert_type query string false "预警类型"
// @Param is_resolved query boolean false "是否已处理"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} ListResponse "成功返回预警列表"
// @Failure 400 {object} ErrorResponse "参数错误"
// @Failure 500 {object} ErrorResponse "服务器错误"
// @Security BearerAuth
// @Router /alerts [get]
func (h *AlertHandler) ListAlerts(c *gin.Context) {
	var merchantID *uuid.UUID
	if merchantIDStr := c.Query("merchant_id"); merchantIDStr != "" {
		id, err := uuid.Parse(merchantIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid merchant_id"})
			return
		}
		merchantID = &id
	}

	alertLevel := c.Query("alert_level")
	alertType := c.Query("alert_type")

	var isResolved *bool
	if isResolvedStr := c.Query("is_resolved"); isResolvedStr != "" {
		val, err := strconv.ParseBool(isResolvedStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid is_resolved value"})
			return
		}
		isResolved = &val
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

	result, err := h.alertService.ListAlerts(c.Request.Context(), merchantID, alertLevel, alertType, isResolved, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, ListResponse{
		Code:       0,
		Message:    "success",
		Data:       result.Alerts,
		Total:      result.Total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
	})
}
