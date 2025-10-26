package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"payment-platform/merchant-service/internal/model"
)

// MerchantUserRepository 商户用户仓储接口
type MerchantUserRepository interface {
	Create(ctx context.Context, user *model.MerchantUser) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.MerchantUser, error)
	GetByEmail(ctx context.Context, email string) (*model.MerchantUser, error)
	GetByMerchantID(ctx context.Context, merchantID uuid.UUID) ([]*model.MerchantUser, error)
	Update(ctx context.Context, user *model.MerchantUser) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type merchantUserRepository struct {
	db *gorm.DB
}

// NewMerchantUserRepository 创建商户用户仓储实例
func NewMerchantUserRepository(db *gorm.DB) MerchantUserRepository {
	return &merchantUserRepository{db: db}
}

func (r *merchantUserRepository) Create(ctx context.Context, user *model.MerchantUser) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *merchantUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.MerchantUser, error) {
	var user model.MerchantUser
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *merchantUserRepository) GetByEmail(ctx context.Context, email string) (*model.MerchantUser, error) {
	var user model.MerchantUser
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *merchantUserRepository) GetByMerchantID(ctx context.Context, merchantID uuid.UUID) ([]*model.MerchantUser, error) {
	var users []*model.MerchantUser
	err := r.db.WithContext(ctx).
		Where("merchant_id = ?", merchantID).
		Order("created_at DESC").
		Find(&users).Error
	return users, err
}

func (r *merchantUserRepository) Update(ctx context.Context, user *model.MerchantUser) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *merchantUserRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.db.WithContext(ctx).Model(&model.MerchantUser{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *merchantUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.MerchantUser{}, "id = ?", id).Error
}
