package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"payment-platform/merchant-quota-service/internal/model"
)

// UsageLogRepository 配额使用日志仓储接口
type UsageLogRepository interface {
	Create(ctx context.Context, log *model.QuotaUsageLog) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.QuotaUsageLog, error)
	GetByOrderNo(ctx context.Context, merchantID uuid.UUID, orderNo string) ([]*model.QuotaUsageLog, error)

	// 按商户查询使用日志
	ListByMerchant(ctx context.Context, merchantID uuid.UUID, actionType string, startTime, endTime *time.Time, offset, limit int) ([]*model.QuotaUsageLog, int64, error)

	// 按时间范围查询（审计用）
	ListByTimeRange(ctx context.Context, startTime, endTime time.Time, offset, limit int) ([]*model.QuotaUsageLog, int64, error)

	// 统计配额操作次数
	CountByAction(ctx context.Context, merchantID uuid.UUID, actionType string, startTime, endTime time.Time) (int64, error)
}

type usageLogRepository struct {
	db *gorm.DB
}

// NewUsageLogRepository 创建使用日志仓储实例
func NewUsageLogRepository(db *gorm.DB) UsageLogRepository {
	return &usageLogRepository{db: db}
}

func (r *usageLogRepository) Create(ctx context.Context, log *model.QuotaUsageLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *usageLogRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.QuotaUsageLog, error) {
	var log model.QuotaUsageLog
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&log).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &log, nil
}

func (r *usageLogRepository) GetByOrderNo(ctx context.Context, merchantID uuid.UUID, orderNo string) ([]*model.QuotaUsageLog, error) {
	var logs []*model.QuotaUsageLog
	err := r.db.WithContext(ctx).
		Where("merchant_id = ? AND order_no = ?", merchantID, orderNo).
		Order("created_at DESC").
		Find(&logs).Error
	if err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *usageLogRepository) ListByMerchant(ctx context.Context, merchantID uuid.UUID, actionType string, startTime, endTime *time.Time, offset, limit int) ([]*model.QuotaUsageLog, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.QuotaUsageLog{}).
		Where("merchant_id = ?", merchantID)

	if actionType != "" {
		query = query.Where("action_type = ?", actionType)
	}
	if startTime != nil {
		query = query.Where("created_at >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("created_at <= ?", *endTime)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var logs []*model.QuotaUsageLog
	err := query.Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *usageLogRepository) ListByTimeRange(ctx context.Context, startTime, endTime time.Time, offset, limit int) ([]*model.QuotaUsageLog, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.QuotaUsageLog{}).
		Where("created_at >= ? AND created_at <= ?", startTime, endTime)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var logs []*model.QuotaUsageLog
	err := query.Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *usageLogRepository) CountByAction(ctx context.Context, merchantID uuid.UUID, actionType string, startTime, endTime time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.QuotaUsageLog{}).
		Where("merchant_id = ? AND action_type = ?", merchantID, actionType).
		Where("created_at >= ? AND created_at <= ?", startTime, endTime).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}
