package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/services/accounting-service/internal/model"
	"gorm.io/gorm"
)

// AccountRepository 账户仓储接口
type AccountRepository interface {
	// 账户管理
	CreateAccount(ctx context.Context, account *model.Account) error
	GetAccountByID(ctx context.Context, id uuid.UUID) (*model.Account, error)
	GetAccountByMerchant(ctx context.Context, merchantID uuid.UUID, accountType, currency string) (*model.Account, error)
	ListAccounts(ctx context.Context, query *AccountQuery) ([]*model.Account, int64, error)
	UpdateAccount(ctx context.Context, account *model.Account) error
	UpdateBalance(ctx context.Context, id uuid.UUID, amount int64) error
	FreezeBalance(ctx context.Context, id uuid.UUID, amount int64) error
	UnfreezeBalance(ctx context.Context, id uuid.UUID, amount int64) error

	// 交易记录
	CreateTransaction(ctx context.Context, tx *model.AccountTransaction) error
	GetTransactionByNo(ctx context.Context, transactionNo string) (*model.AccountTransaction, error)
	ListTransactions(ctx context.Context, query *TransactionQuery) ([]*model.AccountTransaction, int64, error)

	// 结算管理
	CreateSettlement(ctx context.Context, settlement *model.Settlement) error
	GetSettlementByID(ctx context.Context, id uuid.UUID) (*model.Settlement, error)
	GetSettlementByNo(ctx context.Context, settlementNo string) (*model.Settlement, error)
	ListSettlements(ctx context.Context, query *SettlementQuery) ([]*model.Settlement, int64, error)
	UpdateSettlement(ctx context.Context, settlement *model.Settlement) error

	// 复式记账
	CreateDoubleEntry(ctx context.Context, entry *model.DoubleEntry) error
	ListDoubleEntries(ctx context.Context, query *DoubleEntryQuery) ([]*model.DoubleEntry, int64, error)
}

type accountRepository struct {
	db *gorm.DB
}

// NewAccountRepository 创建账户仓储实例
func NewAccountRepository(db *gorm.DB) AccountRepository {
	return &accountRepository{db: db}
}

// AccountQuery 账户查询参数
type AccountQuery struct {
	MerchantID  *uuid.UUID
	AccountType string
	Currency    string
	Status      string
	Page        int
	PageSize    int
}

// TransactionQuery 交易查询参数
type TransactionQuery struct {
	AccountID       *uuid.UUID
	MerchantID      *uuid.UUID
	TransactionType string
	Currency        string
	StartTime       *time.Time
	EndTime         *time.Time
	Page            int
	PageSize        int
}

// SettlementQuery 结算查询参数
type SettlementQuery struct {
	MerchantID *uuid.UUID
	Status     string
	StartTime  *time.Time
	EndTime    *time.Time
	Page       int
	PageSize   int
}

// DoubleEntryQuery 复式记账查询参数
type DoubleEntryQuery struct {
	RelatedID *uuid.UUID
	RelatedNo string
	EntryType string
	StartTime *time.Time
	EndTime   *time.Time
	Page      int
	PageSize  int
}

// CreateAccount 创建账户
func (r *accountRepository) CreateAccount(ctx context.Context, account *model.Account) error {
	return r.db.WithContext(ctx).Create(account).Error
}

