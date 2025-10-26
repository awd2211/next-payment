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
	"payment-platform/admin-service/internal/model"
	"payment-platform/admin-service/internal/repository"
	"github.com/ua-parser/uap-go/uaparser"
)

// SecurityService 安全服务接口
type SecurityService interface {
	// 密码管理
	ChangePassword(ctx context.Context, userID uuid.UUID, userType, oldPassword, newPassword string) error
	ResetPassword(ctx context.Context, userID uuid.UUID, userType, newPassword string) error
	ValidatePasswordStrength(password string) error

	// 2FA管理
	Setup2FA(ctx context.Context, userID uuid.UUID, userType, accountName string) (*Setup2FAResponse, error)
	Verify2FA(ctx context.Context, userID uuid.UUID, userType, code string) error
	Disable2FA(ctx context.Context, userID uuid.UUID, userType, password string) error
	Validate2FACode(ctx context.Context, userID uuid.UUID, userType, code string) (bool, error)
	RegenerateBackupCodes(ctx context.Context, userID uuid.UUID, userType string) ([]string, error)

	// 登录活动
	RecordLoginActivity(ctx context.Context, activity *LoginActivityInput) error
	GetLoginActivities(ctx context.Context, userID uuid.UUID, userType string, limit int) ([]*model.LoginActivity, error)
	GetAbnormalActivities(ctx context.Context, userID uuid.UUID, userType string) ([]*model.LoginActivity, error)
	CheckAbnormalLogin(ctx context.Context, userID uuid.UUID, userType string, input *LoginActivityInput) (bool, []string)

	// 安全设置
	GetSecuritySettings(ctx context.Context, userID uuid.UUID, userType string) (*model.SecuritySettings, error)
	UpdateSecuritySettings(ctx context.Context, userID uuid.UUID, userType string, settings *UpdateSecuritySettingsInput) error
	InitDefaultSecuritySettings(ctx context.Context, userID uuid.UUID, userType string) error

	// 会话管理
	CreateSession(ctx context.Context, userID uuid.UUID, userType, ip, userAgent string, expiresIn time.Duration) (*model.Session, error)
	GetActiveSessions(ctx context.Context, userID uuid.UUID, userType string) ([]*model.Session, error)
	DeactivateSession(ctx context.Context, sessionID string) error
	DeactivateAllOtherSessions(ctx context.Context, userID uuid.UUID, userType, currentSessionID string) error
}

type securityService struct {
	securityRepo     repository.SecurityRepository
	adminRepo        repository.AdminRepository
	totpManager      *auth.TOTPManager
	uaParser         *uaparser.Parser
}

// NewSecurityService 创建安全服务实例
func NewSecurityService(securityRepo repository.SecurityRepository, adminRepo repository.AdminRepository) SecurityService {
	return &securityService{
		securityRepo: securityRepo,
		adminRepo:    adminRepo,
		totpManager:  auth.NewTOTPManager("Payment Platform"),
		uaParser:     uaparser.NewFromSaved(),
	}
}

// Setup2FAResponse 2FA设置响应
type Setup2FAResponse struct {
	Secret      string   `json:"secret"`       // TOTP密钥
	QRCodeURL   string   `json:"qr_code_url"`  // 二维码URL
	BackupCodes []string `json:"backup_codes"` // 备用恢复代码
}

