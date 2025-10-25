package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"payment-platform/payment-gateway/internal/model"
)

// WebhookNotificationRepository Webhook 通知仓储接口
type WebhookNotificationRepository interface {
	Create(ctx context.Context, notification *model.WebhookNotification) error
	Update(ctx context.Context, notification *model.WebhookNotification) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.WebhookNotification, error)
	GetByPaymentNo(ctx context.Context, merchantID uuid.UUID, paymentNo string) ([]*model.WebhookNotification, error)

	// 查询待重试的通知
	GetPendingRetries(ctx context.Context, limit int) ([]*model.WebhookNotification, error)

	// 查询失败的通知
	GetFailedNotifications(ctx context.Context, merchantID uuid.UUID, limit int, offset int) ([]*model.WebhookNotification, int64, error)
}

type webhookNotificationRepository struct {
	db *gorm.DB
}

// NewWebhookNotificationRepository 创建 Webhook 通知仓储
func NewWebhookNotificationRepository(db *gorm.DB) WebhookNotificationRepository {
	return &webhookNotificationRepository{db: db}
}

func (r *webhookNotificationRepository) Create(ctx context.Context, notification *model.WebhookNotification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

func (r *webhookNotificationRepository) Update(ctx context.Context, notification *model.WebhookNotification) error {
	return r.db.WithContext(ctx).Save(notification).Error
}

func (r *webhookNotificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.WebhookNotification, error) {
	var notification model.WebhookNotification
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&notification).Error; err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *webhookNotificationRepository) GetByPaymentNo(ctx context.Context, merchantID uuid.UUID, paymentNo string) ([]*model.WebhookNotification, error) {
	var notifications []*model.WebhookNotification
	if err := r.db.WithContext(ctx).
		Where("merchant_id = ? AND payment_no = ?", merchantID, paymentNo).
		Order("created_at DESC").
		Find(&notifications).Error; err != nil {
		return nil, err
	}
	return notifications, nil
}

func (r *webhookNotificationRepository) GetPendingRetries(ctx context.Context, limit int) ([]*model.WebhookNotification, error) {
	var notifications []*model.WebhookNotification
	now := time.Now()

	if err := r.db.WithContext(ctx).
		Where("status IN (?, ?)", model.WebhookStatusPending, model.WebhookStatusRetrying).
		Where("attempts < max_attempts").
		Where("next_retry_at IS NULL OR next_retry_at <= ?", now).
		Order("created_at ASC").
		Limit(limit).
		Find(&notifications).Error; err != nil {
		return nil, err
	}

	return notifications, nil
}

func (r *webhookNotificationRepository) GetFailedNotifications(ctx context.Context, merchantID uuid.UUID, limit int, offset int) ([]*model.WebhookNotification, int64, error) {
	var notifications []*model.WebhookNotification
	var total int64

	query := r.db.WithContext(ctx).Model(&model.WebhookNotification{}).
		Where("merchant_id = ? AND status = ?", merchantID, model.WebhookStatusFailed)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error; err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}
