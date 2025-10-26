package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"payment-platform/merchant-quota-service/internal/model"
)

// AlertRepository 配额预警仓储接口
type AlertRepository interface {
	Create(ctx context.Context, alert *model.QuotaAlert) error
	Update(ctx context.Context, alert *model.QuotaAlert) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.QuotaAlert, error)

	// 查询商户的活跃预警
	GetActiveAlerts(ctx context.Context, merchantID uuid.UUID, alertLevel string) ([]*model.QuotaAlert, error)

	// 检查是否已存在相同预警（防止重复发送）
	ExistsByMerchantAndType(ctx context.Context, merchantID uuid.UUID, alertType string, since time.Time) (bool, error)

	// 标记预警为已处理
	MarkAsResolved(ctx context.Context, id uuid.UUID, resolvedBy uuid.UUID) error

	// 标记预警为已通知
	MarkAsNotified(ctx context.Context, id uuid.UUID) error

	// 清除过期预警（24小时前的warning级别预警）
	CleanupExpiredAlerts(ctx context.Context, expiredBefore time.Time) error

	// 列表查询
	List(ctx context.Context, merchantID *uuid.UUID, alertLevel, alertType string, isResolved *bool, offset, limit int) ([]*model.QuotaAlert, int64, error)

	// 统计预警数量
	CountByLevel(ctx context.Context, merchantID uuid.UUID, alertLevel string, startTime, endTime time.Time) (int64, error)
}

type alertRepository struct {
	db *gorm.DB
}

// NewAlertRepository 创建预警仓储实例
func NewAlertRepository(db *gorm.DB) AlertRepository {
	return &alertRepository{db: db}
}

func (r *alertRepository) Create(ctx context.Context, alert *model.QuotaAlert) error {
	return r.db.WithContext(ctx).Create(alert).Error
}

func (r *alertRepository) Update(ctx context.Context, alert *model.QuotaAlert) error {
	return r.db.WithContext(ctx).Save(alert).Error
}

func (r *alertRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.QuotaAlert, error) {
	var alert model.QuotaAlert
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&alert).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &alert, nil
}

func (r *alertRepository) GetActiveAlerts(ctx context.Context, merchantID uuid.UUID, alertLevel string) ([]*model.QuotaAlert, error) {
	query := r.db.WithContext(ctx).
		Where("merchant_id = ?", merchantID).
		Where("is_resolved = ?", false)

	if alertLevel != "" {
		query = query.Where("alert_level = ?", alertLevel)
	}

	var alerts []*model.QuotaAlert
	err := query.Order("created_at DESC").Find(&alerts).Error
	if err != nil {
		return nil, err
	}
	return alerts, nil
}

func (r *alertRepository) ExistsByMerchantAndType(ctx context.Context, merchantID uuid.UUID, alertType string, since time.Time) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.QuotaAlert{}).
		Where("merchant_id = ? AND alert_type = ?", merchantID, alertType).
		Where("created_at >= ?", since).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *alertRepository) MarkAsResolved(ctx context.Context, id uuid.UUID, resolvedBy uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.QuotaAlert{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_resolved":  true,
			"resolved_by":  resolvedBy,
			"resolved_at":  &now,
		}).Error
}

func (r *alertRepository) MarkAsNotified(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.QuotaAlert{}).
		Where("id = ?", id).
		Update("notified_at", &now).Error
}

func (r *alertRepository) CleanupExpiredAlerts(ctx context.Context, expiredBefore time.Time) error {
	// 软删除过期的warning级别预警（critical级别预警保留）
	return r.db.WithContext(ctx).
		Where("alert_level = ?", "warning").
		Where("created_at < ?", expiredBefore).
		Delete(&model.QuotaAlert{}).Error
}

func (r *alertRepository) List(ctx context.Context, merchantID *uuid.UUID, alertLevel, alertType string, isResolved *bool, offset, limit int) ([]*model.QuotaAlert, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.QuotaAlert{})

	if merchantID != nil {
		query = query.Where("merchant_id = ?", *merchantID)
	}
	if alertLevel != "" {
		query = query.Where("alert_level = ?", alertLevel)
	}
	if alertType != "" {
		query = query.Where("alert_type = ?", alertType)
	}
	if isResolved != nil {
		query = query.Where("is_resolved = ?", *isResolved)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var alerts []*model.QuotaAlert
	err := query.Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&alerts).Error
	if err != nil {
		return nil, 0, err
	}

	return alerts, total, nil
}

func (r *alertRepository) CountByLevel(ctx context.Context, merchantID uuid.UUID, alertLevel string, startTime, endTime time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.QuotaAlert{}).
		Where("merchant_id = ? AND alert_level = ?", merchantID, alertLevel).
		Where("created_at >= ? AND created_at <= ?", startTime, endTime).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}
