package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"payment-platform/kyc-service/internal/model"
)

// BusinessQualificationRepository 业务资质仓储接口
type BusinessQualificationRepository interface {
	Create(ctx context.Context, qual *model.BusinessQualification) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.BusinessQualification, error)
	GetByMerchantID(ctx context.Context, merchantID uuid.UUID) ([]*model.BusinessQualification, error)
	GetByMerchantIDAndType(ctx context.Context, merchantID uuid.UUID, qualType string) ([]*model.BusinessQualification, error)
	Update(ctx context.Context, qual *model.BusinessQualification) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, status string, limit, offset int) ([]*model.BusinessQualification, int64, error)
}

type businessQualificationRepository struct {
	db *gorm.DB
}

// NewBusinessQualificationRepository 创建业务资质仓储
func NewBusinessQualificationRepository(db *gorm.DB) BusinessQualificationRepository {
	return &businessQualificationRepository{db: db}
}

// Create 创建业务资质
func (r *businessQualificationRepository) Create(ctx context.Context, qual *model.BusinessQualification) error {
	return r.db.WithContext(ctx).Create(qual).Error
}

// GetByID 根据ID获取业务资质
func (r *businessQualificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.BusinessQualification, error) {
	var qual model.BusinessQualification
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&qual).Error
	if err != nil {
		return nil, err
	}
	return &qual, nil
}

// GetByMerchantID 获取商户的所有业务资质
func (r *businessQualificationRepository) GetByMerchantID(ctx context.Context, merchantID uuid.UUID) ([]*model.BusinessQualification, error) {
	var quals []*model.BusinessQualification
	err := r.db.WithContext(ctx).
		Where("merchant_id = ?", merchantID).
		Order("created_at DESC").
		Find(&quals).Error
	return quals, err
}

// GetByMerchantIDAndType 获取商户指定类型的业务资质
func (r *businessQualificationRepository) GetByMerchantIDAndType(ctx context.Context, merchantID uuid.UUID, qualType string) ([]*model.BusinessQualification, error) {
	var quals []*model.BusinessQualification
	err := r.db.WithContext(ctx).
		Where("merchant_id = ? AND qualification_type = ?", merchantID, qualType).
		Order("created_at DESC").
		Find(&quals).Error
	return quals, err
}

// Update 更新业务资质
func (r *businessQualificationRepository) Update(ctx context.Context, qual *model.BusinessQualification) error {
	return r.db.WithContext(ctx).Save(qual).Error
}

// Delete 删除业务资质（软删除）
func (r *businessQualificationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.BusinessQualification{}, "id = ?", id).Error
}

// List 分页列出业务资质
func (r *businessQualificationRepository) List(ctx context.Context, status string, limit, offset int) ([]*model.BusinessQualification, int64, error) {
	var quals []*model.BusinessQualification
	var total int64

	query := r.db.WithContext(ctx).Model(&model.BusinessQualification{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&quals).Error

	return quals, total, err
}
