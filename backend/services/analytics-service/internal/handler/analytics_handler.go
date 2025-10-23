package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/payment-platform/services/analytics-service/internal/repository"
	"github.com/payment-platform/services/analytics-service/internal/service"
)

// AnalyticsHandler 分析处理器
type AnalyticsHandler struct {
	analyticsService service.AnalyticsService
}

// NewAnalyticsHandler 创建分析处理器实例
func NewAnalyticsHandler(analyticsService service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
	}
}

// RegisterRoutes 注册路由
func (h *AnalyticsHandler) RegisterRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		// 支付分析
		payments := v1.Group("/analytics/payments")
		{
			payments.GET("/metrics", h.GetPaymentMetrics)
			payments.GET("/summary", h.GetPaymentSummary)
		}

		// 商户分析
		merchants := v1.Group("/analytics/merchants")
		{
			merchants.GET("/metrics", h.GetMerchantMetrics)
			merchants.GET("/summary", h.GetMerchantSummary)
		}

		// 渠道分析
		channels := v1.Group("/analytics/channels")
		{
			channels.GET("/metrics", h.GetChannelMetrics)
			channels.GET("/summary", h.GetChannelSummary)
		}

		// 实时统计
		realtime := v1.Group("/analytics/realtime")
		{
			realtime.GET("/stats", h.GetRealtimeStats)
		}
	}
}

// Payment Analytics

func (h *AnalyticsHandler) GetPaymentMetrics(c *gin.Context) {
	merchantIDStr := c.Query("merchant_id")
	if merchantIDStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse("merchant_id 不能为空"))
		return
	}

	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("无效的商户ID"))
		return
	}

	startDate, endDate, err := parseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	metrics, err := h.analyticsService.GetPaymentMetrics(c.Request.Context(), merchantID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(metrics))
}

func (h *AnalyticsHandler) GetPaymentSummary(c *gin.Context) {
	merchantIDStr := c.Query("merchant_id")
	if merchantIDStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse("merchant_id 不能为空"))
		return
	}

	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("无效的商户ID"))
		return
	}

	startDate, endDate, err := parseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	summary, err := h.analyticsService.GetPaymentSummary(c.Request.Context(), merchantID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(summary))
}

// Merchant Analytics

func (h *AnalyticsHandler) GetMerchantMetrics(c *gin.Context) {
	merchantIDStr := c.Query("merchant_id")
	if merchantIDStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse("merchant_id 不能为空"))
		return
	}

	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("无效的商户ID"))
		return
	}

	startDate, endDate, err := parseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	metrics, err := h.analyticsService.GetMerchantMetrics(c.Request.Context(), merchantID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(metrics))
}

func (h *AnalyticsHandler) GetMerchantSummary(c *gin.Context) {
	merchantIDStr := c.Query("merchant_id")
	if merchantIDStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse("merchant_id 不能为空"))
		return
	}

	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("无效的商户ID"))
		return
	}

	startDate, endDate, err := parseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	summary, err := h.analyticsService.GetMerchantSummary(c.Request.Context(), merchantID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(summary))
}

// Channel Analytics

func (h *AnalyticsHandler) GetChannelMetrics(c *gin.Context) {
	channelCode := c.Query("channel_code")
	if channelCode == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse("channel_code 不能为空"))
		return
	}

	startDate, endDate, err := parseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	metrics, err := h.analyticsService.GetChannelMetrics(c.Request.Context(), channelCode, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(metrics))
}

func (h *AnalyticsHandler) GetChannelSummary(c *gin.Context) {
	channelCode := c.Query("channel_code")
	if channelCode == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse("channel_code 不能为空"))
		return
	}

	startDate, endDate, err := parseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	summary, err := h.analyticsService.GetChannelSummary(c.Request.Context(), channelCode, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(summary))
}

// Realtime Stats

func (h *AnalyticsHandler) GetRealtimeStats(c *gin.Context) {
	query := &repository.RealtimeStatsQuery{
		StatType: c.Query("stat_type"),
		StatKey:  c.Query("stat_key"),
		Period:   c.Query("period"),
	}

	if merchantIDStr := c.Query("merchant_id"); merchantIDStr != "" {
		merchantID, err := uuid.Parse(merchantIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse("无效的商户ID"))
			return
		}
		query.MerchantID = &merchantID
	}

	stats, err := h.analyticsService.GetRealtimeStats(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(stats))
}

// Helper functions

func parseDateRange(c *gin.Context) (time.Time, time.Time, error) {
	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	return startDate, endDate, nil
}

// Response structures

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func SuccessResponse(data interface{}) Response {
	return Response{
		Code:    0,
		Message: "success",
		Data:    data,
	}
}

func ErrorResponse(message string) Response {
	return Response{
		Code:    -1,
		Message: message,
	}
}
