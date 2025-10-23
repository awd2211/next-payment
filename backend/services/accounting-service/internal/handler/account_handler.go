package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"payment-platform/accounting-service/internal/repository"
	"payment-platform/accounting-service/internal/service"
)

// AccountHandler 账户处理器
type AccountHandler struct {
	accountService service.AccountService
}

// NewAccountHandler 创建账户处理器实例
func NewAccountHandler(accountService service.AccountService) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
	}
}

// RegisterRoutes 注册路由
func (h *AccountHandler) RegisterRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		// 账户管理
		accounts := v1.Group("/accounts")
		{
			accounts.POST("", h.CreateAccount)
			accounts.GET("/:id", h.GetAccount)
			accounts.GET("", h.ListAccounts)
			accounts.POST("/:id/freeze", h.FreezeAccount)
			accounts.POST("/:id/unfreeze", h.UnfreezeAccount)
		}

		// 交易管理
		transactions := v1.Group("/transactions")
		{
			transactions.POST("", h.CreateTransaction)
			transactions.GET("/:transactionNo", h.GetTransaction)
			transactions.GET("", h.ListTransactions)
			transactions.POST("/:transactionNo/reverse", h.ReverseTransaction)
		}

		// 结算管理
		settlements := v1.Group("/settlements")
		{
			settlements.POST("", h.CreateSettlement)
			settlements.GET("/:settlementNo", h.GetSettlement)
			settlements.GET("", h.ListSettlements)
			settlements.POST("/:settlementNo/process", h.ProcessSettlement)
		}

		// 提现管理
		withdrawals := v1.Group("/withdrawals")
		{
			withdrawals.POST("", h.CreateWithdrawal)
			withdrawals.GET("/:withdrawalNo", h.GetWithdrawal)
			withdrawals.GET("", h.ListWithdrawals)
			withdrawals.POST("/:withdrawalNo/approve", h.ApproveWithdrawal)
			withdrawals.POST("/:withdrawalNo/reject", h.RejectWithdrawal)
			withdrawals.POST("/:withdrawalNo/process", h.ProcessWithdrawal)
			withdrawals.POST("/:withdrawalNo/complete", h.CompleteWithdrawal)
			withdrawals.POST("/:withdrawalNo/fail", h.FailWithdrawal)
			withdrawals.POST("/:withdrawalNo/cancel", h.CancelWithdrawal)
		}

		// 账单管理
		invoices := v1.Group("/invoices")
		{
			invoices.POST("", h.CreateInvoice)
			invoices.GET("/:invoiceNo", h.GetInvoice)
			invoices.GET("", h.ListInvoices)
			invoices.POST("/:invoiceNo/pay", h.PayInvoice)
			invoices.POST("/:invoiceNo/cancel", h.CancelInvoice)
			invoices.POST("/:invoiceNo/void", h.VoidInvoice)
		}

		// 对账管理
		reconciliations := v1.Group("/reconciliations")
		{
			reconciliations.POST("", h.CreateReconciliation)
			reconciliations.GET("/:reconciliationNo", h.GetReconciliation)
			reconciliations.GET("", h.ListReconciliations)
			reconciliations.POST("/:reconciliationNo/process", h.ProcessReconciliation)
			reconciliations.POST("/:reconciliationNo/complete", h.CompleteReconciliation)
			reconciliations.POST("/items/:itemID/resolve", h.ResolveReconciliationItem)
		}

		// 余额查询聚合
		balances := v1.Group("/balances")
		{
			balances.GET("/merchants/:merchantID/summary", h.GetMerchantBalanceSummary)
			balances.GET("/merchants/:merchantID/currencies/:currency", h.GetBalanceByCurrency)
			balances.GET("/merchants/:merchantID/account-types/:accountType", h.GetBalanceByAccountType)
			balances.GET("/merchants/:merchantID/currencies", h.GetAllCurrencyBalances)
		}
	}
}

// CreateAccount 创建账户
func (h *AccountHandler) CreateAccount(c *gin.Context) {
	var input service.CreateAccountInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	account, err := h.accountService.CreateAccount(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(account))
}

// GetAccount 获取账户
func (h *AccountHandler) GetAccount(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("无效的账户ID"))
		return
	}

	account, err := h.accountService.GetAccount(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(account))
}

// ListAccounts 账户列表
func (h *AccountHandler) ListAccounts(c *gin.Context) {
	query := &repository.AccountQuery{
		AccountType: c.Query("account_type"),
		Currency:    c.Query("currency"),
		Status:      c.Query("status"),
	}

	if merchantIDStr := c.Query("merchant_id"); merchantIDStr != "" {
		merchantID, err := uuid.Parse(merchantIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse("无效的商户ID"))
			return
		}
		query.MerchantID = &merchantID
	}

	query.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	query.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))

	accounts, total, err := h.accountService.ListAccounts(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(PageResponse{
		List:     accounts,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}))
}

