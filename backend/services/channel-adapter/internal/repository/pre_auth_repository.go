package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"payment-platform/channel-adapter/internal/model"
)

// PreAuthRepository 预授权记录仓储接口
type PreAuthRepository interface {
	Create(ctx context.Context, record *model.PreAuthRecord) error
	GetByChannelPreAuthNo(ctx context.Context, channelPreAuthNo string) (*model.PreAuthRecord, error)
	Update(ctx context.Context, record *model.PreAuthRecord) error
	GetByMerchantID(ctx context.Context, merchantID uuid.UUID, limit, offset int) ([]*model.PreAuthRecord, int64, error)
}

type preAuthRepository struct {
	db *gorm.DB
}

// NewPreAuthRepository 创建预授权记录仓储
func NewPreAuthRepository(db *gorm.DB) PreAuthRepository {
	return &preAuthRepository{db: db}
}

// Create 创建预授权记录
func (r *preAuthRepository) Create(ctx context.Context, record *model.PreAuthRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

// GetByChannelPreAuthNo 根据渠道预授权号查询
func (r *preAuthRepository) GetByChannelPreAuthNo(ctx context.Context, channelPreAuthNo string) (*model.PreAuthRecord, error) {
	var record model.PreAuthRecord
	err := r.db.WithContext(ctx).
		Where("channel_pre_auth_no = ?", channelPreAuthNo).
		First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// Update 更新预授权记录
func (r *preAuthRepository) Update(ctx context.Context, record *model.PreAuthRecord) error {
	return r.db.WithContext(ctx).Save(record).Error
}

// GetByMerchantID 获取商户的预授权记录列表
func (r *preAuthRepository) GetByMerchantID(ctx context.Context, merchantID uuid.UUID, limit, offset int) ([]*model.PreAuthRecord, int64, error) {
	var records []*model.PreAuthRecord
	var total int64

	query := r.db.WithContext(ctx).Model(&model.PreAuthRecord{}).
		Where("merchant_id = ?", merchantID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&records).Error

	return records, total, err
}