// GetAccountByID 根据ID获取账户
func (r *accountRepository) GetAccountByID(ctx context.Context, id uuid.UUID) (*model.Account, error) {
	var account model.Account
	err := r.db.WithContext(ctx).First(&account, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

// GetAccountByMerchant 根据商户获取账户
func (r *accountRepository) GetAccountByMerchant(ctx context.Context, merchantID uuid.UUID, accountType, currency string) (*model.Account, error) {
	var account model.Account
	err := r.db.WithContext(ctx).
		Where("merchant_id = ? AND account_type = ? AND currency = ?", merchantID, accountType, currency).
		First(&account).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

// ListAccounts 账户列表
func (r *accountRepository) ListAccounts(ctx context.Context, query *AccountQuery) ([]*model.Account, int64, error) {
	var accounts []*model.Account
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Account{})

	if query.MerchantID != nil {
		db = db.Where("merchant_id = ?", *query.MerchantID)
	}
	if query.AccountType != "" {
		db = db.Where("account_type = ?", query.AccountType)
	}
	if query.Currency != "" {
		db = db.Where("currency = ?", query.Currency)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	err := db.Offset(offset).Limit(query.PageSize).Find(&accounts).Error
	return accounts, total, err
}

// UpdateAccount 更新账户
func (r *accountRepository) UpdateAccount(ctx context.Context, account *model.Account) error {
	return r.db.WithContext(ctx).Save(account).Error
}

// UpdateBalance 更新余额
func (r *accountRepository) UpdateBalance(ctx context.Context, id uuid.UUID, amount int64) error {
	return r.db.WithContext(ctx).Model(&model.Account{}).
		Where("id = ?", id).
		UpdateColumns(map[string]interface{}{
			"balance":    gorm.Expr("balance + ?", amount),
			"updated_at": time.Now(),
		}).Error
}

// FreezeBalance 冻结余额
func (r *accountRepository) FreezeBalance(ctx context.Context, id uuid.UUID, amount int64) error {
	return r.db.WithContext(ctx).Model(&model.Account{}).
		Where("id = ? AND balance >= ?", id, amount).
		UpdateColumns(map[string]interface{}{
			"balance":        gorm.Expr("balance - ?", amount),
			"frozen_balance": gorm.Expr("frozen_balance + ?", amount),
			"updated_at":     time.Now(),
		}).Error
}

// UnfreezeBalance 解冻余额
func (r *accountRepository) UnfreezeBalance(ctx context.Context, id uuid.UUID, amount int64) error {
	return r.db.WithContext(ctx).Model(&model.Account{}).
		Where("id = ? AND frozen_balance >= ?", id, amount).
		UpdateColumns(map[string]interface{}{
			"balance":        gorm.Expr("balance + ?", amount),
			"frozen_balance": gorm.Expr("frozen_balance - ?", amount),
			"updated_at":     time.Now(),
		}).Error
}

// CreateTransaction 创建交易记录
func (r *accountRepository) CreateTransaction(ctx context.Context, tx *model.AccountTransaction) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

// GetTransactionByNo 根据交易号获取交易
func (r *accountRepository) GetTransactionByNo(ctx context.Context, transactionNo string) (*model.AccountTransaction, error) {
	var tx model.AccountTransaction
	err := r.db.WithContext(ctx).First(&tx, "transaction_no = ?", transactionNo).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &tx, nil
}

// ListTransactions 交易列表
func (r *accountRepository) ListTransactions(ctx context.Context, query *TransactionQuery) ([]*model.AccountTransaction, int64, error) {
	var transactions []*model.AccountTransaction
	var total int64

	db := r.db.WithContext(ctx).Model(&model.AccountTransaction{})

	if query.AccountID != nil {
		db = db.Where("account_id = ?", *query.AccountID)
	}
	if query.MerchantID != nil {
		db = db.Where("merchant_id = ?", *query.MerchantID)
	}
	if query.TransactionType != "" {
		db = db.Where("transaction_type = ?", query.TransactionType)
	}
	if query.Currency != "" {
		db = db.Where("currency = ?", query.Currency)
	}
	if query.StartTime != nil {
		db = db.Where("created_at >= ?", *query.StartTime)
	}
	if query.EndTime != nil {
		db = db.Where("created_at <= ?", *query.EndTime)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	err := db.Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&transactions).Error
	return transactions, total, err
}

// CreateSettlement 创建结算记录
func (r *accountRepository) CreateSettlement(ctx context.Context, settlement *model.Settlement) error {
	return r.db.WithContext(ctx).Create(settlement).Error
}

// GetSettlementByID 根据ID获取结算
func (r *accountRepository) GetSettlementByID(ctx context.Context, id uuid.UUID) (*model.Settlement, error) {
	var settlement model.Settlement
	err := r.db.WithContext(ctx).First(&settlement, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &settlement, nil
}

// GetSettlementByNo 根据结算单号获取结算
func (r *accountRepository) GetSettlementByNo(ctx context.Context, settlementNo string) (*model.Settlement, error) {
	var settlement model.Settlement
	err := r.db.WithContext(ctx).First(&settlement, "settlement_no = ?", settlementNo).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &settlement, nil
}

// ListSettlements 结算列表
func (r *accountRepository) ListSettlements(ctx context.Context, query *SettlementQuery) ([]*model.Settlement, int64, error) {
	var settlements []*model.Settlement
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Settlement{})

	if query.MerchantID != nil {
		db = db.Where("merchant_id = ?", *query.MerchantID)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.StartTime != nil {
		db = db.Where("period_start >= ?", *query.StartTime)
	}
	if query.EndTime != nil {
		db = db.Where("period_end <= ?", *query.EndTime)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	err := db.Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&settlements).Error
	return settlements, total, err
}

// UpdateSettlement 更新结算
func (r *accountRepository) UpdateSettlement(ctx context.Context, settlement *model.Settlement) error {
	return r.db.WithContext(ctx).Save(settlement).Error
}

// CreateDoubleEntry 创建复式记账
func (r *accountRepository) CreateDoubleEntry(ctx context.Context, entry *model.DoubleEntry) error {
	return r.db.WithContext(ctx).Create(entry).Error
}

// ListDoubleEntries 复式记账列表
func (r *accountRepository) ListDoubleEntries(ctx context.Context, query *DoubleEntryQuery) ([]*model.DoubleEntry, int64, error) {
	var entries []*model.DoubleEntry
	var total int64

	db := r.db.WithContext(ctx).Model(&model.DoubleEntry{})

	if query.RelatedID != nil {
		db = db.Where("related_id = ?", *query.RelatedID)
	}
	if query.RelatedNo != "" {
		db = db.Where("related_no = ?", query.RelatedNo)
	}
	if query.EntryType != "" {
		db = db.Where("entry_type = ?", query.EntryType)
	}
	if query.StartTime != nil {
		db = db.Where("created_at >= ?", *query.StartTime)
	}
	if query.EndTime != nil {
		db = db.Where("created_at <= ?", *query.EndTime)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	err := db.Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&entries).Error
	return entries, total, err
}
