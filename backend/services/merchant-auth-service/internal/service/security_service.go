package service

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/auth"
	"github.com/pquerna/otp/totp"
	"payment-platform/merchant-auth-service/internal/client"
	"payment-platform/merchant-auth-service/internal/model"
	"payment-platform/merchant-auth-service/internal/repository"
)

// SecurityService 安全服务接口
type SecurityService interface {
	// 密码管理
	ChangePassword(ctx context.Context, merchantID uuid.UUID, oldPassword, newPassword string) error

	// 2FA管理
	Enable2FA(ctx context.Context, merchantID uuid.UUID) (*Enable2FAResponse, error)
	Verify2FA(ctx context.Context, merchantID uuid.UUID, code string) (*Verify2FAResponse, error)
	Disable2FA(ctx context.Context, merchantID uuid.UUID, password string) error
	Validate2FACode(ctx context.Context, merchantID uuid.UUID, code string) (bool, error)

	// 安全设置
	GetSecuritySettings(ctx context.Context, merchantID uuid.UUID) (*model.SecuritySettings, error)
	UpdateSecuritySettings(ctx context.Context, merchantID uuid.UUID, input *UpdateSecuritySettingsInput) (*model.SecuritySettings, error)

	// 登录活动
	RecordLoginActivity(ctx context.Context, activity *model.LoginActivity) error
	GetLoginActivities(ctx context.Context, merchantID uuid.UUID, page, pageSize int) ([]*model.LoginActivity, int64, error)

	// 会话管理
	CreateSession(ctx context.Context, merchantID uuid.UUID, ip, userAgent string) (*model.Session, error)
	GetActiveSessions(ctx context.Context, merchantID uuid.UUID) ([]*model.Session, error)
	RevokeSession(ctx context.Context, sessionID string) error
	RevokeAllSessions(ctx context.Context, merchantID uuid.UUID) error
	CleanExpiredSessions(ctx context.Context) error
}

type securityService struct {
	securityRepo   repository.SecurityRepository
	merchantClient client.MerchantClient
}

// NewSecurityService 创建安全服务实例
func NewSecurityService(
	securityRepo repository.SecurityRepository,
	merchantClient client.MerchantClient,
) SecurityService {
	return &securityService{
		securityRepo:   securityRepo,
		merchantClient: merchantClient,
	}
}

// Enable2FAResponse 启用2FA响应
type Enable2FAResponse struct {
	Secret      string   `json:"secret"`
	QRCode      string   `json:"qr_code"`
	BackupCodes []string `json:"backup_codes"`
}

// Verify2FAResponse 验证2FA响应
type Verify2FAResponse struct {
	Success bool `json:"success"`
}

// UpdateSecuritySettingsInput 更新安全设置输入
type UpdateSecuritySettingsInput struct {
	PasswordExpiryDays    *int     `json:"password_expiry_days"`
	SessionTimeoutMinutes *int     `json:"session_timeout_minutes"`
	MaxConcurrentSessions *int     `json:"max_concurrent_sessions"`
	IPWhitelist           []string `json:"ip_whitelist"`
	AllowedCountries      []string `json:"allowed_countries"`
	BlockedCountries      []string `json:"blocked_countries"`
	LoginNotification     *bool    `json:"login_notification"`
	AbnormalNotification  *bool    `json:"abnormal_notification"`
}

// ChangePassword 修改密码
func (s *securityService) ChangePassword(ctx context.Context, merchantID uuid.UUID, oldPassword, newPassword string) error {
	// 获取商户信息（包含密码哈希）
	merchant, err := s.merchantClient.GetMerchantWithPassword(ctx, merchantID)
	if err != nil {
		return fmt.Errorf("获取商户失败: %w", err)
	}

	// 验证旧密码
	if err := auth.VerifyPassword(oldPassword, merchant.PasswordHash); err != nil {
		return fmt.Errorf("旧密码错误")
	}

	// 检查新密码是否与旧密码相同
	if err := auth.VerifyPassword(newPassword, merchant.PasswordHash); err == nil {
		return fmt.Errorf("新密码不能与旧密码相同")
	}

	// 检查密码历史（防止重复使用最近的密码）
	passwordHistory, err := s.securityRepo.GetPasswordHistory(ctx, merchantID, 5)
	if err != nil {
		return fmt.Errorf("获取密码历史失败: %w", err)
	}

	for _, history := range passwordHistory {
		if err := auth.VerifyPassword(newPassword, history.PasswordHash); err == nil {
			return fmt.Errorf("不能使用最近使用过的密码")
		}
	}

	// 加密新密码
	newPasswordHash, err := auth.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 保存旧密码到历史记录
	if err := s.securityRepo.CreatePasswordHistory(ctx, &model.PasswordHistory{
		MerchantID:   merchantID,
		PasswordHash: merchant.PasswordHash,
	}); err != nil {
		return fmt.Errorf("保存密码历史失败: %w", err)
	}

	// 更新密码（通过merchant-service）
	if err := s.merchantClient.UpdatePassword(ctx, merchantID, newPasswordHash); err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}

	// 更新安全设置
	settings, err := s.securityRepo.GetSecuritySettings(ctx, merchantID)
	if err != nil {
		return fmt.Errorf("获取安全设置失败: %w", err)
	}

	now := time.Now()
	if settings == nil {
		// 创建默认安全设置
		settings = &model.SecuritySettings{
			MerchantID:        merchantID,
			PasswordChangedAt: &now,
		}
		if err := s.securityRepo.CreateSecuritySettings(ctx, settings); err != nil {
			return fmt.Errorf("创建安全设置失败: %w", err)
		}
	} else {
		settings.PasswordChangedAt = &now
		if err := s.securityRepo.UpdateSecuritySettings(ctx, settings); err != nil {
			return fmt.Errorf("更新安全设置失败: %w", err)
		}
	}

	return nil
}