// LoginActivityInput 登录活动输入
type LoginActivityInput struct {
	UserID     uuid.UUID
	UserType   string
	LoginType  string
	Status     string
	IP         string
	UserAgent  string
	SessionID  string
	FailedReason string
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
func (s *securityService) ChangePassword(ctx context.Context, userID uuid.UUID, userType, oldPassword, newPassword string) error {
	// 验证旧密码
	var admin *model.Admin
	var err error

	if userType == model.UserTypeAdmin {
		admin, err = s.adminRepo.GetByID(ctx, userID)
		if err != nil {
			return fmt.Errorf("获取用户失败: %w", err)
		}
		if admin == nil {
			return fmt.Errorf("用户不存在")
		}

		// 验证旧密码
		if err := auth.VerifyPassword(oldPassword, admin.PasswordHash); err != nil {
			return fmt.Errorf("旧密码错误")
		}
	}

	// 验证新密码强度
	if err := s.ValidatePasswordStrength(newPassword); err != nil {
		return err
	}

	// 检查新密码是否与最近5次使用的密码重复
	newPasswordHash, err := auth.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	history, err := s.securityRepo.GetPasswordHistory(ctx, userID, userType, 5)
	if err != nil {
		return fmt.Errorf("获取密码历史失败: %w", err)
	}

	for _, h := range history {
		if err := auth.VerifyPassword(newPassword, h.PasswordHash); err == nil {
			return fmt.Errorf("新密码不能与最近5次使用的密码相同")
		}
	}

	// 更新密码
	admin.PasswordHash = newPasswordHash
	if err := s.adminRepo.Update(ctx, admin); err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}

	// 记录密码历史
	passwordHistory := &model.PasswordHistory{
		UserID:       userID,
		UserType:     userType,
		PasswordHash: newPasswordHash,
	}
	if err := s.securityRepo.CreatePasswordHistory(ctx, passwordHistory); err != nil {
		return fmt.Errorf("记录密码历史失败: %w", err)
	}

	// 更新安全设置中的密码修改时间
	settings, err := s.securityRepo.GetSecuritySettings(ctx, userID, userType)
	if err != nil {
		return fmt.Errorf("获取安全设置失败: %w", err)
	}

	if settings == nil {
		// 创建默认设置
		if err := s.InitDefaultSecuritySettings(ctx, userID, userType); err != nil {
			return err
		}
		settings, _ = s.securityRepo.GetSecuritySettings(ctx, userID, userType)
	}

	now := time.Now()
	settings.PasswordChangedAt = &now
	settings.RequirePasswordChange = false

	if err := s.securityRepo.UpdateSecuritySettings(ctx, settings); err != nil {
		return fmt.Errorf("更新安全设置失败: %w", err)
	}

	return nil
}

// ResetPassword 重置密码（管理员操作）
func (s *securityService) ResetPassword(ctx context.Context, userID uuid.UUID, userType, newPassword string) error {
	// 验证新密码强度
	if err := s.ValidatePasswordStrength(newPassword); err != nil {
		return err
	}

	// 加密密码
	newPasswordHash, err := auth.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 更新密码
	if userType == model.UserTypeAdmin {
		admin, err := s.adminRepo.GetByID(ctx, userID)
		if err != nil {
			return fmt.Errorf("获取用户失败: %w", err)
		}
		if admin == nil {
			return fmt.Errorf("用户不存在")
		}

		admin.PasswordHash = newPasswordHash
		if err := s.adminRepo.Update(ctx, admin); err != nil {
			return fmt.Errorf("更新密码失败: %w", err)
		}
	}

	// 记录密码历史
	passwordHistory := &model.PasswordHistory{
		UserID:       userID,
		UserType:     userType,
		PasswordHash: newPasswordHash,
	}
	if err := s.securityRepo.CreatePasswordHistory(ctx, passwordHistory); err != nil {
		return fmt.Errorf("记录密码历史失败: %w", err)
	}

	// 要求用户下次登录时修改密码
	settings, _ := s.securityRepo.GetSecuritySettings(ctx, userID, userType)
	if settings != nil {
		settings.RequirePasswordChange = true
		s.securityRepo.UpdateSecuritySettings(ctx, settings)
	}

	return nil
}

// ValidatePasswordStrength 验证密码强度
func (s *securityService) ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("密码长度至少为8个字符")
	}

	if len(password) > 128 {
		return fmt.Errorf("密码长度不能超过128个字符")
	}

	// 至少包含一个大写字母
	hasUpper := false
	for _, c := range password {
		if c >= 'A' && c <= 'Z' {
			hasUpper = true
			break
		}
	}
	if !hasUpper {
		return fmt.Errorf("密码必须包含至少一个大写字母")
	}

	// 至少包含一个小写字母
	hasLower := false
	for _, c := range password {
		if c >= 'a' && c <= 'z' {
			hasLower = true
			break
		}
	}
	if !hasLower {
		return fmt.Errorf("密码必须包含至少一个小写字母")
	}

	// 至少包含一个数字
	hasDigit := false
	for _, c := range password {
		if c >= '0' && c <= '9' {
			hasDigit = true
			break
		}
	}
	if !hasDigit {
		return fmt.Errorf("密码必须包含至少一个数字")
	}

	// 至少包含一个特殊字符
	hasSpecial := false
	specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"
	for _, c := range password {
		for _, s := range specialChars {
			if c == s {
				hasSpecial = true
				break
			}
		}
		if hasSpecial {
			break
		}
	}
	if !hasSpecial {
		return fmt.Errorf("密码必须包含至少一个特殊字符")
	}

	return nil
}

