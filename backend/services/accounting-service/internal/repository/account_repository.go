package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"payment-platform/accounting-service/internal/model"
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

	// 提现管理
	CreateWithdrawal(ctx context.Context, withdrawal *model.Withdrawal) error
	GetWithdrawalByID(ctx context.Context, id uuid.UUID) (*model.Withdrawal, error)
	GetWithdrawalByNo(ctx context.Context, withdrawalNo string) (*model.Withdrawal, error)
	ListWithdrawals(ctx context.Context, query *WithdrawalQuery) ([]*model.Withdrawal, int64, error)
	UpdateWithdrawal(ctx context.Context, withdrawal *model.Withdrawal) error

	// 账单管理
	CreateInvoice(ctx context.Context, invoice *model.Invoice) error
	CreateInvoiceItem(ctx context.Context, item *model.InvoiceItem) error
	GetInvoiceByID(ctx context.Context, id uuid.UUID) (*model.Invoice, error)
	GetInvoiceByNo(ctx context.Context, invoiceNo string) (*model.Invoice, error)
	ListInvoices(ctx context.Context, query *InvoiceQuery) ([]*model.Invoice, int64, error)
	UpdateInvoice(ctx context.Context, invoice *model.Invoice) error
	GetInvoiceItems(ctx context.Context, invoiceID uuid.UUID) ([]*model.InvoiceItem, error)

	// 对账管理
	CreateReconciliation(ctx context.Context, reconciliation *model.Reconciliation) error
	CreateReconciliationItem(ctx context.Context, item *model.ReconciliationItem) error
	GetReconciliationByID(ctx context.Context, id uuid.UUID) (*model.Reconciliation, error)
	GetReconciliationByNo(ctx context.Context, reconciliationNo string) (*model.Reconciliation, error)
	ListReconciliations(ctx context.Context, query *ReconciliationQuery) ([]*model.Reconciliation, int64, error)
	UpdateReconciliation(ctx context.Context, reconciliation *model.Reconciliation) error
	GetReconciliationItems(ctx context.Context, reconciliationID uuid.UUID) ([]*model.ReconciliationItem, error)
	UpdateReconciliationItem(ctx context.Context, item *model.ReconciliationItem) error
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

// WithdrawalQuery 提现查询参数
type WithdrawalQuery struct {
	MerchantID *uuid.UUID
	AccountID  *uuid.UUID
	Status     string
	StartTime  *time.Time
	EndTime    *time.Time
	Page       int
	PageSize   int
}

// InvoiceQuery 账单查询参数
type InvoiceQuery struct {
	MerchantID  *uuid.UUID
	InvoiceType string
	Status      string
	StartTime   *time.Time
	EndTime     *time.Time
	Page        int
	PageSize    int
}