// FreezeAccount 冻结账户
func (h *AccountHandler) FreezeAccount(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("无效的账户ID"))
		return
	}

	if err := h.accountService.FreezeAccount(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(nil))
}

// UnfreezeAccount 解冻账户
func (h *AccountHandler) UnfreezeAccount(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("无效的账户ID"))
		return
	}

	if err := h.accountService.UnfreezeAccount(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(nil))
}

// CreateTransaction 创建交易
func (h *AccountHandler) CreateTransaction(c *gin.Context) {
	var input service.CreateTransactionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	transaction, err := h.accountService.CreateTransaction(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(transaction))
}

// GetTransaction 获取交易
func (h *AccountHandler) GetTransaction(c *gin.Context) {
	transactionNo := c.Param("transactionNo")

	transaction, err := h.accountService.GetTransaction(c.Request.Context(), transactionNo)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(transaction))
}

// ListTransactions 交易列表
func (h *AccountHandler) ListTransactions(c *gin.Context) {
	query := &repository.TransactionQuery{
		TransactionType: c.Query("transaction_type"),
		Currency:        c.Query("currency"),
	}

	if accountIDStr := c.Query("account_id"); accountIDStr != "" {
		accountID, err := uuid.Parse(accountIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse("无效的账户ID"))
			return
		}
		query.AccountID = &accountID
	}

	if merchantIDStr := c.Query("merchant_id"); merchantIDStr != "" {
		merchantID, err := uuid.Parse(merchantIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse("无效的商户ID"))
			return
		}
		query.MerchantID = &merchantID
	}

	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err == nil {
			query.StartTime = &startTime
		}
	}
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		endTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err == nil {
			query.EndTime = &endTime
		}
	}

	query.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	query.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))

	transactions, total, err := h.accountService.ListTransactions(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(PageResponse{
		List:     transactions,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}))
}

// ReverseTransaction 冲正交易
func (h *AccountHandler) ReverseTransaction(c *gin.Context) {
	transactionNo := c.Param("transactionNo")

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	if err := h.accountService.ReverseTransaction(c.Request.Context(), transactionNo, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(nil))
}

// CreateSettlement 创建结算
func (h *AccountHandler) CreateSettlement(c *gin.Context) {
	var input service.CreateSettlementInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	settlement, err := h.accountService.CreateSettlement(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(settlement))
}

// GetSettlement 获取结算
func (h *AccountHandler) GetSettlement(c *gin.Context) {
	settlementNo := c.Param("settlementNo")

	settlement, err := h.accountService.GetSettlement(c.Request.Context(), settlementNo)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(settlement))
}

// ListSettlements 结算列表
func (h *AccountHandler) ListSettlements(c *gin.Context) {
	query := &repository.SettlementQuery{
		Status: c.Query("status"),
	}

	if merchantIDStr := c.Query("merchant_id"); merchantIDStr != "" {
		merchantID, err := uuid.Parse(merchantIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse("无效的商户ID"))
			return
		}
		query.MerchantID = &merchantID
	}

	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err == nil {
			query.StartTime = &startTime
		}
	}
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		endTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err == nil {
			query.EndTime = &endTime
		}
	}

	query.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	query.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))

	settlements, total, err := h.accountService.ListSettlements(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(PageResponse{
		List:     settlements,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}))
}

// ProcessSettlement 处理结算
func (h *AccountHandler) ProcessSettlement(c *gin.Context) {
	settlementNo := c.Param("settlementNo")

	if err := h.accountService.ProcessSettlement(c.Request.Context(), settlementNo); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(nil))
}

// Response structures

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type PageResponse struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
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

// Withdrawal Management Handlers

// CreateWithdrawal 创建提现申请
func (h *AccountHandler) CreateWithdrawal(c *gin.Context) {
	var input service.CreateWithdrawalInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	withdrawal, err := h.accountService.CreateWithdrawal(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(withdrawal))
}

// GetWithdrawal 获取提现记录
func (h *AccountHandler) GetWithdrawal(c *gin.Context) {
	withdrawalNo := c.Param("withdrawalNo")

	withdrawal, err := h.accountService.GetWithdrawal(c.Request.Context(), withdrawalNo)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(withdrawal))
}