// Setup2FA 设置2FA
func (s *securityService) Setup2FA(ctx context.Context, userID uuid.UUID, userType, accountName string) (*Setup2FAResponse, error) {
	// 检查是否已经设置过2FA
	existing, err := s.securityRepo.Get2FA(ctx, userID, userType)
	if err != nil {
		return nil, fmt.Errorf("检查2FA状态失败: %w", err)
	}

	if existing != nil && existing.IsEnabled {
		return nil, fmt.Errorf("2FA已经启用，请先禁用后再重新设置")
	}

	// 生成TOTP密钥
	secret, qrURL, err := s.totpManager.GenerateSecret(accountName)
	if err != nil {
		return nil, fmt.Errorf("生成2FA密钥失败: %w", err)
	}

	// 生成备用恢复代码
	backupCodes, err := s.totpManager.GenerateBackupCodes(8)
	if err != nil {
		return nil, fmt.Errorf("生成备用代码失败: %w", err)
	}

	// 加密备用代码存储
	backupCodesJSON, _ := json.Marshal(backupCodes)

	// 保存2FA记录
	tfa := &model.TwoFactorAuth{
		UserID:      userID,
		UserType:    userType,
		Secret:      secret,
		IsEnabled:   false, // 需要验证后才启用
		IsVerified:  false,
		BackupCodes: string(backupCodesJSON),
	}

	if existing != nil {
		tfa.ID = existing.ID
		err = s.securityRepo.Update2FA(ctx, tfa)
	} else {
		err = s.securityRepo.Create2FA(ctx, tfa)
	}

	if err != nil {
		return nil, fmt.Errorf("保存2FA设置失败: %w", err)
	}

	return &Setup2FAResponse{
		Secret:      secret,
		QRCodeURL:   qrURL,
		BackupCodes: backupCodes,
	}, nil
}

// Verify2FA 验证并启用2FA
func (s *securityService) Verify2FA(ctx context.Context, userID uuid.UUID, userType, code string) error {
	tfa, err := s.securityRepo.Get2FA(ctx, userID, userType)
	if err != nil {
		return fmt.Errorf("获取2FA设置失败: %w", err)
	}

	if tfa == nil {
		return fmt.Errorf("2FA未设置")
	}

	// 验证代码
	if !s.totpManager.VerifyCode(tfa.Secret, code) {
		return fmt.Errorf("验证码错误")
	}

	// 启用2FA
	now := time.Now()
	tfa.IsEnabled = true
	tfa.IsVerified = true
	tfa.VerifiedAt = &now

	if err := s.securityRepo.Update2FA(ctx, tfa); err != nil {
		return fmt.Errorf("启用2FA失败: %w", err)
	}

	return nil
}

// Disable2FA 禁用2FA
func (s *securityService) Disable2FA(ctx context.Context, userID uuid.UUID, userType, password string) error {
	// 验证密码
	if userType == model.UserTypeAdmin {
		admin, err := s.adminRepo.GetByID(ctx, userID)
		if err != nil {
			return fmt.Errorf("获取用户失败: %w", err)
		}
		if admin == nil {
			return fmt.Errorf("用户不存在")
		}

		if err := auth.VerifyPassword(password, admin.PasswordHash); err != nil {
			return fmt.Errorf("密码错误")
		}
	}

	// 删除2FA记录
	if err := s.securityRepo.Delete2FA(ctx, userID, userType); err != nil {
		return fmt.Errorf("禁用2FA失败: %w", err)
	}

	return nil
}

// Validate2FACode 验证2FA代码
func (s *securityService) Validate2FACode(ctx context.Context, userID uuid.UUID, userType, code string) (bool, error) {
	tfa, err := s.securityRepo.Get2FA(ctx, userID, userType)
	if err != nil {
		return false, fmt.Errorf("获取2FA设置失败: %w", err)
	}

	if tfa == nil || !tfa.IsEnabled {
		return false, fmt.Errorf("2FA未启用")
	}

	// 先验证TOTP代码
	if s.totpManager.VerifyCode(tfa.Secret, code) {
		return true, nil
	}

	// 验证备用代码
	var backupCodes []string
	if err := json.Unmarshal([]byte(tfa.BackupCodes), &backupCodes); err == nil {
		for i, bc := range backupCodes {
			if bc == code {
				// 使用后移除该备用代码
				backupCodes = append(backupCodes[:i], backupCodes[i+1:]...)
				updatedCodesJSON, _ := json.Marshal(backupCodes)
				tfa.BackupCodes = string(updatedCodesJSON)
				s.securityRepo.Update2FA(ctx, tfa)
				return true, nil
			}
		}
	}

	return false, nil
}

