package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"payment-platform/notification-service/internal/model"
	"gorm.io/gorm"
)

// NotificationRepository 通知仓储接口
type NotificationRepository interface {
	// 通知管理
	Create(ctx context.Context, notification *model.Notification) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Notification, error)
	List(ctx context.Context, query *NotificationQuery) ([]*model.Notification, int64, error)
	Update(ctx context.Context, notification *model.Notification) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	Delete(ctx context.Context, id uuid.UUID) error

	// 模板管理
	CreateTemplate(ctx context.Context, template *model.NotificationTemplate) error
	GetTemplate(ctx context.Context, code string, merchantID *uuid.UUID) (*model.NotificationTemplate, error)
	ListTemplates(ctx context.Context, merchantID *uuid.UUID) ([]*model.NotificationTemplate, error)
	UpdateTemplate(ctx context.Context, template *model.NotificationTemplate) error
	DeleteTemplate(ctx context.Context, id uuid.UUID) error

	// Webhook 端点管理
	CreateEndpoint(ctx context.Context, endpoint *model.WebhookEndpoint) error
	GetEndpoint(ctx context.Context, id uuid.UUID) (*model.WebhookEndpoint, error)
	ListEndpoints(ctx context.Context, merchantID uuid.UUID) ([]*model.WebhookEndpoint, error)
	UpdateEndpoint(ctx context.Context, endpoint *model.WebhookEndpoint) error
	DeleteEndpoint(ctx context.Context, id uuid.UUID) error

	// Webhook 投递记录
	CreateDelivery(ctx context.Context, delivery *model.WebhookDelivery) error
	GetDelivery(ctx context.Context, id uuid.UUID) (*model.WebhookDelivery, error)
	ListDeliveries(ctx context.Context, query *DeliveryQuery) ([]*model.WebhookDelivery, int64, error)
	UpdateDelivery(ctx context.Context, delivery *model.WebhookDelivery) error

	// 查询待处理的通知
	ListPendingNotifications(ctx context.Context, limit int) ([]*model.Notification, error)
	// 查询待重试的投递
	ListPendingDeliveries(ctx context.Context, limit int) ([]*model.WebhookDelivery, error)

	// 通知偏好管理
	CreatePreference(ctx context.Context, preference *model.NotificationPreference) error
	GetPreference(ctx context.Context, id uuid.UUID) (*model.NotificationPreference, error)
	ListPreferences(ctx context.Context, merchantID uuid.UUID, userID *uuid.UUID) ([]*model.NotificationPreference, error)
	UpdatePreference(ctx context.Context, preference *model.NotificationPreference) error
	DeletePreference(ctx context.Context, id uuid.UUID) error
	CheckPreference(ctx context.Context, merchantID uuid.UUID, userID *uuid.UUID, channel, eventType string) (bool, error)
}

type notificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository 创建通知仓储实例
func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

// NotificationQuery 通知查询参数
type NotificationQuery struct {
	MerchantID *uuid.UUID
	Type       string
	Channel    string
	Status     string
	StartTime  *time.Time
	EndTime    *time.Time
	Page       int
	PageSize   int
}

// DeliveryQuery 投递查询参数
type DeliveryQuery struct {
	EndpointID *uuid.UUID
	MerchantID *uuid.UUID
	EventType  string
	Status     string
	StartTime  *time.Time
	EndTime    *time.Time
	Page       int
	PageSize   int
}

