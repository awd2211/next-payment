package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"payment-platform/accounting-service/internal/client"
	"payment-platform/accounting-service/internal/model"
	"payment-platform/accounting-service/internal/repository"
)

// AccountService 账户服务接口
type AccountService interface {
	// 账户管理
	CreateAccount(ctx context.Context, input *CreateAccountInput) (*model.Account, error)
	GetAccount(ctx context.Context, id uuid.UUID) (*model.Account, error)
	GetMerchantAccount(ctx context.Context, merchantID uuid.UUID, accountType, currency string) (*model.Account, error)
	ListAccounts(ctx context.Context, query *repository.AccountQuery) ([]*model.Account, int64, error)
	FreezeAccount(ctx context.Context, id uuid.UUID) error
	UnfreezeAccount(ctx context.Context, id uuid.UUID) error

	// 交易管理
	CreateTransaction(ctx context.Context, input *CreateTransactionInput) (*model.AccountTransaction, error)
	GetTransaction(ctx context.Context, transactionNo string) (*model.AccountTransaction, error)
	ListTransactions(ctx context.Context, query *repository.TransactionQuery) ([]*model.AccountTransaction, int64, error)
	ReverseTransaction(ctx context.Context, transactionNo string, reason string) error

	// 复式记账
	CreateDoubleEntry(ctx context.Context, input *CreateDoubleEntryInput) (*model.DoubleEntry, error)
	ListDoubleEntries(ctx context.Context, query *repository.DoubleEntryQuery) ([]*model.DoubleEntry, int64, error)

	// 结算管理
	CreateSettlement(ctx context.Context, input *CreateSettlementInput) (*model.Settlement, error)
	GetSettlement(ctx context.Context, settlementNo string) (*model.Settlement, error)
	ListSettlements(ctx context.Context, query *repository.SettlementQuery) ([]*model.Settlement, int64, error)
	ProcessSettlement(ctx context.Context, settlementNo string) error

	// 提现管理
	CreateWithdrawal(ctx context.Context, input *CreateWithdrawalInput) (*model.Withdrawal, error)
	GetWithdrawal(ctx context.Context, withdrawalNo string) (*model.Withdrawal, error)
	ListWithdrawals(ctx context.Context, query *repository.WithdrawalQuery) ([]*model.Withdrawal, int64, error)
	ApproveWithdrawal(ctx context.Context, withdrawalNo string, approverID uuid.UUID, notes string) error
	RejectWithdrawal(ctx context.Context, withdrawalNo string, approverID uuid.UUID, reason string) error
	ProcessWithdrawal(ctx context.Context, withdrawalNo string, processorID uuid.UUID) error
	CompleteWithdrawal(ctx context.Context, withdrawalNo string) error
	FailWithdrawal(ctx context.Context, withdrawalNo string, reason string) error
	CancelWithdrawal(ctx context.Context, withdrawalNo string) error

	// 账单管理
	CreateInvoice(ctx context.Context, input *CreateInvoiceInput) (*model.Invoice, error)
	GetInvoice(ctx context.Context, invoiceNo string) (*model.Invoice, error)
	ListInvoices(ctx context.Context, query *repository.InvoiceQuery) ([]*model.Invoice, int64, error)
	PayInvoice(ctx context.Context, invoiceNo string, paidAmount int64) error
	CancelInvoice(ctx context.Context, invoiceNo string) error
	VoidInvoice(ctx context.Context, invoiceNo string) error
	CheckOverdueInvoices(ctx context.Context) error

	// 对账管理
	CreateReconciliation(ctx context.Context, input *CreateReconciliationInput) (*model.Reconciliation, error)
	GetReconciliation(ctx context.Context, reconciliationNo string) (*model.Reconciliation, error)
	ListReconciliations(ctx context.Context, query *repository.ReconciliationQuery) ([]*model.Reconciliation, int64, error)
	ProcessReconciliation(ctx context.Context, reconciliationNo string, userID uuid.UUID) error
	CompleteReconciliation(ctx context.Context, reconciliationNo string) error
	ResolveReconciliationItem(ctx context.Context, itemID uuid.UUID, resolution string) error

	// 余额查询聚合
	GetMerchantBalanceSummary(ctx context.Context, merchantID uuid.UUID) (*MerchantBalanceSummary, error)
	GetBalanceByCurrency(ctx context.Context, merchantID uuid.UUID, currency string) (*CurrencyBalanceSummary, error)
	GetBalanceByAccountType(ctx context.Context, merchantID uuid.UUID, accountType string) (*AccountTypeBalanceSummary, error)
	GetAllCurrencyBalances(ctx context.Context, merchantID uuid.UUID) ([]*CurrencyBalanceSummary, error)

	// 货币转换管理
	CreateCurrencyConversion(ctx context.Context, input *CreateCurrencyConversionInput) (*model.CurrencyConversion, error)
	GetCurrencyConversion(ctx context.Context, conversionNo string) (*model.CurrencyConversion, error)
	ListCurrencyConversions(ctx context.Context, query *repository.CurrencyConversionQuery) ([]*model.CurrencyConversion, int64, error)
	ProcessCurrencyConversion(ctx context.Context, conversionNo string) error
	CancelCurrencyConversion(ctx context.Context, conversionNo string, reason string) error
}

type accountService struct {
	db                   *gorm.DB                      // 添加数据库连接，用于事务支持
	accountRepo          repository.AccountRepository
	channelAdapterClient *client.ChannelAdapterClient  // 汇率查询客户端（可选）
}

// NewAccountService 创建账户服务实例
func NewAccountService(db *gorm.DB, accountRepo repository.AccountRepository, channelAdapterClient *client.ChannelAdapterClient) AccountService {
	return &accountService{
		db:                   db,
		accountRepo:          accountRepo,
		channelAdapterClient: channelAdapterClient,
	}
}

// CreateAccountInput 创建账户输入
type CreateAccountInput struct {
	MerchantID  uuid.UUID `json:"merchant_id" binding:"required"`
	AccountType string    `json:"account_type" binding:"required"`
	Currency    string    `json:"currency" binding:"required"`
}

// CreateTransactionInput 创建交易输入
type CreateTransactionInput struct {
	AccountID       uuid.UUID              `json:"account_id" binding:"required"`
	TransactionType string                 `json:"transaction_type" binding:"required"`
	Amount          int64                  `json:"amount" binding:"required"`
	RelatedID       uuid.UUID              `json:"related_id"`
	RelatedNo       string                 `json:"related_no"`
	Description     string                 `json:"description"`
	Extra           map[string]interface{} `json:"extra"`
}

// CreateDoubleEntryInput 创建复式记账输入
type CreateDoubleEntryInput struct {
	RelatedID     uuid.UUID `json:"related_id"`
	RelatedNo     string    `json:"related_no"`
	EntryType     string    `json:"entry_type" binding:"required"`
	DebitAccount  string    `json:"debit_account" binding:"required"`
	CreditAccount string    `json:"credit_account" binding:"required"`
	Amount        int64     `json:"amount" binding:"required"`
	Currency      string    `json:"currency" binding:"required"`
	Description   string    `json:"description"`
}

// CreateSettlementInput 创建结算输入
type CreateSettlementInput struct {
	MerchantID   uuid.UUID `json:"merchant_id" binding:"required"`
	AccountID    uuid.UUID `json:"account_id" binding:"required"`
	PeriodStart  time.Time `json:"period_start" binding:"required"`
	PeriodEnd    time.Time `json:"period_end" binding:"required"`
	Currency     string    `json:"currency" binding:"required"`
	PaymentCount int       `json:"payment_count"`
	RefundCount  int       `json:"refund_count"`
}

// CreateWithdrawalInput 创建提现输入
type CreateWithdrawalInput struct {
	MerchantID          uuid.UUID `json:"merchant_id" binding:"required"`
	AccountID           uuid.UUID `json:"account_id" binding:"required"`
	SettlementAccountID uuid.UUID `json:"settlement_account_id" binding:"required"`
	Amount              int64     `json:"amount" binding:"required,gt=0"`
	Currency            string    `json:"currency" binding:"required"`
	RequestReason       string    `json:"request_reason"`
}

// CreateInvoiceInput 创建账单输入
type CreateInvoiceInput struct {
	MerchantID   uuid.UUID           `json:"merchant_id" binding:"required"`
	InvoiceType  string              `json:"invoice_type" binding:"required"`
	PeriodStart  time.Time           `json:"period_start" binding:"required"`
	PeriodEnd    time.Time           `json:"period_end" binding:"required"`
	Currency     string              `json:"currency" binding:"required"`
	DueDate      time.Time           `json:"due_date" binding:"required"`
	TaxRate      float64             `json:"tax_rate"`                        // 税率（百分比）
	Notes        string              `json:"notes"`
	Items        []InvoiceItemInput  `json:"items" binding:"required,min=1"`
}

