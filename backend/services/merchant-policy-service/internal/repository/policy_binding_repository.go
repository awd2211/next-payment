package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"payment-platform/merchant-policy-service/internal/model"
)

// PolicyBindingRepository 策略绑定仓储接口
type PolicyBindingRepository interface {
	Create(ctx context.Context, binding *model.MerchantPolicyBinding) error
	Update(ctx context.Context, binding *model.MerchantPolicyBinding) error
	GetByMerchantID(ctx context.Context, merchantID uuid.UUID) (*model.MerchantPolicyBinding, error)
	GetByMerchantIDWithTier(ctx context.Context, merchantID uuid.UUID) (*model.MerchantPolicyBinding, *model.MerchantTier, error)
	ListByTierID(ctx context.Context, tierID uuid.UUID, offset, limit int) ([]*model.MerchantPolicyBinding, int64, error)
	Delete(ctx context.Context, merchantID uuid.UUID) error
}

type policyBindingRepository struct {
	db *gorm.DB
}

// NewPolicyBindingRepository 创建策略绑定仓储实例
func NewPolicyBindingRepository(db *gorm.DB) PolicyBindingRepository {
	return &policyBindingRepository{db: db}
}

func (r *policyBindingRepository) Create(ctx context.Context, binding *model.MerchantPolicyBinding) error {
	return r.db.WithContext(ctx).Create(binding).Error
}

func (r *policyBindingRepository) Update(ctx context.Context, binding *model.MerchantPolicyBinding) error {
	return r.db.WithContext(ctx).Save(binding).Error
}

func (r *policyBindingRepository) GetByMerchantID(ctx context.Context, merchantID uuid.UUID) (*model.MerchantPolicyBinding, error) {
	var binding model.MerchantPolicyBinding
	err := r.db.WithContext(ctx).
		Where("merchant_id = ?", merchantID).
		First(&binding).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &binding, nil
}

func (r *policyBindingRepository) GetByMerchantIDWithTier(ctx context.Context, merchantID uuid.UUID) (*model.MerchantPolicyBinding, *model.MerchantTier, error) {
	var binding model.MerchantPolicyBinding
	var tier model.MerchantTier

	err := r.db.WithContext(ctx).
		Where("merchant_id = ?", merchantID).
		First(&binding).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	err = r.db.WithContext(ctx).
		Where("id = ?", binding.TierID).
		First(&tier).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &binding, nil, nil
		}
		return &binding, nil, err
	}

	return &binding, &tier, nil
}

func (r *policyBindingRepository) ListByTierID(ctx context.Context, tierID uuid.UUID, offset, limit int) ([]*model.MerchantPolicyBinding, int64, error) {
	var bindings []*model.MerchantPolicyBinding
	var total int64

	query := r.db.WithContext(ctx).Model(&model.MerchantPolicyBinding{}).
		Where("tier_id = ?", tierID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset(offset).Limit(limit).Find(&bindings).Error
	if err != nil {
		return nil, 0, err
	}

	return bindings, total, nil
}

func (r *policyBindingRepository) Delete(ctx context.Context, merchantID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Delete(&model.MerchantPolicyBinding{}, "merchant_id = ?", merchantID).Error
}
