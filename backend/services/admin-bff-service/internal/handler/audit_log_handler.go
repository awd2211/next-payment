package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"payment-platform/admin-service/internal/service"
)

// AuditLogHandler 审计日志HTTP处理器
type AuditLogHandler struct {
	auditLogService service.AuditLogService
}

// NewAuditLogHandler 创建审计日志处理器实例
func NewAuditLogHandler(auditLogService service.AuditLogService) *AuditLogHandler {
	return &AuditLogHandler{
		auditLogService: auditLogService,
	}
}

// RegisterRoutes 注册路由
func (h *AuditLogHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	logs := r.Group("/audit-logs")
	logs.Use(authMiddleware)
	{
		logs.GET("/:id", h.GetLog)
		logs.GET("", h.ListLogs)
		logs.GET("/stats", h.GetLogStats)
	}
}

// GetLog 获取审计日志详情
func (h *AuditLogHandler) GetLog(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "日志ID格式错误"})
		return
	}

	log, err := h.auditLogService.GetLog(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取日志失败", "details": err.Error()})
		return
	}

	if log == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "日志不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": log,
	})
}

// ListLogs 获取审计日志列表
func (h *AuditLogHandler) ListLogs(c *gin.Context) {
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

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	req := &service.ListAuditLogsRequest{
		Action:   c.Query("action"),
		Resource: c.Query("resource"),
		Method:   c.Query("method"),
		IP:       c.Query("ip"),
		Page:     page,
		PageSize: pageSize,
	}

	// 解析 admin_id
	if adminIDStr := c.Query("admin_id"); adminIDStr != "" {
		adminID, err := uuid.Parse(adminIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "管理员ID格式错误"})
			return
		}
		req.AdminID = &adminID
	}

	// 解析 response_code
	if responseCodeStr := c.Query("response_code"); responseCodeStr != "" {
		responseCode, err := strconv.Atoi(responseCodeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "响应码格式错误"})
			return
		}
		req.ResponseCode = &responseCode
	}

	// 解析时间范围
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "开始时间格式错误，请使用RFC3339格式"})
			return
		}
		req.StartTime = &startTime
	}

	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		endTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "结束时间格式错误，请使用RFC3339格式"})
			return
		}
		req.EndTime = &endTime
	}

	logs, total, err := h.auditLogService.ListLogs(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取日志列表失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": logs,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetLogStats 获取审计日志统计信息
func (h *AuditLogHandler) GetLogStats(c *gin.Context) {
	// 默认统计最近7天
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -7)

	// 解析自定义时间范围
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		t, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "开始时间格式错误，请使用RFC3339格式"})
			return
		}
		startTime = t
	}

	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		t, err := time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "结束时间格式错误，请使用RFC3339格式"})
			return
		}
		endTime = t
	}

	stats, err := h.auditLogService.GetLogStats(c.Request.Context(), startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取统计信息失败", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stats,
		"time_range": gin.H{
			"start_time": startTime.Format(time.RFC3339),
			"end_time":   endTime.Format(time.RFC3339),
		},
	})
}
