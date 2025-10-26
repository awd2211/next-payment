package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"payment-platform/admin-service/internal/client"
)

type ReconciliationBFFHandler struct {
	reconciliationClient *client.ServiceClient
}

func NewReconciliationBFFHandler(reconciliationServiceURL string) *ReconciliationBFFHandler {
	return &ReconciliationBFFHandler{
		reconciliationClient: client.NewServiceClient(reconciliationServiceURL),
	}
}

func (h *ReconciliationBFFHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	admin := r.Group("/admin/reconciliation")
	admin.Use(authMiddleware)
	{
		// 对账任务
		tasks := admin.Group("/tasks")
		{
			tasks.GET("", h.ListTasks)
			tasks.GET("/:id", h.GetTask)
			tasks.POST("", h.CreateTask)
			tasks.POST("/:id/start", h.StartTask)
			tasks.POST("/:id/cancel", h.CancelTask)
			tasks.POST("/:id/retry", h.RetryTask)
		}

		// 差异处理
		discrepancies := admin.Group("/discrepancies")
		{
			discrepancies.GET("", h.ListDiscrepancies)
			discrepancies.GET("/:id", h.GetDiscrepancy)
			discrepancies.POST("/:id/resolve", h.ResolveDiscrepancy)
			discrepancies.POST("/:id/ignore", h.IgnoreDiscrepancy)
			discrepancies.POST("/batch-resolve", h.BatchResolveDiscrepancies)
		}

		// 对账报告
		reports := admin.Group("/reports")
		{
			reports.GET("", h.ListReports)
			reports.GET("/:id", h.GetReport)
			reports.GET("/:id/download", h.DownloadReport)
			reports.POST("/:id/regenerate", h.RegenerateReport)
		}

		// 对账配置
		admin.GET("/config", h.GetConfig)
		admin.PUT("/config", h.UpdateConfig)

		// 统计
		admin.GET("/statistics", h.GetStatistics)
		admin.GET("/statistics/trend", h.GetTrendStatistics)
	}
}

// ========== 对账任务 ==========

