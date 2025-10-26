package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"payment-platform/admin-service/internal/client"
)

type WithdrawalBFFHandler struct {
	withdrawalClient *client.ServiceClient
}

func NewWithdrawalBFFHandler(withdrawalServiceURL string) *WithdrawalBFFHandler {
	return &WithdrawalBFFHandler{
		withdrawalClient: client.NewServiceClient(withdrawalServiceURL),
	}
}

func (h *WithdrawalBFFHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	admin := r.Group("/admin/withdrawals")
	admin.Use(authMiddleware)
	{
		// 提现单管理
		withdrawals := admin.Group("/withdrawals")
		{
			withdrawals.GET("", h.ListWithdrawals)
			withdrawals.GET("/:id", h.GetWithdrawal)
			withdrawals.POST("", h.CreateWithdrawal)
		}

		// 提现审批
		approvals := admin.Group("/approvals")
		{
			approvals.POST("/:id/approve", h.ApproveWithdrawal)
			approvals.POST("/:id/reject", h.RejectWithdrawal)
			approvals.GET("/pending", h.ListPendingApprovals)
			approvals.POST("/batch-approve", h.BatchApproveWithdrawals)
		}

		// 银行账户管理
		bankAccounts := admin.Group("/bank-accounts")
		{
			bankAccounts.GET("", h.ListBankAccounts)
			bankAccounts.GET("/:id", h.GetBankAccount)
			bankAccounts.POST("", h.CreateBankAccount)
			bankAccounts.PUT("/:id", h.UpdateBankAccount)
			bankAccounts.DELETE("/:id", h.DeleteBankAccount)
			bankAccounts.POST("/:id/verify", h.VerifyBankAccount)
		}

		// 提现统计
		admin.GET("/statistics", h.GetStatistics)
		admin.GET("/statistics/trend", h.GetTrendStatistics)

		// 提现配置
		admin.GET("/config", h.GetConfig)
		admin.PUT("/config", h.UpdateConfig)

		// 提现渠道
		admin.GET("/channels", h.ListChannels)
		admin.GET("/channels/:id", h.GetChannel)
	}
}

// ========== 提现单管理 ==========

