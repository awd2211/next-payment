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