// ListWithdrawals 提现列表
func (h *AccountHandler) ListWithdrawals(c *gin.Context) {
	query := &repository.WithdrawalQuery{
		Status: c.Query("status"),
	}

	if merchantIDStr := c.Query("merchant_id"); merchantIDStr != "" {
		merchantID, err := uuid.Parse(merchantIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse("无效的商户ID"))
			return
		}
		query.MerchantID = &merchantID
	}

	if accountIDStr := c.Query("account_id"); accountIDStr != "" {
		accountID, err := uuid.Parse(accountIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse("无效的账户ID"))
			return
		}
		query.AccountID = &accountID
	}

	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err == nil {
			query.StartTime = &startTime
		}
	}
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		endTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err == nil {
			query.EndTime = &endTime
		}
	}

	query.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	query.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))

	withdrawals, total, err := h.accountService.ListWithdrawals(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(PageResponse{
		List:     withdrawals,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}))
}

// ApproveWithdrawal 批准提现
func (h *AccountHandler) ApproveWithdrawal(c *gin.Context) {
	withdrawalNo := c.Param("withdrawalNo")

	var req struct {
		ApproverID uuid.UUID `json:"approver_id" binding:"required"`
		Notes      string    `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	if err := h.accountService.ApproveWithdrawal(c.Request.Context(), withdrawalNo, req.ApproverID, req.Notes); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(nil))
}

// RejectWithdrawal 拒绝提现
func (h *AccountHandler) RejectWithdrawal(c *gin.Context) {
	withdrawalNo := c.Param("withdrawalNo")

	var req struct {
		ApproverID uuid.UUID `json:"approver_id" binding:"required"`
		Reason     string    `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	if err := h.accountService.RejectWithdrawal(c.Request.Context(), withdrawalNo, req.ApproverID, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(nil))
}

// ProcessWithdrawal 处理提现
func (h *AccountHandler) ProcessWithdrawal(c *gin.Context) {
	withdrawalNo := c.Param("withdrawalNo")

	var req struct {
		ProcessorID uuid.UUID `json:"processor_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	if err := h.accountService.ProcessWithdrawal(c.Request.Context(), withdrawalNo, req.ProcessorID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(nil))
}

// CompleteWithdrawal 完成提现
func (h *AccountHandler) CompleteWithdrawal(c *gin.Context) {
	withdrawalNo := c.Param("withdrawalNo")

	if err := h.accountService.CompleteWithdrawal(c.Request.Context(), withdrawalNo); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(nil))
}

// FailWithdrawal 提现失败
func (h *AccountHandler) FailWithdrawal(c *gin.Context) {
	withdrawalNo := c.Param("withdrawalNo")

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	if err := h.accountService.FailWithdrawal(c.Request.Context(), withdrawalNo, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(nil))
}

// CancelWithdrawal 取消提现
func (h *AccountHandler) CancelWithdrawal(c *gin.Context) {
	withdrawalNo := c.Param("withdrawalNo")

	if err := h.accountService.CancelWithdrawal(c.Request.Context(), withdrawalNo); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(nil))
}

// Invoice Management Handlers

// CreateInvoice 创建账单
func (h *AccountHandler) CreateInvoice(c *gin.Context) {
	var input service.CreateInvoiceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	invoice, err := h.accountService.CreateInvoice(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(invoice))
}

// GetInvoice 获取账单
func (h *AccountHandler) GetInvoice(c *gin.Context) {
	invoiceNo := c.Param("invoiceNo")

	invoice, err := h.accountService.GetInvoice(c.Request.Context(), invoiceNo)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(invoice))
}

// ListInvoices 账单列表
func (h *AccountHandler) ListInvoices(c *gin.Context) {
	query := &repository.InvoiceQuery{
		InvoiceType: c.Query("invoice_type"),
		Status:      c.Query("status"),
	}

	if merchantIDStr := c.Query("merchant_id"); merchantIDStr != "" {
		merchantID, err := uuid.Parse(merchantIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse("无效的商户ID"))
			return
		}
		query.MerchantID = &merchantID
	}

	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err == nil {
			query.StartTime = &startTime
		}
	}
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		endTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err == nil {
			query.EndTime = &endTime
		}
	}

	query.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	query.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))

	invoices, total, err := h.accountService.ListInvoices(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(PageResponse{
		List:     invoices,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}))
}

// PayInvoice 支付账单
func (h *AccountHandler) PayInvoice(c *gin.Context) {
	invoiceNo := c.Param("invoiceNo")

	var req struct {
		PaidAmount int64 `json:"paid_amount" binding:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	if err := h.accountService.PayInvoice(c.Request.Context(), invoiceNo, req.PaidAmount); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(nil))
}

// CancelInvoice 取消账单
func (h *AccountHandler) CancelInvoice(c *gin.Context) {
	invoiceNo := c.Param("invoiceNo")

	if err := h.accountService.CancelInvoice(c.Request.Context(), invoiceNo); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(nil))
}