func (h *WithdrawalBFFHandler) ListWithdrawals(c *gin.Context) {
	queryParams := make(map[string]string)
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}
	if status := c.Query("status"); status != "" {
		queryParams["status"] = status
	}
	if channel := c.Query("channel"); channel != "" {
		queryParams["channel"] = channel
	}
	if startTime := c.Query("start_time"); startTime != "" {
		queryParams["start_time"] = startTime
	}
	if endTime := c.Query("end_time"); endTime != "" {
		queryParams["end_time"] = endTime
	}
	if page := c.Query("page"); page != "" {
		queryParams["page"] = page
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		queryParams["page_size"] = pageSize
	}

	result, statusCode, err := h.withdrawalClient.Get(c.Request.Context(), "/api/v1/withdrawals", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Withdrawal Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *WithdrawalBFFHandler) GetWithdrawal(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.withdrawalClient.Get(c.Request.Context(), "/api/v1/withdrawals/"+id, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Withdrawal Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *WithdrawalBFFHandler) CreateWithdrawal(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	adminID := c.GetString("user_id")
	req["created_by"] = adminID

	result, statusCode, err := h.withdrawalClient.Post(c.Request.Context(), "/api/v1/withdrawals", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Withdrawal Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== 提现审批 ==========

func (h *WithdrawalBFFHandler) ApproveWithdrawal(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		req = make(map[string]interface{})
	}

	adminID := c.GetString("user_id")
	req["approved_by"] = adminID

	result, statusCode, err := h.withdrawalClient.Post(c.Request.Context(), "/api/v1/withdrawals/"+id+"/approve", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Withdrawal Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *WithdrawalBFFHandler) RejectWithdrawal(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	adminID := c.GetString("user_id")
	req["rejected_by"] = adminID

	result, statusCode, err := h.withdrawalClient.Post(c.Request.Context(), "/api/v1/withdrawals/"+id+"/reject", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Withdrawal Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *WithdrawalBFFHandler) ListPendingApprovals(c *gin.Context) {
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

	result, statusCode, err := h.withdrawalClient.Get(c.Request.Context(), "/api/v1/withdrawals/pending", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Withdrawal Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *WithdrawalBFFHandler) BatchApproveWithdrawals(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	adminID := c.GetString("user_id")
	req["approved_by"] = adminID

	result, statusCode, err := h.withdrawalClient.Post(c.Request.Context(), "/api/v1/withdrawals/batch-approve", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Withdrawal Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== 银行账户管理 ==========

func (h *WithdrawalBFFHandler) ListBankAccounts(c *gin.Context) {
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

	result, statusCode, err := h.withdrawalClient.Get(c.Request.Context(), "/api/v1/bank-accounts", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Withdrawal Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *WithdrawalBFFHandler) GetBankAccount(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.withdrawalClient.Get(c.Request.Context(), "/api/v1/bank-accounts/"+id, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Withdrawal Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *WithdrawalBFFHandler) CreateBankAccount(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	adminID := c.GetString("user_id")
	req["created_by"] = adminID

	result, statusCode, err := h.withdrawalClient.Post(c.Request.Context(), "/api/v1/bank-accounts", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Withdrawal Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *WithdrawalBFFHandler) UpdateBankAccount(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	adminID := c.GetString("user_id")
	req["updated_by"] = adminID

	result, statusCode, err := h.withdrawalClient.Put(c.Request.Context(), "/api/v1/bank-accounts/"+id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Withdrawal Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *WithdrawalBFFHandler) DeleteBankAccount(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.withdrawalClient.Delete(c.Request.Context(), "/api/v1/bank-accounts/"+id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Withdrawal Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *WithdrawalBFFHandler) VerifyBankAccount(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		req = make(map[string]interface{})
	}

	adminID := c.GetString("user_id")
	req["verified_by"] = adminID

	result, statusCode, err := h.withdrawalClient.Post(c.Request.Context(), "/api/v1/bank-accounts/"+id+"/verify", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Withdrawal Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== 提现统计 ==========

func (h *WithdrawalBFFHandler) GetStatistics(c *gin.Context) {
	queryParams := make(map[string]string)
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}
	if startTime := c.Query("start_time"); startTime != "" {
		queryParams["start_time"] = startTime
	}
	if endTime := c.Query("end_time"); endTime != "" {
		queryParams["end_time"] = endTime
	}

	result, statusCode, err := h.withdrawalClient.Get(c.Request.Context(), "/api/v1/statistics", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Withdrawal Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *WithdrawalBFFHandler) GetTrendStatistics(c *gin.Context) {
	queryParams := make(map[string]string)
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}
	if period := c.Query("period"); period != "" {
		queryParams["period"] = period
	}

	result, statusCode, err := h.withdrawalClient.Get(c.Request.Context(), "/api/v1/statistics/trend", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Withdrawal Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== 提现配置 ==========

func (h *WithdrawalBFFHandler) GetConfig(c *gin.Context) {
	queryParams := make(map[string]string)
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}

	result, statusCode, err := h.withdrawalClient.Get(c.Request.Context(), "/api/v1/config", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Withdrawal Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *WithdrawalBFFHandler) UpdateConfig(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	adminID := c.GetString("user_id")
	req["updated_by"] = adminID

	result, statusCode, err := h.withdrawalClient.Put(c.Request.Context(), "/api/v1/config", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Withdrawal Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

// ========== 提现渠道 ==========

func (h *WithdrawalBFFHandler) ListChannels(c *gin.Context) {
	queryParams := make(map[string]string)
	if status := c.Query("status"); status != "" {
		queryParams["status"] = status
	}

	result, statusCode, err := h.withdrawalClient.Get(c.Request.Context(), "/api/v1/channels", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Withdrawal Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}

func (h *WithdrawalBFFHandler) GetChannel(c *gin.Context) {
	id := c.Param("id")

	result, statusCode, err := h.withdrawalClient.Get(c.Request.Context(), "/api/v1/channels/"+id, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "调用Withdrawal Service失败", "details": err.Error()})
		return
	}

	c.JSON(statusCode, result)
}