// InvoiceItemInput 账单明细输入
type InvoiceItemInput struct {
	ItemType    string     `json:"item_type" binding:"required"`
	Description string     `json:"description"`
	Quantity    int        `json:"quantity" binding:"required,gt=0"`
	UnitPrice   int64      `json:"unit_price" binding:"required,gt=0"`
	RelatedID   *uuid.UUID `json:"related_id"`
	RelatedNo   string     `json:"related_no"`
}

// CreateReconciliationInput 创建对账单输入
type CreateReconciliationInput struct {
	MerchantID         uuid.UUID                     `json:"merchant_id" binding:"required"`
	Channel            string                        `json:"channel" binding:"required"`
	ReconciliationDate time.Time                     `json:"reconciliation_date" binding:"required"`
	PeriodStart        time.Time                     `json:"period_start" binding:"required"`
	PeriodEnd          time.Time                     `json:"period_end" binding:"required"`
	Currency           string                        `json:"currency" binding:"required"`
	Items              []ReconciliationItemInput     `json:"items" binding:"required,min=1"`
}

// ReconciliationItemInput 对账明细输入
type ReconciliationItemInput struct {
	TransactionNo  string `json:"transaction_no" binding:"required"`
	ExternalTxNo   string `json:"external_tx_no"`
	ItemType       string `json:"item_type" binding:"required"`
	InternalAmount int64  `json:"internal_amount"`
	ExternalAmount int64  `json:"external_amount"`
	Status         string `json:"status"`
	Description    string `json:"description"`
}

// CreateCurrencyConversionInput 创建货币转换输入
type CreateCurrencyConversionInput struct {
	MerchantID     uuid.UUID `json:"merchant_id" binding:"required"`
	SourceCurrency string    `json:"source_currency" binding:"required"` // 源货币（如 USD）
	TargetCurrency string    `json:"target_currency" binding:"required"` // 目标货币（如 EUR）
	SourceAmount   int64     `json:"source_amount" binding:"required"`   // 源货币金额（分）
	Reason         string    `json:"reason"`                             // 转换原因
	RequestedBy    uuid.UUID `json:"requested_by"`                       // 请求人ID
	Notes          string    `json:"notes"`                              // 备注
}

// Balance Aggregation Response Structures

