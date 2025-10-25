package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"payment-platform/payment-gateway/internal/service"
)

// ExportHandler 导出处理器
type ExportHandler struct {
	exportService *service.PaymentExportService
}

// NewExportHandler 创建导出处理器
func NewExportHandler(exportService *service.PaymentExportService) *ExportHandler {
	return &ExportHandler{
		exportService: exportService,
	}
}

// CreatePaymentExportRequest 创建支付记录导出请求
type CreatePaymentExportRequest struct {
	StartDate string `form:"start_date" binding:"required"` // 开始日期 YYYY-MM-DD
	EndDate   string `form:"end_date" binding:"required"`   // 结束日期 YYYY-MM-DD
	Format    string `form:"format" binding:"required,oneof=csv excel"` // 导出格式
}

// CreatePaymentExport 创建支付记录导出任务
//
//	@Summary		创建支付记录导出任务
//	@Description	创建支付记录导出任务（支持CSV和Excel格式）
//	@Tags			Export
//	@Accept			json
//	@Produce		json
//	@Param			start_date	query		string	true	"开始日期 (YYYY-MM-DD)"
//	@Param			end_date	query		string	true	"结束日期 (YYYY-MM-DD)"
//	@Param			format		query		string	true	"导出格式 (csv 或 excel)"
//	@Success		200			{object}	map[string]interface{}
//	@Failure		400			{object}	map[string]interface{}
//	@Failure		500			{object}	map[string]interface{}
//	@Router			/api/v1/merchant/payments/export [post]
func (h *ExportHandler) CreatePaymentExport(c *gin.Context) {
	var req CreatePaymentExportRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 获取商户ID
	merchantID, exists := c.Get("merchant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未授权",
		})
		return
	}

	// 解析日期
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "开始日期格式错误",
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "结束日期格式错误",
		})
		return
	}

	// 设置结束日期为当天 23:59:59
	endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	// 创建导出任务
	task, err := h.exportService.CreatePaymentExportTask(
		c.Request.Context(),
		merchantID.(uuid.UUID),
		startDate,
		endDate,
		req.Format,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建导出任务失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "导出任务已创建",
		"data": gin.H{
			"task_id":    task.ID,
			"status":     task.Status,
			"created_at": task.CreatedAt,
		},
	})
}

// CreateRefundExport 创建退款记录导出任务
//
//	@Summary		创建退款记录导出任务
//	@Description	创建退款记录导出任务（支持CSV和Excel格式）
//	@Tags			Export
//	@Accept			json
//	@Produce		json
//	@Param			start_date	query		string	true	"开始日期 (YYYY-MM-DD)"
//	@Param			end_date	query		string	true	"结束日期 (YYYY-MM-DD)"
//	@Param			format		query		string	true	"导出格式 (csv 或 excel)"
//	@Success		200			{object}	map[string]interface{}
//	@Failure		400			{object}	map[string]interface{}
//	@Failure		500			{object}	map[string]interface{}
//	@Router			/api/v1/merchant/refunds/export [post]
func (h *ExportHandler) CreateRefundExport(c *gin.Context) {
	var req CreatePaymentExportRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 获取商户ID
	merchantID, exists := c.Get("merchant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未授权",
		})
		return
	}

	// 解析日期
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "开始日期格式错误",
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "结束日期格式错误",
		})
		return
	}

	// 设置结束日期为当天 23:59:59
	endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	// 创建导出任务
	task, err := h.exportService.CreateRefundExportTask(
		c.Request.Context(),
		merchantID.(uuid.UUID),
		startDate,
		endDate,
		req.Format,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建导出任务失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "导出任务已创建",
		"data": gin.H{
			"task_id":    task.ID,
			"status":     task.Status,
			"created_at": task.CreatedAt,
		},
	})
}

// GetExportTask 获取导出任务状态
//
//	@Summary		获取导出任务状态
//	@Description	获取导出任务的详细信息和状态
//	@Tags			Export
//	@Accept			json
//	@Produce		json
//	@Param			task_id	path		string	true	"任务ID"
//	@Success		200		{object}	map[string]interface{}
//	@Failure		404		{object}	map[string]interface{}
//	@Failure		500		{object}	map[string]interface{}
//	@Router			/api/v1/merchant/exports/{task_id} [get]
func (h *ExportHandler) GetExportTask(c *gin.Context) {
	taskIDStr := c.Param("task_id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "任务ID格式错误",
		})
		return
	}

	// 获取商户ID
	merchantID, exists := c.Get("merchant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未授权",
		})
		return
	}

	task, err := h.exportService.GetExportTask(
		c.Request.Context(),
		taskID,
		merchantID.(uuid.UUID),
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功",
		"data":    task,
	})
}

// DownloadExport 下载导出文件
//
//	@Summary		下载导出文件
//	@Description	下载已完成的导出文件
//	@Tags			Export
//	@Produce		application/octet-stream
//	@Param			task_id	path	string	true	"任务ID"
//	@Success		200		{file}	binary
//	@Failure		400		{object}	map[string]interface{}
//	@Failure		404		{object}	map[string]interface{}
//	@Router			/api/v1/merchant/exports/{task_id}/download [get]
func (h *ExportHandler) DownloadExport(c *gin.Context) {
	taskIDStr := c.Param("task_id")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "任务ID格式错误",
		})
		return
	}

	// 获取商户ID
	merchantID, exists := c.Get("merchant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未授权",
		})
		return
	}

	task, err := h.exportService.GetExportTask(
		c.Request.Context(),
		taskID,
		merchantID.(uuid.UUID),
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "导出任务不存在",
		})
		return
	}

	if task.Status != "completed" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "文件尚未生成，请稍后再试",
			"data": gin.H{
				"status": task.Status,
			},
		})
		return
	}

	// 返回文件
	c.Header("Content-Disposition", "attachment; filename="+task.FileName)
	c.Header("Content-Type", "application/octet-stream")
	c.File(task.FilePath)
}

// ListExportTasks 查询导出任务列表
//
//	@Summary		查询导出任务列表
//	@Description	分页查询导出任务列表
//	@Tags			Export
//	@Accept			json
//	@Produce		json
//	@Param			page		query		int	false	"页码"	default(1)
//	@Param			page_size	query		int	false	"每页数量"	default(20)
//	@Success		200			{object}	map[string]interface{}
//	@Failure		500			{object}	map[string]interface{}
//	@Router			/api/v1/merchant/exports [get]
func (h *ExportHandler) ListExportTasks(c *gin.Context) {
	// 获取商户ID
	merchantID, exists := c.Get("merchant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未授权",
		})
		return
	}

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	tasks, total, err := h.exportService.ListExportTasks(
		c.Request.Context(),
		merchantID.(uuid.UUID),
		page,
		pageSize,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功",
		"data": gin.H{
			"list":      tasks,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}
