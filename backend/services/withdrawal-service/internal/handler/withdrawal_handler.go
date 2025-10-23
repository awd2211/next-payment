package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"payment-platform/withdrawal-service/internal/model"
	"payment-platform/withdrawal-service/internal/service"
)

// WithdrawalHandler 提现处理器
type WithdrawalHandler struct {
	withdrawalService service.WithdrawalService
}

// NewWithdrawalHandler 创建提现处理器
func NewWithdrawalHandler(withdrawalService service.WithdrawalService) *WithdrawalHandler {
	return &WithdrawalHandler{
		withdrawalService: withdrawalService,
	}
}

// RegisterRoutes 注册路由
func (h *WithdrawalHandler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		withdrawals := api.Group("/withdrawals")
		{
			withdrawals.POST("", h.CreateWithdrawal)
			withdrawals.GET("", h.ListWithdrawals)
			withdrawals.GET("/:id", h.GetWithdrawal)
			withdrawals.POST("/:id/approve", h.ApproveWithdrawal)
			withdrawals.POST("/:id/reject", h.RejectWithdrawal)
			withdrawals.POST("/:id/execute", h.ExecuteWithdrawal)
			withdrawals.POST("/:id/cancel", h.CancelWithdrawal)
			withdrawals.GET("/reports", h.GetWithdrawalReport)
		}

		bankAccounts := api.Group("/bank-accounts")
		{
			bankAccounts.POST("", h.CreateBankAccount)
			bankAccounts.GET("", h.ListBankAccounts)
			bankAccounts.GET("/:id", h.GetBankAccount)
			bankAccounts.PUT("/:id", h.UpdateBankAccount)
			bankAccounts.POST("/:id/set-default", h.SetDefaultBankAccount)
		}
	}
}

// CreateWithdrawalRequest 创建提现请求
type CreateWithdrawalRequest struct {
	MerchantID    string               `json:"merchant_id" binding:"required"`
	Amount        int64                `json:"amount" binding:"required,min=1"`
	Type          model.WithdrawalType `json:"type" binding:"required"`
	BankAccountID string               `json:"bank_account_id" binding:"required"`
	Remarks       string               `json:"remarks"`
	CreatedBy     string               `json:"created_by" binding:"required"`
}

// CreateWithdrawal 创建提现
// @Summary 创建提现
// @Tags Withdrawals
// @Accept json
// @Produce json
// @Param request body CreateWithdrawalRequest true "创建提现请求"
// @Success 200 {object} map[string]interface{}
// @Router /withdrawals [post]
func (h *WithdrawalHandler) CreateWithdrawal(c *gin.Context) {
	var req CreateWithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	merchantID, err := uuid.Parse(req.MerchantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的商户ID"})
		return
	}

	bankAccountID, err := uuid.Parse(req.BankAccountID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的银行账户ID"})
		return
	}

	createdBy, err := uuid.Parse(req.CreatedBy)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的创建人ID"})
		return
	}

	input := &service.CreateWithdrawalInput{
		MerchantID:    merchantID,
		Amount:        req.Amount,
		Type:          req.Type,
		BankAccountID: bankAccountID,
		Remarks:       req.Remarks,
		CreatedBy:     createdBy,
	}

	withdrawal, err := h.withdrawalService.CreateWithdrawal(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": withdrawal,
	})
}

// GetWithdrawal 获取提现详情
// @Summary 获取提现详情
// @Tags Withdrawals
// @Accept json
// @Produce json
// @Param id path string true "提现ID"
// @Success 200 {object} map[string]interface{}
// @Router /withdrawals/{id} [get]
func (h *WithdrawalHandler) GetWithdrawal(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的提现ID"})
		return
	}

	detail, err := h.withdrawalService.GetWithdrawal(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": detail,
	})
}

// ListWithdrawals 提现列表
// @Summary 提现列表
// @Tags Withdrawals
// @Accept json
// @Produce json
// @Param merchant_id query string false "商户ID"
// @Param status query string false "状态"
// @Param type query string false "类型"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{}
// @Router /withdrawals [get]
func (h *WithdrawalHandler) ListWithdrawals(c *gin.Context) {
	var merchantID *uuid.UUID
	if merchantIDStr := c.Query("merchant_id"); merchantIDStr != "" {
		id, err := uuid.Parse(merchantIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的商户ID"})
			return
		}
		merchantID = &id
	}

	var status *model.WithdrawalStatus
	if statusStr := c.Query("status"); statusStr != "" {
		s := model.WithdrawalStatus(statusStr)
		status = &s
	}

	var withdrawalType *model.WithdrawalType
	if typeStr := c.Query("type"); typeStr != "" {
		t := model.WithdrawalType(typeStr)
		withdrawalType = &t
	}

	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	pageSize := 20
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	query := &service.ListWithdrawalQuery{
		MerchantID: merchantID,
		Status:     status,
		Type:       withdrawalType,
		Page:       page,
		PageSize:   pageSize,
	}

	resp, err := h.withdrawalService.ListWithdrawals(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": resp,
	})
}