// VoidInvoice 作废账单
func (h *AccountHandler) VoidInvoice(c *gin.Context) {
	invoiceNo := c.Param("invoiceNo")

	if err := h.accountService.VoidInvoice(c.Request.Context(), invoiceNo); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(nil))
}

// Reconciliation Management Handlers

// CreateReconciliation 创建对账单
func (h *AccountHandler) CreateReconciliation(c *gin.Context) {
	var input service.CreateReconciliationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	reconciliation, err := h.accountService.CreateReconciliation(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(reconciliation))
}

// GetReconciliation 获取对账单
func (h *AccountHandler) GetReconciliation(c *gin.Context) {
	reconciliationNo := c.Param("reconciliationNo")

	reconciliation, err := h.accountService.GetReconciliation(c.Request.Context(), reconciliationNo)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(reconciliation))
}

// ListReconciliations 对账单列表
func (h *AccountHandler) ListReconciliations(c *gin.Context) {
	query := &repository.ReconciliationQuery{
		Channel: c.Query("channel"),
		Status:  c.Query("status"),
	}

	if merchantIDStr := c.Query("merchant_id"); merchantIDStr != "" {
		merchantID, err := uuid.Parse(merchantIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse("无效的商户ID"))
			return
		}
		query.MerchantID = &merchantID
	}

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			query.StartDate = &startDate
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			query.EndDate = &endDate
		}
	}

	query.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	query.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))

	reconciliations, total, err := h.accountService.ListReconciliations(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(PageResponse{
		List:     reconciliations,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}))
}

// ProcessReconciliation 处理对账单
func (h *AccountHandler) ProcessReconciliation(c *gin.Context) {
	reconciliationNo := c.Param("reconciliationNo")

	var req struct {
		UserID uuid.UUID `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	if err := h.accountService.ProcessReconciliation(c.Request.Context(), reconciliationNo, req.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(nil))
}

// CompleteReconciliation 完成对账
func (h *AccountHandler) CompleteReconciliation(c *gin.Context) {
	reconciliationNo := c.Param("reconciliationNo")

	if err := h.accountService.CompleteReconciliation(c.Request.Context(), reconciliationNo); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(nil))
}

// ResolveReconciliationItem 解决对账明细差异
func (h *AccountHandler) ResolveReconciliationItem(c *gin.Context) {
	itemIDStr := c.Param("itemID")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("无效的明细ID"))
		return
	}

	var req struct {
		Resolution string `json:"resolution" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err.Error()))
		return
	}

	if err := h.accountService.ResolveReconciliationItem(c.Request.Context(), itemID, req.Resolution); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(nil))
}

// Balance Aggregation Handlers

// GetMerchantBalanceSummary 获取商户余额汇总
func (h *AccountHandler) GetMerchantBalanceSummary(c *gin.Context) {
	merchantIDStr := c.Param("merchantID")
	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("无效的商户ID"))
		return
	}

	summary, err := h.accountService.GetMerchantBalanceSummary(c.Request.Context(), merchantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(summary))
}

// GetBalanceByCurrency 按货币获取余额汇总
func (h *AccountHandler) GetBalanceByCurrency(c *gin.Context) {
	merchantIDStr := c.Param("merchantID")
	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("无效的商户ID"))
		return
	}

	currency := c.Param("currency")
	if currency == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse("货币代码不能为空"))
		return
	}

	summary, err := h.accountService.GetBalanceByCurrency(c.Request.Context(), merchantID, currency)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(summary))
}

// GetBalanceByAccountType 按账户类型获取余额汇总
func (h *AccountHandler) GetBalanceByAccountType(c *gin.Context) {
	merchantIDStr := c.Param("merchantID")
	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("无效的商户ID"))
		return
	}

	accountType := c.Param("accountType")
	if accountType == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse("账户类型不能为空"))
		return
	}

	summary, err := h.accountService.GetBalanceByAccountType(c.Request.Context(), merchantID, accountType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(summary))
}

// GetAllCurrencyBalances 获取所有货币的余额汇总
func (h *AccountHandler) GetAllCurrencyBalances(c *gin.Context) {
	merchantIDStr := c.Param("merchantID")
	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse("无效的商户ID"))
		return
	}

	summaries, err := h.accountService.GetAllCurrencyBalances(c.Request.Context(), merchantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(summaries))
}
