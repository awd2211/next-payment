package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"payment-platform/withdrawal-service/internal/model"
)

// WithdrawalRepository 提现仓储接口
type WithdrawalRepository interface {
	Create(ctx context.Context, withdrawal *model.Withdrawal) error
	Update(ctx context.Context, withdrawal *model.Withdrawal) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Withdrawal, error)
	GetByWithdrawalNo(ctx context.Context, withdrawalNo string) (*model.Withdrawal, error)
	List(ctx context.Context, query *WithdrawalQuery) ([]*model.Withdrawal, int64, error)
	CreateApproval(ctx context.Context, approval *model.WithdrawalApproval) error
	GetApprovals(ctx context.Context, withdrawalID uuid.UUID) ([]*model.WithdrawalApproval, error)
	GetPendingWithdrawals(ctx context.Context, merchantID uuid.UUID) ([]*model.Withdrawal, error)

	// Bank Account operations
	CreateBankAccount(ctx context.Context, account *model.WithdrawalBankAccount) error
	UpdateBankAccount(ctx context.Context, account *model.WithdrawalBankAccount) error
	GetBankAccountByID(ctx context.Context, id uuid.UUID) (*model.WithdrawalBankAccount, error)
	ListBankAccounts(ctx context.Context, merchantID uuid.UUID) ([]*model.WithdrawalBankAccount, error)
	GetDefaultBankAccount(ctx context.Context, merchantID uuid.UUID) (*model.WithdrawalBankAccount, error)

	// Batch operations
	CreateBatch(ctx context.Context, batch *model.WithdrawalBatch) error
	UpdateBatch(ctx context.Context, batch *model.WithdrawalBatch) error
	GetBatchByID(ctx context.Context, id uuid.UUID) (*model.WithdrawalBatch, error)
}

// WithdrawalQuery 查询条件
type WithdrawalQuery struct {
	MerchantID    *uuid.UUID
	Status        *model.WithdrawalStatus
	Type          *model.WithdrawalType
	StartDate     *time.Time
	EndDate       *time.Time
	MinAmount     *int64
	MaxAmount     *int64
	Page          int
	PageSize      int
}

type withdrawalRepository struct {
	db *gorm.DB
}

// NewWithdrawalRepository 创建提现仓储
func NewWithdrawalRepository(db *gorm.DB) WithdrawalRepository {
	return &withdrawalRepository{
		db: db,
	}
}

func (r *withdrawalRepository) Create(ctx context.Context, withdrawal *model.Withdrawal) error {
	return r.db.WithContext(ctx).Create(withdrawal).Error
}

func (r *withdrawalRepository) Update(ctx context.Context, withdrawal *model.Withdrawal) error {
	return r.db.WithContext(ctx).Save(withdrawal).Error
}

func (r *withdrawalRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Withdrawal, error) {
	var withdrawal model.Withdrawal
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&withdrawal).Error
	if err != nil {
		return nil, err
	}
	return &withdrawal, nil
}

func (r *withdrawalRepository) GetByWithdrawalNo(ctx context.Context, withdrawalNo string) (*model.Withdrawal, error) {
	var withdrawal model.Withdrawal
	err := r.db.WithContext(ctx).Where("withdrawal_no = ?", withdrawalNo).First(&withdrawal).Error
	if err != nil {
		return nil, err
	}
	return &withdrawal, nil
}

func (r *withdrawalRepository) List(ctx context.Context, query *WithdrawalQuery) ([]*model.Withdrawal, int64, error) {
	var withdrawals []*model.Withdrawal
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Withdrawal{})

	// Apply filters
	if query.MerchantID != nil {
		db = db.Where("merchant_id = ?", *query.MerchantID)
	}
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}
	if query.Type != nil {
		db = db.Where("type = ?", *query.Type)
	}
	if query.StartDate != nil {
		db = db.Where("created_at >= ?", *query.StartDate)
	}
	if query.EndDate != nil {
		db = db.Where("created_at <= ?", *query.EndDate)
	}
	if query.MinAmount != nil {
		db = db.Where("amount >= ?", *query.MinAmount)
	}
	if query.MaxAmount != nil {
		db = db.Where("amount <= ?", *query.MaxAmount)
	}

	// Count total
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Paginate
	offset := (query.Page - 1) * query.PageSize
	err := db.Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&withdrawals).Error
	if err != nil {
		return nil, 0, err
	}

	return withdrawals, total, nil
}

func (r *withdrawalRepository) CreateApproval(ctx context.Context, approval *model.WithdrawalApproval) error {
	return r.db.WithContext(ctx).Create(approval).Error
}

func (r *withdrawalRepository) GetApprovals(ctx context.Context, withdrawalID uuid.UUID) ([]*model.WithdrawalApproval, error) {
	var approvals []*model.WithdrawalApproval
	err := r.db.WithContext(ctx).Where("withdrawal_id = ?", withdrawalID).Order("level ASC, created_at ASC").Find(&approvals).Error
	if err != nil {
		return nil, err
	}
	return approvals, nil
}

func (r *withdrawalRepository) GetPendingWithdrawals(ctx context.Context, merchantID uuid.UUID) ([]*model.Withdrawal, error) {
	var withdrawals []*model.Withdrawal
	err := r.db.WithContext(ctx).
		Where("merchant_id = ? AND status = ?", merchantID, model.WithdrawalStatusPending).
		Order("created_at ASC").
		Find(&withdrawals).Error
	if err != nil {
		return nil, err
	}
	return withdrawals, nil
}

// Bank Account operations
func (r *withdrawalRepository) CreateBankAccount(ctx context.Context, account *model.WithdrawalBankAccount) error {
	return r.db.WithContext(ctx).Create(account).Error
}

func (r *withdrawalRepository) UpdateBankAccount(ctx context.Context, account *model.WithdrawalBankAccount) error {
	return r.db.WithContext(ctx).Save(account).Error
}

func (r *withdrawalRepository) GetBankAccountByID(ctx context.Context, id uuid.UUID) (*model.WithdrawalBankAccount, error) {
	var account model.WithdrawalBankAccount
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *withdrawalRepository) ListBankAccounts(ctx context.Context, merchantID uuid.UUID) ([]*model.WithdrawalBankAccount, error) {
	var accounts []*model.WithdrawalBankAccount
	err := r.db.WithContext(ctx).
		Where("merchant_id = ? AND status = ?", merchantID, "active").
		Order("is_default DESC, created_at DESC").
		Find(&accounts).Error
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

func (r *withdrawalRepository) GetDefaultBankAccount(ctx context.Context, merchantID uuid.UUID) (*model.WithdrawalBankAccount, error) {
	var account model.WithdrawalBankAccount
	err := r.db.WithContext(ctx).
		Where("merchant_id = ? AND is_default = ? AND status = ?", merchantID, true, "active").
		First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// Batch operations
func (r *withdrawalRepository) CreateBatch(ctx context.Context, batch *model.WithdrawalBatch) error {
	return r.db.WithContext(ctx).Create(batch).Error
}

func (r *withdrawalRepository) UpdateBatch(ctx context.Context, batch *model.WithdrawalBatch) error {
	return r.db.WithContext(ctx).Save(batch).Error
}

func (r *withdrawalRepository) GetBatchByID(ctx context.Context, id uuid.UUID) (*model.WithdrawalBatch, error) {
	var batch model.WithdrawalBatch
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&batch).Error
	if err != nil {
		return nil, err
	}
	return &batch, nil
}
