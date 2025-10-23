package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/services/admin-service/internal/model"
	"gorm.io/gorm"
)

// SecurityRepository 安全功能仓储接口
type SecurityRepository interface {
	// 2FA管理
	Create2FA(ctx context.Context, tfa *model.TwoFactorAuth) error
	Get2FA(ctx context.Context, userID uuid.UUID, userType string) (*model.TwoFactorAuth, error)
	Update2FA(ctx context.Context, tfa *model.TwoFactorAuth) error
	Delete2FA(ctx context.Context, userID uuid.UUID, userType string) error

	// 登录活动
	CreateLoginActivity(ctx context.Context, activity *model.LoginActivity) error
	GetLoginActivities(ctx context.Context, userID uuid.UUID, userType string, limit int) ([]*model.LoginActivity, error)
	GetLoginActivitiesByTimeRange(ctx context.Context, userID uuid.UUID, userType string, startTime, endTime time.Time) ([]*model.LoginActivity, error)
	GetAbnormalActivities(ctx context.Context, userID uuid.UUID, userType string, limit int) ([]*model.LoginActivity, error)
	UpdateLoginActivity(ctx context.Context, activity *model.LoginActivity) error

	// 安全设置
	CreateSecuritySettings(ctx context.Context, settings *model.SecuritySettings) error
	GetSecuritySettings(ctx context.Context, userID uuid.UUID, userType string) (*model.SecuritySettings, error)
	UpdateSecuritySettings(ctx context.Context, settings *model.SecuritySettings) error

	// 密码历史
	CreatePasswordHistory(ctx context.Context, history *model.PasswordHistory) error
	GetPasswordHistory(ctx context.Context, userID uuid.UUID, userType string, limit int) ([]*model.PasswordHistory, error)
	CheckPasswordInHistory(ctx context.Context, userID uuid.UUID, userType string, passwordHash string) (bool, error)

	// 会话管理
	CreateSession(ctx context.Context, session *model.Session) error
	GetSession(ctx context.Context, sessionID string) (*model.Session, error)
	GetActiveSessions(ctx context.Context, userID uuid.UUID, userType string) ([]*model.Session, error)
	UpdateSession(ctx context.Context, session *model.Session) error
	DeactivateSession(ctx context.Context, sessionID string) error
	DeactivateAllSessions(ctx context.Context, userID uuid.UUID, userType string) error
	CleanupExpiredSessions(ctx context.Context) error
}

type securityRepository struct {
	db *gorm.DB
}

// NewSecurityRepository 创建安全功能仓储实例
func NewSecurityRepository(db *gorm.DB) SecurityRepository {
	return &securityRepository{db: db}
}

// Create2FA 创建2FA记录
func (r *securityRepository) Create2FA(ctx context.Context, tfa *model.TwoFactorAuth) error {
	return r.db.WithContext(ctx).Create(tfa).Error
}

// Get2FA 获取2FA记录
func (r *securityRepository) Get2FA(ctx context.Context, userID uuid.UUID, userType string) (*model.TwoFactorAuth, error) {
	var tfa model.TwoFactorAuth
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND user_type = ?", userID, userType).
		First(&tfa).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &tfa, nil
}

// Update2FA 更新2FA记录
func (r *securityRepository) Update2FA(ctx context.Context, tfa *model.TwoFactorAuth) error {
	return r.db.WithContext(ctx).Save(tfa).Error
}

// Delete2FA 删除2FA记录
func (r *securityRepository) Delete2FA(ctx context.Context, userID uuid.UUID, userType string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND user_type = ?", userID, userType).
		Delete(&model.TwoFactorAuth{}).Error
}

// CreateLoginActivity 创建登录活动记录
func (r *securityRepository) CreateLoginActivity(ctx context.Context, activity *model.LoginActivity) error {
	return r.db.WithContext(ctx).Create(activity).Error
}

// GetLoginActivities 获取登录活动记录
func (r *securityRepository) GetLoginActivities(ctx context.Context, userID uuid.UUID, userType string, limit int) ([]*model.LoginActivity, error) {
	var activities []*model.LoginActivity
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND user_type = ?", userID, userType).
		Order("login_at DESC").
		Limit(limit).
		Find(&activities).Error
	return activities, err
}