// ApproveWithdrawalRequest 审批提现请求
type ApproveWithdrawalRequest struct {
	ApproverID   string `json:"approver_id" binding:"required"`
	ApproverName string `json:"approver_name" binding:"required"`
	Comments     string `json:"comments"`
}

// ApproveWithdrawal 审批通过提现
// @Summary 审批通过提现
// @Tags Withdrawals
// @Accept json
// @Produce json
// @Param id path string true "提现ID"
// @Param request body ApproveWithdrawalRequest true "审批请求"
// @Success 200 {object} map[string]interface{}
// @Router /withdrawals/{id}/approve [post]
func (h *WithdrawalHandler) ApproveWithdrawal(c *gin.Context) {
	idStr := c.Param("id")
	withdrawalID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的提现ID"})
		return
	}

	var req ApproveWithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	approverID, err := uuid.Parse(req.ApproverID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的审批人ID"})
		return
	}

	err = h.withdrawalService.ApproveWithdrawal(c.Request.Context(), withdrawalID, approverID, req.ApproverName, req.Comments)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "审批通过",
	})
}

// RejectWithdrawal 拒绝提现
// @Summary 拒绝提现
// @Tags Withdrawals
// @Accept json
// @Produce json
// @Param id path string true "提现ID"
// @Param request body ApproveWithdrawalRequest true "拒绝请求"
// @Success 200 {object} map[string]interface{}
// @Router /withdrawals/{id}/reject [post]
func (h *WithdrawalHandler) RejectWithdrawal(c *gin.Context) {
	idStr := c.Param("id")
	withdrawalID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的提现ID"})
		return
	}

	var req ApproveWithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	approverID, err := uuid.Parse(req.ApproverID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的审批人ID"})
		return
	}

	err = h.withdrawalService.RejectWithdrawal(c.Request.Context(), withdrawalID, approverID, req.ApproverName, req.Comments)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "已拒绝",
	})
}

// ExecuteWithdrawal 执行提现
// @Summary 执行提现
// @Tags Withdrawals
// @Accept json
// @Produce json
// @Param id path string true "提现ID"
// @Success 200 {object} map[string]interface{}
// @Router /withdrawals/{id}/execute [post]
func (h *WithdrawalHandler) ExecuteWithdrawal(c *gin.Context) {
	idStr := c.Param("id")
	withdrawalID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的提现ID"})
		return
	}

	err = h.withdrawalService.ExecuteWithdrawal(c.Request.Context(), withdrawalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "提现执行成功",
	})
}

// CancelWithdrawalRequest 取消提现请求
type CancelWithdrawalRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// CancelWithdrawal 取消提现
// @Summary 取消提现
// @Tags Withdrawals
// @Accept json
// @Produce json
// @Param id path string true "提现ID"
// @Param request body CancelWithdrawalRequest true "取消请求"
// @Success 200 {object} map[string]interface{}
// @Router /withdrawals/{id}/cancel [post]
func (h *WithdrawalHandler) CancelWithdrawal(c *gin.Context) {
	idStr := c.Param("id")
	withdrawalID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的提现ID"})
		return
	}

	var req CancelWithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.withdrawalService.CancelWithdrawal(c.Request.Context(), withdrawalID, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "提现已取消",
	})
}

// GetWithdrawalReport 获取提现报表
// @Summary 获取提现报表
// @Tags Withdrawals
// @Accept json
// @Produce json
// @Param merchant_id query string true "商户ID"
// @Param start_date query string true "开始日期"
// @Param end_date query string true "结束日期"
// @Success 200 {object} map[string]interface{}
// @Router /withdrawals/reports [get]
func (h *WithdrawalHandler) GetWithdrawalReport(c *gin.Context) {
	merchantIDStr := c.Query("merchant_id")
	if merchantIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "商户ID不能为空"})
		return
	}

	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的商户ID"})
		return
	}

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的开始日期"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的结束日期"})
		return
	}

	report, err := h.withdrawalService.GetWithdrawalReport(c.Request.Context(), merchantID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": report,
	})
}

// Bank Account Handlers

