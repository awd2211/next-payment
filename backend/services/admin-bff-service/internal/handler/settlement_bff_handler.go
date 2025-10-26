package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"payment-platform/admin-service/internal/client"
	localMiddleware "payment-platform/admin-service/internal/middleware"
	"payment-platform/admin-service/internal/service"
	"payment-platform/admin-service/internal/utils"
)

type SettlementBFFHandler struct {
	settlementClient *client.ServiceClient
	auditLogService  service.AuditLogService
	auditHelper      *utils.AuditHelper
}

func NewSettlementBFFHandler(settlementServiceURL string, auditLogService service.AuditLogService) *SettlementBFFHandler {
	return &SettlementBFFHandler{
		settlementClient: client.NewServiceClient(settlementServiceURL),
		auditLogService:  auditLogService,
		auditHelper:      utils.NewAuditHelper(auditLogService),
	}
}

func (h *SettlementBFFHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	admin := r.Group("/admin/settlement")
	admin.Use(authMiddleware)
	{
		// 结算单管理
		settlements := admin.Group("/settlements")
		{
			settlements.GET("",
				localMiddleware.RequirePermission("settlements.view"),
				localMiddleware.RequireReason,
				h.ListSettlements,
			)
			settlements.GET("/:id",
				localMiddleware.RequirePermission("settlements.view"),
				localMiddleware.RequireReason,
				h.GetSettlement,
			)
			settlements.POST("",
				localMiddleware.RequirePermission("settlements.create"),
				h.CreateSettlement,
			)
			settlements.GET("/:id/details",
				localMiddleware.RequirePermission("settlements.view"),
				localMiddleware.RequireReason,
				h.GetSettlementDetails,
			)
		}

		// 结算审批
		approvals := admin.Group("/approvals")
		{
			approvals.POST("/:id/approve",
				localMiddleware.RequirePermission("settlements.approve"),
				localMiddleware.RequireReason,
				h.ApproveSettlement,
			)
			approvals.POST("/:id/reject",
				localMiddleware.RequirePermission("settlements.approve"),
				localMiddleware.RequireReason,
				h.RejectSettlement,
			)
			approvals.GET("/pending",
				localMiddleware.RequirePermission("settlements.view"),
				h.ListPendingApprovals,
			)
			approvals.POST("/batch-approve",
				localMiddleware.RequirePermission("settlements.approve"),
				localMiddleware.RequireReason,
				h.BatchApproveSettlements,
			)
		}

		// 结算统计
		admin.GET("/statistics",
			localMiddleware.RequirePermission("settlements.view"),
			h.GetStatistics,
		)
		admin.GET("/statistics/trend",
			localMiddleware.RequirePermission("settlements.view"),
			h.GetTrendStatistics,
		)
		admin.GET("/statistics/merchant",
			localMiddleware.RequirePermission("settlements.view"),
			h.GetMerchantStatistics,
		)

		// 结算配置
		admin.GET("/config",
			localMiddleware.RequirePermission("settlements.view"),
			h.GetConfig,
		)
		admin.PUT("/config",
			localMiddleware.RequirePermission("settlements.update"),
			h.UpdateConfig,
		)

		// 自动结算任务
		admin.GET("/auto-tasks",
			localMiddleware.RequirePermission("settlements.view"),
			h.ListAutoTasks,
		)
		admin.GET("/auto-tasks/:id",
			localMiddleware.RequirePermission("settlements.view"),
			h.GetAutoTask,
		)
		admin.POST("/auto-tasks/:id/pause",
			localMiddleware.RequirePermission("settlements.manage"),
			localMiddleware.RequireReason,
			h.PauseAutoTask,
		)
		admin.POST("/auto-tasks/:id/resume",
			localMiddleware.RequirePermission("settlements.manage"),
			localMiddleware.RequireReason,
			h.ResumeAutoTask,
		)
	}
}

// ========== 结算单管理 ==========

