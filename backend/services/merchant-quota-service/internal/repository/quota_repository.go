package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"payment-platform/merchant-quota-service/internal/model"
)

// QuotaRepository 配额仓储接口
type QuotaRepository interface {
	Create(ctx context.Context, quota *model.MerchantQuota) error
	Update(ctx context.Context, quota *model.MerchantQuota) error
	GetByMerchantAndCurrency(ctx context.Context, merchantID uuid.UUID, currency string) (*model.MerchantQuota, error)
	GetByMerchantID(ctx context.Context, merchantID uuid.UUID) ([]*model.MerchantQuota, error)

	// 配额消耗（使用乐观锁）
	ConsumeQuota(ctx context.Context, merchantID uuid.UUID, currency string, amount int64, orderNo string) error

	// 配额释放（退款时调用）
	ReleaseQuota(ctx context.Context, merchantID uuid.UUID, currency string, amount int64, orderNo string) error

	// 调整配额（管理员操作）
	AdjustQuota(ctx context.Context, merchantID uuid.UUID, currency string, dailyAdjust, monthlyAdjust, yearlyAdjust int64) error

	// 重置日配额
	ResetDailyQuotas(ctx context.Context) error

	// 重置月配额
	ResetMonthlyQuotas(ctx context.Context) error

	// 暂停/恢复商户配额
	SuspendQuota(ctx context.Context, merchantID uuid.UUID, currency string) error
	ResumeQuota(ctx context.Context, merchantID uuid.UUID, currency string) error

	// 列表查询
	List(ctx context.Context, merchantID *uuid.UUID, currency string, isSuspended *bool, offset, limit int) ([]*model.MerchantQuota, int64, error)
}

type quotaRepository struct {
	db *gorm.DB
}

// NewQuotaRepository 创建配额仓储实例
func NewQuotaRepository(db *gorm.DB) QuotaRepository {
	return &quotaRepository{db: db}
}

func (r *quotaRepository) Create(ctx context.Context, quota *model.MerchantQuota) error {
	return r.db.WithContext(ctx).Create(quota).Error
}

func (r *quotaRepository) Update(ctx context.Context, quota *model.MerchantQuota) error {
	return r.db.WithContext(ctx).Save(quota).Error
}

func (r *quotaRepository) GetByMerchantAndCurrency(ctx context.Context, merchantID uuid.UUID, currency string) (*model.MerchantQuota, error) {
	var quota model.MerchantQuota
	err := r.db.WithContext(ctx).
		Where("merchant_id = ? AND currency = ?", merchantID, currency).
		First(&quota).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &quota, nil
}

func (r *quotaRepository) GetByMerchantID(ctx context.Context, merchantID uuid.UUID) ([]*model.MerchantQuota, error) {
	var quotas []*model.MerchantQuota
	err := r.db.WithContext(ctx).
		Where("merchant_id = ?", merchantID).
		Find(&quotas).Error
	if err != nil {
		return nil, err
	}
	return quotas, nil
}

func (r *quotaRepository) ConsumeQuota(ctx context.Context, merchantID uuid.UUID, currency string, amount int64, orderNo string) error {
	// 使用乐观锁更新配额
	result := r.db.WithContext(ctx).
		Model(&model.MerchantQuota{}).
		Where("merchant_id = ? AND currency = ?", merchantID, currency).
		Where("is_suspended = ?", false).
		Updates(map[string]interface{}{
			"daily_used":    gorm.Expr("daily_used + ?", amount),
			"monthly_used":  gorm.Expr("monthly_used + ?", amount),
			"yearly_used":   gorm.Expr("yearly_used + ?", amount),
			"pending_amount": gorm.Expr("pending_amount + ?", amount),
			"last_order_no": orderNo,
			"version":       gorm.Expr("version + 1"),
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("配额已暂停或不存在")
	}

	return nil
}

func (r *quotaRepository) ReleaseQuota(ctx context.Context, merchantID uuid.UUID, currency string, amount int64, orderNo string) error {
	// 释放待结算金额（退款时调用）
	result := r.db.WithContext(ctx).
		Model(&model.MerchantQuota{}).
		Where("merchant_id = ? AND currency = ?", merchantID, currency).
		Updates(map[string]interface{}{
			"pending_amount": gorm.Expr("pending_amount - ?", amount),
			"last_order_no":  orderNo,
			"version":        gorm.Expr("version + 1"),
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("配额不存在")
	}

	return nil
}

func (r *quotaRepository) AdjustQuota(ctx context.Context, merchantID uuid.UUID, currency string, dailyAdjust, monthlyAdjust, yearlyAdjust int64) error {
	// 管理员手动调整配额使用量（可为负数）
	result := r.db.WithContext(ctx).
		Model(&model.MerchantQuota{}).
		Where("merchant_id = ? AND currency = ?", merchantID, currency).
		Updates(map[string]interface{}{
			"daily_used":   gorm.Expr("daily_used + ?", dailyAdjust),
			"monthly_used": gorm.Expr("monthly_used + ?", monthlyAdjust),
			"yearly_used":  gorm.Expr("yearly_used + ?", yearlyAdjust),
			"version":      gorm.Expr("version + 1"),
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("配额不存在")
	}

	return nil
}

func (r *quotaRepository) ResetDailyQuotas(ctx context.Context) error {
	// 每日00:00重置所有商户的日配额
	return r.db.WithContext(ctx).
		Model(&model.MerchantQuota{}).
		Updates(map[string]interface{}{
			"daily_used":     0,
			"daily_reset_at": gorm.Expr("NOW()"),
			"version":        gorm.Expr("version + 1"),
		}).Error
}

func (r *quotaRepository) ResetMonthlyQuotas(ctx context.Context) error {
	// 每月1日00:00重置所有商户的月配额
	return r.db.WithContext(ctx).
		Model(&model.MerchantQuota{}).
		Updates(map[string]interface{}{
			"monthly_used":     0,
			"monthly_reset_at": gorm.Expr("NOW()"),
			"version":          gorm.Expr("version + 1"),
		}).Error
}

func (r *quotaRepository) SuspendQuota(ctx context.Context, merchantID uuid.UUID, currency string) error {
	return r.db.WithContext(ctx).
		Model(&model.MerchantQuota{}).
		Where("merchant_id = ? AND currency = ?", merchantID, currency).
		Update("is_suspended", true).Error
}

func (r *quotaRepository) ResumeQuota(ctx context.Context, merchantID uuid.UUID, currency string) error {
	return r.db.WithContext(ctx).
		Model(&model.MerchantQuota{}).
		Where("merchant_id = ? AND currency = ?", merchantID, currency).
		Update("is_suspended", false).Error
}

func (r *quotaRepository) List(ctx context.Context, merchantID *uuid.UUID, currency string, isSuspended *bool, offset, limit int) ([]*model.MerchantQuota, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.MerchantQuota{})

	if merchantID != nil {
		query = query.Where("merchant_id = ?", *merchantID)
	}
	if currency != "" {
		query = query.Where("currency = ?", currency)
	}
	if isSuspended != nil {
		query = query.Where("is_suspended = ?", *isSuspended)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var quotas []*model.MerchantQuota
	err := query.Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&quotas).Error
	if err != nil {
		return nil, 0, err
	}

	return quotas, total, nil
}
