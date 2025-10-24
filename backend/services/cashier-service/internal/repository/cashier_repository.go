package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"payment-platform/cashier-service/internal/model"
)

// CashierRepository 收银台仓储接口
type CashierRepository interface {
	// 配置管理
	CreateOrUpdateConfig(ctx context.Context, config *model.CashierConfig) error
	GetConfig(ctx context.Context, merchantID uuid.UUID) (*model.CashierConfig, error)
	DeleteConfig(ctx context.Context, merchantID uuid.UUID) error

	// 会话管理
	CreateSession(ctx context.Context, session *model.CashierSession) error
	GetSession(ctx context.Context, sessionToken string) (*model.CashierSession, error)
	GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*model.CashierSession, error)
	UpdateSession(ctx context.Context, session *model.CashierSession) error
	DeleteSession(ctx context.Context, sessionToken string) error
	ExpireSessions(ctx context.Context) error

	// 日志管理
	CreateLog(ctx context.Context, log *model.CashierLog) error
	GetLogs(ctx context.Context, merchantID uuid.UUID, limit int) ([]*model.CashierLog, error)
	GetLogsBySession(ctx context.Context, sessionID uuid.UUID) ([]*model.CashierLog, error)

	// 模板管理
	CreateTemplate(ctx context.Context, template *model.CashierTemplate) error
	GetTemplate(ctx context.Context, id uuid.UUID) (*model.CashierTemplate, error)
	ListTemplates(ctx context.Context) ([]*model.CashierTemplate, error)
	UpdateTemplate(ctx context.Context, template *model.CashierTemplate) error
	DeleteTemplate(ctx context.Context, id uuid.UUID) error

	// 统计分析
	GetConversionRate(ctx context.Context, merchantID uuid.UUID, startTime, endTime time.Time) (float64, error)
	GetChannelStats(ctx context.Context, merchantID uuid.UUID, startTime, endTime time.Time) (map[string]int, error)

	// 平台统计
	GetActiveMerchantCount(ctx context.Context) (int, error)
	GetSessionCount(ctx context.Context, startTime, endTime time.Time) (int64, error)
	GetCompletedSessionCount(ctx context.Context, startTime, endTime time.Time) (int64, error)
	GetAverageConversionRate(ctx context.Context, startTime, endTime time.Time) (float64, error)
}

type cashierRepository struct {
	db *gorm.DB
}

// NewCashierRepository 创建收银台仓储实例
func NewCashierRepository(db *gorm.DB) CashierRepository {
	return &cashierRepository{db: db}
}

// CreateOrUpdateConfig 创建或更新配置
func (r *cashierRepository) CreateOrUpdateConfig(ctx context.Context, config *model.CashierConfig) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "merchant_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"theme_color", "logo_url", "background_image_url", "custom_css",
				"enabled_channels", "default_channel", "enabled_languages", "default_language",
				"auto_submit", "show_amount_breakdown", "allow_channel_switch", "session_timeout_minutes",
				"require_cvv", "enable_3d_secure", "allowed_countries",
				"success_redirect_url", "cancel_redirect_url", "updated_at",
			}),
		}).
		Create(config).Error
}

// GetConfig 获取配置
func (r *cashierRepository) GetConfig(ctx context.Context, merchantID uuid.UUID) (*model.CashierConfig, error) {
	var config model.CashierConfig
	err := r.db.WithContext(ctx).Where("merchant_id = ?", merchantID).First(&config).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &config, err
}

// DeleteConfig 删除配置
func (r *cashierRepository) DeleteConfig(ctx context.Context, merchantID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("merchant_id = ?", merchantID).Delete(&model.CashierConfig{}).Error
}

// CreateSession 创建会话
func (r *cashierRepository) CreateSession(ctx context.Context, session *model.CashierSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

// GetSession 通过token获取会话
func (r *cashierRepository) GetSession(ctx context.Context, sessionToken string) (*model.CashierSession, error) {
	var session model.CashierSession
	err := r.db.WithContext(ctx).Where("session_token = ?", sessionToken).First(&session).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &session, err
}

// GetSessionByID 通过ID获取会话
func (r *cashierRepository) GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*model.CashierSession, error) {
	var session model.CashierSession
	err := r.db.WithContext(ctx).Where("id = ?", sessionID).First(&session).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &session, err
}

// UpdateSession 更新会话
func (r *cashierRepository) UpdateSession(ctx context.Context, session *model.CashierSession) error {
	return r.db.WithContext(ctx).Save(session).Error
}

// DeleteSession 删除会话
func (r *cashierRepository) DeleteSession(ctx context.Context, sessionToken string) error {
	return r.db.WithContext(ctx).Where("session_token = ?", sessionToken).Delete(&model.CashierSession{}).Error
}

