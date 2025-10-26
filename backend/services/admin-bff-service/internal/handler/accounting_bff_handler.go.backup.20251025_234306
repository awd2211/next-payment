package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"payment-platform/admin-service/internal/client"
)

// AccountingBFFHandler Accounting Service BFF处理器
type AccountingBFFHandler struct {
	accountingClient *client.ServiceClient
}

// NewAccountingBFFHandler 创建Accounting BFF处理器
func NewAccountingBFFHandler(accountingServiceURL string) *AccountingBFFHandler {
	return &AccountingBFFHandler{
		accountingClient: client.NewServiceClient(accountingServiceURL),
	}
}

// RegisterRoutes 注册路由
func (h *AccountingBFFHandler) RegisterRoutes(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	admin := r.Group("/admin/accounting")
	admin.Use(authMiddleware)
	{
		// 账户管理
		accounts := admin.Group("/accounts")
		{
			accounts.POST("", h.CreateAccount)
			accounts.GET("/:id", h.GetAccount)
			accounts.GET("", h.ListAccounts)
			accounts.POST("/:id/freeze", h.FreezeAccount)
			accounts.POST("/:id/unfreeze", h.UnfreezeAccount)
		}

		// 交易管理
		transactions := admin.Group("/transactions")
		{
			transactions.POST("", h.CreateTransaction)
			transactions.GET("/:transactionNo", h.GetTransaction)
			transactions.GET("", h.ListTransactions)
			transactions.POST("/:transactionNo/reverse", h.ReverseTransaction)
		}

		// 结算管理
		settlements := admin.Group("/settlements")
		{
			settlements.POST("", h.CreateSettlement)
			settlements.GET("/:settlementNo", h.GetSettlement)
			settlements.GET("", h.ListSettlements)
			settlements.POST("/:settlementNo/process", h.ProcessSettlement)
		}

		// 提现管理
		withdrawals := admin.Group("/withdrawals")
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

		// 发票管理
		invoices := admin.Group("/invoices")
		{
			invoices.POST("", h.CreateInvoice)
			invoices.GET("/:invoiceNo", h.GetInvoice)
			invoices.GET("", h.ListInvoices)
			invoices.POST("/:invoiceNo/pay", h.PayInvoice)
			invoices.POST("/:invoiceNo/cancel", h.CancelInvoice)
			invoices.POST("/:invoiceNo/void", h.VoidInvoice)
		}

		// 对账管理
		reconciliations := admin.Group("/reconciliations")
		{
			reconciliations.POST("", h.CreateReconciliation)
			reconciliations.GET("/:reconciliationNo", h.GetReconciliation)
			reconciliations.GET("", h.ListReconciliations)
			reconciliations.POST("/:reconciliationNo/process", h.ProcessReconciliation)
			reconciliations.POST("/:reconciliationNo/complete", h.CompleteReconciliation)
			reconciliations.POST("/items/:itemID/resolve", h.ResolveReconciliationItem)
		}

		// 余额查询
		balances := admin.Group("/balances")
		{
			balances.GET("/merchants/:merchantID/summary", h.GetMerchantBalanceSummary)
			balances.GET("/merchants/:merchantID/currencies/:currency", h.GetBalanceByCurrency)
			balances.GET("/merchants/:merchantID/account-types/:accountType", h.GetBalanceByAccountType)
			balances.GET("/merchants/:merchantID/currencies", h.GetAllCurrencyBalances)
		}

		// 币种转换
		conversions := admin.Group("/conversions")
		{
			conversions.POST("", h.CreateCurrencyConversion)
			conversions.GET("/:conversionNo", h.GetCurrencyConversion)
			conversions.GET("", h.ListCurrencyConversions)
			conversions.POST("/:conversionNo/process", h.ProcessCurrencyConversion)
			conversions.POST("/:conversionNo/cancel", h.CancelCurrencyConversion)
		}
	}
}

