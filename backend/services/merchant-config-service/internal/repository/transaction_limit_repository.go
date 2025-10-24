package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"payment-platform/merchant-config-service/internal/model"
)

// TransactionLimitRepository 交易限额仓储接口
type TransactionLimitRepository interface {
	Create(ctx context.Context, limit *model.MerchantTransactionLimit) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.MerchantTransactionLimit, error)
	GetByMerchantID(ctx context.Context, merchantID uuid.UUID) ([]*model.MerchantTransactionLimit, error)
	GetEffectiveLimits(ctx context.Context, merchantID uuid.UUID, limitType, channel, paymentMethod string, queryTime time.Time) ([]*model.MerchantTransactionLimit, error)
	Update(ctx context.Context, limit *model.MerchantTransactionLimit) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, status string, limit, offset int) ([]*model.MerchantTransactionLimit, int64, error)
}

type transactionLimitRepository struct {
	db *gorm.DB
}

// NewTransactionLimitRepository 创建交易限额仓储实例
func NewTransactionLimitRepository(db *gorm.DB) TransactionLimitRepository {
	return &transactionLimitRepository{db: db}
}

// Create 创建交易限额
func (r *transactionLimitRepository) Create(ctx context.Context, limit *model.MerchantTransactionLimit) error {
	return r.db.WithContext(ctx).Create(limit).Error
}

// GetByID 根据ID获取交易限额
func (r *transactionLimitRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.MerchantTransactionLimit, error) {
	var limit model.MerchantTransactionLimit
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&limit).Error
	if err != nil {
		return nil, err
	}
	return &limit, nil
}

// GetByMerchantID 获取商户的所有交易限额
func (r *transactionLimitRepository) GetByMerchantID(ctx context.Context, merchantID uuid.UUID) ([]*model.MerchantTransactionLimit, error) {
	var limits []*model.MerchantTransactionLimit
	err := r.db.WithContext(ctx).
		Where("merchant_id = ?", merchantID).
		Order("limit_type, created_at DESC").
		Find(&limits).Error
	return limits, err
}

// GetEffectiveLimits 获取生效的交易限额（根据类型、渠道、支付方式、时间查询）
func (r *transactionLimitRepository) GetEffectiveLimits(ctx context.Context, merchantID uuid.UUID, limitType, channel, paymentMethod string, queryTime time.Time) ([]*model.MerchantTransactionLimit, error) {
	var limits []*model.MerchantTransactionLimit

	// 查询条件：
	// 1. merchant_id 匹配
	// 2. limit_type 匹配
	// 3. status = 'active'
	// 4. effective_date <= queryTime
	// 5. (expiry_date IS NULL OR expiry_date > queryTime)
	// 6. channel 匹配（精确匹配或'all'）
	// 7. payment_method 匹配（精确匹配或'all'）

	query := r.db.WithContext(ctx).
		Where("merchant_id = ?", merchantID).
		Where("limit_type = ?", limitType).
		Where("status = ?", model.LimitStatusActive).
		Where("effective_date <= ?", queryTime).
		Where("(expiry_date IS NULL OR expiry_date > ?)", queryTime)

	if channel != "" {
		query = query.Where("(channel = ? OR channel = ?)", channel, model.ChannelAll)
	}

	if paymentMethod != "" {
		query = query.Where("(payment_method = ? OR payment_method = ?)", paymentMethod, model.PaymentMethodAll)
	}

	err := query.Find(&limits).Error
	return limits, err
}

// Update 更新交易限额
func (r *transactionLimitRepository) Update(ctx context.Context, limit *model.MerchantTransactionLimit) error {
	return r.db.WithContext(ctx).Save(limit).Error
}

// Delete 删除交易限额（软删除）
func (r *transactionLimitRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.MerchantTransactionLimit{}, "id = ?", id).Error
}

// List 列出交易限额（管理员功能）
func (r *transactionLimitRepository) List(ctx context.Context, status string, limit, offset int) ([]*model.MerchantTransactionLimit, int64, error) {
	var limits []*model.MerchantTransactionLimit
	var total int64

	query := r.db.WithContext(ctx).Model(&model.MerchantTransactionLimit{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&limits).Error
	return limits, total, err
}