// GetLoginActivitiesByTimeRange 按时间范围获取登录活动
func (r *securityRepository) GetLoginActivitiesByTimeRange(ctx context.Context, userID uuid.UUID, userType string, startTime, endTime time.Time) ([]*model.LoginActivity, error) {
	var activities []*model.LoginActivity
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND user_type = ? AND login_at BETWEEN ? AND ?", userID, userType, startTime, endTime).
		Order("login_at DESC").
		Find(&activities).Error
	return activities, err
}

// GetAbnormalActivities 获取异常登录活动
func (r *securityRepository) GetAbnormalActivities(ctx context.Context, userID uuid.UUID, userType string, limit int) ([]*model.LoginActivity, error) {
	var activities []*model.LoginActivity
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND user_type = ? AND is_abnormal = true", userID, userType).
		Order("login_at DESC").
		Limit(limit).
		Find(&activities).Error
	return activities, err
}

// UpdateLoginActivity 更新登录活动记录
func (r *securityRepository) UpdateLoginActivity(ctx context.Context, activity *model.LoginActivity) error {
	return r.db.WithContext(ctx).Save(activity).Error
}

// CreateSecuritySettings 创建安全设置
func (r *securityRepository) CreateSecuritySettings(ctx context.Context, settings *model.SecuritySettings) error {
	return r.db.WithContext(ctx).Create(settings).Error
}

// GetSecuritySettings 获取安全设置
func (r *securityRepository) GetSecuritySettings(ctx context.Context, userID uuid.UUID, userType string) (*model.SecuritySettings, error) {
	var settings model.SecuritySettings
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND user_type = ?", userID, userType).
		First(&settings).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &settings, nil
}

// UpdateSecuritySettings 更新安全设置
func (r *securityRepository) UpdateSecuritySettings(ctx context.Context, settings *model.SecuritySettings) error {
	return r.db.WithContext(ctx).Save(settings).Error
}

// CreatePasswordHistory 创建密码历史记录
func (r *securityRepository) CreatePasswordHistory(ctx context.Context, history *model.PasswordHistory) error {
	return r.db.WithContext(ctx).Create(history).Error
}

// GetPasswordHistory 获取密码历史记录
func (r *securityRepository) GetPasswordHistory(ctx context.Context, userID uuid.UUID, userType string, limit int) ([]*model.PasswordHistory, error) {
	var history []*model.PasswordHistory
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND user_type = ?", userID, userType).
		Order("created_at DESC").
		Limit(limit).
		Find(&history).Error
	return history, err
}

// CheckPasswordInHistory 检查密码是否在历史记录中
func (r *securityRepository) CheckPasswordInHistory(ctx context.Context, userID uuid.UUID, userType string, passwordHash string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.PasswordHistory{}).
		Where("user_id = ? AND user_type = ? AND password_hash = ?", userID, userType, passwordHash).
		Count(&count).Error
	return count > 0, err
}

// CreateSession 创建会话
func (r *securityRepository) CreateSession(ctx context.Context, session *model.Session) error {
	return r.db.WithContext(ctx).Create(session).Error
}

// GetSession 获取会话
func (r *securityRepository) GetSession(ctx context.Context, sessionID string) (*model.Session, error) {
	var session model.Session
	err := r.db.WithContext(ctx).
		Where("session_id = ? AND is_active = true AND expires_at > NOW()", sessionID).
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
func (r *securityRepository) GetActiveSessions(ctx context.Context, userID uuid.UUID, userType string) ([]*model.Session, error) {
	var sessions []*model.Session
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND user_type = ? AND is_active = true AND expires_at > NOW()", userID, userType).
		Order("last_seen_at DESC").
		Find(&sessions).Error
	return sessions, err
}

// UpdateSession 更新会话
func (r *securityRepository) UpdateSession(ctx context.Context, session *model.Session) error {
	return r.db.WithContext(ctx).Save(session).Error
}

// DeactivateSession 停用会话
func (r *securityRepository) DeactivateSession(ctx context.Context, sessionID string) error {
	return r.db.WithContext(ctx).
		Model(&model.Session{}).
		Where("session_id = ?", sessionID).
		Update("is_active", false).Error
}

// DeactivateAllSessions 停用所有会话
func (r *securityRepository) DeactivateAllSessions(ctx context.Context, userID uuid.UUID, userType string) error {
	return r.db.WithContext(ctx).
		Model(&model.Session{}).
		Where("user_id = ? AND user_type = ?", userID, userType).
		Update("is_active", false).Error
}

// CleanupExpiredSessions 清理过期会话
func (r *securityRepository) CleanupExpiredSessions(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < NOW() OR is_active = false").
		Delete(&model.Session{}).Error
}