// ExpireSessions 过期会话清理
func (r *cashierRepository) ExpireSessions(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Model(&model.CashierSession{}).
		Where("status = ? AND expires_at < ?", "pending", time.Now()).
		Update("status", "expired").Error
}

// CreateLog 创建日志
func (r *cashierRepository) CreateLog(ctx context.Context, log *model.CashierLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// GetLogs 获取日志列表
func (r *cashierRepository) GetLogs(ctx context.Context, merchantID uuid.UUID, limit int) ([]*model.CashierLog, error) {
	var logs []*model.CashierLog
	err := r.db.WithContext(ctx).
		Where("merchant_id = ?", merchantID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// GetLogsBySession 获取会话日志
func (r *cashierRepository) GetLogsBySession(ctx context.Context, sessionID uuid.UUID) ([]*model.CashierLog, error) {
	var logs []*model.CashierLog
	err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Find(&logs).Error
	return logs, err
}

// CreateTemplate 创建模板
func (r *cashierRepository) CreateTemplate(ctx context.Context, template *model.CashierTemplate) error {
	return r.db.WithContext(ctx).Create(template).Error
}

// GetTemplate 获取模板
func (r *cashierRepository) GetTemplate(ctx context.Context, id uuid.UUID) (*model.CashierTemplate, error) {
	var template model.CashierTemplate
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&template).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &template, err
}

// ListTemplates 获取模板列表
func (r *cashierRepository) ListTemplates(ctx context.Context) ([]*model.CashierTemplate, error) {
	var templates []*model.CashierTemplate
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&templates).Error
	return templates, err
}

// UpdateTemplate 更新模板
func (r *cashierRepository) UpdateTemplate(ctx context.Context, template *model.CashierTemplate) error {
	return r.db.WithContext(ctx).Save(template).Error
}

// DeleteTemplate 删除模板
func (r *cashierRepository) DeleteTemplate(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.CashierTemplate{}, id).Error
}

// GetConversionRate 获取转化率
func (r *cashierRepository) GetConversionRate(ctx context.Context, merchantID uuid.UUID, startTime, endTime time.Time) (float64, error) {
	var total, completed int64

	// 总会话数
	r.db.WithContext(ctx).
		Model(&model.CashierSession{}).
		Where("merchant_id = ? AND created_at BETWEEN ? AND ?", merchantID, startTime, endTime).
		Count(&total)

	// 已完成会话数
	r.db.WithContext(ctx).
		Model(&model.CashierSession{}).
		Where("merchant_id = ? AND status = ? AND created_at BETWEEN ? AND ?", merchantID, "completed", startTime, endTime).
		Count(&completed)

	if total == 0 {
		return 0, nil
	}

	return float64(completed) / float64(total) * 100, nil
}

// GetChannelStats 获取渠道统计
func (r *cashierRepository) GetChannelStats(ctx context.Context, merchantID uuid.UUID, startTime, endTime time.Time) (map[string]int, error) {
	type Result struct {
		Channel string
		Count   int
	}

	var results []Result
	err := r.db.WithContext(ctx).
		Model(&model.CashierLog{}).
		Select("selected_channel as channel, COUNT(*) as count").
		Where("merchant_id = ? AND created_at BETWEEN ? AND ? AND selected_channel != ''", merchantID, startTime, endTime).
		Group("selected_channel").
		Find(&results).Error

	stats := make(map[string]int)
	for _, r := range results {
		stats[r.Channel] = r.Count
	}

	return stats, err
}

// GetActiveMerchantCount 获取活跃商户数
func (r *cashierRepository) GetActiveMerchantCount(ctx context.Context) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.CashierConfig{}).
		Count(&count).Error
	return int(count), err
}

// GetSessionCount 获取会话总数
func (r *cashierRepository) GetSessionCount(ctx context.Context, startTime, endTime time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.CashierSession{}).
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Count(&count).Error
	return count, err
}

// GetCompletedSessionCount 获取已完成会话数
func (r *cashierRepository) GetCompletedSessionCount(ctx context.Context, startTime, endTime time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.CashierSession{}).
		Where("status = ? AND created_at BETWEEN ? AND ?", "completed", startTime, endTime).
		Count(&count).Error
	return count, err
}

// GetAverageConversionRate 获取平均转化率
func (r *cashierRepository) GetAverageConversionRate(ctx context.Context, startTime, endTime time.Time) (float64, error) {
	var total, completed int64

	// 总会话数
	r.db.WithContext(ctx).
		Model(&model.CashierSession{}).
		Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Count(&total)

	// 已完成会话数
	r.db.WithContext(ctx).
		Model(&model.CashierSession{}).
		Where("status = ? AND created_at BETWEEN ? AND ?", "completed", startTime, endTime).
		Count(&completed)

	if total == 0 {
		return 0, nil
	}

	return float64(completed) / float64(total) * 100, nil
}
