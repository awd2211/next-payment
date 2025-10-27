package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	_ "payment-platform/reconciliation-service/internal/model" // imported for Swagger
	"payment-platform/reconciliation-service/internal/service"
)

// ReconciliationHandler 对账HTTP处理器
type ReconciliationHandler struct {
	service service.ReconciliationService
}

// NewReconciliationHandler 创建处理器实例
func NewReconciliationHandler(service service.ReconciliationService) *ReconciliationHandler {
	return &ReconciliationHandler{
		service: service,
	}
}

// RegisterRoutes 注册路由
func (h *ReconciliationHandler) RegisterRoutes(router *gin.RouterGroup) {
	reconciliation := router.Group("/reconciliation")
	{
		// Task management
		reconciliation.POST("/tasks", h.CreateTask)
		reconciliation.GET("/tasks", h.ListTasks)
		reconciliation.GET("/tasks/:task_id", h.GetTaskDetails)
		reconciliation.POST("/tasks/:task_id/execute", h.ExecuteTask)
		reconciliation.POST("/tasks/:task_id/retry", h.RetryTask)

		// Record management
		reconciliation.GET("/records", h.ListRecords)
		reconciliation.GET("/records/:record_id", h.GetRecordDetails)
		reconciliation.POST("/records/:record_id/resolve", h.ResolveRecord)

		// File management
		reconciliation.GET("/settlement-files", h.ListFiles)
		reconciliation.POST("/settlement-files/download", h.DownloadSettlementFile)
		reconciliation.GET("/settlement-files/:file_no", h.GetFileDetails)

		// Report generation
		reconciliation.GET("/reports/:task_id", h.GenerateReport)
	}
}

