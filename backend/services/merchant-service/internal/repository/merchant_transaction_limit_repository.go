package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"payment-platform/merchant-service/internal/model"
)

// MerchantTransactionLimitRepository 商户交易限额仓储接口
type MerchantTransactionLimitRepository interface {
	Create(ctx context.Context, limit *model.MerchantTransactionLimit) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.MerchantTransactionLimit, error)
	GetByMerchantID(ctx context.Context, merchantID uuid.UUID) ([]*model.MerchantTransactionLimit, error)
	GetActiveByType(ctx context.Context, merchantID uuid.UUID, limitType, paymentMethod, channel string) (*model.MerchantTransactionLimit, error)
	Update(ctx context.Context, limit *model.MerchantTransactionLimit) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type merchantTransactionLimitRepository struct {
	db *gorm.DB
}

// NewMerchantTransactionLimitRepository 创建商户交易限额仓储实例
func NewMerchantTransactionLimitRepository(db *gorm.DB) MerchantTransactionLimitRepository {
	return &merchantTransactionLimitRepository{db: db}
}

func (r *merchantTransactionLimitRepository) Create(ctx context.Context, limit *model.MerchantTransactionLimit) error {
	return r.db.WithContext(ctx).Create(limit).Error
}

func (r *merchantTransactionLimitRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.MerchantTransactionLimit, error) {
	var limit model.MerchantTransactionLimit
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&limit).Error
	if err != nil {
		return nil, err
	}
	return &limit, nil
}

func (r *merchantTransactionLimitRepository) GetByMerchantID(ctx context.Context, merchantID uuid.UUID) ([]*model.MerchantTransactionLimit, error) {
	var limits []*model.MerchantTransactionLimit
	err := r.db.WithContext(ctx).
		Where("merchant_id = ?", merchantID).
		Order("created_at DESC").
		Find(&limits).Error
	return limits, err
}

func (r *merchantTransactionLimitRepository) GetActiveByType(ctx context.Context, merchantID uuid.UUID, limitType, paymentMethod, channel string) (*model.MerchantTransactionLimit, error) {
	var limit model.MerchantTransactionLimit
	query := r.db.WithContext(ctx).
		Where("merchant_id = ? AND limit_type = ? AND status = ?", merchantID, limitType, "active").
		Where("effective_date <= NOW()").
		Where("(expiry_date IS NULL OR expiry_date > NOW())")

	if paymentMethod != "" {
		query = query.Where("(payment_method = ? OR payment_method = 'all')", paymentMethod)
	}

	if channel != "" {
		query = query.Where("(channel = ? OR channel = 'all')", channel)
	}

	err := query.First(&limit).Error
	if err != nil {
		return nil, err
	}
	return &limit, nil
}

func (r *merchantTransactionLimitRepository) Update(ctx context.Context, limit *model.MerchantTransactionLimit) error {
	return r.db.WithContext(ctx).Save(limit).Error
}

func (r *merchantTransactionLimitRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.MerchantTransactionLimit{}, "id = ?", id).Error
}