// RegenerateBackupCodes 重新生成备用代码
func (s *securityService) RegenerateBackupCodes(ctx context.Context, userID uuid.UUID, userType string) ([]string, error) {
	tfa, err := s.securityRepo.Get2FA(ctx, userID, userType)
	if err != nil {
		return nil, fmt.Errorf("获取2FA设置失败: %w", err)
	}

	if tfa == nil {
		return nil, fmt.Errorf("2FA未设置")
	}

	// 生成新的备用代码
	backupCodes, err := s.totpManager.GenerateBackupCodes(8)
	if err != nil {
		return nil, fmt.Errorf("生成备用代码失败: %w", err)
	}

	backupCodesJSON, _ := json.Marshal(backupCodes)
	tfa.BackupCodes = string(backupCodesJSON)

	if err := s.securityRepo.Update2FA(ctx, tfa); err != nil {
		return nil, fmt.Errorf("更新备用代码失败: %w", err)
	}

	return backupCodes, nil
}

// RecordLoginActivity 记录登录活动
func (s *securityService) RecordLoginActivity(ctx context.Context, input *LoginActivityInput) error {
	activity := &model.LoginActivity{
		UserID:     input.UserID,
		UserType:   input.UserType,
		LoginType:  input.LoginType,
		Status:     input.Status,
		IP:         input.IP,
		UserAgent:  input.UserAgent,
		SessionID:  input.SessionID,
		FailedReason: input.FailedReason,
		LoginAt:    time.Now(),
	}

	// 解析User-Agent
	if input.UserAgent != "" {
		client := s.uaParser.Parse(input.UserAgent)
		if client.UserAgent != nil {
			activity.Browser = client.UserAgent.Family
		}
		if client.Os != nil {
			activity.OS = client.Os.Family
		}
		if client.Device != nil {
			activity.DeviceType = s.getDeviceType(client.Device.Family)
		}
	}

	// 检查是否异常登录
	if input.Status == model.LoginStatusSuccess {
		isAbnormal, reasons := s.CheckAbnormalLogin(ctx, input.UserID, input.UserType, input)
		activity.IsAbnormal = isAbnormal
		if len(reasons) > 0 {
			activity.AbnormalReason = strings.Join(reasons, ", ")
		}
	}

	return s.securityRepo.CreateLoginActivity(ctx, activity)
}

// GetLoginActivities 获取登录活动记录
func (s *securityService) GetLoginActivities(ctx context.Context, userID uuid.UUID, userType string, limit int) ([]*model.LoginActivity, error) {
	return s.securityRepo.GetLoginActivities(ctx, userID, userType, limit)
}

// GetAbnormalActivities 获取异常登录活动
func (s *securityService) GetAbnormalActivities(ctx context.Context, userID uuid.UUID, userType string) ([]*model.LoginActivity, error) {
	return s.securityRepo.GetAbnormalActivities(ctx, userID, userType, 50)
}

// CheckAbnormalLogin 检查异常登录
func (s *securityService) CheckAbnormalLogin(ctx context.Context, userID uuid.UUID, userType string, input *LoginActivityInput) (bool, []string) {
	reasons := []string{}

	// 获取最近的登录记录
	recentActivities, err := s.securityRepo.GetLoginActivities(ctx, userID, userType, 10)
	if err != nil || len(recentActivities) == 0 {
		return false, reasons
	}

	// 检查新IP
	newIP := true
	for _, activity := range recentActivities {
		if activity.IP == input.IP {
			newIP = false
			break
		}
	}
	if newIP {
		reasons = append(reasons, model.AbnormalReasonNewIP)
	}

	// 检查新设备
	newDevice := true
	currentUA := input.UserAgent
	for _, activity := range recentActivities {
		if activity.UserAgent == currentUA {
			newDevice = false
			break
		}
	}
	if newDevice {
		reasons = append(reasons, model.AbnormalReasonNewDevice)
	}

	// 如果有异常原因，标记为异常
	return len(reasons) > 0, reasons
}