func (h *SettlementBFFHandler) ListSettlements(c *gin.Context) {
	queryParams := make(map[string]string)
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}
	if status := c.Query("status"); status != "" {
		queryParams["status"] = status
	}
	if settlementType := c.Query("type"); settlementType != "" {
		queryParams["type"] = settlementType
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

	result, statusCode, err := h.settlementClient.Get(c.Request.Context(), "/api/v1/settlements", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Settlement Service失败", "details": err.Error()})
		return
	}

	// 数据脱敏
	if data, ok := result["data"].(map[string]interface{}); ok {
		result["data"] = utils.MaskSensitiveData(data)
	}

	// 记录审计日志
	h.auditHelper.LogCrossTenantAccess(c, "VIEW_SETTLEMENTS", "settlement", "", queryParams["merchant_id"], statusCode)

	c.JSON(statusCode, result)
}

func (h *SettlementBFFHandler) GetSettlement(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.settlementClient.Get(c.Request.Context(), "/api/v1/settlements/"+id, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Settlement Service失败", "details": err.Error()})
		return
	}

	// 数据脱敏
	if data, ok := result["data"].(map[string]interface{}); ok {
		result["data"] = utils.MaskSensitiveData(data)
	}

	// 记录审计日志
	h.auditHelper.LogCrossTenantAccess(c, "VIEW_SETTLEMENT_DETAIL", "settlement", id, "", statusCode)

	c.JSON(statusCode, result)
}

func (h *SettlementBFFHandler) CreateSettlement(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	adminID := c.GetString("user_id")
	req["created_by"] = adminID

	result, statusCode, err := h.settlementClient.Post(c.Request.Context(), "/api/v1/settlements", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Settlement Service失败", "details": err.Error()})
		return
	}

	// 记录审计日志
	h.auditHelper.LogSensitiveOperation(c, "CREATE_SETTLEMENT", adminID, statusCode == http.StatusOK)

	c.JSON(statusCode, result)
}

func (h *SettlementBFFHandler) GetSettlementDetails(c *gin.Context) {
	id := c.Param("id")

	queryParams := make(map[string]string)
	if page := c.Query("page"); page != "" {
		queryParams["page"] = page
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		queryParams["page_size"] = pageSize
	}

	result, statusCode, err := h.settlementClient.Get(c.Request.Context(), "/api/v1/settlements/"+id+"/details", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Settlement Service失败", "details": err.Error()})
		return
	}

	// 数据脱敏
	if data, ok := result["data"].(map[string]interface{}); ok {
		result["data"] = utils.MaskSensitiveData(data)
	}

	// 记录审计日志
	h.auditHelper.LogCrossTenantAccess(c, "VIEW_SETTLEMENT_DETAILS", "settlement", id, "", statusCode)

	c.JSON(statusCode, result)
}

// ========== 结算审批 ==========

func (h *SettlementBFFHandler) ApproveSettlement(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		req = make(map[string]interface{})
	}

	adminID := c.GetString("user_id")
	req["approved_by"] = adminID

	result, statusCode, err := h.settlementClient.Post(c.Request.Context(), "/api/v1/settlements/"+id+"/approve", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Settlement Service失败", "details": err.Error()})
		return
	}

	// 记录审计日志
	h.auditHelper.LogSensitiveOperation(c, "APPROVE_SETTLEMENT", id, statusCode == http.StatusOK)

	c.JSON(statusCode, result)
}

func (h *SettlementBFFHandler) RejectSettlement(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	adminID := c.GetString("user_id")
	req["rejected_by"] = adminID

	result, statusCode, err := h.settlementClient.Post(c.Request.Context(), "/api/v1/settlements/"+id+"/reject", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Settlement Service失败", "details": err.Error()})
		return
	}

	// 记录审计日志
	h.auditHelper.LogSensitiveOperation(c, "REJECT_SETTLEMENT", id, statusCode == http.StatusOK)

	c.JSON(statusCode, result)
}