// Enable2FA 启用2FA
func (s *securityService) Enable2FA(ctx context.Context, merchantID uuid.UUID) (*Enable2FAResponse, error) {
	// 检查是否已启用
	existing, err := s.securityRepo.GetTwoFactorAuth(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("检查2FA状态失败: %w", err)
	}
	if existing != nil && existing.IsEnabled {
		return nil, fmt.Errorf("2FA已启用")
	}

	// 获取商户信息
	merchant, err := s.merchantClient.GetMerchant(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("获取商户失败: %w", err)
	}

	// 生成TOTP密钥
	secret := generateTOTPSecret()

	// 生成备用恢复代码
	backupCodes := generateBackupCodes(10)
	backupCodesJSON, _ := json.Marshal(backupCodes)

	// 保存2FA配置
	tfa := &model.TwoFactorAuth{
		MerchantID:  merchantID,
		Secret:      secret,
		IsEnabled:   false, // 需要验证后才启用
		IsVerified:  false,
		BackupCodes: string(backupCodesJSON),
	}

	if existing == nil {
		if err := s.securityRepo.CreateTwoFactorAuth(ctx, tfa); err != nil {
			return nil, fmt.Errorf("创建2FA配置失败: %w", err)
		}
	} else {
		tfa.ID = existing.ID
		if err := s.securityRepo.UpdateTwoFactorAuth(ctx, tfa); err != nil {
			return nil, fmt.Errorf("更新2FA配置失败: %w", err)
		}
	}

	// 生成QR码URL
	qrCode, err := totp.GenerateCodeCustom(secret, time.Now(), totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    6,
		Algorithm: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("生成QR码失败: %w", err)
	}

	qrCodeURL := fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s",
		"PaymentPlatform",
		merchant.Email,
		secret,
		"PaymentPlatform",
	)

	_ = qrCode // 临时忽略未使用变量

	return &Enable2FAResponse{
		Secret:      secret,
		QRCode:      qrCodeURL,
		BackupCodes: backupCodes,
	}, nil
}

// Verify2FA 验证2FA
func (s *securityService) Verify2FA(ctx context.Context, merchantID uuid.UUID, code string) (*Verify2FAResponse, error) {
	// 获取2FA配置
	tfa, err := s.securityRepo.GetTwoFactorAuth(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("获取2FA配置失败: %w", err)
	}
	if tfa == nil {
		return nil, fmt.Errorf("2FA未配置")
	}

	// 验证TOTP代码
	valid := totp.Validate(code, tfa.Secret)
	if !valid {
		// 检查备用代码
		var backupCodes []string
		if err := json.Unmarshal([]byte(tfa.BackupCodes), &backupCodes); err == nil {
			for i, backupCode := range backupCodes {
				if backupCode == code {
					// 使用后删除备用代码
					backupCodes = append(backupCodes[:i], backupCodes[i+1:]...)
					backupCodesJSON, _ := json.Marshal(backupCodes)
					tfa.BackupCodes = string(backupCodesJSON)
					valid = true
					break
				}
			}
		}
	}

	if !valid {
		return &Verify2FAResponse{Success: false}, nil
	}

	// 启用2FA
	now := time.Now()
	tfa.IsEnabled = true
	tfa.IsVerified = true
	tfa.VerifiedAt = &now

	if err := s.securityRepo.UpdateTwoFactorAuth(ctx, tfa); err != nil {
		return nil, fmt.Errorf("更新2FA配置失败: %w", err)
	}

	return &Verify2FAResponse{Success: true}, nil
}