// CreateTask 创建对账任务
// @Summary 创建对账任务
// @Tags Reconciliation
// @Accept json
// @Produce json
// @Param body body CreateTaskRequest true "创建任务请求"
// @Success 200 {object} Response{data=model.ReconciliationTask}
// @Router /reconciliation/tasks [post]
func (h *ReconciliationHandler) CreateTask(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_REQUEST", err.Error()))
		return
	}

	// Parse task date
	taskDate, err := time.Parse("2006-01-02", req.TaskDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_DATE", "Invalid task_date format, expected YYYY-MM-DD"))
		return
	}

	input := &service.CreateTaskInput{
		TaskDate: taskDate,
		Channel:  req.Channel,
		TaskType: req.TaskType,
	}

	task, err := h.service.CreateTask(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("CREATE_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(task))
}

// ExecuteTask 执行对账任务
// @Summary 执行对账任务
// @Tags Reconciliation
// @Produce json
// @Param task_id path string true "任务ID"
// @Success 200 {object} Response
// @Router /reconciliation/tasks/{task_id}/execute [post]
func (h *ReconciliationHandler) ExecuteTask(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("task_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_TASK_ID", "Invalid task ID format"))
		return
	}

	if err := h.service.ExecuteTask(c.Request.Context(), taskID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("EXECUTE_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "Task execution started"}))
}

// GetTaskDetails 获取任务详情
// @Summary 获取任务详情
// @Tags Reconciliation
// @Produce json
// @Param task_id path string true "任务ID"
// @Success 200 {object} Response{data=service.TaskDetails}
// @Router /reconciliation/tasks/{task_id} [get]
func (h *ReconciliationHandler) GetTaskDetails(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("task_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_TASK_ID", "Invalid task ID format"))
		return
	}

	details, err := h.service.GetTaskDetails(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("GET_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(details))
}

// ListTasks 查询任务列表
// @Summary 查询任务列表
// @Tags Reconciliation
// @Produce json
// @Param task_date query string false "任务日期 (YYYY-MM-DD)"
// @Param channel query string false "支付渠道"
// @Param status query string false "任务状态"
// @Param start_date query string false "开始日期 (YYYY-MM-DD)"
// @Param end_date query string false "结束日期 (YYYY-MM-DD)"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} Response{data=service.TaskListResult}
// @Router /reconciliation/tasks [get]
func (h *ReconciliationHandler) ListTasks(c *gin.Context) {
	filters := &service.TaskFilters{}

	// Parse filters
	if taskDateStr := c.Query("task_date"); taskDateStr != "" {
		taskDate, err := time.Parse("2006-01-02", taskDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_DATE", "Invalid task_date format"))
			return
		}
		filters.TaskDate = &taskDate
	}

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_DATE", "Invalid start_date format"))
			return
		}
		filters.StartDate = &startDate
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_DATE", "Invalid end_date format"))
			return
		}
		filters.EndDate = &endDate
	}

	filters.Channel = c.Query("channel")
	filters.Status = c.Query("status")

	// Parse pagination
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

	result, err := h.service.ListTasks(c.Request.Context(), filters, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("LIST_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(result))
}

// RetryTask 重试失败的任务
// @Summary 重试失败的任务
// @Tags Reconciliation
// @Produce json
// @Param task_id path string true "任务ID"
// @Success 200 {object} Response
// @Router /reconciliation/tasks/{task_id}/retry [post]
func (h *ReconciliationHandler) RetryTask(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("task_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_TASK_ID", "Invalid task ID format"))
		return
	}

	if err := h.service.RetryTask(c.Request.Context(), taskID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("RETRY_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "Task retry started"}))
}

// ListRecords 查询差异记录列表
// @Summary 查询差异记录列表
// @Tags Reconciliation
// @Produce json
// @Param task_id query string false "任务ID"
// @Param diff_type query string false "差异类型"
// @Param is_resolved query bool false "是否已解决"
// @Param merchant_id query string false "商户ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} Response{data=service.RecordListResult}
// @Router /reconciliation/records [get]
func (h *ReconciliationHandler) ListRecords(c *gin.Context) {
	filters := &service.RecordFilters{}

	// Parse filters
	if taskIDStr := c.Query("task_id"); taskIDStr != "" {
		taskID, err := uuid.Parse(taskIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_TASK_ID", "Invalid task_id format"))
			return
		}
		filters.TaskID = &taskID
	}

	if merchantIDStr := c.Query("merchant_id"); merchantIDStr != "" {
		merchantID, err := uuid.Parse(merchantIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_MERCHANT_ID", "Invalid merchant_id format"))
			return
		}
		filters.MerchantID = &merchantID
	}

	if isResolvedStr := c.Query("is_resolved"); isResolvedStr != "" {
		isResolved, err := strconv.ParseBool(isResolvedStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_PARAM", "Invalid is_resolved format"))
			return
		}
		filters.IsResolved = &isResolved
	}

	filters.DiffType = c.Query("diff_type")

	// Parse pagination
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

	result, err := h.service.ListRecords(c.Request.Context(), filters, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("LIST_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(result))
}

// GetRecordDetails 获取差异记录详情
// @Summary 获取差异记录详情
// @Tags Reconciliation
// @Produce json
// @Param record_id path string true "记录ID"
// @Success 200 {object} Response{data=model.ReconciliationRecord}
// @Router /reconciliation/records/{record_id} [get]
func (h *ReconciliationHandler) GetRecordDetails(c *gin.Context) {
	recordID, err := uuid.Parse(c.Param("record_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_RECORD_ID", "Invalid record ID format"))
		return
	}

	record, err := h.service.GetRecordDetails(c.Request.Context(), recordID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("GET_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(record))
}

// ResolveRecord 标记差异已解决
// @Summary 标记差异已解决
// @Tags Reconciliation
// @Accept json
// @Produce json
// @Param record_id path string true "记录ID"
// @Param body body ResolveRecordRequest true "解决记录请求"
// @Success 200 {object} Response
// @Router /reconciliation/records/{record_id}/resolve [post]
func (h *ReconciliationHandler) ResolveRecord(c *gin.Context) {
	recordID, err := uuid.Parse(c.Param("record_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_RECORD_ID", "Invalid record ID format"))
		return
	}

	var req ResolveRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_REQUEST", err.Error()))
		return
	}

	resolvedBy, err := uuid.Parse(req.ResolvedBy)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_USER_ID", "Invalid resolved_by format"))
		return
	}

	if err := h.service.ResolveRecord(c.Request.Context(), recordID, resolvedBy, req.Note); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("RESOLVE_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "Record resolved successfully"}))
}

// ListFiles 查询文件列表
// @Summary 查询文件列表
// @Tags Reconciliation
// @Produce json
// @Param channel query string false "支付渠道"
// @Param settlement_date query string false "结算日期 (YYYY-MM-DD)"
// @Param status query string false "文件状态"
// @Param start_date query string false "开始日期 (YYYY-MM-DD)"
// @Param end_date query string false "结束日期 (YYYY-MM-DD)"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} Response{data=service.FileListResult}
// @Router /reconciliation/settlement-files [get]
func (h *ReconciliationHandler) ListFiles(c *gin.Context) {
	filters := &service.FileFilters{}

	// Parse filters
	if settlementDateStr := c.Query("settlement_date"); settlementDateStr != "" {
		settlementDate, err := time.Parse("2006-01-02", settlementDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_DATE", "Invalid settlement_date format"))
			return
		}
		filters.SettlementDate = &settlementDate
	}

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_DATE", "Invalid start_date format"))
			return
		}
		filters.StartDate = &startDate
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_DATE", "Invalid end_date format"))
			return
		}
		filters.EndDate = &endDate
	}

	filters.Channel = c.Query("channel")
	filters.Status = c.Query("status")

	// Parse pagination
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

	result, err := h.service.ListFiles(c.Request.Context(), filters, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("LIST_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(result))
}

// DownloadSettlementFile 下载渠道账单文件
// @Summary 下载渠道账单文件
// @Tags Reconciliation
// @Accept json
// @Produce json
// @Param body body DownloadFileRequest true "下载文件请求"
// @Success 200 {object} Response{data=model.ChannelSettlementFile}
// @Router /reconciliation/settlement-files/download [post]
func (h *ReconciliationHandler) DownloadSettlementFile(c *gin.Context) {
	var req DownloadFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_REQUEST", err.Error()))
		return
	}

	settlementDate, err := time.Parse("2006-01-02", req.SettlementDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_DATE", "Invalid settlement_date format"))
		return
	}

	file, err := h.service.DownloadSettlementFile(c.Request.Context(), req.Channel, settlementDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("DOWNLOAD_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(file))
}

// GetFileDetails 获取文件详情
// @Summary 获取文件详情
// @Tags Reconciliation
// @Produce json
// @Param file_no path string true "文件编号"
// @Success 200 {object} Response{data=model.ChannelSettlementFile}
// @Router /reconciliation/settlement-files/{file_no} [get]
func (h *ReconciliationHandler) GetFileDetails(c *gin.Context) {
	fileNo := c.Param("file_no")

	file, err := h.service.GetFileDetails(c.Request.Context(), fileNo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("GET_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(file))
}

// GenerateReport 生成对账报告
// @Summary 生成对账报告
// @Tags Reconciliation
// @Produce json
// @Param task_id path string true "任务ID"
// @Success 200 {object} Response{data=map[string]string}
// @Router /reconciliation/reports/{task_id} [get]
func (h *ReconciliationHandler) GenerateReport(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("task_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("INVALID_TASK_ID", "Invalid task ID format"))
		return
	}

	reportURL, err := h.service.GenerateReport(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse("GENERATE_FAILED", err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(gin.H{
		"report_url": reportURL,
		"message":    "Report generated successfully",
	}))
}

// Request DTOs

type CreateTaskRequest struct {
	TaskDate string `json:"task_date" binding:"required"`
	Channel  string `json:"channel" binding:"required"`
	TaskType string `json:"task_type" binding:"required"`
}

type ResolveRecordRequest struct {
	ResolvedBy string `json:"resolved_by" binding:"required"`
	Note       string `json:"note"`
}

type DownloadFileRequest struct {
	Channel        string `json:"channel" binding:"required"`
	SettlementDate string `json:"settlement_date" binding:"required"`
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
