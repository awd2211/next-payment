package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"payment-platform/merchant-service/internal/model"
	"gorm.io/gorm"
)

// SecurityRepository 安全仓储接口
type SecurityRepository interface {
	// 安全设置
	GetSecuritySettings(ctx context.Context, merchantID uuid.UUID) (*model.SecuritySettings, error)
	CreateSecuritySettings(ctx context.Context, settings *model.SecuritySettings) error
	UpdateSecuritySettings(ctx context.Context, settings *model.SecuritySettings) error

	// 密码历史
	CreatePasswordHistory(ctx context.Context, history *model.PasswordHistory) error
	GetPasswordHistory(ctx context.Context, merchantID uuid.UUID, limit int) ([]*model.PasswordHistory, error)

	// 2FA
	GetTwoFactorAuth(ctx context.Context, merchantID uuid.UUID) (*model.TwoFactorAuth, error)
	CreateTwoFactorAuth(ctx context.Context, tfa *model.TwoFactorAuth) error
	UpdateTwoFactorAuth(ctx context.Context, tfa *model.TwoFactorAuth) error
	DeleteTwoFactorAuth(ctx context.Context, merchantID uuid.UUID) error

	// 登录活动
	CreateLoginActivity(ctx context.Context, activity *model.LoginActivity) error
	GetLoginActivities(ctx context.Context, merchantID uuid.UUID, page, pageSize int) ([]*model.LoginActivity, int64, error)
	GetRecentLoginActivities(ctx context.Context, merchantID uuid.UUID, limit int) ([]*model.LoginActivity, error)

	// 会话
	CreateSession(ctx context.Context, session *model.Session) error
	GetSession(ctx context.Context, sessionID string) (*model.Session, error)
	GetActiveSessions(ctx context.Context, merchantID uuid.UUID) ([]*model.Session, error)
	UpdateSession(ctx context.Context, session *model.Session) error
	DeleteSession(ctx context.Context, sessionID string) error
	DeleteExpiredSessions(ctx context.Context) error
	CountActiveSessions(ctx context.Context, merchantID uuid.UUID) (int64, error)
}

type securityRepository struct {
	db *gorm.DB
}

// NewSecurityRepository 创建安全仓储实例
func NewSecurityRepository(db *gorm.DB) SecurityRepository {
	return &securityRepository{db: db}
}

// GetSecuritySettings 获取安全设置
func (r *securityRepository) GetSecuritySettings(ctx context.Context, merchantID uuid.UUID) (*model.SecuritySettings, error) {
	var settings model.SecuritySettings
	err := r.db.WithContext(ctx).
		Where("merchant_id = ?", merchantID).
		First(&settings).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &settings, nil
}

// CreateSecuritySettings 创建安全设置
func (r *securityRepository) CreateSecuritySettings(ctx context.Context, settings *model.SecuritySettings) error {
	return r.db.WithContext(ctx).Create(settings).Error
}

// UpdateSecuritySettings 更新安全设置
func (r *securityRepository) UpdateSecuritySettings(ctx context.Context, settings *model.SecuritySettings) error {
	return r.db.WithContext(ctx).Save(settings).Error
}

// CreatePasswordHistory 创建密码历史
func (r *securityRepository) CreatePasswordHistory(ctx context.Context, history *model.PasswordHistory) error {
	return r.db.WithContext(ctx).Create(history).Error
}

// GetPasswordHistory 获取密码历史
func (r *securityRepository) GetPasswordHistory(ctx context.Context, merchantID uuid.UUID, limit int) ([]*model.PasswordHistory, error) {
	var history []*model.PasswordHistory
	err := r.db.WithContext(ctx).
		Where("merchant_id = ?", merchantID).
		Order("created_at DESC").
		Limit(limit).
		Find(&history).Error
	return history, err
}

// GetTwoFactorAuth 获取2FA配置
func (r *securityRepository) GetTwoFactorAuth(ctx context.Context, merchantID uuid.UUID) (*model.TwoFactorAuth, error) {
	var tfa model.TwoFactorAuth
	err := r.db.WithContext(ctx).
		Where("merchant_id = ?", merchantID).
		First(&tfa).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &tfa, nil
}

