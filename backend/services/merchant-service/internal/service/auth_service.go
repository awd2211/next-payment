package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/auth"
	"payment-platform/merchant-service/internal/model"
	"payment-platform/merchant-service/internal/repository"
)

// AuthService 认证服务接口
type AuthService interface {
	// 登录
	LoginWithPassword(ctx context.Context, email, password, ip, userAgent string) (*LoginResponse, error)
	LoginWith2FA(ctx context.Context, tempToken, code, ip, userAgent string) (*LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*RefreshTokenResponse, error)
	Logout(ctx context.Context, sessionID string) error
}

type authService struct {
	merchantRepo repository.MerchantRepository
	securityRepo repository.SecurityRepository
	jwtManager   *auth.JWTManager
}

// NewAuthService 创建认证服务实例
func NewAuthService(
	merchantRepo repository.MerchantRepository,
	securityRepo repository.SecurityRepository,
	jwtManager *auth.JWTManager,
) AuthService {
	return &authService{
		merchantRepo: merchantRepo,
		securityRepo: securityRepo,
		jwtManager:   jwtManager,
	}
}

// RefreshTokenResponse 刷新Token响应
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// LoginWithPassword 密码登录
func (s *authService) LoginWithPassword(ctx context.Context, email, password, ip, userAgent string) (*LoginResponse, error) {
	// 查找商户
	merchant, err := s.merchantRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("登录失败: %w", err)
	}
	if merchant == nil {
		// 记录失败的登录尝试
		s.recordFailedLogin(ctx, uuid.Nil, email, ip, userAgent, "邮箱或密码错误")
		return nil, fmt.Errorf("邮箱或密码错误")
	}

	// 验证密码
	if err := auth.VerifyPassword(password, merchant.PasswordHash); err != nil {
		// 记录失败的登录尝试
		s.recordFailedLogin(ctx, merchant.ID, email, ip, userAgent, "密码错误")
		return nil, fmt.Errorf("邮箱或密码错误")
	}

	// 检查商户状态
	if merchant.Status == model.MerchantStatusSuspended {
		s.recordFailedLogin(ctx, merchant.ID, email, ip, userAgent, "账户已被暂停")
		return nil, fmt.Errorf("账户已被暂停")
	}
	if merchant.Status == model.MerchantStatusRejected {
		s.recordFailedLogin(ctx, merchant.ID, email, ip, userAgent, "账户已被拒绝")
		return nil, fmt.Errorf("账户已被拒绝")
	}

	// 检查是否启用2FA
	tfa, err := s.securityRepo.GetTwoFactorAuth(ctx, merchant.ID)
	if err != nil {
		return nil, fmt.Errorf("检查2FA状态失败: %w", err)
	}

	if tfa != nil && tfa.IsEnabled {
		// 需要2FA验证，生成临时token
		tempToken := generateTempToken()

		// 这里应该将tempToken和merchantID存储到Redis中，设置短期过期时间（如5分钟）
		// 暂时直接返回，实际应用中需要实现临时token存储

		return &LoginResponse{
			Require2FA: true,
			TempToken:  tempToken,
		}, nil
	}

	// 生成会话ID
	sessionID := generateSessionID()

	// 生成JWT Token
	token, err := s.jwtManager.GenerateToken(
		merchant.ID,
		merchant.Email,
		"merchant",
		&merchant.ID,
		[]string{"merchant"},
		[]string{},
	)
	if err != nil {
		return nil, fmt.Errorf("生成Token失败: %w", err)
	}

	// 创建会话
	settings, _ := s.securityRepo.GetSecuritySettings(ctx, merchant.ID)
	sessionTimeout := 60
	if settings != nil {
		sessionTimeout = settings.SessionTimeoutMinutes
	}

	session := &model.Session{
		SessionID:  sessionID,
		MerchantID: merchant.ID,
		IP:         ip,
		UserAgent:  userAgent,
		ExpiresAt:  time.Now().Add(time.Duration(sessionTimeout) * time.Minute),
		IsActive:   true,
	}
	if err := s.securityRepo.CreateSession(ctx, session); err != nil {
		// 不影响登录，只记录错误
		fmt.Printf("创建会话失败: %v\n", err)
	}

	// 记录成功的登录活动
	s.recordSuccessLogin(ctx, merchant.ID, sessionID, ip, userAgent)

	// 清除密码字段
	merchant.PasswordHash = ""

	return &LoginResponse{
		Token:      token,
		Merchant:   merchant,
		Require2FA: false,
	}, nil
}

