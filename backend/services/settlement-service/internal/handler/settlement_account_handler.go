package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"payment-platform/settlement-service/internal/model"
	"payment-platform/settlement-service/internal/service"
)

// SettlementAccountHandler 结算账户处理器
type SettlementAccountHandler struct {
	service service.SettlementAccountService
}

// NewSettlementAccountHandler 创建结算账户处理器
func NewSettlementAccountHandler(service service.SettlementAccountService) *SettlementAccountHandler {
	return &SettlementAccountHandler{service: service}
}

// CreateAccount 创建结算账户
// @Summary		创建结算账户
// @Description	商户创建新的结算账户
// @Tags		Settlement Accounts
// @Accept		json
// @Produce		json
// @Security	BearerAuth
// @Param		request	body	CreateAccountRequest	true	"创建请求"
// @Success		201		{object}	SettlementAccountResponse
// @Failure		400		{object}	ErrorResponse
// @Failure		401		{object}	ErrorResponse
// @Router		/settlement-accounts [post]
func (h *SettlementAccountHandler) CreateAccount(c *gin.Context) {
	merchantIDStr, exists := c.Get("merchant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	merchantID, err := uuid.Parse(merchantIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid merchant_id"})
		return
	}

	var req CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	account := &model.SettlementAccount{
		MerchantID:    merchantID,
		AccountType:   req.AccountType,
		BankName:      req.BankName,
		BankCode:      req.BankCode,
		AccountNumber: req.AccountNumber,
		AccountName:   req.AccountName,
		SwiftCode:     req.SwiftCode,
		IBAN:          req.IBAN,
		BankAddress:   req.BankAddress,
		Currency:      req.Currency,
		Country:       req.Country,
	}

	if err := h.service.CreateAccount(c.Request.Context(), account); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, SettlementAccountResponse{
		ID:             account.ID.String(),
		MerchantID:     account.MerchantID.String(),
		AccountType:    account.AccountType,
		BankName:       account.BankName,
		AccountNumber:  maskAccountNumber(account.AccountNumber),
		AccountName:    account.AccountName,
		Currency:       account.Currency,
		Country:        account.Country,
		IsDefault:      account.IsDefault,
		Status:         account.Status,
		CreatedAt:      account.CreatedAt,
	})
}

// GetAccount 获取结算账户详情
// @Summary		获取结算账户详情
// @Description	获取指定结算账户的详细信息
// @Tags		Settlement Accounts
// @Produce		json
// @Security	BearerAuth
// @Param		id	path	string	true	"账户ID"
// @Success		200	{object}	SettlementAccountResponse
// @Failure		404	{object}	ErrorResponse
// @Router		/settlement-accounts/{id} [get]
func (h *SettlementAccountHandler) GetAccount(c *gin.Context) {
	accountID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid account id"})
		return
	}

	account, err := h.service.GetAccount(c.Request.Context(), accountID)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "account not found"})
		return
	}

	c.JSON(http.StatusOK, SettlementAccountResponse{
		ID:             account.ID.String(),
		MerchantID:     account.MerchantID.String(),
		AccountType:    account.AccountType,
		BankName:       account.BankName,
		AccountNumber:  maskAccountNumber(account.AccountNumber),
		AccountName:    account.AccountName,
		SwiftCode:      account.SwiftCode,
		IBAN:           account.IBAN,
		Currency:       account.Currency,
		Country:        account.Country,
		IsDefault:      account.IsDefault,
		Status:         account.Status,
		VerifiedAt:     account.VerifiedAt,
		CreatedAt:      account.CreatedAt,
	})
}