// ==================== 账户管理 ====================
func (h *AccountingBFFHandler) CreateAccount(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, statusCode, err := h.accountingClient.Post(c.Request.Context(), "/api/v1/accounts", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *AccountingBFFHandler) GetAccount(c *gin.Context) {
	id := c.Param("id")
	result, statusCode, err := h.accountingClient.Get(c.Request.Context(), "/api/v1/accounts/"+id, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *AccountingBFFHandler) ListAccounts(c *gin.Context) {
	queryParams := map[string]string{
		"page":      c.DefaultQuery("page", "1"),
		"page_size": c.DefaultQuery("page_size", "10"),
	}
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}
	if accountType := c.Query("account_type"); accountType != "" {
		queryParams["account_type"] = accountType
	}
	if currency := c.Query("currency"); currency != "" {
		queryParams["currency"] = currency
	}
	result, statusCode, err := h.accountingClient.Get(c.Request.Context(), "/api/v1/accounts", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *AccountingBFFHandler) FreezeAccount(c *gin.Context) {
	id := c.Param("id")
	var req map[string]interface{}
	c.ShouldBindJSON(&req)
	result, statusCode, err := h.accountingClient.Post(c.Request.Context(), "/api/v1/accounts/"+id+"/freeze", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *AccountingBFFHandler) UnfreezeAccount(c *gin.Context) {
	id := c.Param("id")
	var req map[string]interface{}
	c.ShouldBindJSON(&req)
	result, statusCode, err := h.accountingClient.Post(c.Request.Context(), "/api/v1/accounts/"+id+"/unfreeze", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

// ==================== 交易管理 ====================
func (h *AccountingBFFHandler) CreateTransaction(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, statusCode, err := h.accountingClient.Post(c.Request.Context(), "/api/v1/transactions", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *AccountingBFFHandler) GetTransaction(c *gin.Context) {
	transactionNo := c.Param("transactionNo")
	result, statusCode, err := h.accountingClient.Get(c.Request.Context(), "/api/v1/transactions/"+transactionNo, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *AccountingBFFHandler) ListTransactions(c *gin.Context) {
	queryParams := map[string]string{
		"page":      c.DefaultQuery("page", "1"),
		"page_size": c.DefaultQuery("page_size", "10"),
	}
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}
	if accountID := c.Query("account_id"); accountID != "" {
		queryParams["account_id"] = accountID
	}
	if transactionType := c.Query("transaction_type"); transactionType != "" {
		queryParams["transaction_type"] = transactionType
	}
	result, statusCode, err := h.accountingClient.Get(c.Request.Context(), "/api/v1/transactions", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *AccountingBFFHandler) ReverseTransaction(c *gin.Context) {
	transactionNo := c.Param("transactionNo")
	var req map[string]interface{}
	c.ShouldBindJSON(&req)
	result, statusCode, err := h.accountingClient.Post(c.Request.Context(), "/api/v1/transactions/"+transactionNo+"/reverse", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

// ==================== 结算管理 ====================
func (h *AccountingBFFHandler) CreateSettlement(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, statusCode, err := h.accountingClient.Post(c.Request.Context(), "/api/v1/settlements", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *AccountingBFFHandler) GetSettlement(c *gin.Context) {
	settlementNo := c.Param("settlementNo")
	result, statusCode, err := h.accountingClient.Get(c.Request.Context(), "/api/v1/settlements/"+settlementNo, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *AccountingBFFHandler) ListSettlements(c *gin.Context) {
	queryParams := map[string]string{
		"page":      c.DefaultQuery("page", "1"),
		"page_size": c.DefaultQuery("page_size", "10"),
	}
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}
	result, statusCode, err := h.accountingClient.Get(c.Request.Context(), "/api/v1/settlements", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *AccountingBFFHandler) ProcessSettlement(c *gin.Context) {
	settlementNo := c.Param("settlementNo")
	var req map[string]interface{}
	c.ShouldBindJSON(&req)
	result, statusCode, err := h.accountingClient.Post(c.Request.Context(), "/api/v1/settlements/"+settlementNo+"/process", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

// ==================== 提现管理 ====================
func (h *AccountingBFFHandler) CreateWithdrawal(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, statusCode, err := h.accountingClient.Post(c.Request.Context(), "/api/v1/withdrawals", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *AccountingBFFHandler) GetWithdrawal(c *gin.Context) {
	withdrawalNo := c.Param("withdrawalNo")
	result, statusCode, err := h.accountingClient.Get(c.Request.Context(), "/api/v1/withdrawals/"+withdrawalNo, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *AccountingBFFHandler) ListWithdrawals(c *gin.Context) {
	queryParams := map[string]string{
		"page":      c.DefaultQuery("page", "1"),
		"page_size": c.DefaultQuery("page_size", "10"),
	}
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}
	result, statusCode, err := h.accountingClient.Get(c.Request.Context(), "/api/v1/withdrawals", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *AccountingBFFHandler) ApproveWithdrawal(c *gin.Context) {
	withdrawalNo := c.Param("withdrawalNo")
	var req map[string]interface{}
	c.ShouldBindJSON(&req)
	result, statusCode, err := h.accountingClient.Post(c.Request.Context(), "/api/v1/withdrawals/"+withdrawalNo+"/approve", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *AccountingBFFHandler) RejectWithdrawal(c *gin.Context) {
	withdrawalNo := c.Param("withdrawalNo")
	var req map[string]interface{}
	c.ShouldBindJSON(&req)
	result, statusCode, err := h.accountingClient.Post(c.Request.Context(), "/api/v1/withdrawals/"+withdrawalNo+"/reject", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *AccountingBFFHandler) ProcessWithdrawal(c *gin.Context) {
	withdrawalNo := c.Param("withdrawalNo")
	var req map[string]interface{}
	c.ShouldBindJSON(&req)
	result, statusCode, err := h.accountingClient.Post(c.Request.Context(), "/api/v1/withdrawals/"+withdrawalNo+"/process", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *AccountingBFFHandler) CompleteWithdrawal(c *gin.Context) {
	withdrawalNo := c.Param("withdrawalNo")
	var req map[string]interface{}
	c.ShouldBindJSON(&req)
	result, statusCode, err := h.accountingClient.Post(c.Request.Context(), "/api/v1/withdrawals/"+withdrawalNo+"/complete", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *AccountingBFFHandler) FailWithdrawal(c *gin.Context) {
	withdrawalNo := c.Param("withdrawalNo")
	var req map[string]interface{}
	c.ShouldBindJSON(&req)
	result, statusCode, err := h.accountingClient.Post(c.Request.Context(), "/api/v1/withdrawals/"+withdrawalNo+"/fail", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *AccountingBFFHandler) CancelWithdrawal(c *gin.Context) {
	withdrawalNo := c.Param("withdrawalNo")
	var req map[string]interface{}
	c.ShouldBindJSON(&req)
	result, statusCode, err := h.accountingClient.Post(c.Request.Context(), "/api/v1/withdrawals/"+withdrawalNo+"/cancel", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

// ==================== 发票管理 ====================
func (h *AccountingBFFHandler) CreateInvoice(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, statusCode, err := h.accountingClient.Post(c.Request.Context(), "/api/v1/invoices", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *AccountingBFFHandler) GetInvoice(c *gin.Context) {
	invoiceNo := c.Param("invoiceNo")
	result, statusCode, err := h.accountingClient.Get(c.Request.Context(), "/api/v1/invoices/"+invoiceNo, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *AccountingBFFHandler) ListInvoices(c *gin.Context) {
	queryParams := map[string]string{
		"page":      c.DefaultQuery("page", "1"),
		"page_size": c.DefaultQuery("page_size", "10"),
	}
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}
	result, statusCode, err := h.accountingClient.Get(c.Request.Context(), "/api/v1/invoices", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *AccountingBFFHandler) PayInvoice(c *gin.Context) {
	invoiceNo := c.Param("invoiceNo")
	var req map[string]interface{}
	c.ShouldBindJSON(&req)
	result, statusCode, err := h.accountingClient.Post(c.Request.Context(), "/api/v1/invoices/"+invoiceNo+"/pay", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *AccountingBFFHandler) CancelInvoice(c *gin.Context) {
	invoiceNo := c.Param("invoiceNo")
	var req map[string]interface{}
	c.ShouldBindJSON(&req)
	result, statusCode, err := h.accountingClient.Post(c.Request.Context(), "/api/v1/invoices/"+invoiceNo+"/cancel", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *AccountingBFFHandler) VoidInvoice(c *gin.Context) {
	invoiceNo := c.Param("invoiceNo")
	var req map[string]interface{}
	c.ShouldBindJSON(&req)
	result, statusCode, err := h.accountingClient.Post(c.Request.Context(), "/api/v1/invoices/"+invoiceNo+"/void", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

// ==================== 对账管理 ====================
func (h *AccountingBFFHandler) CreateReconciliation(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, statusCode, err := h.accountingClient.Post(c.Request.Context(), "/api/v1/reconciliations", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *AccountingBFFHandler) GetReconciliation(c *gin.Context) {
	reconciliationNo := c.Param("reconciliationNo")
	result, statusCode, err := h.accountingClient.Get(c.Request.Context(), "/api/v1/reconciliations/"+reconciliationNo, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *AccountingBFFHandler) ListReconciliations(c *gin.Context) {
	queryParams := map[string]string{
		"page":      c.DefaultQuery("page", "1"),
		"page_size": c.DefaultQuery("page_size", "10"),
	}
	result, statusCode, err := h.accountingClient.Get(c.Request.Context(), "/api/v1/reconciliations", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *AccountingBFFHandler) ProcessReconciliation(c *gin.Context) {
	reconciliationNo := c.Param("reconciliationNo")
	var req map[string]interface{}
	c.ShouldBindJSON(&req)
	result, statusCode, err := h.accountingClient.Post(c.Request.Context(), "/api/v1/reconciliations/"+reconciliationNo+"/process", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *AccountingBFFHandler) CompleteReconciliation(c *gin.Context) {
	reconciliationNo := c.Param("reconciliationNo")
	var req map[string]interface{}
	c.ShouldBindJSON(&req)
	result, statusCode, err := h.accountingClient.Post(c.Request.Context(), "/api/v1/reconciliations/"+reconciliationNo+"/complete", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *AccountingBFFHandler) ResolveReconciliationItem(c *gin.Context) {
	itemID := c.Param("itemID")
	var req map[string]interface{}
	c.ShouldBindJSON(&req)
	result, statusCode, err := h.accountingClient.Post(c.Request.Context(), "/api/v1/reconciliations/items/"+itemID+"/resolve", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

// ==================== 余额查询 ====================
func (h *AccountingBFFHandler) GetMerchantBalanceSummary(c *gin.Context) {
	merchantID := c.Param("merchantID")
	result, statusCode, err := h.accountingClient.Get(c.Request.Context(), "/api/v1/balances/merchants/"+merchantID+"/summary", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *AccountingBFFHandler) GetBalanceByCurrency(c *gin.Context) {
	merchantID := c.Param("merchantID")
	currency := c.Param("currency")
	result, statusCode, err := h.accountingClient.Get(c.Request.Context(), "/api/v1/balances/merchants/"+merchantID+"/currencies/"+currency, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *AccountingBFFHandler) GetBalanceByAccountType(c *gin.Context) {
	merchantID := c.Param("merchantID")
	accountType := c.Param("accountType")
	result, statusCode, err := h.accountingClient.Get(c.Request.Context(), "/api/v1/balances/merchants/"+merchantID+"/account-types/"+accountType, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *AccountingBFFHandler) GetAllCurrencyBalances(c *gin.Context) {
	merchantID := c.Param("merchantID")
	result, statusCode, err := h.accountingClient.Get(c.Request.Context(), "/api/v1/balances/merchants/"+merchantID+"/currencies", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

// ==================== 币种转换 ====================
func (h *AccountingBFFHandler) CreateCurrencyConversion(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, statusCode, err := h.accountingClient.Post(c.Request.Context(), "/api/v1/conversions", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *AccountingBFFHandler) GetCurrencyConversion(c *gin.Context) {
	conversionNo := c.Param("conversionNo")
	result, statusCode, err := h.accountingClient.Get(c.Request.Context(), "/api/v1/conversions/"+conversionNo, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *AccountingBFFHandler) ListCurrencyConversions(c *gin.Context) {
	queryParams := map[string]string{
		"page":      c.DefaultQuery("page", "1"),
		"page_size": c.DefaultQuery("page_size", "10"),
	}
	if merchantID := c.Query("merchant_id"); merchantID != "" {
		queryParams["merchant_id"] = merchantID
	}
	result, statusCode, err := h.accountingClient.Get(c.Request.Context(), "/api/v1/conversions", queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *AccountingBFFHandler) ProcessCurrencyConversion(c *gin.Context) {
	conversionNo := c.Param("conversionNo")
	var req map[string]interface{}
	c.ShouldBindJSON(&req)
	result, statusCode, err := h.accountingClient.Post(c.Request.Context(), "/api/v1/conversions/"+conversionNo+"/process", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}

func (h *AccountingBFFHandler) CancelCurrencyConversion(c *gin.Context) {
	conversionNo := c.Param("conversionNo")
	var req map[string]interface{}
	c.ShouldBindJSON(&req)
	result, statusCode, err := h.accountingClient.Post(c.Request.Context(), "/api/v1/conversions/"+conversionNo+"/cancel", req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(statusCode, result)
}
