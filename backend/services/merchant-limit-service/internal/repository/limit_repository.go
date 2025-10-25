package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"payment-platform/merchant-limit-service/internal/model"
)

// LimitRepository 额度仓库接口
type LimitRepository interface {
	// Tier operations
	CreateTier(ctx context.Context, tier *model.MerchantTier) error
	GetTierByID(ctx context.Context, id uuid.UUID) (*model.MerchantTier, error)
	GetTierByCode(ctx context.Context, code string) (*model.MerchantTier, error)
	UpdateTier(ctx context.Context, tier *model.MerchantTier) error
	ListTiers(ctx context.Context) ([]*model.MerchantTier, error)
	DeleteTier(ctx context.Context, id uuid.UUID) error

	// Limit operations
	CreateLimit(ctx context.Context, limit *model.MerchantLimit) error
	GetLimitByMerchantID(ctx context.Context, merchantID uuid.UUID) (*model.MerchantLimit, error)
	UpdateLimit(ctx context.Context, limit *model.MerchantLimit) error
	UpdateUsage(ctx context.Context, merchantID uuid.UUID, dailyDelta, monthlyDelta int64) error
	SuspendMerchant(ctx context.Context, merchantID uuid.UUID, reason string) error
	UnsuspendMerchant(ctx context.Context, merchantID uuid.UUID) error
	ResetDailyUsage(ctx context.Context, merchantID uuid.UUID) error
	ResetMonthlyUsage(ctx context.Context, merchantID uuid.UUID) error

	// Usage log operations
	CreateUsageLog(ctx context.Context, log *model.LimitUsageLog) error
	ListUsageLogs(ctx context.Context, merchantID uuid.UUID, startTime, endTime *time.Time, page, pageSize int) ([]*model.LimitUsageLog, int64, error)

	// Statistics
	GetMerchantStatistics(ctx context.Context, merchantID uuid.UUID) (*MerchantStatistics, error)
}

// DTOs

type MerchantStatistics struct {
	MerchantID         uuid.UUID `json:"merchant_id"`
	TierCode           string    `json:"tier_code"`
	DailyLimit         int64     `json:"daily_limit"`
	DailyUsed          int64     `json:"daily_used"`
	DailyRemaining     int64     `json:"daily_remaining"`
	DailyUsageRate     float64   `json:"daily_usage_rate"`
	MonthlyLimit       int64     `json:"monthly_limit"`
	MonthlyUsed        int64     `json:"monthly_used"`
	MonthlyRemaining   int64     `json:"monthly_remaining"`
	MonthlyUsageRate   float64   `json:"monthly_usage_rate"`
	SingleTransLimit   int64     `json:"single_trans_limit"`
	IsSuspended        bool      `json:"is_suspended"`
	TotalTransactions  int       `json:"total_transactions"`
	SuccessCount       int       `json:"success_count"`
	FailureCount       int       `json:"failure_count"`
}

// limitRepository 额度仓库实现
type limitRepository struct {
	db *gorm.DB
}

// NewLimitRepository 创建额度仓库实例
func NewLimitRepository(db *gorm.DB) LimitRepository {
	return &limitRepository{db: db}
}

// CreateTier 创建等级
func (r *limitRepository) CreateTier(ctx context.Context, tier *model.MerchantTier) error {
	if err := r.db.WithContext(ctx).Create(tier).Error; err != nil {
		return fmt.Errorf("create tier failed: %w", err)
	}
	return nil
}

// GetTierByID 根据ID获取等级
func (r *limitRepository) GetTierByID(ctx context.Context, id uuid.UUID) (*model.MerchantTier, error) {
	var tier model.MerchantTier
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&tier).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get tier by id failed: %w", err)
	}
	return &tier, nil
}

// GetTierByCode 根据Code获取等级
func (r *limitRepository) GetTierByCode(ctx context.Context, code string) (*model.MerchantTier, error) {
	var tier model.MerchantTier
	if err := r.db.WithContext(ctx).Where("tier_code = ?", code).First(&tier).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get tier by code failed: %w", err)
	}
	return &tier, nil
}

// UpdateTier 更新等级
func (r *limitRepository) UpdateTier(ctx context.Context, tier *model.MerchantTier) error {
	if err := r.db.WithContext(ctx).Save(tier).Error; err != nil {
		return fmt.Errorf("update tier failed: %w", err)
	}
	return nil
}