// ListAccounts 获取商户的结算账户列表
// @Summary		获取结算账户列表
// @Description	获取当前商户的所有结算账户
// @Tags		Settlement Accounts
// @Produce		json
// @Security	BearerAuth
// @Success		200	{array}		SettlementAccountResponse
// @Failure		401	{object}	ErrorResponse
// @Router		/settlement-accounts [get]
func (h *SettlementAccountHandler) ListAccounts(c *gin.Context) {
	merchantIDStr, exists := c.Get("merchant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	merchantID, err := uuid.Parse(merchantIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid merchant_id"})
		return
	}

	accounts, err := h.service.GetMerchantAccounts(c.Request.Context(), merchantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	var response []SettlementAccountResponse
	for _, account := range accounts {
		response = append(response, SettlementAccountResponse{
			ID:             account.ID.String(),
			MerchantID:     account.MerchantID.String(),
			AccountType:    account.AccountType,
			BankName:       account.BankName,
			AccountNumber:  maskAccountNumber(account.AccountNumber),
			AccountName:    account.AccountName,
			Currency:       account.Currency,
			Country:        account.Country,
			IsDefault:      account.IsDefault,
			Status:         account.Status,
			VerifiedAt:     account.VerifiedAt,
			CreatedAt:      account.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, response)
}

// SetDefaultAccount 设置默认结算账户
// @Summary		设置默认结算账户
// @Description	设置指定账户为默认结算账户
// @Tags		Settlement Accounts
// @Produce		json
// @Security	BearerAuth
// @Param		id	path	string	true	"账户ID"
// @Success		200	{object}	SuccessResponse
// @Failure		400	{object}	ErrorResponse
// @Router		/settlement-accounts/{id}/set-default [post]
func (h *SettlementAccountHandler) SetDefaultAccount(c *gin.Context) {
	merchantIDStr, exists := c.Get("merchant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	merchantID, err := uuid.Parse(merchantIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid merchant_id"})
		return
	}

	accountID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid account id"})
		return
	}

	if err := h.service.SetDefaultAccount(c.Request.Context(), merchantID, accountID); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Message: "Default account set successfully"})
}

// DeleteAccount 删除结算账户
// @Summary		删除结算账户
// @Description	删除指定的结算账户
// @Tags		Settlement Accounts
// @Produce		json
// @Security	BearerAuth
// @Param		id	path	string	true	"账户ID"
// @Success		200	{object}	SuccessResponse
// @Failure		400	{object}	ErrorResponse
// @Router		/settlement-accounts/{id} [delete]
func (h *SettlementAccountHandler) DeleteAccount(c *gin.Context) {
	accountID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid account id"})
		return
	}

	if err := h.service.DeleteAccount(c.Request.Context(), accountID); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Message: "Account deleted successfully"})
}

// RegisterSettlementAccountRoutes 注册结算账户路由
func RegisterSettlementAccountRoutes(r *gin.RouterGroup, handler *SettlementAccountHandler, authMiddleware gin.HandlerFunc) {
	accounts := r.Group("/settlement-accounts")
	accounts.Use(authMiddleware)
	{
		accounts.POST("", handler.CreateAccount)
		accounts.GET("", handler.ListAccounts)
		accounts.GET("/:id", handler.GetAccount)
		accounts.POST("/:id/set-default", handler.SetDefaultAccount)
		accounts.DELETE("/:id", handler.DeleteAccount)
	}
}

// maskAccountNumber 遮蔽账号敏感信息
func maskAccountNumber(accountNumber string) string {
	if len(accountNumber) <= 8 {
		return "****"
	}
	return accountNumber[:4] + "****" + accountNumber[len(accountNumber)-4:]
}

// DTO定义

type CreateAccountRequest struct {
	AccountType   string `json:"account_type" binding:"required"`
	BankName      string `json:"bank_name"`
	BankCode      string `json:"bank_code"`
	AccountNumber string `json:"account_number" binding:"required"`
	AccountName   string `json:"account_name" binding:"required"`
	SwiftCode     string `json:"swift_code"`
	IBAN          string `json:"iban"`
	BankAddress   string `json:"bank_address"`
	Currency      string `json:"currency" binding:"required"`
	Country       string `json:"country"`
}

type SettlementAccountResponse struct {
	ID             string     `json:"id"`
	MerchantID     string     `json:"merchant_id"`
	AccountType    string     `json:"account_type"`
	BankName       string     `json:"bank_name"`
	AccountNumber  string     `json:"account_number"` // 已遮蔽
	AccountName    string     `json:"account_name"`
	SwiftCode      string     `json:"swift_code"`
	IBAN           string     `json:"iban"`
	Currency       string     `json:"currency"`
	Country        string     `json:"country"`
	IsDefault      bool       `json:"is_default"`
	Status         string     `json:"status"`
	VerifiedAt     *time.Time `json:"verified_at"`
	CreatedAt      time.Time  `json:"created_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}