// GetSecuritySettings 获取安全设置
func (s *securityService) GetSecuritySettings(ctx context.Context, userID uuid.UUID, userType string) (*model.SecuritySettings, error) {
	settings, err := s.securityRepo.GetSecuritySettings(ctx, userID, userType)
	if err != nil {
		return nil, err
	}

	if settings == nil {
		// 创建默认设置
		if err := s.InitDefaultSecuritySettings(ctx, userID, userType); err != nil {
			return nil, err
		}
		return s.securityRepo.GetSecuritySettings(ctx, userID, userType)
	}

	return settings, nil
}

// UpdateSecuritySettings 更新安全设置
func (s *securityService) UpdateSecuritySettings(ctx context.Context, userID uuid.UUID, userType string, input *UpdateSecuritySettingsInput) error {
	settings, err := s.GetSecuritySettings(ctx, userID, userType)
	if err != nil {
		return err
	}

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
		whitelistJSON, _ := json.Marshal(input.IPWhitelist)
		settings.IPWhitelist = string(whitelistJSON)
	}
	if input.AllowedCountries != nil {
		allowedJSON, _ := json.Marshal(input.AllowedCountries)
		settings.AllowedCountries = string(allowedJSON)
	}
	if input.BlockedCountries != nil {
		blockedJSON, _ := json.Marshal(input.BlockedCountries)
		settings.BlockedCountries = string(blockedJSON)
	}
	if input.LoginNotification != nil {
		settings.LoginNotification = *input.LoginNotification
	}
	if input.AbnormalNotification != nil {
		settings.AbnormalNotification = *input.AbnormalNotification
	}

	return s.securityRepo.UpdateSecuritySettings(ctx, settings)
}

// InitDefaultSecuritySettings 初始化默认安全设置
func (s *securityService) InitDefaultSecuritySettings(ctx context.Context, userID uuid.UUID, userType string) error {
	settings := &model.SecuritySettings{
		UserID:                userID,
		UserType:              userType,
		PasswordExpiryDays:    90,
		SessionTimeoutMinutes: 60,
		MaxConcurrentSessions: 5,
		LoginNotification:     true,
		AbnormalNotification:  true,
	}

	return s.securityRepo.CreateSecuritySettings(ctx, settings)
}

// CreateSession 创建会话
func (s *securityService) CreateSession(ctx context.Context, userID uuid.UUID, userType, ip, userAgent string, expiresIn time.Duration) (*model.Session, error) {
	// 生成会话ID
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("生成会话ID失败: %w", err)
	}

	session := &model.Session{
		SessionID: sessionID,
		UserID:    userID,
		UserType:  userType,
		IP:        ip,
		UserAgent: userAgent,
		ExpiresAt: time.Now().Add(expiresIn),
		IsActive:  true,
	}

	if err := s.securityRepo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("创建会话失败: %w", err)
	}

	return session, nil
}

// GetActiveSessions 获取活跃会话
func (s *securityService) GetActiveSessions(ctx context.Context, userID uuid.UUID, userType string) ([]*model.Session, error) {
	return s.securityRepo.GetActiveSessions(ctx, userID, userType)
}

// DeactivateSession 停用会话
func (s *securityService) DeactivateSession(ctx context.Context, sessionID string) error {
	return s.securityRepo.DeactivateSession(ctx, sessionID)
}

// DeactivateAllOtherSessions 停用其他所有会话
func (s *securityService) DeactivateAllOtherSessions(ctx context.Context, userID uuid.UUID, userType, currentSessionID string) error {
	sessions, err := s.securityRepo.GetActiveSessions(ctx, userID, userType)
	if err != nil {
		return err
	}

	for _, session := range sessions {
		if session.SessionID != currentSessionID {
			s.securityRepo.DeactivateSession(ctx, session.SessionID)
		}
	}

	return nil
}

// getDeviceType 根据设备名称判断设备类型
func (s *securityService) getDeviceType(deviceFamily string) string {
	deviceLower := strings.ToLower(deviceFamily)

	if strings.Contains(deviceLower, "iphone") || strings.Contains(deviceLower, "android") || strings.Contains(deviceLower, "mobile") {
		return model.DeviceTypeMobile
	}
	if strings.Contains(deviceLower, "ipad") || strings.Contains(deviceLower, "tablet") {
		return model.DeviceTypeTablet
	}
	if deviceFamily != "" && deviceFamily != "Other" {
		return model.DeviceTypeDesktop
	}

	return model.DeviceTypeUnknown
}

// generateSessionID 生成会话ID
func generateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base32.StdEncoding.EncodeToString(b), nil
}
