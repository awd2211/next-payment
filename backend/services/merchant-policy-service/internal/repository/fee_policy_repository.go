package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"payment-platform/merchant-policy-service/internal/model"
)

// FeePolicyRepository 费率策略仓储接口
type FeePolicyRepository interface {
	Create(ctx context.Context, policy *model.MerchantFeePolicy) error
	Update(ctx context.Context, policy *model.MerchantFeePolicy) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.MerchantFeePolicy, error)

	// 查询商户自定义费率策略
	GetByMerchantID(ctx context.Context, merchantID uuid.UUID, channel, paymentMethod, currency string) ([]*model.MerchantFeePolicy, error)

	// 查询等级默认费率策略
	GetByTierID(ctx context.Context, tierID uuid.UUID, channel, paymentMethod, currency string) ([]*model.MerchantFeePolicy, error)

	// 获取有效的费率策略（按优先级排序）
	GetEffectivePolicies(ctx context.Context, merchantID *uuid.UUID, tierID *uuid.UUID, channel, paymentMethod, currency string, now time.Time) ([]*model.MerchantFeePolicy, error)

	// 列表查询
	List(ctx context.Context, merchantID *uuid.UUID, tierID *uuid.UUID, status string, offset, limit int) ([]*model.MerchantFeePolicy, int64, error)
}

type feePolicyRepository struct {
	db *gorm.DB
}

// NewFeePolicyRepository 创建费率策略仓储实例
func NewFeePolicyRepository(db *gorm.DB) FeePolicyRepository {
	return &feePolicyRepository{db: db}
}

func (r *feePolicyRepository) Create(ctx context.Context, policy *model.MerchantFeePolicy) error {
	return r.db.WithContext(ctx).Create(policy).Error
}

func (r *feePolicyRepository) Update(ctx context.Context, policy *model.MerchantFeePolicy) error {
	return r.db.WithContext(ctx).Save(policy).Error
}

func (r *feePolicyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.MerchantFeePolicy{}, "id = ?", id).Error
}

func (r *feePolicyRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.MerchantFeePolicy, error) {
	var policy model.MerchantFeePolicy
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&policy).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &policy, nil
}

func (r *feePolicyRepository) GetByMerchantID(ctx context.Context, merchantID uuid.UUID, channel, paymentMethod, currency string) ([]*model.MerchantFeePolicy, error) {
	query := r.db.WithContext(ctx).
		Where("merchant_id = ?", merchantID).
		Where("status = ?", model.FeeStatusActive)

	if channel != "" {
		query = query.Where("(channel = ? OR channel = ?)", channel, model.ChannelAll)
	}
	if paymentMethod != "" {
		query = query.Where("(payment_method = ? OR payment_method = ?)", paymentMethod, model.PaymentMethodAll)
	}
	if currency != "" {
		query = query.Where("currency = ?", currency)
	}

	var policies []*model.MerchantFeePolicy
	err := query.Order("priority DESC").Find(&policies).Error
	if err != nil {
		return nil, err
	}
	return policies, nil
}

func (r *feePolicyRepository) GetByTierID(ctx context.Context, tierID uuid.UUID, channel, paymentMethod, currency string) ([]*model.MerchantFeePolicy, error) {
	query := r.db.WithContext(ctx).
		Where("tier_id = ?", tierID).
		Where("merchant_id IS NULL").
		Where("status = ?", model.FeeStatusActive)

	if channel != "" {
		query = query.Where("(channel = ? OR channel = ?)", channel, model.ChannelAll)
	}
	if paymentMethod != "" {
		query = query.Where("(payment_method = ? OR payment_method = ?)", paymentMethod, model.PaymentMethodAll)
	}
	if currency != "" {
		query = query.Where("currency = ?", currency)
	}

	var policies []*model.MerchantFeePolicy
	err := query.Order("priority DESC").Find(&policies).Error
	if err != nil {
		return nil, err
	}
	return policies, nil
}

func (r *feePolicyRepository) GetEffectivePolicies(ctx context.Context, merchantID *uuid.UUID, tierID *uuid.UUID, channel, paymentMethod, currency string, now time.Time) ([]*model.MerchantFeePolicy, error) {
	query := r.db.WithContext(ctx).Where("status = ?", model.FeeStatusActive)

	// 时间范围过滤
	query = query.Where("effective_date <= ?", now).
		Where("(expiry_date IS NULL OR expiry_date > ?)", now)

	// 商户或等级过滤
	if merchantID != nil {
		query = query.Where("merchant_id = ?", *merchantID)
	} else if tierID != nil {
		query = query.Where("tier_id = ? AND merchant_id IS NULL", *tierID)
	}

	// 渠道、支付方式、币种过滤
	if channel != "" {
		query = query.Where("(channel = ? OR channel = ?)", channel, model.ChannelAll)
	}
	if paymentMethod != "" {
		query = query.Where("(payment_method = ? OR payment_method = ?)", paymentMethod, model.PaymentMethodAll)
	}
	if currency != "" {
		query = query.Where("currency = ?", currency)
	}

	var policies []*model.MerchantFeePolicy
	err := query.Order("priority DESC, created_at DESC").Find(&policies).Error
	if err != nil {
		return nil, err
	}
	return policies, nil
}

func (r *feePolicyRepository) List(ctx context.Context, merchantID *uuid.UUID, tierID *uuid.UUID, status string, offset, limit int) ([]*model.MerchantFeePolicy, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.MerchantFeePolicy{})

	if merchantID != nil {
		query = query.Where("merchant_id = ?", *merchantID)
	}
	if tierID != nil {
		query = query.Where("tier_id = ?", *tierID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var policies []*model.MerchantFeePolicy
	err := query.Order("priority DESC, created_at DESC").
		Offset(offset).Limit(limit).
		Find(&policies).Error
	if err != nil {
		return nil, 0, err
	}

	return policies, total, nil
}