// Disable2FA 禁用2FA
func (s *securityService) Disable2FA(ctx context.Context, merchantID uuid.UUID, password string) error {
	// 验证密码
	merchant, err := s.merchantClient.GetMerchantWithPassword(ctx, merchantID)
	if err != nil {
		return fmt.Errorf("获取商户失败: %w", err)
	}

	if err := auth.VerifyPassword(password, merchant.PasswordHash); err != nil {
		return fmt.Errorf("密码错误")
	}

	// 删除2FA配置
	if err := s.securityRepo.DeleteTwoFactorAuth(ctx, merchantID); err != nil {
		return fmt.Errorf("删除2FA配置失败: %w", err)
	}

	return nil
}

// Validate2FACode 验证2FA代码（登录时使用）
func (s *securityService) Validate2FACode(ctx context.Context, merchantID uuid.UUID, code string) (bool, error) {
	// 获取2FA配置
	tfa, err := s.securityRepo.GetTwoFactorAuth(ctx, merchantID)
	if err != nil {
		return false, fmt.Errorf("获取2FA配置失败: %w", err)
	}
	if tfa == nil || !tfa.IsEnabled {
		return false, fmt.Errorf("2FA未启用")
	}

	// 验证TOTP代码
	valid := totp.Validate(code, tfa.Secret)
	if !valid {
		// 检查备用代码
		var backupCodes []string
		if err := json.Unmarshal([]byte(tfa.BackupCodes), &backupCodes); err == nil {
			for i, backupCode := range backupCodes {
				if backupCode == code {
					// 使用后删除备用代码
					backupCodes = append(backupCodes[:i], backupCodes[i+1:]...)
					backupCodesJSON, _ := json.Marshal(backupCodes)
					tfa.BackupCodes = string(backupCodesJSON)
					if err := s.securityRepo.UpdateTwoFactorAuth(ctx, tfa); err != nil {
						return false, fmt.Errorf("更新2FA配置失败: %w", err)
					}
					valid = true
					break
				}
			}
		}
	}

	return valid, nil
}

// GetSecuritySettings 获取安全设置
func (s *securityService) GetSecuritySettings(ctx context.Context, merchantID uuid.UUID) (*model.SecuritySettings, error) {
	settings, err := s.securityRepo.GetSecuritySettings(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("获取安全设置失败: %w", err)
	}

	if settings == nil {
		// 创建默认安全设置
		settings = &model.SecuritySettings{
			MerchantID:            merchantID,
			PasswordExpiryDays:    90,
			SessionTimeoutMinutes: 60,
			MaxConcurrentSessions: 5,
			LoginNotification:     true,
			AbnormalNotification:  true,
			IPWhitelist:           "[]",  // 空数组JSON
			AllowedCountries:      "[]",  // 空数组JSON
			BlockedCountries:      "[]",  // 空数组JSON
		}
		if err := s.securityRepo.CreateSecuritySettings(ctx, settings); err != nil {
			return nil, fmt.Errorf("创建安全设置失败: %w", err)
		}
	}

	return settings, nil
}

// UpdateSecuritySettings 更新安全设置
func (s *securityService) UpdateSecuritySettings(ctx context.Context, merchantID uuid.UUID, input *UpdateSecuritySettingsInput) (*model.SecuritySettings, error) {
	settings, err := s.GetSecuritySettings(ctx, merchantID)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if input.PasswordExpiryDays != nil {
		settings.PasswordExpiryDays = *input.PasswordExpiryDays
	}
	if input.SessionTimeoutMinutes != nil {
		settings.SessionTimeoutMinutes = *input.SessionTimeoutMinutes
	}
	if input.MaxConcurrentSessions != nil {
		settings.MaxConcurrentSessions = *input.MaxConcurrentSessions
	}
	if input.IPWhitelist != nil {
		ipWhitelistJSON, _ := json.Marshal(input.IPWhitelist)
		settings.IPWhitelist = string(ipWhitelistJSON)
	}
	if input.AllowedCountries != nil {
		allowedCountriesJSON, _ := json.Marshal(input.AllowedCountries)
		settings.AllowedCountries = string(allowedCountriesJSON)
	}
	if input.BlockedCountries != nil {
		blockedCountriesJSON, _ := json.Marshal(input.BlockedCountries)
		settings.BlockedCountries = string(blockedCountriesJSON)
	}
	if input.LoginNotification != nil {
		settings.LoginNotification = *input.LoginNotification
	}
	if input.AbnormalNotification != nil {
		settings.AbnormalNotification = *input.AbnormalNotification
	}

	if err := s.securityRepo.UpdateSecuritySettings(ctx, settings); err != nil {
		return nil, fmt.Errorf("更新安全设置失败: %w", err)
	}

	return settings, nil
}