// ReconciliationQuery 对账查询参数
type ReconciliationQuery struct {
	MerchantID *uuid.UUID
	Channel    string
	Status     string
	StartDate  *time.Time
	EndDate    *time.Time
	Page       int
	PageSize   int
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

// CreateWithdrawal 创建提现记录
func (r *accountRepository) CreateWithdrawal(ctx context.Context, withdrawal *model.Withdrawal) error {
	return r.db.WithContext(ctx).Create(withdrawal).Error
}

// GetWithdrawalByID 根据ID获取提现记录
func (r *accountRepository) GetWithdrawalByID(ctx context.Context, id uuid.UUID) (*model.Withdrawal, error) {
	var withdrawal model.Withdrawal
	err := r.db.WithContext(ctx).First(&withdrawal, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &withdrawal, nil
}

// GetWithdrawalByNo 根据提现单号获取提现记录
func (r *accountRepository) GetWithdrawalByNo(ctx context.Context, withdrawalNo string) (*model.Withdrawal, error) {
	var withdrawal model.Withdrawal
	err := r.db.WithContext(ctx).First(&withdrawal, "withdrawal_no = ?", withdrawalNo).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &withdrawal, nil
}

// ListWithdrawals 提现列表
func (r *accountRepository) ListWithdrawals(ctx context.Context, query *WithdrawalQuery) ([]*model.Withdrawal, int64, error) {
	var withdrawals []*model.Withdrawal
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Withdrawal{})

	if query.MerchantID != nil {
		db = db.Where("merchant_id = ?", *query.MerchantID)
	}
	if query.AccountID != nil {
		db = db.Where("account_id = ?", *query.AccountID)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
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
	err := db.Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&withdrawals).Error
	return withdrawals, total, err
}

// UpdateWithdrawal 更新提现记录
func (r *accountRepository) UpdateWithdrawal(ctx context.Context, withdrawal *model.Withdrawal) error {
	return r.db.WithContext(ctx).Save(withdrawal).Error
}

// Invoice Management Methods

// CreateInvoice 创建账单
func (r *accountRepository) CreateInvoice(ctx context.Context, invoice *model.Invoice) error {
	return r.db.WithContext(ctx).Create(invoice).Error
}

// CreateInvoiceItem 创建账单明细
func (r *accountRepository) CreateInvoiceItem(ctx context.Context, item *model.InvoiceItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

// GetInvoiceByID 根据ID获取账单
func (r *accountRepository) GetInvoiceByID(ctx context.Context, id uuid.UUID) (*model.Invoice, error) {
	var invoice model.Invoice
	err := r.db.WithContext(ctx).Preload("Items").First(&invoice, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &invoice, nil
}

// GetInvoiceByNo 根据账单号获取账单
func (r *accountRepository) GetInvoiceByNo(ctx context.Context, invoiceNo string) (*model.Invoice, error) {
	var invoice model.Invoice
	err := r.db.WithContext(ctx).Preload("Items").First(&invoice, "invoice_no = ?", invoiceNo).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &invoice, nil
}

// ListInvoices 账单列表
func (r *accountRepository) ListInvoices(ctx context.Context, query *InvoiceQuery) ([]*model.Invoice, int64, error) {
	var invoices []*model.Invoice
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Invoice{})

	if query.MerchantID != nil {
		db = db.Where("merchant_id = ?", *query.MerchantID)
	}
	if query.InvoiceType != "" {
		db = db.Where("invoice_type = ?", query.InvoiceType)
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
	err := db.Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&invoices).Error
	return invoices, total, err
}

// UpdateInvoice 更新账单
func (r *accountRepository) UpdateInvoice(ctx context.Context, invoice *model.Invoice) error {
	return r.db.WithContext(ctx).Save(invoice).Error
}

// GetInvoiceItems 获取账单明细列表
func (r *accountRepository) GetInvoiceItems(ctx context.Context, invoiceID uuid.UUID) ([]*model.InvoiceItem, error) {
	var items []*model.InvoiceItem
	err := r.db.WithContext(ctx).Where("invoice_id = ?", invoiceID).Find(&items).Error
	return items, err
}

// Reconciliation Management Methods

// CreateReconciliation 创建对账单
func (r *accountRepository) CreateReconciliation(ctx context.Context, reconciliation *model.Reconciliation) error {
	return r.db.WithContext(ctx).Create(reconciliation).Error
}

// CreateReconciliationItem 创建对账明细
func (r *accountRepository) CreateReconciliationItem(ctx context.Context, item *model.ReconciliationItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

// GetReconciliationByID 根据ID获取对账单
func (r *accountRepository) GetReconciliationByID(ctx context.Context, id uuid.UUID) (*model.Reconciliation, error) {
	var reconciliation model.Reconciliation
	err := r.db.WithContext(ctx).Preload("Items").First(&reconciliation, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &reconciliation, nil
}

// GetReconciliationByNo 根据对账单号获取对账单
func (r *accountRepository) GetReconciliationByNo(ctx context.Context, reconciliationNo string) (*model.Reconciliation, error) {
	var reconciliation model.Reconciliation
	err := r.db.WithContext(ctx).Preload("Items").First(&reconciliation, "reconciliation_no = ?", reconciliationNo).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &reconciliation, nil
}

// ListReconciliations 对账单列表
func (r *accountRepository) ListReconciliations(ctx context.Context, query *ReconciliationQuery) ([]*model.Reconciliation, int64, error) {
	var reconciliations []*model.Reconciliation
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Reconciliation{})

	if query.MerchantID != nil {
		db = db.Where("merchant_id = ?", *query.MerchantID)
	}
	if query.Channel != "" {
		db = db.Where("channel = ?", query.Channel)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.StartDate != nil {
		db = db.Where("reconciliation_date >= ?", *query.StartDate)
	}
	if query.EndDate != nil {
		db = db.Where("reconciliation_date <= ?", *query.EndDate)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	err := db.Order("reconciliation_date DESC, created_at DESC").Offset(offset).Limit(query.PageSize).Find(&reconciliations).Error
	return reconciliations, total, err
}

// UpdateReconciliation 更新对账单
func (r *accountRepository) UpdateReconciliation(ctx context.Context, reconciliation *model.Reconciliation) error {
	return r.db.WithContext(ctx).Save(reconciliation).Error
}

// GetReconciliationItems 获取对账明细列表
func (r *accountRepository) GetReconciliationItems(ctx context.Context, reconciliationID uuid.UUID) ([]*model.ReconciliationItem, error) {
	var items []*model.ReconciliationItem
	err := r.db.WithContext(ctx).Where("reconciliation_id = ?", reconciliationID).Find(&items).Error
	return items, err
}

// UpdateReconciliationItem 更新对账明细
func (r *accountRepository) UpdateReconciliationItem(ctx context.Context, item *model.ReconciliationItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}
