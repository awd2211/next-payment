package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"payment-platform/merchant-service/internal/service"
)

// DashboardHandler Dashboard处理器
type DashboardHandler struct {
	dashboardService service.DashboardService
}

// NewDashboardHandler 创建Dashboard处理器实例
func NewDashboardHandler(dashboardService service.DashboardService) *DashboardHandler {
	return &DashboardHandler{
		dashboardService: dashboardService,
	}
}

// RegisterRoutes 注册Dashboard路由
func (h *DashboardHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	dashboard := r.Group("/dashboard")
	dashboard.Use(authMiddleware)
	{
		dashboard.GET("", h.GetDashboard)
		dashboard.GET("/transaction-summary", h.GetTransactionSummary)
		dashboard.GET("/balance", h.GetBalanceInfo)
	}
}

// GetDashboard 获取Dashboard数据
// @Summary 获取Dashboard概览
// @Tags Dashboard
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/dashboard [get]
func (h *DashboardHandler) GetDashboard(c *gin.Context) {
	merchantID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	dashboard, err := h.dashboardService.GetDashboard(c.Request.Context(), merchantID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": dashboard,
	})
}

// GetTransactionSummary 获取交易汇总
// @Summary 获取交易汇总
// @Tags Dashboard
// @Produce json
// @Param start_date query string false "开始日期 YYYY-MM-DD"
// @Param end_date query string false "结束日期 YYYY-MM-DD"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/dashboard/transaction-summary [get]
func (h *DashboardHandler) GetTransactionSummary(c *gin.Context) {
	merchantID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	summary, err := h.dashboardService.GetTransactionSummary(c.Request.Context(), merchantID.(uuid.UUID), startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": summary,
	})
}

// GetBalanceInfo 获取余额信息
// @Summary 获取余额信息
// @Tags Dashboard
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/dashboard/balance [get]
func (h *DashboardHandler) GetBalanceInfo(c *gin.Context) {
	merchantID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	balanceInfo, err := h.dashboardService.GetBalanceInfo(c.Request.Context(), merchantID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": balanceInfo,
	})
}