func (h *ReconciliationBFFHandler) ListTasks(c *gin.Context) {
	queryParams := make(map[string]string)
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}
	if channel := c.Query("channel"); channel != "" {
		queryParams["channel"] = channel
	}
	if status := c.Query("status"); status != "" {
		queryParams["status"] = status
	}
	if startDate := c.Query("start_date"); startDate != "" {
		queryParams["start_date"] = startDate
	}
	if endDate := c.Query("end_date"); endDate != "" {
		queryParams["end_date"] = endDate
	}
	if page := c.Query("page"); page != "" {
		queryParams["page"] = page
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		queryParams["page_size"] = pageSize
	}

	result, statusCode, err := h.reconciliationClient.Get(c.Request.Context(), "/api/v1/tasks", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Reconciliation Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *ReconciliationBFFHandler) GetTask(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.reconciliationClient.Get(c.Request.Context(), "/api/v1/tasks/"+id, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Reconciliation Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *ReconciliationBFFHandler) CreateTask(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	adminID := c.GetString("user_id")
	req["created_by"] = adminID

	result, statusCode, err := h.reconciliationClient.Post(c.Request.Context(), "/api/v1/tasks", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Reconciliation Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *ReconciliationBFFHandler) StartTask(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		req = make(map[string]interface{})
	}

	adminID := c.GetString("user_id")
	req["started_by"] = adminID

	result, statusCode, err := h.reconciliationClient.Post(c.Request.Context(), "/api/v1/tasks/"+id+"/start", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Reconciliation Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *ReconciliationBFFHandler) CancelTask(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		req = make(map[string]interface{})
	}

	adminID := c.GetString("user_id")
	req["cancelled_by"] = adminID

	result, statusCode, err := h.reconciliationClient.Post(c.Request.Context(), "/api/v1/tasks/"+id+"/cancel", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Reconciliation Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *ReconciliationBFFHandler) RetryTask(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		req = make(map[string]interface{})
	}

	adminID := c.GetString("user_id")
	req["retried_by"] = adminID

	result, statusCode, err := h.reconciliationClient.Post(c.Request.Context(), "/api/v1/tasks/"+id+"/retry", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Reconciliation Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== 差异处理 ==========

func (h *ReconciliationBFFHandler) ListDiscrepancies(c *gin.Context) {
	queryParams := make(map[string]string)
	if taskID := c.Query("task_id"); taskID != "" {
		queryParams["task_id"] = taskID
	}
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}
	if discrepancyType := c.Query("type"); discrepancyType != "" {
		queryParams["type"] = discrepancyType
	}
	if status := c.Query("status"); status != "" {
		queryParams["status"] = status
	}
	if page := c.Query("page"); page != "" {
		queryParams["page"] = page
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		queryParams["page_size"] = pageSize
	}

	result, statusCode, err := h.reconciliationClient.Get(c.Request.Context(), "/api/v1/discrepancies", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Reconciliation Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *ReconciliationBFFHandler) GetDiscrepancy(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.reconciliationClient.Get(c.Request.Context(), "/api/v1/discrepancies/"+id, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Reconciliation Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *ReconciliationBFFHandler) ResolveDiscrepancy(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	adminID := c.GetString("user_id")
	req["resolved_by"] = adminID

	result, statusCode, err := h.reconciliationClient.Post(c.Request.Context(), "/api/v1/discrepancies/"+id+"/resolve", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Reconciliation Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *ReconciliationBFFHandler) IgnoreDiscrepancy(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	adminID := c.GetString("user_id")
	req["ignored_by"] = adminID

	result, statusCode, err := h.reconciliationClient.Post(c.Request.Context(), "/api/v1/discrepancies/"+id+"/ignore", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Reconciliation Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *ReconciliationBFFHandler) BatchResolveDiscrepancies(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	adminID := c.GetString("user_id")
	req["resolved_by"] = adminID

	result, statusCode, err := h.reconciliationClient.Post(c.Request.Context(), "/api/v1/discrepancies/batch-resolve", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Reconciliation Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== 对账报告 ==========

func (h *ReconciliationBFFHandler) ListReports(c *gin.Context) {
	queryParams := make(map[string]string)
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}
	if channel := c.Query("channel"); channel != "" {
		queryParams["channel"] = channel
	}
	if startDate := c.Query("start_date"); startDate != "" {
		queryParams["start_date"] = startDate
	}
	if endDate := c.Query("end_date"); endDate != "" {
		queryParams["end_date"] = endDate
	}
	if page := c.Query("page"); page != "" {
		queryParams["page"] = page
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		queryParams["page_size"] = pageSize
	}

	result, statusCode, err := h.reconciliationClient.Get(c.Request.Context(), "/api/v1/reports", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Reconciliation Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *ReconciliationBFFHandler) GetReport(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.reconciliationClient.Get(c.Request.Context(), "/api/v1/reports/"+id, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Reconciliation Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *ReconciliationBFFHandler) DownloadReport(c *gin.Context) {
	id := c.Param("id")

	queryParams := make(map[string]string)
	if format := c.Query("format"); format != "" {
		queryParams["format"] = format
	}

	result, statusCode, err := h.reconciliationClient.Get(c.Request.Context(), "/api/v1/reports/"+id+"/download", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Reconciliation Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *ReconciliationBFFHandler) RegenerateReport(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		req = make(map[string]interface{})
	}

	adminID := c.GetString("user_id")
	req["regenerated_by"] = adminID

	result, statusCode, err := h.reconciliationClient.Post(c.Request.Context(), "/api/v1/reports/"+id+"/regenerate", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Reconciliation Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== 对账配置 ==========

func (h *ReconciliationBFFHandler) GetConfig(c *gin.Context) {
	queryParams := make(map[string]string)
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}

	result, statusCode, err := h.reconciliationClient.Get(c.Request.Context(), "/api/v1/config", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Reconciliation Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *ReconciliationBFFHandler) UpdateConfig(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	adminID := c.GetString("user_id")
	req["updated_by"] = adminID

	result, statusCode, err := h.reconciliationClient.Put(c.Request.Context(), "/api/v1/config", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Reconciliation Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== 统计 ==========

func (h *ReconciliationBFFHandler) GetStatistics(c *gin.Context) {
	queryParams := make(map[string]string)
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}
	if startDate := c.Query("start_date"); startDate != "" {
		queryParams["start_date"] = startDate
	}
	if endDate := c.Query("end_date"); endDate != "" {
		queryParams["end_date"] = endDate
	}

	result, statusCode, err := h.reconciliationClient.Get(c.Request.Context(), "/api/v1/statistics", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Reconciliation Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *ReconciliationBFFHandler) GetTrendStatistics(c *gin.Context) {
	queryParams := make(map[string]string)
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}
	if period := c.Query("period"); period != "" {
		queryParams["period"] = period
	}

	result, statusCode, err := h.reconciliationClient.Get(c.Request.Context(), "/api/v1/statistics/trend", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Reconciliation Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}