// Create 创建通知
func (r *notificationRepository) Create(ctx context.Context, notification *model.Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

// GetByID 根据ID获取通知
func (r *notificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Notification, error) {
	var notification model.Notification
	err := r.db.WithContext(ctx).First(&notification, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &notification, nil
}

// List 查询通知列表
func (r *notificationRepository) List(ctx context.Context, query *NotificationQuery) ([]*model.Notification, int64, error) {
	var notifications []*model.Notification
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Notification{})

	// 构建查询条件
	if query.MerchantID != nil {
		db = db.Where("merchant_id = ?", *query.MerchantID)
	}
	if query.Type != "" {
		db = db.Where("type = ?", query.Type)
	}
	if query.Channel != "" {
		db = db.Where("channel = ?", query.Channel)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.StartTime != nil {
		db = db.Where("created_at >= ?", *query.StartTime)
	}
	if query.EndTime != nil {
		db = db.Where("created_at <= ?", *query.EndTime)
	}

	// 计算总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		query.PageSize = 20
	}
	offset := (query.Page - 1) * query.PageSize

	err := db.Order("created_at DESC").
		Offset(offset).
		Limit(query.PageSize).
		Find(&notifications).Error

	return notifications, total, err
}

// Update 更新通知
func (r *notificationRepository) Update(ctx context.Context, notification *model.Notification) error {
	return r.db.WithContext(ctx).Save(notification).Error
}

// UpdateStatus 更新通知状态
func (r *notificationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if status == model.StatusSent {
		updates["sent_at"] = time.Now()
	}

	return r.db.WithContext(ctx).
		Model(&model.Notification{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// Delete 删除通知
func (r *notificationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Notification{}, "id = ?", id).Error
}

// CreateTemplate 创建模板
func (r *notificationRepository) CreateTemplate(ctx context.Context, template *model.NotificationTemplate) error {
	return r.db.WithContext(ctx).Create(template).Error
}

// GetTemplate 获取模板
func (r *notificationRepository) GetTemplate(ctx context.Context, code string, merchantID *uuid.UUID) (*model.NotificationTemplate, error) {
	var template model.NotificationTemplate
	db := r.db.WithContext(ctx).Where("code = ? AND is_enabled = true", code)

	// 优先查找商户模板，如果没有则查找系统模板
	if merchantID != nil {
		db = db.Where("(merchant_id = ? OR is_system = true)", *merchantID).
			Order("is_system ASC") // 商户模板优先
	} else {
		db = db.Where("is_system = true")
	}

	err := db.First(&template).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &template, nil
}

// ListTemplates 列出模板
func (r *notificationRepository) ListTemplates(ctx context.Context, merchantID *uuid.UUID) ([]*model.NotificationTemplate, error) {
	var templates []*model.NotificationTemplate
	db := r.db.WithContext(ctx)

	if merchantID != nil {
		db = db.Where("merchant_id = ? OR is_system = true", *merchantID)
	} else {
		db = db.Where("is_system = true")
	}

	err := db.Order("is_system ASC, created_at DESC").Find(&templates).Error
	return templates, err
}

// UpdateTemplate 更新模板
func (r *notificationRepository) UpdateTemplate(ctx context.Context, template *model.NotificationTemplate) error {
	return r.db.WithContext(ctx).Save(template).Error
}

// DeleteTemplate 删除模板
func (r *notificationRepository) DeleteTemplate(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.NotificationTemplate{}, "id = ?", id).Error
}

// CreateEndpoint 创建 Webhook 端点
func (r *notificationRepository) CreateEndpoint(ctx context.Context, endpoint *model.WebhookEndpoint) error {
	return r.db.WithContext(ctx).Create(endpoint).Error
}

// GetEndpoint 获取 Webhook 端点
func (r *notificationRepository) GetEndpoint(ctx context.Context, id uuid.UUID) (*model.WebhookEndpoint, error) {
	var endpoint model.WebhookEndpoint
	err := r.db.WithContext(ctx).First(&endpoint, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &endpoint, nil
}

// ListEndpoints 列出 Webhook 端点
func (r *notificationRepository) ListEndpoints(ctx context.Context, merchantID uuid.UUID) ([]*model.WebhookEndpoint, error) {
	var endpoints []*model.WebhookEndpoint
	err := r.db.WithContext(ctx).
		Where("merchant_id = ?", merchantID).
		Order("created_at DESC").
		Find(&endpoints).Error
	return endpoints, err
}

// UpdateEndpoint 更新 Webhook 端点
func (r *notificationRepository) UpdateEndpoint(ctx context.Context, endpoint *model.WebhookEndpoint) error {
	return r.db.WithContext(ctx).Save(endpoint).Error
}

// DeleteEndpoint 删除 Webhook 端点
func (r *notificationRepository) DeleteEndpoint(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.WebhookEndpoint{}, "id = ?", id).Error
}

// CreateDelivery 创建投递记录
func (r *notificationRepository) CreateDelivery(ctx context.Context, delivery *model.WebhookDelivery) error {
	return r.db.WithContext(ctx).Create(delivery).Error
}

// GetDelivery 获取投递记录
func (r *notificationRepository) GetDelivery(ctx context.Context, id uuid.UUID) (*model.WebhookDelivery, error) {
	var delivery model.WebhookDelivery
	err := r.db.WithContext(ctx).First(&delivery, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &delivery, nil
}

// ListDeliveries 列出投递记录
func (r *notificationRepository) ListDeliveries(ctx context.Context, query *DeliveryQuery) ([]*model.WebhookDelivery, int64, error) {
	var deliveries []*model.WebhookDelivery
	var total int64

	db := r.db.WithContext(ctx).Model(&model.WebhookDelivery{})

	// 构建查询条件
	if query.EndpointID != nil {
		db = db.Where("endpoint_id = ?", *query.EndpointID)
	}
	if query.MerchantID != nil {
		db = db.Where("merchant_id = ?", *query.MerchantID)
	}
	if query.EventType != "" {
		db = db.Where("event_type = ?", query.EventType)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.StartTime != nil {
		db = db.Where("created_at >= ?", *query.StartTime)
	}
	if query.EndTime != nil {
		db = db.Where("created_at <= ?", *query.EndTime)
	}

	// 计算总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		query.PageSize = 20
	}
	offset := (query.Page - 1) * query.PageSize

	err := db.Order("created_at DESC").
		Offset(offset).
		Limit(query.PageSize).
		Find(&deliveries).Error

	return deliveries, total, err
}

// UpdateDelivery 更新投递记录
func (r *notificationRepository) UpdateDelivery(ctx context.Context, delivery *model.WebhookDelivery) error {
	return r.db.WithContext(ctx).Save(delivery).Error
}

// ListPendingNotifications 列出待处理的通知
func (r *notificationRepository) ListPendingNotifications(ctx context.Context, limit int) ([]*model.Notification, error) {
	var notifications []*model.Notification

	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("status = ? AND retry_count < max_retry", model.StatusPending).
		Where("scheduled_at IS NULL OR scheduled_at <= ?", now).
		Order("priority DESC, created_at ASC").
		Limit(limit).
		Find(&notifications).Error

	return notifications, err
}

// ListPendingDeliveries 列出待重试的投递
func (r *notificationRepository) ListPendingDeliveries(ctx context.Context, limit int) ([]*model.WebhookDelivery, error) {
	var deliveries []*model.WebhookDelivery

	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("status IN (?, ?)", model.DeliveryStatusPending, model.DeliveryStatusRetrying).
		Where("next_retry_at IS NULL OR next_retry_at <= ?", now).
		Order("created_at ASC").
		Limit(limit).
		Find(&deliveries).Error

	return deliveries, err
}

// CreatePreference 创建通知偏好
func (r *notificationRepository) CreatePreference(ctx context.Context, preference *model.NotificationPreference) error {
	return r.db.WithContext(ctx).Create(preference).Error
}

// GetPreference 获取通知偏好
func (r *notificationRepository) GetPreference(ctx context.Context, id uuid.UUID) (*model.NotificationPreference, error) {
	var preference model.NotificationPreference
	err := r.db.WithContext(ctx).First(&preference, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &preference, nil
}

// ListPreferences 列出通知偏好
func (r *notificationRepository) ListPreferences(ctx context.Context, merchantID uuid.UUID, userID *uuid.UUID) ([]*model.NotificationPreference, error) {
	var preferences []*model.NotificationPreference
	db := r.db.WithContext(ctx).Where("merchant_id = ?", merchantID)

	if userID != nil {
		db = db.Where("user_id = ?", *userID)
	}

	err := db.Order("created_at DESC").Find(&preferences).Error
	return preferences, err
}

// UpdatePreference 更新通知偏好
func (r *notificationRepository) UpdatePreference(ctx context.Context, preference *model.NotificationPreference) error {
	return r.db.WithContext(ctx).Save(preference).Error
}

// DeletePreference 删除通知偏好
func (r *notificationRepository) DeletePreference(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.NotificationPreference{}, "id = ?", id).Error
}

// CheckPreference 检查是否允许发送通知
func (r *notificationRepository) CheckPreference(ctx context.Context, merchantID uuid.UUID, userID *uuid.UUID, channel, eventType string) (bool, error) {
	var preference model.NotificationPreference
	db := r.db.WithContext(ctx).
		Where("merchant_id = ? AND channel = ? AND event_type = ?", merchantID, channel, eventType)

	if userID != nil {
		db = db.Where("user_id = ?", *userID)
	}

	err := db.First(&preference).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 没有找到偏好设置，默认允许发送
			return true, nil
		}
		return false, err
	}

	// 返回偏好设置的 is_enabled 状态
	return preference.IsEnabled, nil
}
