package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/payment-platform/services/merchant-service/internal/model"
	"gorm.io/gorm"
)

// MerchantRepository 商户仓储接口
type MerchantRepository interface {
	Create(ctx context.Context, merchant *model.Merchant) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Merchant, error)
	GetByEmail(ctx context.Context, email string) (*model.Merchant, error)
	List(ctx context.Context, page, pageSize int, status, kycStatus, keyword string) ([]*model.Merchant, int64, error)
	Update(ctx context.Context, merchant *model.Merchant) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateKYCStatus(ctx context.Context, id uuid.UUID, kycStatus string) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type merchantRepository struct {
	db *gorm.DB
}

// NewMerchantRepository 创建商户仓储实例
func NewMerchantRepository(db *gorm.DB) MerchantRepository {
	return &merchantRepository{db: db}
}

// Create 创建商户
func (r *merchantRepository) Create(ctx context.Context, merchant *model.Merchant) error {
	return r.db.WithContext(ctx).Create(merchant).Error
}

// GetByID 根据ID获取商户
func (r *merchantRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Merchant, error) {
	var merchant model.Merchant
	err := r.db.WithContext(ctx).
		Preload("APIKeys").
		Preload("WebhookConfig").
		Preload("ChannelConfigs").
		First(&merchant, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &merchant, nil
}

// GetByEmail 根据邮箱获取商户
func (r *merchantRepository) GetByEmail(ctx context.Context, email string) (*model.Merchant, error) {
	var merchant model.Merchant
	err := r.db.WithContext(ctx).First(&merchant, "email = ?", email).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &merchant, nil
}

// List 分页查询商户列表
func (r *merchantRepository) List(ctx context.Context, page, pageSize int, status, kycStatus, keyword string) ([]*model.Merchant, int64, error) {
	var merchants []*model.Merchant
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Merchant{})

	// 状态筛选
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// KYC状态筛选
	if kycStatus != "" {
		query = query.Where("kyc_status = ?", kycStatus)
	}

	// 关键词搜索（商户名称、邮箱、公司名称）
	if keyword != "" {
		query = query.Where("name LIKE ? OR email LIKE ? OR company_name LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&merchants).Error

	return merchants, total, err
}

// Update 更新商户
func (r *merchantRepository) Update(ctx context.Context, merchant *model.Merchant) error {
	return r.db.WithContext(ctx).Save(merchant).Error
}

// UpdateStatus 更新商户状态
func (r *merchantRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.db.WithContext(ctx).
		Model(&model.Merchant{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// UpdateKYCStatus 更新KYC状态
func (r *merchantRepository) UpdateKYCStatus(ctx context.Context, id uuid.UUID, kycStatus string) error {
	return r.db.WithContext(ctx).
		Model(&model.Merchant{}).
		Where("id = ?", id).
		Update("kyc_status", kycStatus).Error
}

// Delete 软删除商户
func (r *merchantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Merchant{}, "id = ?", id).Error
}