// ListTiers 查询所有等级
func (r *limitRepository) ListTiers(ctx context.Context) ([]*model.MerchantTier, error) {
	var tiers []*model.MerchantTier
	if err := r.db.WithContext(ctx).Order("tier_level ASC").Find(&tiers).Error; err != nil {
		return nil, fmt.Errorf("list tiers failed: %w", err)
	}
	return tiers, nil
}

// DeleteTier 删除等级
func (r *limitRepository) DeleteTier(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&model.MerchantTier{}, id).Error; err != nil {
		return fmt.Errorf("delete tier failed: %w", err)
	}
	return nil
}

// CreateLimit 创建商户额度
func (r *limitRepository) CreateLimit(ctx context.Context, limit *model.MerchantLimit) error {
	if err := r.db.WithContext(ctx).Create(limit).Error; err != nil {
		return fmt.Errorf("create limit failed: %w", err)
	}
	return nil
}

// GetLimitByMerchantID 根据商户ID获取额度
func (r *limitRepository) GetLimitByMerchantID(ctx context.Context, merchantID uuid.UUID) (*model.MerchantLimit, error) {
	var limit model.MerchantLimit
	if err := r.db.WithContext(ctx).Where("merchant_id = ?", merchantID).First(&limit).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get limit by merchant id failed: %w", err)
	}
	return &limit, nil
}

// UpdateLimit 更新商户额度
func (r *limitRepository) UpdateLimit(ctx context.Context, limit *model.MerchantLimit) error {
	if err := r.db.WithContext(ctx).Save(limit).Error; err != nil {
		return fmt.Errorf("update limit failed: %w", err)
	}
	return nil
}

// UpdateUsage 更新使用量 (原子操作)
func (r *limitRepository) UpdateUsage(ctx context.Context, merchantID uuid.UUID, dailyDelta, monthlyDelta int64) error {
	if err := r.db.WithContext(ctx).
		Model(&model.MerchantLimit{}).
		Where("merchant_id = ?", merchantID).
		Updates(map[string]interface{}{
			"daily_used":   gorm.Expr("daily_used + ?", dailyDelta),
			"monthly_used": gorm.Expr("monthly_used + ?", monthlyDelta),
			"updated_at":   time.Now(),
		}).Error; err != nil {
		return fmt.Errorf("update usage failed: %w", err)
	}
	return nil
}

// SuspendMerchant 暂停商户
func (r *limitRepository) SuspendMerchant(ctx context.Context, merchantID uuid.UUID, reason string) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).
		Model(&model.MerchantLimit{}).
		Where("merchant_id = ?", merchantID).
		Updates(map[string]interface{}{
			"is_suspended":     true,
			"suspended_at":     now,
			"suspended_reason": reason,
			"updated_at":       now,
		}).Error; err != nil {
		return fmt.Errorf("suspend merchant failed: %w", err)
	}
	return nil
}

// UnsuspendMerchant 恢复商户
func (r *limitRepository) UnsuspendMerchant(ctx context.Context, merchantID uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Model(&model.MerchantLimit{}).
		Where("merchant_id = ?", merchantID).
		Updates(map[string]interface{}{
			"is_suspended":     false,
			"suspended_at":     nil,
			"suspended_reason": "",
			"updated_at":       time.Now(),
		}).Error; err != nil {
		return fmt.Errorf("unsuspend merchant failed: %w", err)
	}
	return nil
}

// ResetDailyUsage 重置日使用量
func (r *limitRepository) ResetDailyUsage(ctx context.Context, merchantID uuid.UUID) error {
	now := time.Now()
	nextReset := now.Add(24 * time.Hour)

	if err := r.db.WithContext(ctx).
		Model(&model.MerchantLimit{}).
		Where("merchant_id = ?", merchantID).
		Updates(map[string]interface{}{
			"daily_used":      0,
			"refunded_today":  0,
			"daily_reset_at":  nextReset,
			"updated_at":      now,
		}).Error; err != nil {
		return fmt.Errorf("reset daily usage failed: %w", err)
	}
	return nil
}

// ResetMonthlyUsage 重置月使用量
func (r *limitRepository) ResetMonthlyUsage(ctx context.Context, merchantID uuid.UUID) error {
	now := time.Now()
	nextReset := now.AddDate(0, 1, 0)

	if err := r.db.WithContext(ctx).
		Model(&model.MerchantLimit{}).
		Where("merchant_id = ?", merchantID).
		Updates(map[string]interface{}{
			"monthly_used":     0,
			"monthly_reset_at": nextReset,
			"updated_at":       now,
		}).Error; err != nil {
		return fmt.Errorf("reset monthly usage failed: %w", err)
	}
	return nil
}

