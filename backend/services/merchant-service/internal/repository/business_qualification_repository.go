package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"payment-platform/merchant-service/internal/model"
)

// BusinessQualificationRepository 业务资质仓储接口
type BusinessQualificationRepository interface {
	Create(ctx context.Context, qualification *model.BusinessQualification) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.BusinessQualification, error)
	GetByMerchantID(ctx context.Context, merchantID uuid.UUID) ([]*model.BusinessQualification, error)
	Update(ctx context.Context, qualification *model.BusinessQualification) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, verifiedBy uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type businessQualificationRepository struct {
	db *gorm.DB
}

// NewBusinessQualificationRepository 创建业务资质仓储实例
func NewBusinessQualificationRepository(db *gorm.DB) BusinessQualificationRepository {
	return &businessQualificationRepository{db: db}
}

func (r *businessQualificationRepository) Create(ctx context.Context, qualification *model.BusinessQualification) error {
	return r.db.WithContext(ctx).Create(qualification).Error
}

func (r *businessQualificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.BusinessQualification, error) {
	var qualification model.BusinessQualification
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&qualification).Error
	if err != nil {
		return nil, err
	}
	return &qualification, nil
}

func (r *businessQualificationRepository) GetByMerchantID(ctx context.Context, merchantID uuid.UUID) ([]*model.BusinessQualification, error) {
	var qualifications []*model.BusinessQualification
	err := r.db.WithContext(ctx).
		Where("merchant_id = ?", merchantID).
		Order("created_at DESC").
		Find(&qualifications).Error
	return qualifications, err
}

func (r *businessQualificationRepository) Update(ctx context.Context, qualification *model.BusinessQualification) error {
	return r.db.WithContext(ctx).Save(qualification).Error
}

func (r *businessQualificationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, verifiedBy uuid.UUID) error {
	updates := map[string]interface{}{
		"status":      status,
		"verified_by": verifiedBy,
		"verified_at": gorm.Expr("NOW()"),
	}
	return r.db.WithContext(ctx).Model(&model.BusinessQualification{}).Where("id = ?", id).Updates(updates).Error
}

func (r *businessQualificationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.BusinessQualification{}, "id = ?", id).Error
}
