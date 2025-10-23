package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"payment-platform/merchant-service/internal/model"
)

// SettlementAccountRepository 结算账户仓储接口
type SettlementAccountRepository interface {
	Create(ctx context.Context, account *model.SettlementAccount) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.SettlementAccount, error)
	GetByMerchantID(ctx context.Context, merchantID uuid.UUID) ([]*model.SettlementAccount, error)
	GetDefaultByMerchant(ctx context.Context, merchantID uuid.UUID) (*model.SettlementAccount, error)
	Update(ctx context.Context, account *model.SettlementAccount) error
	Delete(ctx context.Context, id uuid.UUID) error
	SetDefault(ctx context.Context, merchantID, accountID uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
}

type settlementAccountRepository struct {
	db *gorm.DB
}

// NewSettlementAccountRepository 创建结算账户仓储实例
func NewSettlementAccountRepository(db *gorm.DB) SettlementAccountRepository {
	return &settlementAccountRepository{db: db}
}

// Create 创建结算账户
func (r *settlementAccountRepository) Create(ctx context.Context, account *model.SettlementAccount) error {
	return r.db.WithContext(ctx).Create(account).Error
}

// GetByID 根据ID获取结算账户
func (r *settlementAccountRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.SettlementAccount, error) {
	var account model.SettlementAccount
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// GetByMerchantID 获取商户的所有结算账户
func (r *settlementAccountRepository) GetByMerchantID(ctx context.Context, merchantID uuid.UUID) ([]*model.SettlementAccount, error) {
	var accounts []*model.SettlementAccount
	err := r.db.WithContext(ctx).
		Where("merchant_id = ?", merchantID).
		Order("is_default DESC, created_at DESC").
		Find(&accounts).Error
	return accounts, err
}

// GetDefaultByMerchant 获取商户的默认结算账户
func (r *settlementAccountRepository) GetDefaultByMerchant(ctx context.Context, merchantID uuid.UUID) (*model.SettlementAccount, error) {
	var account model.SettlementAccount
	err := r.db.WithContext(ctx).
		Where("merchant_id = ? AND is_default = ?", merchantID, true).
		First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// Update 更新结算账户
func (r *settlementAccountRepository) Update(ctx context.Context, account *model.SettlementAccount) error {
	return r.db.WithContext(ctx).Save(account).Error
}

// Delete 删除结算账户（软删除）
func (r *settlementAccountRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.SettlementAccount{}, "id = ?", id).Error
}

// SetDefault 设置默认结算账户
func (r *settlementAccountRepository) SetDefault(ctx context.Context, merchantID, accountID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 取消该商户所有账户的默认状态
		if err := tx.Model(&model.SettlementAccount{}).
			Where("merchant_id = ?", merchantID).
			Update("is_default", false).Error; err != nil {
			return err
		}

		// 设置指定账户为默认
		return tx.Model(&model.SettlementAccount{}).
			Where("id = ? AND merchant_id = ?", accountID, merchantID).
			Update("is_default", true).Error
	})
}

// UpdateStatus 更新结算账户状态
func (r *settlementAccountRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.db.WithContext(ctx).Model(&model.SettlementAccount{}).
		Where("id = ?", id).
		Update("status", status).Error
}
