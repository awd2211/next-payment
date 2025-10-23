package repository

import (
	"context"

	"github.com/google/uuid"
	"payment-platform/channel-adapter/internal/model"
	"gorm.io/gorm"
)

// ChannelRepository 渠道配置仓储接口
type ChannelRepository interface {
	// 配置管理
	GetConfig(ctx context.Context, merchantID uuid.UUID, channel string) (*model.ChannelConfig, error)
	GetConfigByID(ctx context.Context, id uuid.UUID) (*model.ChannelConfig, error)
	ListConfigs(ctx context.Context, merchantID uuid.UUID) ([]*model.ChannelConfig, error)
	CreateConfig(ctx context.Context, config *model.ChannelConfig) error
	UpdateConfig(ctx context.Context, config *model.ChannelConfig) error
	DeleteConfig(ctx context.Context, id uuid.UUID) error

	// 交易记录管理
	CreateTransaction(ctx context.Context, tx *model.Transaction) error
	GetTransaction(ctx context.Context, paymentNo string) (*model.Transaction, error)
	GetTransactionByChannelTradeNo(ctx context.Context, channelTradeNo string) (*model.Transaction, error)
	UpdateTransaction(ctx context.Context, tx *model.Transaction) error
	ListTransactions(ctx context.Context, query *TransactionQuery) ([]*model.Transaction, int64, error)

	// Webhook 日志管理
	CreateWebhookLog(ctx context.Context, log *model.WebhookLog) error
	GetWebhookLog(ctx context.Context, eventID string) (*model.WebhookLog, error)
	UpdateWebhookLog(ctx context.Context, log *model.WebhookLog) error
	ListUnprocessedWebhooks(ctx context.Context, limit int) ([]*model.WebhookLog, error)
}

type channelRepository struct {
	db *gorm.DB
}

// NewChannelRepository 创建渠道配置仓储实例
func NewChannelRepository(db *gorm.DB) ChannelRepository {
	return &channelRepository{db: db}
}

// TransactionQuery 交易查询参数
type TransactionQuery struct {
	MerchantID      *uuid.UUID
	Channel         string
	TransactionType string
	Status          string
	Page            int
	PageSize        int
}

// GetConfig 获取渠道配置
func (r *channelRepository) GetConfig(ctx context.Context, merchantID uuid.UUID, channel string) (*model.ChannelConfig, error) {
	var config model.ChannelConfig
	err := r.db.WithContext(ctx).
		Where("merchant_id = ? AND channel = ? AND is_enabled = true", merchantID, channel).
		First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

// GetConfigByID 根据ID获取渠道配置
func (r *channelRepository) GetConfigByID(ctx context.Context, id uuid.UUID) (*model.ChannelConfig, error) {
	var config model.ChannelConfig
	err := r.db.WithContext(ctx).First(&config, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

// ListConfigs 列出商户的所有渠道配置
func (r *channelRepository) ListConfigs(ctx context.Context, merchantID uuid.UUID) ([]*model.ChannelConfig, error) {
	var configs []*model.ChannelConfig
	err := r.db.WithContext(ctx).
		Where("merchant_id = ?", merchantID).
		Order("priority DESC, created_at DESC").
		Find(&configs).Error
	return configs, err
}

// CreateConfig 创建渠道配置
func (r *channelRepository) CreateConfig(ctx context.Context, config *model.ChannelConfig) error {
	return r.db.WithContext(ctx).Create(config).Error
}

// UpdateConfig 更新渠道配置
func (r *channelRepository) UpdateConfig(ctx context.Context, config *model.ChannelConfig) error {
	return r.db.WithContext(ctx).Save(config).Error
}

// DeleteConfig 删除渠道配置
func (r *channelRepository) DeleteConfig(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.ChannelConfig{}, "id = ?", id).Error
}

// CreateTransaction 创建交易记录
func (r *channelRepository) CreateTransaction(ctx context.Context, tx *model.Transaction) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

// GetTransaction 根据支付流水号获取交易记录
func (r *channelRepository) GetTransaction(ctx context.Context, paymentNo string) (*model.Transaction, error) {
	var tx model.Transaction
	err := r.db.WithContext(ctx).
		Where("payment_no = ?", paymentNo).
		First(&tx).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &tx, nil
}

// GetTransactionByChannelTradeNo 根据渠道交易号获取交易记录
func (r *channelRepository) GetTransactionByChannelTradeNo(ctx context.Context, channelTradeNo string) (*model.Transaction, error) {
	var tx model.Transaction
	err := r.db.WithContext(ctx).
		Where("channel_trade_no = ?", channelTradeNo).
		First(&tx).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &tx, nil
}

// UpdateTransaction 更新交易记录
func (r *channelRepository) UpdateTransaction(ctx context.Context, tx *model.Transaction) error {
	return r.db.WithContext(ctx).Save(tx).Error
}

// ListTransactions 查询交易记录列表
func (r *channelRepository) ListTransactions(ctx context.Context, query *TransactionQuery) ([]*model.Transaction, int64, error) {
	var transactions []*model.Transaction
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Transaction{})

	// 构建查询条件
	if query.MerchantID != nil {
		db = db.Where("merchant_id = ?", *query.MerchantID)
	}
	if query.Channel != "" {
		db = db.Where("channel = ?", query.Channel)
	}
	if query.TransactionType != "" {
		db = db.Where("transaction_type = ?", query.TransactionType)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
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
		Find(&transactions).Error

	return transactions, total, err
}

// CreateWebhookLog 创建 Webhook 日志
func (r *channelRepository) CreateWebhookLog(ctx context.Context, log *model.WebhookLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// GetWebhookLog 根据事件ID获取 Webhook 日志
func (r *channelRepository) GetWebhookLog(ctx context.Context, eventID string) (*model.WebhookLog, error) {
	var log model.WebhookLog
	err := r.db.WithContext(ctx).
		Where("event_id = ?", eventID).
		First(&log).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &log, nil
}

// UpdateWebhookLog 更新 Webhook 日志
func (r *channelRepository) UpdateWebhookLog(ctx context.Context, log *model.WebhookLog) error {
	return r.db.WithContext(ctx).Save(log).Error
}

// ListUnprocessedWebhooks 获取未处理的 Webhook 列表
func (r *channelRepository) ListUnprocessedWebhooks(ctx context.Context, limit int) ([]*model.WebhookLog, error) {
	var logs []*model.WebhookLog
	err := r.db.WithContext(ctx).
		Where("is_processed = false AND is_verified = true AND retry_count < ?", 3).
		Order("created_at ASC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}