// RecordLoginActivity 记录登录活动
func (s *securityService) RecordLoginActivity(ctx context.Context, activity *model.LoginActivity) error {
	return s.securityRepo.CreateLoginActivity(ctx, activity)
}

// GetLoginActivities 获取登录活动
func (s *securityService) GetLoginActivities(ctx context.Context, merchantID uuid.UUID, page, pageSize int) ([]*model.LoginActivity, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return s.securityRepo.GetLoginActivities(ctx, merchantID, page, pageSize)
}

// CreateSession 创建会话
func (s *securityService) CreateSession(ctx context.Context, merchantID uuid.UUID, ip, userAgent string) (*model.Session, error) {
	// 检查并发会话数限制
	settings, err := s.GetSecuritySettings(ctx, merchantID)
	if err != nil {
		return nil, err
	}

	activeCount, err := s.securityRepo.CountActiveSessions(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("统计会话数失败: %w", err)
	}

	if int(activeCount) >= settings.MaxConcurrentSessions {
		return nil, fmt.Errorf("超过最大并发会话数限制")
	}

	// 生成会话ID
	sessionID := generateSessionID()
	expiresAt := time.Now().Add(time.Duration(settings.SessionTimeoutMinutes) * time.Minute)

	session := &model.Session{
		SessionID:  sessionID,
		MerchantID: merchantID,
		IP:         ip,
		UserAgent:  userAgent,
		ExpiresAt:  expiresAt,
		IsActive:   true,
	}

	if err := s.securityRepo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("创建会话失败: %w", err)
	}

	return session, nil
}

// GetActiveSessions 获取活跃会话
func (s *securityService) GetActiveSessions(ctx context.Context, merchantID uuid.UUID) ([]*model.Session, error) {
	return s.securityRepo.GetActiveSessions(ctx, merchantID)
}

// RevokeSession 撤销会话
func (s *securityService) RevokeSession(ctx context.Context, sessionID string) error {
	return s.securityRepo.DeleteSession(ctx, sessionID)
}

// RevokeAllSessions 撤销所有会话
func (s *securityService) RevokeAllSessions(ctx context.Context, merchantID uuid.UUID) error {
	sessions, err := s.securityRepo.GetActiveSessions(ctx, merchantID)
	if err != nil {
		return fmt.Errorf("获取会话失败: %w", err)
	}

	for _, session := range sessions {
		if err := s.securityRepo.DeleteSession(ctx, session.SessionID); err != nil {
			return fmt.Errorf("撤销会话失败: %w", err)
		}
	}

	return nil
}

// CleanExpiredSessions 清理过期会话
func (s *securityService) CleanExpiredSessions(ctx context.Context) error {
	return s.securityRepo.DeleteExpiredSessions(ctx)
}

// 辅助函数

// generateTOTPSecret 生成TOTP密钥
func generateTOTPSecret() string {
	secret := make([]byte, 20)
	rand.Read(secret)
	return base32.StdEncoding.EncodeToString(secret)
}

// generateBackupCodes 生成备用恢复代码
func generateBackupCodes(count int) []string {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		b := make([]byte, 4)
		rand.Read(b)
		codes[i] = fmt.Sprintf("%08X", b)
	}
	return codes
}

// generateSessionID 生成会话ID
func generateSessionID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return fmt.Sprintf("%X", b)
}

// DetectAbnormalLogin 检测异常登录
func DetectAbnormalLogin(ctx context.Context, merchantID uuid.UUID, ip, country string, securityRepo repository.SecurityRepository) (bool, string) {
	reasons := []string{}

	// 获取最近的登录活动
	recentActivities, err := securityRepo.GetRecentLoginActivities(ctx, merchantID, 10)
	if err != nil || len(recentActivities) == 0 {
		return false, ""
	}

	// 检查新IP
	isNewIP := true
	for _, activity := range recentActivities {
		if activity.IP == ip {
			isNewIP = false
			break
		}
	}
	if isNewIP {
		reasons = append(reasons, model.AbnormalReasonNewIP)
	}

	// 检查新国家
	isNewCountry := true
	for _, activity := range recentActivities {
		if activity.Country == country {
			isNewCountry = false
			break
		}
	}
	if isNewCountry && country != "" {
		reasons = append(reasons, model.AbnormalReasonNewCountry)
	}

	isAbnormal := len(reasons) > 0
	return isAbnormal, strings.Join(reasons, ",")
}
