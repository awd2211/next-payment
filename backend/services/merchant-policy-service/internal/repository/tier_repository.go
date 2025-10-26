package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"payment-platform/merchant-policy-service/internal/model"
)

// TierRepository 商户等级仓储接口
type TierRepository interface {
	Create(ctx context.Context, tier *model.MerchantTier) error
	Update(ctx context.Context, tier *model.MerchantTier) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.MerchantTier, error)
	GetByCode(ctx context.Context, tierCode string) (*model.MerchantTier, error)
	List(ctx context.Context, isActive *bool, offset, limit int) ([]*model.MerchantTier, int64, error)
	GetAllActive(ctx context.Context) ([]*model.MerchantTier, error)
}

type tierRepository struct {
	db *gorm.DB
}

// NewTierRepository 创建等级仓储实例
func NewTierRepository(db *gorm.DB) TierRepository {
	return &tierRepository{db: db}
}

func (r *tierRepository) Create(ctx context.Context, tier *model.MerchantTier) error {
	return r.db.WithContext(ctx).Create(tier).Error
}

func (r *tierRepository) Update(ctx context.Context, tier *model.MerchantTier) error {
	return r.db.WithContext(ctx).Save(tier).Error
}

func (r *tierRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.MerchantTier{}, "id = ?", id).Error
}

func (r *tierRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.MerchantTier, error) {
	var tier model.MerchantTier
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&tier).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &tier, nil
}

func (r *tierRepository) GetByCode(ctx context.Context, tierCode string) (*model.MerchantTier, error) {
	var tier model.MerchantTier
	err := r.db.WithContext(ctx).Where("tier_code = ?", tierCode).First(&tier).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &tier, nil
}

func (r *tierRepository) List(ctx context.Context, isActive *bool, offset, limit int) ([]*model.MerchantTier, int64, error) {
	var tiers []*model.MerchantTier
	var total int64

	query := r.db.WithContext(ctx).Model(&model.MerchantTier{})

	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("tier_level ASC").Offset(offset).Limit(limit).Find(&tiers).Error
	if err != nil {
		return nil, 0, err
	}

	return tiers, total, nil
}

func (r *tierRepository) GetAllActive(ctx context.Context) ([]*model.MerchantTier, error) {
	var tiers []*model.MerchantTier
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("tier_level ASC").
		Find(&tiers).Error
	if err != nil {
		return nil, err
	}
	return tiers, nil
}