// CreateBankAccountRequest 创建银行账户请求
type CreateBankAccountRequest struct {
	MerchantID      string `json:"merchant_id" binding:"required"`
	BankName        string `json:"bank_name" binding:"required"`
	BankCode        string `json:"bank_code" binding:"required"`
	BankBranch      string `json:"bank_branch"`
	AccountName     string `json:"account_name" binding:"required"`
	AccountNo       string `json:"account_no" binding:"required"`
	AccountType     string `json:"account_type" binding:"required"`
	IsDefault       bool   `json:"is_default"`
	VerificationDoc string `json:"verification_doc"`
}

// CreateBankAccount 创建银行账户
// @Summary 创建银行账户
// @Tags BankAccounts
// @Accept json
// @Produce json
// @Param request body CreateBankAccountRequest true "创建银行账户请求"
// @Success 200 {object} map[string]interface{}
// @Router /bank-accounts [post]
func (h *WithdrawalHandler) CreateBankAccount(c *gin.Context) {
	var req CreateBankAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	merchantID, err := uuid.Parse(req.MerchantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的商户ID"})
		return
	}

	input := &service.CreateBankAccountInput{
		MerchantID:      merchantID,
		BankName:        req.BankName,
		BankCode:        req.BankCode,
		BankBranch:      req.BankBranch,
		AccountName:     req.AccountName,
		AccountNo:       req.AccountNo,
		AccountType:     req.AccountType,
		IsDefault:       req.IsDefault,
		VerificationDoc: req.VerificationDoc,
	}

	account, err := h.withdrawalService.CreateBankAccount(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": account,
	})
}

// GetBankAccount 获取银行账户
// @Summary 获取银行账户
// @Tags BankAccounts
// @Accept json
// @Produce json
// @Param id path string true "账户ID"
// @Success 200 {object} map[string]interface{}
// @Router /bank-accounts/{id} [get]
func (h *WithdrawalHandler) GetBankAccount(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的账户ID"})
		return
	}

	account, err := h.withdrawalService.GetBankAccount(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": account,
	})
}

// ListBankAccounts 银行账户列表
// @Summary 银行账户列表
// @Tags BankAccounts
// @Accept json
// @Produce json
// @Param merchant_id query string true "商户ID"
// @Success 200 {object} map[string]interface{}
// @Router /bank-accounts [get]
func (h *WithdrawalHandler) ListBankAccounts(c *gin.Context) {
	merchantIDStr := c.Query("merchant_id")
	if merchantIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "商户ID不能为空"})
		return
	}

	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的商户ID"})
		return
	}

	accounts, err := h.withdrawalService.ListBankAccounts(c.Request.Context(), merchantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": accounts,
	})
}

// UpdateBankAccountRequest 更新银行账户请求
type UpdateBankAccountRequest struct {
	BankBranch      *string `json:"bank_branch"`
	IsDefault       *bool   `json:"is_default"`
	Status          *string `json:"status"`
	VerificationDoc *string `json:"verification_doc"`
	IsVerified      *bool   `json:"is_verified"`
}

// UpdateBankAccount 更新银行账户
// @Summary 更新银行账户
// @Tags BankAccounts
// @Accept json
// @Produce json
// @Param id path string true "账户ID"
// @Param request body UpdateBankAccountRequest true "更新请求"
// @Success 200 {object} map[string]interface{}
// @Router /bank-accounts/{id} [put]
func (h *WithdrawalHandler) UpdateBankAccount(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的账户ID"})
		return
	}

	var req UpdateBankAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := &service.UpdateBankAccountInput{
		BankBranch:      req.BankBranch,
		IsDefault:       req.IsDefault,
		Status:          req.Status,
		VerificationDoc: req.VerificationDoc,
		IsVerified:      req.IsVerified,
	}

	err = h.withdrawalService.UpdateBankAccount(c.Request.Context(), id, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "更新成功",
	})
}

// SetDefaultBankAccount 设置默认银行账户
// @Summary 设置默认银行账户
// @Tags BankAccounts
// @Accept json
// @Produce json
// @Param id path string true "账户ID"
// @Param merchant_id query string true "商户ID"
// @Success 200 {object} map[string]interface{}
// @Router /bank-accounts/{id}/set-default [post]
func (h *WithdrawalHandler) SetDefaultBankAccount(c *gin.Context) {
	idStr := c.Param("id")
	accountID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的账户ID"})
		return
	}

	merchantIDStr := c.Query("merchant_id")
	if merchantIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "商户ID不能为空"})
		return
	}

	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的商户ID"})
		return
	}

	err = h.withdrawalService.SetDefaultBankAccount(c.Request.Context(), merchantID, accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "已设置为默认账户",
	})
}
