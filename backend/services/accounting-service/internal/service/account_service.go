package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/services/accounting-service/internal/model"
	"github.com/payment-platform/services/accounting-service/internal/repository"
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

	// 结算管理
	CreateSettlement(ctx context.Context, input *CreateSettlementInput) (*model.Settlement, error)
	GetSettlement(ctx context.Context, settlementNo string) (*model.Settlement, error)
	ListSettlements(ctx context.Context, query *repository.SettlementQuery) ([]*model.Settlement, int64, error)
	ProcessSettlement(ctx context.Context, settlementNo string) error
}

type accountService struct {
	accountRepo repository.AccountRepository
}

// NewAccountService 创建账户服务实例
func NewAccountService(accountRepo repository.AccountRepository) AccountService {
	return &accountService{
		accountRepo: accountRepo,
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

	// TODO: 计算结算金额、手续费等

	settlement := &model.Settlement{
		MerchantID:   input.MerchantID,
		SettlementNo: settlementNo,
		AccountID:    input.AccountID,
		PeriodStart:  input.PeriodStart,
		PeriodEnd:    input.PeriodEnd,
		TotalAmount:  0,
		FeeAmount:    0,
		NetAmount:    0,
		Currency:     input.Currency,
		Status:       model.SettlementStatusPending,
		PaymentCount: input.PaymentCount,
		RefundCount:  input.RefundCount,
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

// ProcessSettlement 处理结算
func (s *accountService) ProcessSettlement(ctx context.Context, settlementNo string) error {
	settlement, err := s.GetSettlement(ctx, settlementNo)
	if err != nil {
		return err
	}

	if settlement.Status != model.SettlementStatusPending {
		return fmt.Errorf("结算状态不正确: %s", settlement.Status)
	}

	settlement.Status = model.SettlementStatusProcessing
	if err := s.accountRepo.UpdateSettlement(ctx, settlement); err != nil {
		return err
	}

	// TODO: 执行结算逻辑

	now := time.Now()
	settlement.Status = model.SettlementStatusCompleted
	settlement.SettledAt = &now
	return s.accountRepo.UpdateSettlement(ctx, settlement)
}

// 工具函数

func (s *accountService) generateTransactionNo() string {
	timestamp := time.Now().Format("20060102150405")
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	randomStr := base64.URLEncoding.EncodeToString(randomBytes)[:10]
	return fmt.Sprintf("TX%s%s", timestamp, randomStr)
}

func (s *accountService) generateSettlementNo() string {
	timestamp := time.Now().Format("20060102150405")
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	randomStr := base64.URLEncoding.EncodeToString(randomBytes)[:10]
	return fmt.Sprintf("ST%s%s", timestamp, randomStr)
}

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