// CreateUsageLog 创建使用日志
func (r *limitRepository) CreateUsageLog(ctx context.Context, log *model.LimitUsageLog) error {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		return fmt.Errorf("create usage log failed: %w", err)
	}
	return nil
}

// ListUsageLogs 查询使用日志
func (r *limitRepository) ListUsageLogs(ctx context.Context, merchantID uuid.UUID, startTime, endTime *time.Time, page, pageSize int) ([]*model.LimitUsageLog, int64, error) {
	var logs []*model.LimitUsageLog
	var total int64

	query := r.db.WithContext(ctx).Model(&model.LimitUsageLog{}).Where("merchant_id = ?", merchantID)

	if startTime != nil {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime != nil {
		query = query.Where("created_at <= ?", endTime)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count usage logs failed: %w", err)
	}

	// Paginate
	offset := (page - 1) * pageSize
	if err := query.
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("list usage logs failed: %w", err)
	}

	return logs, total, nil
}

// GetMerchantStatistics 获取商户统计信息
func (r *limitRepository) GetMerchantStatistics(ctx context.Context, merchantID uuid.UUID) (*MerchantStatistics, error) {
	// Get limit
	limit, err := r.GetLimitByMerchantID(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("get limit failed: %w", err)
	}
	if limit == nil {
		return nil, fmt.Errorf("merchant limit not found")
	}

	// Get tier
	tier, err := r.GetTierByID(ctx, limit.TierID)
	if err != nil {
		return nil, fmt.Errorf("get tier failed: %w", err)
	}
	if tier == nil {
		return nil, fmt.Errorf("merchant tier not found")
	}

	// Calculate effective limits (custom overrides tier)
	dailyLimit := tier.DailyLimit
	if limit.CustomDailyLimit != nil {
		dailyLimit = *limit.CustomDailyLimit
	}

	monthlyLimit := tier.MonthlyLimit
	if limit.CustomMonthlyLimit != nil {
		monthlyLimit = *limit.CustomMonthlyLimit
	}

	singleTransLimit := tier.SingleTransLimit
	if limit.CustomSingleTransLimit != nil {
		singleTransLimit = *limit.CustomSingleTransLimit
	}

	// Calculate remaining amounts
	dailyRemaining := dailyLimit - limit.DailyUsed
	if dailyRemaining < 0 {
		dailyRemaining = 0
	}

	monthlyRemaining := monthlyLimit - limit.MonthlyUsed
	if monthlyRemaining < 0 {
		monthlyRemaining = 0
	}

	// Calculate usage rates
	dailyUsageRate := 0.0
	if dailyLimit > 0 {
		dailyUsageRate = float64(limit.DailyUsed) / float64(dailyLimit) * 100
	}

	monthlyUsageRate := 0.0
	if monthlyLimit > 0 {
		monthlyUsageRate = float64(limit.MonthlyUsed) / float64(monthlyLimit) * 100
	}

	// Get transaction statistics
	var totalTrans, successCount, failureCount int64
	r.db.WithContext(ctx).Model(&model.LimitUsageLog{}).
		Where("merchant_id = ?", merchantID).
		Count(&totalTrans)

	r.db.WithContext(ctx).Model(&model.LimitUsageLog{}).
		Where("merchant_id = ? AND success = true", merchantID).
		Count(&successCount)

	r.db.WithContext(ctx).Model(&model.LimitUsageLog{}).
		Where("merchant_id = ? AND success = false", merchantID).
		Count(&failureCount)

	return &MerchantStatistics{
		MerchantID:        merchantID,
		TierCode:          tier.TierCode,
		DailyLimit:        dailyLimit,
		DailyUsed:         limit.DailyUsed,
		DailyRemaining:    dailyRemaining,
		DailyUsageRate:    dailyUsageRate,
		MonthlyLimit:      monthlyLimit,
		MonthlyUsed:       limit.MonthlyUsed,
		MonthlyRemaining:  monthlyRemaining,
		MonthlyUsageRate:  monthlyUsageRate,
		SingleTransLimit:  singleTransLimit,
		IsSuspended:       limit.IsSuspended,
		TotalTransactions: int(totalTrans),
		SuccessCount:      int(successCount),
		FailureCount:      int(failureCount),
	}, nil
}