// LoginWith2FA 2FA登录
func (s *authService) LoginWith2FA(ctx context.Context, tempToken, code, ip, userAgent string) (*LoginResponse, error) {
	// 这里应该从Redis中获取tempToken对应的merchantID
	// 暂时返回错误，实际应用中需要实现

	// 示例实现（需要完善）:
	// 1. 从Redis获取merchantID
	// 2. 验证2FA代码
	// 3. 生成正式token
	// 4. 记录登录活动
	// 5. 创建会话

	return nil, fmt.Errorf("2FA登录功能需要Redis支持，暂未实现")
}

// RefreshToken 刷新Token
func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*RefreshTokenResponse, error) {
	// 验证refresh token
	claims, err := s.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("无效的refresh token: %w", err)
	}

	// 获取商户信息
	merchantID := claims.UserID

	merchant, err := s.merchantRepo.GetByID(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("获取商户失败: %w", err)
	}
	if merchant == nil {
		return nil, fmt.Errorf("商户不存在")
	}

	// 检查商户状态
	if merchant.Status != model.MerchantStatusActive && merchant.Status != model.MerchantStatusPending {
		return nil, fmt.Errorf("账户已被禁用")
	}

	// 生成新的access token
	accessToken, err := s.jwtManager.GenerateToken(
		merchant.ID,
		merchant.Email,
		"merchant",
		&merchant.ID,
		[]string{"merchant"},
		[]string{},
	)
	if err != nil {
		return nil, fmt.Errorf("生成Token失败: %w", err)
	}

	// 生成新的refresh token
	newRefreshToken, err := s.jwtManager.GenerateToken(
		merchant.ID,
		merchant.Email,
		"merchant",
		&merchant.ID,
		[]string{"merchant"},
		[]string{},
	)
	if err != nil {
		return nil, fmt.Errorf("生成RefreshToken失败: %w", err)
	}

	return &RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// Logout 登出
func (s *authService) Logout(ctx context.Context, sessionID string) error {
	// 删除会话
	if err := s.securityRepo.DeleteSession(ctx, sessionID); err != nil {
		return fmt.Errorf("登出失败: %w", err)
	}

	// 可以在这里记录登出活动
	// ...

	return nil
}

// recordSuccessLogin 记录成功的登录
func (s *authService) recordSuccessLogin(ctx context.Context, merchantID uuid.UUID, sessionID, ip, userAgent string) {
	// 检测异常登录
	isAbnormal, abnormalReason := DetectAbnormalLogin(ctx, merchantID, ip, "", s.securityRepo)

	activity := &model.LoginActivity{
		MerchantID:     merchantID,
		LoginType:      model.LoginTypePassword,
		Status:         model.LoginStatusSuccess,
		IP:             ip,
		UserAgent:      userAgent,
		SessionID:      sessionID,
		IsAbnormal:     isAbnormal,
		AbnormalReason: abnormalReason,
		LoginAt:        time.Now(),
	}

	// 解析User-Agent
	parseUserAgent(activity, userAgent)

	if err := s.securityRepo.CreateLoginActivity(ctx, activity); err != nil {
		fmt.Printf("记录登录活动失败: %v\n", err)
	}

	// 如果是异常登录，发送通知（需要实现邮件服务）
	if isAbnormal {
		// TODO: 发送异常登录通知邮件
		fmt.Printf("检测到异常登录: merchantID=%s, reason=%s\n", merchantID, abnormalReason)
	}
}

// recordFailedLogin 记录失败的登录
func (s *authService) recordFailedLogin(ctx context.Context, merchantID uuid.UUID, email, ip, userAgent, reason string) {
	activity := &model.LoginActivity{
		MerchantID:   merchantID,
		LoginType:    model.LoginTypePassword,
		Status:       model.LoginStatusFailed,
		IP:           ip,
		UserAgent:    userAgent,
		FailedReason: reason,
		LoginAt:      time.Now(),
	}

	// 解析User-Agent
	parseUserAgent(activity, userAgent)

	s.securityRepo.CreateLoginActivity(ctx, activity)
}

// parseUserAgent 解析User-Agent
func parseUserAgent(activity *model.LoginActivity, userAgent string) {
	// 简单的解析，实际应该使用专门的库
	activity.DeviceType = model.DeviceTypeDesktop
	activity.Browser = "Unknown"
	activity.OS = "Unknown"

	// 这里可以使用 user-agent 解析库来提取详细信息
	// 如: github.com/mssola/user_agent
}

// generateTempToken 生成临时token
func generateTempToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
