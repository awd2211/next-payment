package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"payment-platform/merchant-policy-service/internal/model"
)

// LimitPolicyRepository 限额策略仓储接口
type LimitPolicyRepository interface {
	Create(ctx context.Context, policy *model.MerchantLimitPolicy) error
	Update(ctx context.Context, policy *model.MerchantLimitPolicy) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.MerchantLimitPolicy, error)

	// 查询商户自定义限额策略
	GetByMerchantID(ctx context.Context, merchantID uuid.UUID, channel, currency string) ([]*model.MerchantLimitPolicy, error)

	// 查询等级默认限额策略
	GetByTierID(ctx context.Context, tierID uuid.UUID, channel, currency string) ([]*model.MerchantLimitPolicy, error)

	// 获取有效的限额策略（按优先级排序）
	GetEffectivePolicies(ctx context.Context, merchantID *uuid.UUID, tierID *uuid.UUID, channel, currency string, now time.Time) ([]*model.MerchantLimitPolicy, error)

	// 列表查询
	List(ctx context.Context, merchantID *uuid.UUID, tierID *uuid.UUID, status string, offset, limit int) ([]*model.MerchantLimitPolicy, int64, error)
}

type limitPolicyRepository struct {
	db *gorm.DB
}

// NewLimitPolicyRepository 创建限额策略仓储实例
func NewLimitPolicyRepository(db *gorm.DB) LimitPolicyRepository {
	return &limitPolicyRepository{db: db}
}

func (r *limitPolicyRepository) Create(ctx context.Context, policy *model.MerchantLimitPolicy) error {
	return r.db.WithContext(ctx).Create(policy).Error
}

func (r *limitPolicyRepository) Update(ctx context.Context, policy *model.MerchantLimitPolicy) error {
	return r.db.WithContext(ctx).Save(policy).Error
}

func (r *limitPolicyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.MerchantLimitPolicy{}, "id = ?", id).Error
}

func (r *limitPolicyRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.MerchantLimitPolicy, error) {
	var policy model.MerchantLimitPolicy
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&policy).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &policy, nil
}

func (r *limitPolicyRepository) GetByMerchantID(ctx context.Context, merchantID uuid.UUID, channel, currency string) ([]*model.MerchantLimitPolicy, error) {
	query := r.db.WithContext(ctx).
		Where("merchant_id = ?", merchantID).
		Where("status = ?", model.LimitStatusActive)

	if channel != "" {
		query = query.Where("(channel = ? OR channel = ?)", channel, "all")
	}
	if currency != "" {
		query = query.Where("currency = ?", currency)
	}

	var policies []*model.MerchantLimitPolicy
	err := query.Order("priority DESC").Find(&policies).Error
	if err != nil {
		return nil, err
	}
	return policies, nil
}

func (r *limitPolicyRepository) GetByTierID(ctx context.Context, tierID uuid.UUID, channel, currency string) ([]*model.MerchantLimitPolicy, error) {
	query := r.db.WithContext(ctx).
		Where("tier_id = ?", tierID).
		Where("merchant_id IS NULL").
		Where("status = ?", model.LimitStatusActive)

	if channel != "" {
		query = query.Where("(channel = ? OR channel = ?)", channel, "all")
	}
	if currency != "" {
		query = query.Where("currency = ?", currency)
	}

	var policies []*model.MerchantLimitPolicy
	err := query.Order("priority DESC").Find(&policies).Error
	if err != nil {
		return nil, err
	}
	return policies, nil
}

func (r *limitPolicyRepository) GetEffectivePolicies(ctx context.Context, merchantID *uuid.UUID, tierID *uuid.UUID, channel, currency string, now time.Time) ([]*model.MerchantLimitPolicy, error) {
	query := r.db.WithContext(ctx).Where("status = ?", model.LimitStatusActive)

	// 时间范围过滤
	query = query.Where("effective_date <= ?", now).
		Where("(expiry_date IS NULL OR expiry_date > ?)", now)

	// 商户或等级过滤
	if merchantID != nil {
		query = query.Where("merchant_id = ?", *merchantID)
	} else if tierID != nil {
		query = query.Where("tier_id = ? AND merchant_id IS NULL", *tierID)
	}

	// 渠道、币种过滤
	if channel != "" {
		query = query.Where("(channel = ? OR channel = ?)", channel, "all")
	}
	if currency != "" {
		query = query.Where("currency = ?", currency)
	}

	var policies []*model.MerchantLimitPolicy
	err := query.Order("priority DESC, created_at DESC").Find(&policies).Error
	if err != nil {
		return nil, err
	}
	return policies, nil
}

func (r *limitPolicyRepository) List(ctx context.Context, merchantID *uuid.UUID, tierID *uuid.UUID, status string, offset, limit int) ([]*model.MerchantLimitPolicy, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.MerchantLimitPolicy{})

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

	var policies []*model.MerchantLimitPolicy
	err := query.Order("priority DESC, created_at DESC").
		Offset(offset).Limit(limit).
		Find(&policies).Error
	if err != nil {
		return nil, 0, err
	}

	return policies, total, nil
}