func (h *SettlementBFFHandler) ListPendingApprovals(c *gin.Context) {
	queryParams := make(map[string]string)
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}
	if page := c.Query("page"); page != "" {
		queryParams["page"] = page
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		queryParams["page_size"] = pageSize
	}

	result, statusCode, err := h.settlementClient.Get(c.Request.Context(), "/api/v1/settlements/pending", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Settlement Service失败", "details": err.Error()})
		return
	}

	// 数据脱敏
	if data, ok := result["data"].(map[string]interface{}); ok {
		result["data"] = utils.MaskSensitiveData(data)
	}

	c.JSON(statusCode, result)
}

func (h *SettlementBFFHandler) BatchApproveSettlements(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	adminID := c.GetString("user_id")
	req["approved_by"] = adminID

	result, statusCode, err := h.settlementClient.Post(c.Request.Context(), "/api/v1/settlements/batch-approve", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Settlement Service失败", "details": err.Error()})
		return
	}

	// 记录审计日志
	h.auditHelper.LogSensitiveOperation(c, "BATCH_APPROVE_SETTLEMENTS", adminID, statusCode == http.StatusOK)

	c.JSON(statusCode, result)
}

// ========== 结算统计 ==========

func (h *SettlementBFFHandler) GetStatistics(c *gin.Context) {
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

	result, statusCode, err := h.settlementClient.Get(c.Request.Context(), "/api/v1/statistics", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Settlement Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *SettlementBFFHandler) GetTrendStatistics(c *gin.Context) {
	queryParams := make(map[string]string)
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}
	if period := c.Query("period"); period != "" {
		queryParams["period"] = period
	}

	result, statusCode, err := h.settlementClient.Get(c.Request.Context(), "/api/v1/statistics/trend", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Settlement Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *SettlementBFFHandler) GetMerchantStatistics(c *gin.Context) {
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

	result, statusCode, err := h.settlementClient.Get(c.Request.Context(), "/api/v1/statistics/merchant", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Settlement Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== 结算配置 ==========

func (h *SettlementBFFHandler) GetConfig(c *gin.Context) {
	queryParams := make(map[string]string)
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}

	result, statusCode, err := h.settlementClient.Get(c.Request.Context(), "/api/v1/config", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Settlement Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *SettlementBFFHandler) UpdateConfig(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	adminID := c.GetString("user_id")
	req["updated_by"] = adminID

	result, statusCode, err := h.settlementClient.Put(c.Request.Context(), "/api/v1/config", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Settlement Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== 自动结算任务 ==========

func (h *SettlementBFFHandler) ListAutoTasks(c *gin.Context) {
	queryParams := make(map[string]string)
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
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

	result, statusCode, err := h.settlementClient.Get(c.Request.Context(), "/api/v1/auto-tasks", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Settlement Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *SettlementBFFHandler) GetAutoTask(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.settlementClient.Get(c.Request.Context(), "/api/v1/auto-tasks/"+id, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Settlement Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *SettlementBFFHandler) PauseAutoTask(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		req = make(map[string]interface{})
	}

	adminID := c.GetString("user_id")
	req["paused_by"] = adminID

	result, statusCode, err := h.settlementClient.Post(c.Request.Context(), "/api/v1/auto-tasks/"+id+"/pause", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Settlement Service失败", "details": err.Error()})
		return
	}

	// 记录审计日志
	h.auditHelper.LogSensitiveOperation(c, "PAUSE_AUTO_SETTLEMENT", id, statusCode == http.StatusOK)

	c.JSON(statusCode, result)
}

func (h *SettlementBFFHandler) ResumeAutoTask(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		req = make(map[string]interface{})
	}

	adminID := c.GetString("user_id")
	req["resumed_by"] = adminID

	result, statusCode, err := h.settlementClient.Post(c.Request.Context(), "/api/v1/auto-tasks/"+id+"/resume", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Settlement Service失败", "details": err.Error()})
		return
	}

	// 记录审计日志
	h.auditHelper.LogSensitiveOperation(c, "RESUME_AUTO_SETTLEMENT", id, statusCode == http.StatusOK)

	c.JSON(statusCode, result)
}