// CreateTwoFactorAuth 创建2FA配置
func (r *securityRepository) CreateTwoFactorAuth(ctx context.Context, tfa *model.TwoFactorAuth) error {
	return r.db.WithContext(ctx).Create(tfa).Error
}

// UpdateTwoFactorAuth 更新2FA配置
func (r *securityRepository) UpdateTwoFactorAuth(ctx context.Context, tfa *model.TwoFactorAuth) error {
	return r.db.WithContext(ctx).Save(tfa).Error
}

// DeleteTwoFactorAuth 删除2FA配置
func (r *securityRepository) DeleteTwoFactorAuth(ctx context.Context, merchantID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("merchant_id = ?", merchantID).
		Delete(&model.TwoFactorAuth{}).Error
}

// CreateLoginActivity 创建登录活动记录
func (r *securityRepository) CreateLoginActivity(ctx context.Context, activity *model.LoginActivity) error {
	return r.db.WithContext(ctx).Create(activity).Error
}

// GetLoginActivities 分页获取登录活动
func (r *securityRepository) GetLoginActivities(ctx context.Context, merchantID uuid.UUID, page, pageSize int) ([]*model.LoginActivity, int64, error) {
	var activities []*model.LoginActivity
	var total int64

	query := r.db.WithContext(ctx).Model(&model.LoginActivity{}).Where("merchant_id = ?", merchantID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("login_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&activities).Error

	return activities, total, err
}

// GetRecentLoginActivities 获取最近的登录活动
func (r *securityRepository) GetRecentLoginActivities(ctx context.Context, merchantID uuid.UUID, limit int) ([]*model.LoginActivity, error) {
	var activities []*model.LoginActivity
	err := r.db.WithContext(ctx).
		Where("merchant_id = ? AND status = ?", merchantID, model.LoginStatusSuccess).
		Order("login_at DESC").
		Limit(limit).
		Find(&activities).Error
	return activities, err
}

// CreateSession 创建会话
func (r *securityRepository) CreateSession(ctx context.Context, session *model.Session) error {
	return r.db.WithContext(ctx).Create(session).Error
}

// GetSession 获取会话
func (r *securityRepository) GetSession(ctx context.Context, sessionID string) (*model.Session, error) {
	var session model.Session
	err := r.db.WithContext(ctx).
		Where("session_id = ? AND is_active = ?", sessionID, true).
		First(&session).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

// GetActiveSessions 获取活跃会话
func (r *securityRepository) GetActiveSessions(ctx context.Context, merchantID uuid.UUID) ([]*model.Session, error) {
	var sessions []*model.Session
	err := r.db.WithContext(ctx).
		Where("merchant_id = ? AND is_active = ? AND expires_at > ?", merchantID, true, time.Now()).
		Order("last_seen_at DESC").
		Find(&sessions).Error
	return sessions, err
}

// UpdateSession 更新会话
func (r *securityRepository) UpdateSession(ctx context.Context, session *model.Session) error {
	return r.db.WithContext(ctx).Save(session).Error
}

// DeleteSession 删除会话
func (r *securityRepository) DeleteSession(ctx context.Context, sessionID string) error {
	return r.db.WithContext(ctx).
		Model(&model.Session{}).
		Where("session_id = ?", sessionID).
		Update("is_active", false).Error
}

// DeleteExpiredSessions 删除过期会话
func (r *securityRepository) DeleteExpiredSessions(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Model(&model.Session{}).
		Where("expires_at < ? OR is_active = ?", time.Now(), false).
		Update("is_active", false).Error
}

// CountActiveSessions 统计活跃会话数
func (r *securityRepository) CountActiveSessions(ctx context.Context, merchantID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Session{}).
		Where("merchant_id = ? AND is_active = ? AND expires_at > ?", merchantID, true, time.Now()).
		Count(&count).Error
	return count, err
}