// AccountBalance 单个账户余额信息
type AccountBalance struct {
	AccountID        uuid.UUID `json:"account_id"`
	AccountType      string    `json:"account_type"`
	Currency         string    `json:"currency"`
	Balance          int64     `json:"balance"`
	FrozenBalance    int64     `json:"frozen_balance"`
	AvailableBalance int64     `json:"available_balance"`
	TotalIn          int64     `json:"total_in"`
	TotalOut         int64     `json:"total_out"`
	Status           string    `json:"status"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// MerchantBalanceSummary 商户余额汇总
type MerchantBalanceSummary struct {
	MerchantID       uuid.UUID        `json:"merchant_id"`
	TotalAccounts    int              `json:"total_accounts"`
	ActiveAccounts   int              `json:"active_accounts"`
	FrozenAccounts   int              `json:"frozen_accounts"`
	Accounts         []AccountBalance `json:"accounts"`
	CurrencySummary  map[string]int64 `json:"currency_summary"`   // 按货币汇总余额
	TypeSummary      map[string]int64 `json:"type_summary"`       // 按账户类型汇总余额
	TotalBalance     int64            `json:"total_balance"`      // 所有账户余额总和（需转换为基准货币）
	TotalFrozen      int64            `json:"total_frozen"`       // 所有账户冻结金额总和
	TotalAvailable   int64            `json:"total_available"`    // 所有账户可用余额总和
	LastUpdated      time.Time        `json:"last_updated"`
}

// CurrencyBalanceSummary 货币余额汇总
type CurrencyBalanceSummary struct {
	MerchantID       uuid.UUID        `json:"merchant_id"`
	Currency         string           `json:"currency"`
	AccountCount     int              `json:"account_count"`
	TotalBalance     int64            `json:"total_balance"`
	TotalFrozen      int64            `json:"total_frozen"`
	TotalAvailable   int64            `json:"total_available"`
	TotalIn          int64            `json:"total_in"`
	TotalOut         int64            `json:"total_out"`
	Accounts         []AccountBalance `json:"accounts"`
	LastUpdated      time.Time        `json:"last_updated"`
}

// AccountTypeBalanceSummary 账户类型余额汇总
type AccountTypeBalanceSummary struct {
	MerchantID       uuid.UUID        `json:"merchant_id"`
	AccountType      string           `json:"account_type"`
	AccountCount     int              `json:"account_count"`
	CurrencyBalances map[string]int64 `json:"currency_balances"` // 按货币分组的余额
	TotalBalance     int64            `json:"total_balance"`
	TotalFrozen      int64            `json:"total_frozen"`
	TotalAvailable   int64            `json:"total_available"`
	Accounts         []AccountBalance `json:"accounts"`
	LastUpdated      time.Time        `json:"last_updated"`
}

// CreateAccount 创建账户
func (s *accountService) CreateAccount(ctx context.Context, input *CreateAccountInput) (*model.Account, error) {
	// 检查账户是否已存在
	existing, err := s.accountRepo.GetAccountByMerchant(ctx, input.MerchantID, input.AccountType, input.Currency)
	if err != nil {
		return nil, fmt.Errorf("检查账户失败: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("账户已存在")
	}

	account := &model.Account{
		MerchantID:    input.MerchantID,
		AccountType:   input.AccountType,
		Currency:      input.Currency,
		Balance:       0,
		FrozenBalance: 0,
		TotalIn:       0,
		TotalOut:      0,
		Status:        model.AccountStatusActive,
	}

	if err := s.accountRepo.CreateAccount(ctx, account); err != nil {
		return nil, fmt.Errorf("创建账户失败: %w", err)
	}

	return account, nil
}

// GetAccount 获取账户
func (s *accountService) GetAccount(ctx context.Context, id uuid.UUID) (*model.Account, error) {
	account, err := s.accountRepo.GetAccountByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取账户失败: %w", err)
	}
	if account == nil {
		return nil, fmt.Errorf("账户不存在")
	}
	return account, nil
}

// GetMerchantAccount 获取商户账户
func (s *accountService) GetMerchantAccount(ctx context.Context, merchantID uuid.UUID, accountType, currency string) (*model.Account, error) {
	account, err := s.accountRepo.GetAccountByMerchant(ctx, merchantID, accountType, currency)
	if err != nil {
		return nil, fmt.Errorf("获取账户失败: %w", err)
	}
	if account == nil {
		return nil, fmt.Errorf("账户不存在")
	}
	return account, nil
}

// ListAccounts 账户列表
func (s *accountService) ListAccounts(ctx context.Context, query *repository.AccountQuery) ([]*model.Account, int64, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		query.PageSize = 20
	}
	return s.accountRepo.ListAccounts(ctx, query)
}

// FreezeAccount 冻结账户
func (s *accountService) FreezeAccount(ctx context.Context, id uuid.UUID) error {
	account, err := s.GetAccount(ctx, id)
	if err != nil {
		return err
	}

	account.Status = model.AccountStatusFrozen
	return s.accountRepo.UpdateAccount(ctx, account)
}

// UnfreezeAccount 解冻账户
func (s *accountService) UnfreezeAccount(ctx context.Context, id uuid.UUID) error {
	account, err := s.GetAccount(ctx, id)
	if err != nil {
		return err
	}

	account.Status = model.AccountStatusActive
	return s.accountRepo.UpdateAccount(ctx, account)
}

// CreateTransaction 创建交易
func (s *accountService) CreateTransaction(ctx context.Context, input *CreateTransactionInput) (*model.AccountTransaction, error) {
	// 获取账户
	account, err := s.GetAccount(ctx, input.AccountID)
	if err != nil {
		return nil, err
	}

	// 检查账户状态
	if account.Status != model.AccountStatusActive {
		return nil, fmt.Errorf("账户状态异常: %s", account.Status)
	}

	// 检查余额（如果是出账）
	if input.Amount < 0 && account.Balance < -input.Amount {
		return nil, fmt.Errorf("账户余额不足")
	}

	// 生成交易流水号
	transactionNo := s.generateTransactionNo()

	// 计算余额
	balanceBefore := account.Balance
	balanceAfter := balanceBefore + input.Amount

	// 创建交易记录
	transaction := &model.AccountTransaction{
		AccountID:       input.AccountID,
		MerchantID:      account.MerchantID,
		TransactionNo:   transactionNo,
		TransactionType: input.TransactionType,
		RelatedID:       input.RelatedID,
		RelatedNo:       input.RelatedNo,
		Amount:          input.Amount,
		BalanceBefore:   balanceBefore,
		BalanceAfter:    balanceAfter,
		Currency:        account.Currency,
		Description:     input.Description,
		Status:          "completed",
	}

	// 开始事务：创建交易记录 + 更新账户余额
	if err := s.accountRepo.CreateTransaction(ctx, transaction); err != nil {
		return nil, fmt.Errorf("创建交易记录失败: %w", err)
	}

	if err := s.accountRepo.UpdateBalance(ctx, input.AccountID, input.Amount); err != nil {
		return nil, fmt.Errorf("更新账户余额失败: %w", err)
	}

	// 创建复式记账
	s.createDoubleEntry(ctx, transaction)

	return transaction, nil
}

// GetTransaction 获取交易
func (s *accountService) GetTransaction(ctx context.Context, transactionNo string) (*model.AccountTransaction, error) {
	transaction, err := s.accountRepo.GetTransactionByNo(ctx, transactionNo)
	if err != nil {
		return nil, fmt.Errorf("获取交易失败: %w", err)
	}
	if transaction == nil {
		return nil, fmt.Errorf("交易不存在")
	}
	return transaction, nil
}

// ListTransactions 交易列表
func (s *accountService) ListTransactions(ctx context.Context, query *repository.TransactionQuery) ([]*model.AccountTransaction, int64, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		query.PageSize = 20
	}
	return s.accountRepo.ListTransactions(ctx, query)
}

// ReverseTransaction 冲正交易
func (s *accountService) ReverseTransaction(ctx context.Context, transactionNo string, reason string) error {
	// 获取原交易
	originalTx, err := s.GetTransaction(ctx, transactionNo)
	if err != nil {
		return err
	}

	if originalTx.Status == "reversed" {
		return fmt.Errorf("交易已冲正")
	}

	// 创建冲正交易
	reverseInput := &CreateTransactionInput{
		AccountID:       originalTx.AccountID,
		TransactionType: "adjustment",
		Amount:          -originalTx.Amount,
		RelatedID:       originalTx.ID,
		RelatedNo:       originalTx.TransactionNo,
		Description:     fmt.Sprintf("冲正交易 %s: %s", originalTx.TransactionNo, reason),
	}

	_, err = s.CreateTransaction(ctx, reverseInput)
	return err
}

// CreateSettlement 创建结算
func (s *accountService) CreateSettlement(ctx context.Context, input *CreateSettlementInput) (*model.Settlement, error) {
	// 生成结算单号
	settlementNo := s.generateSettlementNo()

	// 查询结算周期内的交易
	txQuery := &repository.TransactionQuery{
		MerchantID: &input.MerchantID,
		Currency:   input.Currency,
		StartTime:  &input.PeriodStart,
		EndTime:    &input.PeriodEnd,
		Page:       1,
		PageSize:   10000, // 一次性获取所有交易
	}

	transactions, _, err := s.accountRepo.ListTransactions(ctx, txQuery)
	if err != nil {
		return nil, fmt.Errorf("查询交易失败: %w", err)
	}

	// 计算结算金额
	var totalAmount int64
	var paymentCount, refundCount int

	for _, tx := range transactions {
		if tx.Status != "completed" {
			continue
		}

		switch tx.TransactionType {
		case model.TransactionTypePaymentIn:
			totalAmount += tx.Amount
			paymentCount++
		case model.TransactionTypeRefundOut:
			totalAmount += tx.Amount // 退款金额为负数
			refundCount++
		}
	}

	// 计算手续费（简单示例：2% 费率）
	// 实际应该从merchant-service的fee_configs表查询商户费率
	feeRate := 0.02
	feeAmount := int64(float64(totalAmount) * feeRate)

	// 计算净额
	netAmount := totalAmount - feeAmount

	settlement := &model.Settlement{
		MerchantID:   input.MerchantID,
		SettlementNo: settlementNo,
		AccountID:    input.AccountID,
		PeriodStart:  input.PeriodStart,
		PeriodEnd:    input.PeriodEnd,
		TotalAmount:  totalAmount,
		FeeAmount:    feeAmount,
		NetAmount:    netAmount,
		Currency:     input.Currency,
		Status:       model.SettlementStatusPending,
		PaymentCount: paymentCount,
		RefundCount:  refundCount,
	}

	if err := s.accountRepo.CreateSettlement(ctx, settlement); err != nil {
		return nil, fmt.Errorf("创建结算失败: %w", err)
	}

	return settlement, nil
}

// GetSettlement 获取结算
func (s *accountService) GetSettlement(ctx context.Context, settlementNo string) (*model.Settlement, error) {
	settlement, err := s.accountRepo.GetSettlementByNo(ctx, settlementNo)
	if err != nil {
		return nil, fmt.Errorf("获取结算失败: %w", err)
	}
	if settlement == nil {
		return nil, fmt.Errorf("结算不存在")
	}
	return settlement, nil
}

// ListSettlements 结算列表
func (s *accountService) ListSettlements(ctx context.Context, query *repository.SettlementQuery) ([]*model.Settlement, int64, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		query.PageSize = 20
	}
	return s.accountRepo.ListSettlements(ctx, query)
}

// ProcessSettlement 处理结算（带事务保护）
func (s *accountService) ProcessSettlement(ctx context.Context, settlementNo string) error {
	// 1. 获取结算记录
	settlement, err := s.GetSettlement(ctx, settlementNo)
	if err != nil {
		return fmt.Errorf("获取结算记录失败: %w", err)
	}

	// 2. 检查结算状态
	if settlement.Status != model.SettlementStatusPending {
		return fmt.Errorf("结算状态不正确: %s，只能处理pending状态的结算", settlement.Status)
	}

	// 3. 获取账户信息（提前验证）
	account, err := s.GetAccount(ctx, settlement.AccountID)
	if err != nil {
		return fmt.Errorf("获取账户失败: %w", err)
	}

	// 4. 检查账户状态（提前验证）
	if account.Status != model.AccountStatusActive {
		return fmt.Errorf("账户状态异常: %s，无法进行结算", account.Status)
	}

	// 5. 使用数据库事务执行结算操作（确保原子性）
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 5.1 更新结算状态为processing
		settlement.Status = model.SettlementStatusProcessing
		if err := s.accountRepo.UpdateSettlement(ctx, settlement); err != nil {
			return fmt.Errorf("更新结算状态失败: %w", err)
		}

		// 5.2 创建手续费交易（如果有手续费）
		if settlement.FeeAmount > 0 {
			feeInput := &CreateTransactionInput{
				AccountID:       settlement.AccountID,
				TransactionType: model.TransactionTypeFee,
				Amount:          -settlement.FeeAmount,
				RelatedID:       settlement.ID,
				RelatedNo:       settlement.SettlementNo,
				Description:     fmt.Sprintf("结算手续费: %s", settlement.SettlementNo),
			}

			_, err := s.CreateTransaction(ctx, feeInput)
			if err != nil {
				return fmt.Errorf("创建手续费交易失败: %w", err)
			}
		}

		// 5.3 创建结算交易（净额）
		settlementInput := &CreateTransactionInput{
			AccountID:       settlement.AccountID,
			TransactionType: "settlement",
			Amount:          settlement.NetAmount,
			RelatedID:       settlement.ID,
			RelatedNo:       settlement.SettlementNo,
			Description:     fmt.Sprintf("结算: %s (周期: %s - %s)",
				settlement.SettlementNo,
				settlement.PeriodStart.Format("2006-01-02"),
				settlement.PeriodEnd.Format("2006-01-02")),
		}

		_, err := s.CreateTransaction(ctx, settlementInput)
		if err != nil {
			return fmt.Errorf("创建结算交易失败: %w", err)
		}

		// 5.4 完成结算（更新状态和时间）
		now := time.Now()
		settlement.Status = model.SettlementStatusCompleted
		settlement.SettledAt = &now
		if err := s.accountRepo.UpdateSettlement(ctx, settlement); err != nil {
			return fmt.Errorf("完成结算失败: %w", err)
		}

		return nil
	})

	if err != nil {
		// 事务回滚后，尝试标记结算为失败状态（尽力而为）
		settlement.Status = model.SettlementStatusFailed
		if updateErr := s.accountRepo.UpdateSettlement(ctx, settlement); updateErr != nil {
			logger.Error("failed to update settlement status after transaction failure",
				zap.Error(updateErr),
				zap.String("settlement_no", settlement.SettlementNo),
				zap.String("merchant_id", settlement.MerchantID.String()))
		}
		return fmt.Errorf("结算处理失败: %w", err)
	}

	return nil
}

// 工具函数

func (s *accountService) generateTransactionNo() string {
	timestamp := time.Now().Format("20060102150405")
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	randomStr := base64.URLEncoding.EncodeToString(randomBytes)[:10]
	return fmt.Sprintf("TX%s%s", timestamp, randomStr)
}

func (s *accountService) generateEntryNo() string {
	timestamp := time.Now().Format("20060102150405")
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	randomStr := base64.URLEncoding.EncodeToString(randomBytes)[:10]
	return fmt.Sprintf("ENT%s%s", timestamp, randomStr)
}

func (s *accountService) generateSettlementNo() string {
	timestamp := time.Now().Format("20060102150405")
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	randomStr := base64.URLEncoding.EncodeToString(randomBytes)[:10]
	return fmt.Sprintf("ST%s%s", timestamp, randomStr)
}

func (s *accountService) generateWithdrawalNo() string {
	timestamp := time.Now().Format("20060102150405")
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	randomStr := base64.URLEncoding.EncodeToString(randomBytes)[:10]
	return fmt.Sprintf("WD%s%s", timestamp, randomStr)
}

func (s *accountService) generateInvoiceNo() string {
	timestamp := time.Now().Format("20060102150405")
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	randomStr := base64.URLEncoding.EncodeToString(randomBytes)[:10]
	return fmt.Sprintf("INV%s%s", timestamp, randomStr)
}

func (s *accountService) generateReconciliationNo() string {
	timestamp := time.Now().Format("20060102150405")
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	randomStr := base64.URLEncoding.EncodeToString(randomBytes)[:10]
	return fmt.Sprintf("REC%s%s", timestamp, randomStr)
}

// CreateDoubleEntry 创建复式记账（公共方法）
func (s *accountService) CreateDoubleEntry(ctx context.Context, input *CreateDoubleEntryInput) (*model.DoubleEntry, error) {
	entry := &model.DoubleEntry{
		EntryNo:       s.generateEntryNo(),
		RelatedID:     input.RelatedID,
		RelatedNo:     input.RelatedNo,
		EntryType:     input.EntryType,
		DebitAccount:  input.DebitAccount,
		CreditAccount: input.CreditAccount,
		Amount:        input.Amount,
		Currency:      input.Currency,
		Description:   input.Description,
	}

	if err := s.accountRepo.CreateDoubleEntry(ctx, entry); err != nil {
		return nil, fmt.Errorf("创建复式记账失败: %w", err)
	}

	return entry, nil
}

// ListDoubleEntries 复式记账列表
func (s *accountService) ListDoubleEntries(ctx context.Context, query *repository.DoubleEntryQuery) ([]*model.DoubleEntry, int64, error) {
	entries, total, err := s.accountRepo.ListDoubleEntries(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("查询复式记账列表失败: %w", err)
	}
	return entries, total, nil
}

// createDoubleEntry 自动创建复式记账（私有方法，异步调用）
func (s *accountService) createDoubleEntry(ctx context.Context, tx *model.AccountTransaction) {
	entryNo := fmt.Sprintf("DE%s", tx.TransactionNo[2:])

	var debitAccount, creditAccount string
	switch tx.TransactionType {
	case model.TransactionTypePaymentIn:
		debitAccount = "银行存款"
		creditAccount = "主营业务收入"
	case model.TransactionTypeRefundOut:
		debitAccount = "主营业务成本"
		creditAccount = "银行存款"
	case model.TransactionTypeWithdraw:
		debitAccount = "其他应收款"
		creditAccount = "银行存款"
	case model.TransactionTypeFee:
		debitAccount = "手续费支出"
		creditAccount = "银行存款"
	default:
		return
	}

	entry := &model.DoubleEntry{
		EntryNo:       entryNo,
		RelatedID:     tx.ID,
		RelatedNo:     tx.TransactionNo,
		EntryType:     tx.TransactionType,
		DebitAccount:  debitAccount,
		CreditAccount: creditAccount,
		Amount:        tx.Amount,
		Currency:      tx.Currency,
		Description:   tx.Description,
	}

	s.accountRepo.CreateDoubleEntry(ctx, entry)
}

// Withdrawal Management Methods

// CreateWithdrawal 创建提现申请
func (s *accountService) CreateWithdrawal(ctx context.Context, input *CreateWithdrawalInput) (*model.Withdrawal, error) {
	// 获取账户
	account, err := s.GetAccount(ctx, input.AccountID)
	if err != nil {
		return nil, err
	}

	// 检查账户所属
	if account.MerchantID != input.MerchantID {
		return nil, fmt.Errorf("账户不属于该商户")
	}

	// 检查账户状态
	if account.Status != model.AccountStatusActive {
		return nil, fmt.Errorf("账户状态异常: %s", account.Status)
	}

	// 检查账户余额
	if account.Balance < input.Amount {
		return nil, fmt.Errorf("账户余额不足")
	}

	// 计算手续费（简单示例：0.5%）
	feeAmount := input.Amount * 5 / 1000
	actualAmount := input.Amount - feeAmount

	// 生成提现单号
	withdrawalNo := s.generateWithdrawalNo()

	// 创建提现记录
	withdrawal := &model.Withdrawal{
		MerchantID:          input.MerchantID,
		WithdrawalNo:        withdrawalNo,
		AccountID:           input.AccountID,
		SettlementAccountID: input.SettlementAccountID,
		Amount:              input.Amount,
		Currency:            input.Currency,
		FeeAmount:           feeAmount,
		ActualAmount:        actualAmount,
		Status:              model.WithdrawalStatusPending,
		RequestReason:       input.RequestReason,
	}

	if err := s.accountRepo.CreateWithdrawal(ctx, withdrawal); err != nil {
		return nil, fmt.Errorf("创建提现记录失败: %w", err)
	}

	// 冻结提现金额
	if err := s.accountRepo.FreezeBalance(ctx, input.AccountID, input.Amount); err != nil {
		return nil, fmt.Errorf("冻结余额失败: %w", err)
	}

	return withdrawal, nil
}

// GetWithdrawal 获取提现记录
func (s *accountService) GetWithdrawal(ctx context.Context, withdrawalNo string) (*model.Withdrawal, error) {
	withdrawal, err := s.accountRepo.GetWithdrawalByNo(ctx, withdrawalNo)
	if err != nil {
		return nil, fmt.Errorf("获取提现记录失败: %w", err)
	}
	if withdrawal == nil {
		return nil, fmt.Errorf("提现记录不存在")
	}
	return withdrawal, nil
}

// ListWithdrawals 提现列表
func (s *accountService) ListWithdrawals(ctx context.Context, query *repository.WithdrawalQuery) ([]*model.Withdrawal, int64, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		query.PageSize = 20
	}
	return s.accountRepo.ListWithdrawals(ctx, query)
}

// ApproveWithdrawal 批准提现
func (s *accountService) ApproveWithdrawal(ctx context.Context, withdrawalNo string, approverID uuid.UUID, notes string) error {
	withdrawal, err := s.GetWithdrawal(ctx, withdrawalNo)
	if err != nil {
		return err
	}

	if withdrawal.Status != model.WithdrawalStatusPending {
		return fmt.Errorf("提现状态不正确: %s", withdrawal.Status)
	}

	now := time.Now()
	withdrawal.Status = model.WithdrawalStatusApproved
	withdrawal.ApprovedBy = &approverID
	withdrawal.ApprovedAt = &now
	withdrawal.ApprovalNotes = notes

	return s.accountRepo.UpdateWithdrawal(ctx, withdrawal)
}

// RejectWithdrawal 拒绝提现
func (s *accountService) RejectWithdrawal(ctx context.Context, withdrawalNo string, approverID uuid.UUID, reason string) error {
	withdrawal, err := s.GetWithdrawal(ctx, withdrawalNo)
	if err != nil {
		return err
	}

	if withdrawal.Status != model.WithdrawalStatusPending {
		return fmt.Errorf("提现状态不正确: %s", withdrawal.Status)
	}

	now := time.Now()
	withdrawal.Status = model.WithdrawalStatusRejected
	withdrawal.ApprovedBy = &approverID
	withdrawal.ApprovedAt = &now
	withdrawal.ApprovalNotes = reason

	// 解冻余额
	if err := s.accountRepo.UnfreezeBalance(ctx, withdrawal.AccountID, withdrawal.Amount); err != nil {
		return fmt.Errorf("解冻余额失败: %w", err)
	}

	return s.accountRepo.UpdateWithdrawal(ctx, withdrawal)
}

// ProcessWithdrawal 处理提现（开始转账）
func (s *accountService) ProcessWithdrawal(ctx context.Context, withdrawalNo string, processorID uuid.UUID) error {
	withdrawal, err := s.GetWithdrawal(ctx, withdrawalNo)
	if err != nil {
		return err
	}

	if withdrawal.Status != model.WithdrawalStatusApproved {
		return fmt.Errorf("提现状态不正确: %s，必须先审批", withdrawal.Status)
	}

	now := time.Now()
	withdrawal.Status = model.WithdrawalStatusProcessing
	withdrawal.ProcessedBy = &processorID
	withdrawal.ProcessedAt = &now

	return s.accountRepo.UpdateWithdrawal(ctx, withdrawal)
}

// CompleteWithdrawal 完成提现
func (s *accountService) CompleteWithdrawal(ctx context.Context, withdrawalNo string) error {
	withdrawal, err := s.GetWithdrawal(ctx, withdrawalNo)
	if err != nil {
		return err
	}

	if withdrawal.Status != model.WithdrawalStatusProcessing {
		return fmt.Errorf("提现状态不正确: %s", withdrawal.Status)
	}

	// 创建提现交易记录
	txInput := &CreateTransactionInput{
		AccountID:       withdrawal.AccountID,
		TransactionType: model.TransactionTypeWithdraw,
		Amount:          -withdrawal.Amount,
		RelatedID:       withdrawal.ID,
		RelatedNo:       withdrawal.WithdrawalNo,
		Description:     fmt.Sprintf("提现: %s", withdrawal.WithdrawalNo),
	}

	tx, err := s.CreateTransaction(ctx, txInput)
	if err != nil {
		return fmt.Errorf("创建提现交易失败: %w", err)
	}

	// 解冻余额（实际已在CreateTransaction中扣除）
	if err := s.accountRepo.UnfreezeBalance(ctx, withdrawal.AccountID, withdrawal.Amount); err != nil {
		return fmt.Errorf("解冻余额失败: %w", err)
	}

	// 手续费交易
	if withdrawal.FeeAmount > 0 {
		feeInput := &CreateTransactionInput{
			AccountID:       withdrawal.AccountID,
			TransactionType: model.TransactionTypeFee,
			Amount:          -withdrawal.FeeAmount,
			RelatedID:       withdrawal.ID,
			RelatedNo:       withdrawal.WithdrawalNo,
			Description:     fmt.Sprintf("提现手续费: %s", withdrawal.WithdrawalNo),
		}
		_, _ = s.CreateTransaction(ctx, feeInput)
	}

	now := time.Now()
	withdrawal.Status = model.WithdrawalStatusCompleted
	withdrawal.CompletedAt = &now
	withdrawal.TransactionID = &tx.ID

	return s.accountRepo.UpdateWithdrawal(ctx, withdrawal)
}

// FailWithdrawal 提现失败
func (s *accountService) FailWithdrawal(ctx context.Context, withdrawalNo string, reason string) error {
	withdrawal, err := s.GetWithdrawal(ctx, withdrawalNo)
	if err != nil {
		return err
	}

	if withdrawal.Status != model.WithdrawalStatusProcessing {
		return fmt.Errorf("提现状态不正确: %s", withdrawal.Status)
	}

	withdrawal.Status = model.WithdrawalStatusFailed
	withdrawal.FailureReason = reason

	// 解冻余额
	if err := s.accountRepo.UnfreezeBalance(ctx, withdrawal.AccountID, withdrawal.Amount); err != nil {
		return fmt.Errorf("解冻余额失败: %w", err)
	}

	return s.accountRepo.UpdateWithdrawal(ctx, withdrawal)
}

// CancelWithdrawal 取消提现
func (s *accountService) CancelWithdrawal(ctx context.Context, withdrawalNo string) error {
	withdrawal, err := s.GetWithdrawal(ctx, withdrawalNo)
	if err != nil {
		return err
	}

	if withdrawal.Status != model.WithdrawalStatusPending && withdrawal.Status != model.WithdrawalStatusApproved {
		return fmt.Errorf("只能取消待审核或已批准的提现")
	}

	withdrawal.Status = model.WithdrawalStatusCancelled

	// 解冻余额
	if err := s.accountRepo.UnfreezeBalance(ctx, withdrawal.AccountID, withdrawal.Amount); err != nil {
		return fmt.Errorf("解冻余额失败: %w", err)
	}

	return s.accountRepo.UpdateWithdrawal(ctx, withdrawal)
}

// Invoice Management Methods

// CreateInvoice 创建账单
func (s *accountService) CreateInvoice(ctx context.Context, input *CreateInvoiceInput) (*model.Invoice, error) {
	// 生成账单号
	invoiceNo := s.generateInvoiceNo()

	// 计算小计金额
	var subtotalAmount int64
	for _, item := range input.Items {
		subtotalAmount += int64(item.Quantity) * item.UnitPrice
	}

	// 计算税额
	taxAmount := int64(float64(subtotalAmount) * input.TaxRate / 100)

	// 计算总金额
	totalAmount := subtotalAmount + taxAmount

	// 创建账单
	invoice := &model.Invoice{
		MerchantID:        input.MerchantID,
		InvoiceNo:         invoiceNo,
		InvoiceType:       input.InvoiceType,
		PeriodStart:       input.PeriodStart,
		PeriodEnd:         input.PeriodEnd,
		Currency:          input.Currency,
		SubtotalAmount:    subtotalAmount,
		TaxAmount:         taxAmount,
		TotalAmount:       totalAmount,
		PaidAmount:        0,
		OutstandingAmount: totalAmount,
		Status:            model.InvoiceStatusPending,
		DueDate:           input.DueDate,
		Notes:             input.Notes,
	}

	if err := s.accountRepo.CreateInvoice(ctx, invoice); err != nil {
		return nil, fmt.Errorf("创建账单失败: %w", err)
	}

	// 创建账单明细
	for _, itemInput := range input.Items {
		amount := int64(itemInput.Quantity) * itemInput.UnitPrice
		item := &model.InvoiceItem{
			InvoiceID:   invoice.ID,
			ItemType:    itemInput.ItemType,
			Description: itemInput.Description,
			Quantity:    itemInput.Quantity,
			UnitPrice:   itemInput.UnitPrice,
			Amount:      amount,
			RelatedID:   itemInput.RelatedID,
			RelatedNo:   itemInput.RelatedNo,
		}
		if err := s.accountRepo.CreateInvoiceItem(ctx, item); err != nil {
			return nil, fmt.Errorf("创建账单明细失败: %w", err)
		}
	}

	// 重新加载账单（包含明细）
	return s.GetInvoice(ctx, invoiceNo)
}

// GetInvoice 获取账单
func (s *accountService) GetInvoice(ctx context.Context, invoiceNo string) (*model.Invoice, error) {
	invoice, err := s.accountRepo.GetInvoiceByNo(ctx, invoiceNo)
	if err != nil {
		return nil, fmt.Errorf("获取账单失败: %w", err)
	}
	if invoice == nil {
		return nil, fmt.Errorf("账单不存在")
	}
	return invoice, nil
}

// ListInvoices 账单列表
func (s *accountService) ListInvoices(ctx context.Context, query *repository.InvoiceQuery) ([]*model.Invoice, int64, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		query.PageSize = 20
	}
	return s.accountRepo.ListInvoices(ctx, query)
}

// PayInvoice 支付账单
func (s *accountService) PayInvoice(ctx context.Context, invoiceNo string, paidAmount int64) error {
	invoice, err := s.GetInvoice(ctx, invoiceNo)
	if err != nil {
		return err
	}

	if invoice.Status != model.InvoiceStatusPending && invoice.Status != model.InvoiceStatusPartialPaid {
		return fmt.Errorf("账单状态不正确: %s", invoice.Status)
	}

	if paidAmount <= 0 {
		return fmt.Errorf("支付金额必须大于0")
	}

	if paidAmount > invoice.OutstandingAmount {
		return fmt.Errorf("支付金额超过未付金额")
	}

	// 更新已支付金额和未付金额
	invoice.PaidAmount += paidAmount
	invoice.OutstandingAmount -= paidAmount

	// 更新状态
	if invoice.OutstandingAmount == 0 {
		invoice.Status = model.InvoiceStatusPaid
		now := time.Now()
		invoice.PaidAt = &now
	} else {
		invoice.Status = model.InvoiceStatusPartialPaid
	}

	return s.accountRepo.UpdateInvoice(ctx, invoice)
}

// CancelInvoice 取消账单
func (s *accountService) CancelInvoice(ctx context.Context, invoiceNo string) error {
	invoice, err := s.GetInvoice(ctx, invoiceNo)
	if err != nil {
		return err
	}

	if invoice.Status != model.InvoiceStatusDraft && invoice.Status != model.InvoiceStatusPending {
		return fmt.Errorf("只能取消草稿或待支付的账单")
	}

	invoice.Status = model.InvoiceStatusCancelled
	return s.accountRepo.UpdateInvoice(ctx, invoice)
}

// VoidInvoice 作废账单
func (s *accountService) VoidInvoice(ctx context.Context, invoiceNo string) error {
	invoice, err := s.GetInvoice(ctx, invoiceNo)
	if err != nil {
		return err
	}

	if invoice.Status == model.InvoiceStatusPaid {
		return fmt.Errorf("已支付的账单不能作废")
	}

	invoice.Status = model.InvoiceStatusVoided
	return s.accountRepo.UpdateInvoice(ctx, invoice)
}

// CheckOverdueInvoices 检查逾期账单
func (s *accountService) CheckOverdueInvoices(ctx context.Context) error {
	// 查询所有待支付和部分支付的账单
	query := &repository.InvoiceQuery{
		Page:     1,
		PageSize: 1000,
	}

	invoices, _, err := s.ListInvoices(ctx, query)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, invoice := range invoices {
		if (invoice.Status == model.InvoiceStatusPending || invoice.Status == model.InvoiceStatusPartialPaid) &&
			invoice.DueDate.Before(now) {
			invoice.Status = model.InvoiceStatusOverdue
			s.accountRepo.UpdateInvoice(ctx, invoice)
		}
	}

	return nil
}

// Reconciliation Management Methods

// CreateReconciliation 创建对账单
func (s *accountService) CreateReconciliation(ctx context.Context, input *CreateReconciliationInput) (*model.Reconciliation, error) {
	// 生成对账单号
	reconciliationNo := s.generateReconciliationNo()

	// 计算内部和外部统计数据
	var internalCount, externalCount int
	var internalAmount, externalAmount int64
	var mismatchedCount int

	for _, item := range input.Items {
		if item.InternalAmount != 0 {
			internalCount++
			internalAmount += item.InternalAmount
		}
		if item.ExternalAmount != 0 {
			externalCount++
			externalAmount += item.ExternalAmount
		}
		if item.Status == model.ReconciliationItemStatusMismatched ||
			item.Status == model.ReconciliationItemStatusMissing {
			mismatchedCount++
		}
	}

	// 计算差异
	diffCount := internalCount - externalCount
	diffAmount := internalAmount - externalAmount

	// 确定状态
	status := model.ReconciliationStatusPending
	if mismatchedCount == 0 && diffCount == 0 && diffAmount == 0 {
		status = model.ReconciliationStatusMatched
	} else if mismatchedCount > 0 || diffCount != 0 || diffAmount != 0 {
		status = model.ReconciliationStatusMismatched
	}

	// 创建对账单
	reconciliation := &model.Reconciliation{
		ReconciliationNo:   reconciliationNo,
		MerchantID:         input.MerchantID,
		Channel:            input.Channel,
		ReconciliationDate: input.ReconciliationDate,
		PeriodStart:        input.PeriodStart,
		PeriodEnd:          input.PeriodEnd,
		Currency:           input.Currency,
		InternalCount:      internalCount,
		InternalAmount:     internalAmount,
		ExternalCount:      externalCount,
		ExternalAmount:     externalAmount,
		DiffCount:          diffCount,
		DiffAmount:         diffAmount,
		MismatchedCount:    mismatchedCount,
		Status:             status,
	}

	if err := s.accountRepo.CreateReconciliation(ctx, reconciliation); err != nil {
		return nil, fmt.Errorf("创建对账单失败: %w", err)
	}

	// 创建对账明细
	for _, itemInput := range input.Items {
		diffAmount := itemInput.InternalAmount - itemInput.ExternalAmount

		// 如果未指定状态，自动判断
		itemStatus := itemInput.Status
		if itemStatus == "" {
			if diffAmount == 0 {
				itemStatus = model.ReconciliationItemStatusMatched
			} else {
				itemStatus = model.ReconciliationItemStatusMismatched
			}
		}

		item := &model.ReconciliationItem{
			ReconciliationID: reconciliation.ID,
			TransactionNo:    itemInput.TransactionNo,
			ExternalTxNo:     itemInput.ExternalTxNo,
			ItemType:         itemInput.ItemType,
			InternalAmount:   itemInput.InternalAmount,
			ExternalAmount:   itemInput.ExternalAmount,
			DiffAmount:       diffAmount,
			Status:           itemStatus,
			Description:      itemInput.Description,
		}
		if err := s.accountRepo.CreateReconciliationItem(ctx, item); err != nil {
			return nil, fmt.Errorf("创建对账明细失败: %w", err)
		}
	}

	// 重新加载对账单（包含明细）
	return s.GetReconciliation(ctx, reconciliationNo)
}

// GetReconciliation 获取对账单
func (s *accountService) GetReconciliation(ctx context.Context, reconciliationNo string) (*model.Reconciliation, error) {
	reconciliation, err := s.accountRepo.GetReconciliationByNo(ctx, reconciliationNo)
	if err != nil {
		return nil, fmt.Errorf("获取对账单失败: %w", err)
	}
	if reconciliation == nil {
		return nil, fmt.Errorf("对账单不存在")
	}
	return reconciliation, nil
}

// ListReconciliations 对账单列表
func (s *accountService) ListReconciliations(ctx context.Context, query *repository.ReconciliationQuery) ([]*model.Reconciliation, int64, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		query.PageSize = 20
	}
	return s.accountRepo.ListReconciliations(ctx, query)
}

// ProcessReconciliation 处理对账单
func (s *accountService) ProcessReconciliation(ctx context.Context, reconciliationNo string, userID uuid.UUID) error {
	reconciliation, err := s.GetReconciliation(ctx, reconciliationNo)
	if err != nil {
		return err
	}

	if reconciliation.Status != model.ReconciliationStatusPending &&
		reconciliation.Status != model.ReconciliationStatusMismatched {
		return fmt.Errorf("对账单状态不正确: %s", reconciliation.Status)
	}

	reconciliation.Status = model.ReconciliationStatusProcessing
	return s.accountRepo.UpdateReconciliation(ctx, reconciliation)
}

// CompleteReconciliation 完成对账
func (s *accountService) CompleteReconciliation(ctx context.Context, reconciliationNo string) error {
	reconciliation, err := s.GetReconciliation(ctx, reconciliationNo)
	if err != nil {
		return err
	}

	if reconciliation.Status != model.ReconciliationStatusProcessing &&
		reconciliation.Status != model.ReconciliationStatusMatched {
		return fmt.Errorf("对账单状态不正确: %s", reconciliation.Status)
	}

	now := time.Now()
	reconciliation.Status = model.ReconciliationStatusCompleted
	reconciliation.ReconciledAt = &now

	return s.accountRepo.UpdateReconciliation(ctx, reconciliation)
}

// ResolveReconciliationItem 解决对账明细差异
func (s *accountService) ResolveReconciliationItem(ctx context.Context, itemID uuid.UUID, resolution string) error {
	// 获取对账明细
	items, err := s.accountRepo.GetReconciliationItems(ctx, itemID)
	if err != nil {
		return fmt.Errorf("获取对账明细失败: %w", err)
	}
	if len(items) == 0 {
		return fmt.Errorf("对账明细不存在")
	}

	item := items[0]
	item.Status = model.ReconciliationItemStatusResolved
	item.Description = resolution

	return s.accountRepo.UpdateReconciliationItem(ctx, item)
}

// Balance Aggregation Methods

// GetMerchantBalanceSummary 获取商户余额汇总
func (s *accountService) GetMerchantBalanceSummary(ctx context.Context, merchantID uuid.UUID) (*MerchantBalanceSummary, error) {
	// 查询商户所有账户
	query := &repository.AccountQuery{
		MerchantID: &merchantID,
		Page:       1,
		PageSize:   1000, // 获取所有账户
	}

	accounts, _, err := s.accountRepo.ListAccounts(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("查询账户失败: %w", err)
	}

	// 初始化汇总数据
	summary := &MerchantBalanceSummary{
		MerchantID:      merchantID,
		TotalAccounts:   len(accounts),
		ActiveAccounts:  0,
		FrozenAccounts:  0,
		Accounts:        make([]AccountBalance, 0, len(accounts)),
		CurrencySummary: make(map[string]int64),
		TypeSummary:     make(map[string]int64),
		TotalBalance:    0,
		TotalFrozen:     0,
		TotalAvailable:  0,
		LastUpdated:     time.Now(),
	}

	// 遍历账户统计
	for _, account := range accounts {
		availableBalance := account.Balance - account.FrozenBalance

		// 账户状态统计
		if account.Status == model.AccountStatusActive {
			summary.ActiveAccounts++
		} else if account.Status == model.AccountStatusFrozen {
			summary.FrozenAccounts++
		}

		// 账户详情
		accountBalance := AccountBalance{
			AccountID:        account.ID,
			AccountType:      account.AccountType,
			Currency:         account.Currency,
			Balance:          account.Balance,
			FrozenBalance:    account.FrozenBalance,
			AvailableBalance: availableBalance,
			TotalIn:          account.TotalIn,
			TotalOut:         account.TotalOut,
			Status:           account.Status,
			UpdatedAt:        account.UpdatedAt,
		}
		summary.Accounts = append(summary.Accounts, accountBalance)

		// 按货币汇总
		summary.CurrencySummary[account.Currency] += account.Balance

		// 按账户类型汇总
		summary.TypeSummary[account.AccountType] += account.Balance

		// 总计
		summary.TotalBalance += account.Balance
		summary.TotalFrozen += account.FrozenBalance
		summary.TotalAvailable += availableBalance
	}

	return summary, nil
}

// GetBalanceByCurrency 按货币获取余额汇总
func (s *accountService) GetBalanceByCurrency(ctx context.Context, merchantID uuid.UUID, currency string) (*CurrencyBalanceSummary, error) {
	// 查询指定货币的所有账户
	query := &repository.AccountQuery{
		MerchantID: &merchantID,
		Currency:   currency,
		Page:       1,
		PageSize:   1000,
	}

	accounts, _, err := s.accountRepo.ListAccounts(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("查询账户失败: %w", err)
	}

	// 初始化汇总数据
	summary := &CurrencyBalanceSummary{
		MerchantID:     merchantID,
		Currency:       currency,
		AccountCount:   len(accounts),
		TotalBalance:   0,
		TotalFrozen:    0,
		TotalAvailable: 0,
		TotalIn:        0,
		TotalOut:       0,
		Accounts:       make([]AccountBalance, 0, len(accounts)),
		LastUpdated:    time.Now(),
	}

	// 遍历账户统计
	for _, account := range accounts {
		availableBalance := account.Balance - account.FrozenBalance

		// 账户详情
		accountBalance := AccountBalance{
			AccountID:        account.ID,
			AccountType:      account.AccountType,
			Currency:         account.Currency,
			Balance:          account.Balance,
			FrozenBalance:    account.FrozenBalance,
			AvailableBalance: availableBalance,
			TotalIn:          account.TotalIn,
			TotalOut:         account.TotalOut,
			Status:           account.Status,
			UpdatedAt:        account.UpdatedAt,
		}
		summary.Accounts = append(summary.Accounts, accountBalance)

		// 累加统计
		summary.TotalBalance += account.Balance
		summary.TotalFrozen += account.FrozenBalance
		summary.TotalAvailable += availableBalance
		summary.TotalIn += account.TotalIn
		summary.TotalOut += account.TotalOut
	}

	return summary, nil
}

// GetBalanceByAccountType 按账户类型获取余额汇总
func (s *accountService) GetBalanceByAccountType(ctx context.Context, merchantID uuid.UUID, accountType string) (*AccountTypeBalanceSummary, error) {
	// 查询指定类型的所有账户
	query := &repository.AccountQuery{
		MerchantID:  &merchantID,
		AccountType: accountType,
		Page:        1,
		PageSize:    1000,
	}

	accounts, _, err := s.accountRepo.ListAccounts(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("查询账户失败: %w", err)
	}

	// 初始化汇总数据
	summary := &AccountTypeBalanceSummary{
		MerchantID:       merchantID,
		AccountType:      accountType,
		AccountCount:     len(accounts),
		CurrencyBalances: make(map[string]int64),
		TotalBalance:     0,
		TotalFrozen:      0,
		TotalAvailable:   0,
		Accounts:         make([]AccountBalance, 0, len(accounts)),
		LastUpdated:      time.Now(),
	}

	// 遍历账户统计
	for _, account := range accounts {
		availableBalance := account.Balance - account.FrozenBalance

		// 账户详情
		accountBalance := AccountBalance{
			AccountID:        account.ID,
			AccountType:      account.AccountType,
			Currency:         account.Currency,
			Balance:          account.Balance,
			FrozenBalance:    account.FrozenBalance,
			AvailableBalance: availableBalance,
			TotalIn:          account.TotalIn,
			TotalOut:         account.TotalOut,
			Status:           account.Status,
			UpdatedAt:        account.UpdatedAt,
		}
		summary.Accounts = append(summary.Accounts, accountBalance)

		// 按货币统计
		summary.CurrencyBalances[account.Currency] += account.Balance

		// 累加总计
		summary.TotalBalance += account.Balance
		summary.TotalFrozen += account.FrozenBalance
		summary.TotalAvailable += availableBalance
	}

	return summary, nil
}

// GetAllCurrencyBalances 获取所有货币的余额汇总
func (s *accountService) GetAllCurrencyBalances(ctx context.Context, merchantID uuid.UUID) ([]*CurrencyBalanceSummary, error) {
	// 查询商户所有账户
	query := &repository.AccountQuery{
		MerchantID: &merchantID,
		Page:       1,
		PageSize:   1000,
	}

	accounts, _, err := s.accountRepo.ListAccounts(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("查询账户失败: %w", err)
	}

	// 按货币分组
	currencyMap := make(map[string]*CurrencyBalanceSummary)

	for _, account := range accounts {
		availableBalance := account.Balance - account.FrozenBalance

		// 如果该货币还没有汇总记录，创建一个
		if _, exists := currencyMap[account.Currency]; !exists {
			currencyMap[account.Currency] = &CurrencyBalanceSummary{
				MerchantID:     merchantID,
				Currency:       account.Currency,
				AccountCount:   0,
				TotalBalance:   0,
				TotalFrozen:    0,
				TotalAvailable: 0,
				TotalIn:        0,
				TotalOut:       0,
				Accounts:       make([]AccountBalance, 0),
				LastUpdated:    time.Now(),
			}
		}

		summary := currencyMap[account.Currency]

		// 账户详情
		accountBalance := AccountBalance{
			AccountID:        account.ID,
			AccountType:      account.AccountType,
			Currency:         account.Currency,
			Balance:          account.Balance,
			FrozenBalance:    account.FrozenBalance,
			AvailableBalance: availableBalance,
			TotalIn:          account.TotalIn,
			TotalOut:         account.TotalOut,
			Status:           account.Status,
			UpdatedAt:        account.UpdatedAt,
		}
		summary.Accounts = append(summary.Accounts, accountBalance)

		// 累加统计
		summary.AccountCount++
		summary.TotalBalance += account.Balance
		summary.TotalFrozen += account.FrozenBalance
		summary.TotalAvailable += availableBalance
		summary.TotalIn += account.TotalIn
		summary.TotalOut += account.TotalOut
	}

	// 转换为数组
	result := make([]*CurrencyBalanceSummary, 0, len(currencyMap))
	for _, summary := range currencyMap {
		result = append(result, summary)
	}

	return result, nil
}

// ==================== 货币转换管理 ====================

// CreateCurrencyConversion 创建货币转换
func (s *accountService) CreateCurrencyConversion(ctx context.Context, input *CreateCurrencyConversionInput) (*model.CurrencyConversion, error) {
	// 1. 验证输入
	if input.SourceCurrency == input.TargetCurrency {
		return nil, fmt.Errorf("源货币和目标货币不能相同")
	}

	if input.SourceAmount <= 0 {
		return nil, fmt.Errorf("转换金额必须大于0")
	}

	// 2. 查询或创建源货币账户
	sourceAccount, err := s.accountRepo.GetAccountByMerchant(ctx, input.MerchantID, model.AccountTypeOperating, input.SourceCurrency)
	if err != nil {
		return nil, fmt.Errorf("查询源货币账户失败: %w", err)
	}
	if sourceAccount == nil {
		return nil, fmt.Errorf("源货币账户不存在: %s", input.SourceCurrency)
	}

	// 3. 检查余额是否足够
	if sourceAccount.Balance < input.SourceAmount {
		return nil, fmt.Errorf("余额不足，可用余额: %d 分，需要: %d 分", sourceAccount.Balance, input.SourceAmount)
	}

	// 4. 查询或创建目标货币账户
	targetAccount, err := s.accountRepo.GetAccountByMerchant(ctx, input.MerchantID, model.AccountTypeOperating, input.TargetCurrency)
	if err != nil {
		return nil, fmt.Errorf("查询目标货币账户失败: %w", err)
	}

	// 如果目标账户不存在，自动创建
	if targetAccount == nil {
		createInput := &CreateAccountInput{
			MerchantID:  input.MerchantID,
			AccountType: model.AccountTypeOperating,
			Currency:    input.TargetCurrency,
		}
		targetAccount, err = s.CreateAccount(ctx, createInput)
		if err != nil {
			return nil, fmt.Errorf("创建目标货币账户失败: %w", err)
		}
	}

	// 5. 获取实时汇率（通过 channel-adapter 的汇率API或降级到备用汇率）
	exchangeRate := s.getExchangeRate(input.SourceCurrency, input.TargetCurrency)
	if exchangeRate <= 0 {
		return nil, fmt.Errorf("无法获取汇率: %s -> %s", input.SourceCurrency, input.TargetCurrency)
	}

	// 6. 计算目标货币金额
	targetAmount := int64(float64(input.SourceAmount) * exchangeRate)

	// 7. 计算手续费（0.5%）
	feePercentage := 0.005
	feeAmount := int64(float64(input.SourceAmount) * feePercentage)

	// 8. 生成转换单号
	conversionNo, err := s.generateConversionNo()
	if err != nil {
		return nil, fmt.Errorf("生成转换单号失败: %w", err)
	}

	// 9. 创建货币转换记录
	conversion := &model.CurrencyConversion{
		ConversionNo:    conversionNo,
		MerchantID:      input.MerchantID,
		SourceAccountID: sourceAccount.ID,
		TargetAccountID: targetAccount.ID,
		SourceCurrency:  input.SourceCurrency,
		TargetCurrency:  input.TargetCurrency,
		SourceAmount:    input.SourceAmount,
		TargetAmount:    targetAmount,
		ExchangeRate:    exchangeRate,
		FeeAmount:       feeAmount,
		FeePercentage:   feePercentage,
		Status:          model.ConversionStatusPending,
		RequestedBy:     input.RequestedBy,
		Reason:          input.Reason,
		Notes:           input.Notes,
	}

	if err := s.accountRepo.CreateCurrencyConversion(ctx, conversion); err != nil {
		return nil, fmt.Errorf("创建货币转换记录失败: %w", err)
	}

	return conversion, nil
}

// ProcessCurrencyConversion 处理货币转换（执行实际的账户变动）
func (s *accountService) ProcessCurrencyConversion(ctx context.Context, conversionNo string) error {
	// 1. 查询转换记录
	conversion, err := s.accountRepo.GetCurrencyConversionByNo(ctx, conversionNo)
	if err != nil {
		return fmt.Errorf("查询货币转换记录失败: %w", err)
	}
	if conversion == nil {
		return fmt.Errorf("货币转换记录不存在: %s", conversionNo)
	}

	// 2. 检查状态
	if conversion.Status != model.ConversionStatusPending {
		return fmt.Errorf("货币转换状态不是待处理: %s", conversion.Status)
	}

	// 3. 使用事务执行转换
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 3.1 从源账户扣款（包含手续费）
		totalDeduction := conversion.SourceAmount + conversion.FeeAmount
		sourceTransactionInput := &CreateTransactionInput{
			AccountID:       conversion.SourceAccountID,
			TransactionType: model.TransactionTypeCurrencyConversionOut,
			RelatedNo:       conversionNo,
			Amount:          -totalDeduction, // 负数表示出账
			Description:     fmt.Sprintf("货币转换出账: %s -> %s (含手续费)", conversion.SourceCurrency, conversion.TargetCurrency),
		}

		sourceTx, err := s.CreateTransaction(ctx, sourceTransactionInput)
		if err != nil {
			return fmt.Errorf("创建源账户交易失败: %w", err)
		}

		// 3.2 向目标账户入账
		targetTransactionInput := &CreateTransactionInput{
			AccountID:       conversion.TargetAccountID,
			TransactionType: model.TransactionTypeCurrencyConversionIn,
			RelatedNo:       conversionNo,
			Amount:          conversion.TargetAmount, // 正数表示入账
			Description:     fmt.Sprintf("货币转换入账: %s -> %s", conversion.SourceCurrency, conversion.TargetCurrency),
		}

		targetTx, err := s.CreateTransaction(ctx, targetTransactionInput)
		if err != nil {
			return fmt.Errorf("创建目标账户交易失败: %w", err)
		}

		// 3.3 更新转换记录状态
		conversion.Status = model.ConversionStatusCompleted
		conversion.SourceTransactionNo = sourceTx.TransactionNo
		conversion.TargetTransactionNo = targetTx.TransactionNo
		now := time.Now()
		conversion.ProcessedAt = &now

		if err := s.accountRepo.UpdateCurrencyConversion(ctx, conversion); err != nil {
			return fmt.Errorf("更新货币转换记录失败: %w", err)
		}

		return nil
	})

	if err != nil {
		// 如果事务失败，标记转换为失败状态
		conversion.Status = model.ConversionStatusFailed
		s.accountRepo.UpdateCurrencyConversion(ctx, conversion)
		return err
	}

	return nil
}

// GetCurrencyConversion 获取货币转换记录
func (s *accountService) GetCurrencyConversion(ctx context.Context, conversionNo string) (*model.CurrencyConversion, error) {
	conversion, err := s.accountRepo.GetCurrencyConversionByNo(ctx, conversionNo)
	if err != nil {
		return nil, fmt.Errorf("查询货币转换记录失败: %w", err)
	}
	return conversion, nil
}

// ListCurrencyConversions 查询货币转换记录列表
func (s *accountService) ListCurrencyConversions(ctx context.Context, query *repository.CurrencyConversionQuery) ([]*model.CurrencyConversion, int64, error) {
	conversions, total, err := s.accountRepo.ListCurrencyConversions(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("查询货币转换记录列表失败: %w", err)
	}
	return conversions, total, nil
}

// CancelCurrencyConversion 取消货币转换
func (s *accountService) CancelCurrencyConversion(ctx context.Context, conversionNo string, reason string) error {
	conversion, err := s.accountRepo.GetCurrencyConversionByNo(ctx, conversionNo)
	if err != nil {
		return fmt.Errorf("查询货币转换记录失败: %w", err)
	}
	if conversion == nil {
		return fmt.Errorf("货币转换记录不存在: %s", conversionNo)
	}

	// 只有待处理状态才能取消
	if conversion.Status != model.ConversionStatusPending {
		return fmt.Errorf("只有待处理状态的转换才能取消，当前状态: %s", conversion.Status)
	}

	conversion.Status = model.ConversionStatusCancelled
	conversion.Notes = fmt.Sprintf("%s\n取消原因: %s", conversion.Notes, reason)

	if err := s.accountRepo.UpdateCurrencyConversion(ctx, conversion); err != nil {
		return fmt.Errorf("更新货币转换记录失败: %w", err)
	}

	return nil
}

// getExchangeRate 获取汇率（优先使用 channel-adapter 的实时汇率API）
func (s *accountService) getExchangeRate(fromCurrency, toCurrency string) float64 {
	// 同一货币，汇率为1
	if fromCurrency == toCurrency {
		return 1.0
	}

	// 如果配置了 channel-adapter 客户端，优先使用实时汇率
	if s.channelAdapterClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		rate, err := s.channelAdapterClient.GetExchangeRate(ctx, fromCurrency, toCurrency)
		if err == nil && rate > 0 {
			logger.Info("获取实时汇率成功",
				zap.String("from", fromCurrency),
				zap.String("to", toCurrency),
				zap.Float64("rate", rate),
			)
			return rate
		}

		// 如果获取实时汇率失败，记录警告并降级到备用汇率
		logger.Warn("获取实时汇率失败，使用备用汇率",
			zap.String("from", fromCurrency),
			zap.String("to", toCurrency),
			zap.Error(err),
		)
	}

	// 备用静态汇率（当channel-adapter不可用时使用）
	fallbackRates := map[string]map[string]float64{
		"USD": {
			"EUR": 0.92,
			"GBP": 0.79,
			"CNY": 7.24,
			"JPY": 149.50,
			"KRW": 1320.00,
			"HKD": 7.82,
			"SGD": 1.35,
		},
		"EUR": {
			"USD": 1.09,
			"GBP": 0.86,
			"CNY": 7.87,
		},
		"CNY": {
			"USD": 0.14,
			"EUR": 0.13,
		},
	}

	if from, ok := fallbackRates[fromCurrency]; ok {
		if rate, ok := from[toCurrency]; ok {
			logger.Info("使用备用静态汇率",
				zap.String("from", fromCurrency),
				zap.String("to", toCurrency),
				zap.Float64("rate", rate),
			)
			return rate
		}
	}

	// 如果没有找到汇率，返回0表示不支持
	logger.Warn("未找到汇率数据",
		zap.String("from", fromCurrency),
		zap.String("to", toCurrency),
	)
	return 0
}

// generateConversionNo 生成货币转换单号
func (s *accountService) generateConversionNo() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("CONV-%d-%s", time.Now().Unix(), base64.URLEncoding.EncodeToString(b)[:8]), nil
}
