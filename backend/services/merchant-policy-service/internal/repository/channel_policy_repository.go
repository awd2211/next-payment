package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"payment-platform/merchant-policy-service/internal/model"
)

// ChannelPolicyRepository 渠道策略仓储接口
type ChannelPolicyRepository interface {
	Create(ctx context.Context, policy *model.ChannelPolicy) error
	Update(ctx context.Context, policy *model.ChannelPolicy) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.ChannelPolicy, error)

	// 查询商户的所有渠道配置
	GetByMerchantID(ctx context.Context, merchantID uuid.UUID, isEnabled *bool) ([]*model.ChannelPolicy, error)

	// 查询商户的特定渠道配置
	GetByMerchantAndChannel(ctx context.Context, merchantID uuid.UUID, channel string) (*model.ChannelPolicy, error)

	// 获取有效的渠道策略（时间范围内）
	GetEffectivePolicies(ctx context.Context, merchantID uuid.UUID, now time.Time) ([]*model.ChannelPolicy, error)

	// 批量启用/禁用渠道
	BatchUpdateStatus(ctx context.Context, merchantID uuid.UUID, channels []string, isEnabled bool) error

	// 列表查询
	List(ctx context.Context, merchantID *uuid.UUID, channel string, isEnabled *bool, offset, limit int) ([]*model.ChannelPolicy, int64, error)
}

type channelPolicyRepository struct {
	db *gorm.DB
}

// NewChannelPolicyRepository 创建渠道策略仓储实例
func NewChannelPolicyRepository(db *gorm.DB) ChannelPolicyRepository {
	return &channelPolicyRepository{db: db}
}

func (r *channelPolicyRepository) Create(ctx context.Context, policy *model.ChannelPolicy) error {
	return r.db.WithContext(ctx).Create(policy).Error
}

func (r *channelPolicyRepository) Update(ctx context.Context, policy *model.ChannelPolicy) error {
	return r.db.WithContext(ctx).Save(policy).Error
}

func (r *channelPolicyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.ChannelPolicy{}, "id = ?", id).Error
}

func (r *channelPolicyRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.ChannelPolicy, error) {
	var policy model.ChannelPolicy
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&policy).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &policy, nil
}

func (r *channelPolicyRepository) GetByMerchantID(ctx context.Context, merchantID uuid.UUID, isEnabled *bool) ([]*model.ChannelPolicy, error) {
	query := r.db.WithContext(ctx).Where("merchant_id = ?", merchantID)

	if isEnabled != nil {
		query = query.Where("is_enabled = ?", *isEnabled)
	}

	var policies []*model.ChannelPolicy
	err := query.Order("priority DESC, created_at DESC").Find(&policies).Error
	if err != nil {
		return nil, err
	}
	return policies, nil
}

func (r *channelPolicyRepository) GetByMerchantAndChannel(ctx context.Context, merchantID uuid.UUID, channel string) (*model.ChannelPolicy, error) {
	var policy model.ChannelPolicy
	err := r.db.WithContext(ctx).
		Where("merchant_id = ? AND channel = ?", merchantID, channel).
		First(&policy).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &policy, nil
}

func (r *channelPolicyRepository) GetEffectivePolicies(ctx context.Context, merchantID uuid.UUID, now time.Time) ([]*model.ChannelPolicy, error) {
	var policies []*model.ChannelPolicy
	err := r.db.WithContext(ctx).
		Where("merchant_id = ?", merchantID).
		Where("is_enabled = ?", true).
		Where("effective_date <= ?", now).
		Where("(expiry_date IS NULL OR expiry_date > ?)", now).
		Order("priority DESC").
		Find(&policies).Error
	if err != nil {
		return nil, err
	}
	return policies, nil
}

func (r *channelPolicyRepository) BatchUpdateStatus(ctx context.Context, merchantID uuid.UUID, channels []string, isEnabled bool) error {
	return r.db.WithContext(ctx).
		Model(&model.ChannelPolicy{}).
		Where("merchant_id = ? AND channel IN ?", merchantID, channels).
		Update("is_enabled", isEnabled).Error
}

func (r *channelPolicyRepository) List(ctx context.Context, merchantID *uuid.UUID, channel string, isEnabled *bool, offset, limit int) ([]*model.ChannelPolicy, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.ChannelPolicy{})

	if merchantID != nil {
		query = query.Where("merchant_id = ?", *merchantID)
	}
	if channel != "" {
		query = query.Where("channel = ?", channel)
	}
	if isEnabled != nil {
		query = query.Where("is_enabled = ?", *isEnabled)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var policies []*model.ChannelPolicy
	err := query.Order("priority DESC, created_at DESC").
		Offset(offset).Limit(limit).
		Find(&policies).Error
	if err != nil {
		return nil, 0, err
	}

	return policies, total, nil
}
