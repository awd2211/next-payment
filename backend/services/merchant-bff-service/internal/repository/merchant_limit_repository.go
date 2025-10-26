package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"payment-platform/merchant-service/internal/model"
)

// MerchantLimitRepository 商户额度仓库接口
type MerchantLimitRepository interface {
	// 创建商户额度
	Create(ctx context.Context, limit *model.MerchantLimit) error

	// 获取商户额度
	GetByMerchantID(ctx context.Context, merchantID uuid.UUID) (*model.MerchantLimit, error)

	// 更新商户额度配置
	UpdateLimitConfig(ctx context.Context, merchantID uuid.UUID, dailyLimit, monthlyLimit, singleLimit int64) error

	// 更新已使用额度（DB同步，主要依赖Redis）
	UpdateUsedAmount(ctx context.Context, merchantID uuid.UUID, usedToday, usedMonth int64) error

	// 重置日限额使用量
	ResetDailyUsage(ctx context.Context, merchantID uuid.UUID) error

	// 重置月限额使用量
	ResetMonthlyUsage(ctx context.Context, merchantID uuid.UUID) error

	// 设置限额状态
	SetLimitStatus(ctx context.Context, merchantID uuid.UUID, isLimited bool, reason string) error
}

// merchantLimitRepository 商户额度仓库实现
type merchantLimitRepository struct {
	db *gorm.DB
}

// NewMerchantLimitRepository 创建商户额度仓库
func NewMerchantLimitRepository(db *gorm.DB) MerchantLimitRepository {
	return &merchantLimitRepository{db: db}
}

// Create 创建商户额度
func (r *merchantLimitRepository) Create(ctx context.Context, limit *model.MerchantLimit) error {
	return r.db.WithContext(ctx).Create(limit).Error
}

// GetByMerchantID 获取商户额度
func (r *merchantLimitRepository) GetByMerchantID(ctx context.Context, merchantID uuid.UUID) (*model.MerchantLimit, error) {
	var limit model.MerchantLimit
	err := r.db.WithContext(ctx).
		Where("merchant_id = ?", merchantID).
		First(&limit).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &limit, err
}

// UpdateLimitConfig 更新商户额度配置
func (r *merchantLimitRepository) UpdateLimitConfig(ctx context.Context, merchantID uuid.UUID, dailyLimit, monthlyLimit, singleLimit int64) error {
	return r.db.WithContext(ctx).
		Model(&model.MerchantLimit{}).
		Where("merchant_id = ?", merchantID).
		Updates(map[string]interface{}{
			"daily_limit":   dailyLimit,
			"monthly_limit": monthlyLimit,
			"single_limit":  singleLimit,
			"updated_at":    time.Now(),
		}).Error
}

// UpdateUsedAmount 更新已使用额度
func (r *merchantLimitRepository) UpdateUsedAmount(ctx context.Context, merchantID uuid.UUID, usedToday, usedMonth int64) error {
	return r.db.WithContext(ctx).
		Model(&model.MerchantLimit{}).
		Where("merchant_id = ?", merchantID).
		Updates(map[string]interface{}{
			"used_today":  usedToday,
			"used_month":  usedMonth,
			"updated_at":  time.Now(),
		}).Error
}

// ResetDailyUsage 重置日限额使用量
func (r *merchantLimitRepository) ResetDailyUsage(ctx context.Context, merchantID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.MerchantLimit{}).
		Where("merchant_id = ?", merchantID).
		Updates(map[string]interface{}{
			"used_today":      0,
			"today_count":     0,
			"last_reset_day":  now,
			"updated_at":      now,
		}).Error
}

// ResetMonthlyUsage 重置月限额使用量
func (r *merchantLimitRepository) ResetMonthlyUsage(ctx context.Context, merchantID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.MerchantLimit{}).
		Where("merchant_id = ?", merchantID).
		Updates(map[string]interface{}{
			"used_month":   0,
			"month_count":  0,
			"updated_at":   time.Now(),
		}).Error
}

// SetLimitStatus 设置限额状态
func (r *merchantLimitRepository) SetLimitStatus(ctx context.Context, merchantID uuid.UUID, isLimited bool, reason string) error {
	return r.db.WithContext(ctx).
		Model(&model.MerchantLimit{}).
		Where("merchant_id = ?", merchantID).
		Updates(map[string]interface{}{
			"is_limited":   isLimited,
			"limit_reason": reason,
			"updated_at":   time.Now(),
		}).Error
}
