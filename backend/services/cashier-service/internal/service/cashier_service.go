package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/logger"
	"go.uber.org/zap"
	"payment-platform/cashier-service/internal/model"
	"payment-platform/cashier-service/internal/repository"
)

// CashierService 收银台服务接口
type CashierService interface {
	// 配置管理
	CreateOrUpdateConfig(ctx context.Context, merchantID uuid.UUID, input *ConfigInput) (*model.CashierConfig, error)
	GetConfig(ctx context.Context, merchantID uuid.UUID) (*model.CashierConfig, error)
	DeleteConfig(ctx context.Context, merchantID uuid.UUID) error

	// 会话管理
	CreateSession(ctx context.Context, input *SessionInput) (*model.CashierSession, string, error)
	GetSession(ctx context.Context, sessionToken string) (*model.CashierSession, error)
	CompleteSession(ctx context.Context, sessionToken string, paymentNo string) error
	CancelSession(ctx context.Context, sessionToken string) error

	// 日志记录
	RecordLog(ctx context.Context, input *LogInput) error

	// 统计分析
	GetAnalytics(ctx context.Context, merchantID uuid.UUID, startTime, endTime time.Time) (*AnalyticsData, error)

	// 管理员API
	ListTemplates(ctx context.Context) ([]*model.CashierTemplate, error)
	CreateTemplate(ctx context.Context, input *TemplateInput) (*model.CashierTemplate, error)
	UpdateTemplate(ctx context.Context, id uuid.UUID, input *TemplateInput) (*model.CashierTemplate, error)
	DeleteTemplate(ctx context.Context, id uuid.UUID) error
	GetPlatformStats(ctx context.Context) (*PlatformStats, error)
}

type cashierService struct {
	repo repository.CashierRepository
}

// NewCashierService 创建收银台服务实例
func NewCashierService(repo repository.CashierRepository) CashierService {
	return &cashierService{repo: repo}
}

// ConfigInput 配置输入
type ConfigInput struct {
	ThemeColor            string   `json:"theme_color"`
	LogoURL               string   `json:"logo_url"`
	BackgroundImageURL    string   `json:"background_image_url"`
	CustomCSS             string   `json:"custom_css"`
	EnabledChannels       []string `json:"enabled_channels"`
	DefaultChannel        string   `json:"default_channel"`
	EnabledLanguages      []string `json:"enabled_languages"`
	DefaultLanguage       string   `json:"default_language"`
	AutoSubmit            bool     `json:"auto_submit"`
	ShowAmountBreakdown   bool     `json:"show_amount_breakdown"`
	AllowChannelSwitch    bool     `json:"allow_channel_switch"`
	SessionTimeoutMinutes int      `json:"session_timeout_minutes"`
	RequireCVV            bool     `json:"require_cvv"`
	Enable3DSecure        bool     `json:"enable_3d_secure"`
	AllowedCountries      []string `json:"allowed_countries"`
	SuccessRedirectURL    string   `json:"success_redirect_url"`
	CancelRedirectURL     string   `json:"cancel_redirect_url"`
}

// SessionInput 会话输入
type SessionInput struct {
	MerchantID      uuid.UUID              `json:"merchant_id"`
	OrderNo         string                 `json:"order_no"`
	Amount          int64                  `json:"amount"`
	Currency        string                 `json:"currency"`
	Description     string                 `json:"description"`
	CustomerEmail   string                 `json:"customer_email"`
	CustomerName    string                 `json:"customer_name"`
	CustomerIP      string                 `json:"customer_ip"`
	AllowedChannels []string               `json:"allowed_channels"`
	AllowedMethods  []string               `json:"allowed_methods"`
	Metadata        map[string]interface{} `json:"metadata"`
	ExpiresInMinutes int                    `json:"expires_in_minutes"`
}

// LogInput 日志输入
type LogInput struct {
	SessionToken     string `json:"session_token"`
	UserIP           string `json:"user_ip"`
	UserAgent        string `json:"user_agent"`
	DeviceType       string `json:"device_type"`
	Browser          string `json:"browser"`
	SelectedChannel  string `json:"selected_channel"`
	SelectedMethod   string `json:"selected_method"`
	FormFilled       bool   `json:"form_filled"`
	PaymentSubmitted bool   `json:"payment_submitted"`
	PageLoadTime     int    `json:"page_load_time"`
	TimeToSubmit     int    `json:"time_to_submit"`
	DroppedAtStep    string `json:"dropped_at_step"`
	ErrorMessage     string `json:"error_message"`
}

// AnalyticsData 统计数据
type AnalyticsData struct {
	ConversionRate float64           `json:"conversion_rate"`
	ChannelStats   map[string]int    `json:"channel_stats"`
	TotalSessions  int64             `json:"total_sessions"`
}

// TemplateInput 模板输入
type TemplateInput struct {
	Name              string                 `json:"name"`
	Description       string                 `json:"description"`
	TemplateType      string                 `json:"template_type"`
	Config            map[string]interface{} `json:"config"`
	PreviewImageURL   string                 `json:"preview_image_url"`
	IsActive          bool                   `json:"is_active"`
}

// PlatformStats 平台统计
type PlatformStats struct {
	TotalMerchants          int     `json:"total_merchants"`
	ActiveCashiers          int     `json:"active_cashiers"`
	TotalSessions           int64   `json:"total_sessions"`
	AvgConversionRate       float64 `json:"avg_conversion_rate"`
	TotalSessionsToday      int64   `json:"total_sessions_today"`
	CompletedSessionsToday  int64   `json:"completed_sessions_today"`
}

// CreateOrUpdateConfig 创建或更新配置
func (s *cashierService) CreateOrUpdateConfig(ctx context.Context, merchantID uuid.UUID, input *ConfigInput) (*model.CashierConfig, error) {
	config := &model.CashierConfig{
		MerchantID:            merchantID,
		TenantID:              merchantID, // 暂时使用 merchantID 作为 tenantID
		ThemeColor:            input.ThemeColor,
		LogoURL:               input.LogoURL,
		BackgroundImageURL:    input.BackgroundImageURL,
		CustomCSS:             input.CustomCSS,
		EnabledChannels:       input.EnabledChannels,
		DefaultChannel:        input.DefaultChannel,
		EnabledLanguages:      input.EnabledLanguages,
		DefaultLanguage:       input.DefaultLanguage,
		AutoSubmit:            input.AutoSubmit,
		ShowAmountBreakdown:   input.ShowAmountBreakdown,
		AllowChannelSwitch:    input.AllowChannelSwitch,
		SessionTimeoutMinutes: input.SessionTimeoutMinutes,
		RequireCVV:            input.RequireCVV,
		Enable3DSecure:        input.Enable3DSecure,
		AllowedCountries:      input.AllowedCountries,
		SuccessRedirectURL:    input.SuccessRedirectURL,
		CancelRedirectURL:     input.CancelRedirectURL,
	}

	err := s.repo.CreateOrUpdateConfig(ctx, config)
	if err != nil {
		logger.Error("Failed to create/update cashier config",
			zap.String("merchant_id", merchantID.String()),
			zap.Error(err))
		return nil, fmt.Errorf("创建/更新配置失败: %w", err)
	}

	return config, nil
}

// GetConfig 获取配置
func (s *cashierService) GetConfig(ctx context.Context, merchantID uuid.UUID) (*model.CashierConfig, error) {
	config, err := s.repo.GetConfig(ctx, merchantID)
	if err != nil {
		logger.Error("Failed to get cashier config",
			zap.String("merchant_id", merchantID.String()),
			zap.Error(err))
		return nil, fmt.Errorf("获取配置失败: %w", err)
	}

	// 如果没有配置,返回默认配置
	if config == nil {
		config = &model.CashierConfig{
			MerchantID:            merchantID,
			TenantID:              merchantID,
			ThemeColor:            "#1890ff",
			DefaultLanguage:       "en",
			ShowAmountBreakdown:   true,
			AllowChannelSwitch:    true,
			SessionTimeoutMinutes: 30,
			RequireCVV:            true,
			Enable3DSecure:        true,
		}
	}

	return config, nil
}

// DeleteConfig 删除配置
func (s *cashierService) DeleteConfig(ctx context.Context, merchantID uuid.UUID) error {
	err := s.repo.DeleteConfig(ctx, merchantID)
	if err != nil {
		logger.Error("Failed to delete cashier config",
			zap.String("merchant_id", merchantID.String()),
			zap.Error(err))
		return fmt.Errorf("删除配置失败: %w", err)
	}
	return nil
}

// CreateSession 创建会话
func (s *cashierService) CreateSession(ctx context.Context, input *SessionInput) (*model.CashierSession, string, error) {
	// 生成会话token
	sessionToken, err := generateSessionToken()
	if err != nil {
		return nil, "", fmt.Errorf("生成会话token失败: %w", err)
	}

	// 计算过期时间
	expiresInMinutes := input.ExpiresInMinutes
	if expiresInMinutes <= 0 {
		expiresInMinutes = 30 // 默认30分钟
	}
	expiresAt := time.Now().Add(time.Duration(expiresInMinutes) * time.Minute)

	session := &model.CashierSession{
		SessionToken:    sessionToken,
		MerchantID:      input.MerchantID,
		OrderNo:         input.OrderNo,
		Amount:          input.Amount,
		Currency:        input.Currency,
		Description:     input.Description,
		CustomerEmail:   input.CustomerEmail,
		CustomerName:    input.CustomerName,
		CustomerIP:      input.CustomerIP,
		AllowedChannels: input.AllowedChannels,
		AllowedMethods:  input.AllowedMethods,
		Metadata:        input.Metadata,
		Status:          "pending",
		ExpiresAt:       expiresAt,
	}

	err = s.repo.CreateSession(ctx, session)
	if err != nil {
		logger.Error("Failed to create cashier session",
			zap.String("merchant_id", input.MerchantID.String()),
			zap.Error(err))
		return nil, "", fmt.Errorf("创建会话失败: %w", err)
	}

	logger.Info("Created cashier session",
		zap.String("session_token", sessionToken),
		zap.String("merchant_id", input.MerchantID.String()))

	return session, sessionToken, nil
}

// GetSession 获取会话
func (s *cashierService) GetSession(ctx context.Context, sessionToken string) (*model.CashierSession, error) {
	session, err := s.repo.GetSession(ctx, sessionToken)
	if err != nil {
		logger.Error("Failed to get cashier session",
			zap.String("session_token", sessionToken),
			zap.Error(err))
		return nil, fmt.Errorf("获取会话失败: %w", err)
	}

	if session == nil {
		return nil, fmt.Errorf("会话不存在")
	}

	// 检查是否过期
	if session.Status == "pending" && time.Now().After(session.ExpiresAt) {
		session.Status = "expired"
		s.repo.UpdateSession(ctx, session)
		return nil, fmt.Errorf("会话已过期")
	}

	return session, nil
}

// CompleteSession 完成会话
func (s *cashierService) CompleteSession(ctx context.Context, sessionToken string, paymentNo string) error {
	session, err := s.repo.GetSession(ctx, sessionToken)
	if err != nil {
		return fmt.Errorf("获取会话失败: %w", err)
	}

	if session == nil {
		return fmt.Errorf("会话不存在")
	}

	now := time.Now()
	session.Status = "completed"
	session.PaymentNo = paymentNo
	session.CompletedAt = &now

	err = s.repo.UpdateSession(ctx, session)
	if err != nil {
		logger.Error("Failed to complete cashier session",
			zap.String("session_token", sessionToken),
			zap.Error(err))
		return fmt.Errorf("完成会话失败: %w", err)
	}

	logger.Info("Completed cashier session",
		zap.String("session_token", sessionToken),
		zap.String("payment_no", paymentNo))

	return nil
}

// CancelSession 取消会话
func (s *cashierService) CancelSession(ctx context.Context, sessionToken string) error {
	return s.repo.DeleteSession(ctx, sessionToken)
}

// RecordLog 记录日志
func (s *cashierService) RecordLog(ctx context.Context, input *LogInput) error {
	// 获取会话信息
	session, err := s.repo.GetSession(ctx, input.SessionToken)
	if err != nil || session == nil {
		logger.Error("Session not found for log recording",
			zap.String("session_token", input.SessionToken))
		return fmt.Errorf("会话不存在")
	}

	log := &model.CashierLog{
		SessionID:        session.ID,
		MerchantID:       session.MerchantID,
		UserIP:           input.UserIP,
		UserAgent:        input.UserAgent,
		DeviceType:       input.DeviceType,
		Browser:          input.Browser,
		SelectedChannel:  input.SelectedChannel,
		SelectedMethod:   input.SelectedMethod,
		FormFilled:       input.FormFilled,
		PaymentSubmitted: input.PaymentSubmitted,
		PageLoadTime:     input.PageLoadTime,
		TimeToSubmit:     input.TimeToSubmit,
		DroppedAtStep:    input.DroppedAtStep,
		ErrorMessage:     input.ErrorMessage,
	}

	err = s.repo.CreateLog(ctx, log)
	if err != nil {
		logger.Error("Failed to create cashier log",
			zap.String("session_token", input.SessionToken),
			zap.Error(err))
		return fmt.Errorf("记录日志失败: %w", err)
	}

	return nil
}

// GetAnalytics 获取统计数据
func (s *cashierService) GetAnalytics(ctx context.Context, merchantID uuid.UUID, startTime, endTime time.Time) (*AnalyticsData, error) {
	// 获取转化率
	conversionRate, err := s.repo.GetConversionRate(ctx, merchantID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("获取转化率失败: %w", err)
	}

	// 获取渠道统计
	channelStats, err := s.repo.GetChannelStats(ctx, merchantID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("获取渠道统计失败: %w", err)
	}

	return &AnalyticsData{
		ConversionRate: conversionRate,
		ChannelStats:   channelStats,
	}, nil
}

// generateSessionToken 生成会话token
func generateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// ListTemplates 列出所有模板
func (s *cashierService) ListTemplates(ctx context.Context) ([]*model.CashierTemplate, error) {
	templates, err := s.repo.ListTemplates(ctx)
	if err != nil {
		logger.Error("Failed to list templates", zap.Error(err))
		return nil, fmt.Errorf("列出模板失败: %w", err)
	}
	return templates, nil
}

// CreateTemplate 创建模板
func (s *cashierService) CreateTemplate(ctx context.Context, input *TemplateInput) (*model.CashierTemplate, error) {
	template := &model.CashierTemplate{
		Name:            input.Name,
		Description:     input.Description,
		TemplateType:    input.TemplateType,
		Config:          input.Config,
		PreviewImageURL: input.PreviewImageURL,
		IsActive:        input.IsActive,
	}

	err := s.repo.CreateTemplate(ctx, template)
	if err != nil {
		logger.Error("Failed to create template", zap.String("name", input.Name), zap.Error(err))
		return nil, fmt.Errorf("创建模板失败: %w", err)
	}

	logger.Info("Created template", zap.String("name", input.Name), zap.String("id", template.ID.String()))
	return template, nil
}

// UpdateTemplate 更新模板
func (s *cashierService) UpdateTemplate(ctx context.Context, id uuid.UUID, input *TemplateInput) (*model.CashierTemplate, error) {
	template, err := s.repo.GetTemplate(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取模板失败: %w", err)
	}
	if template == nil {
		return nil, fmt.Errorf("模板不存在")
	}

	template.Name = input.Name
	template.Description = input.Description
	template.TemplateType = input.TemplateType
	template.Config = input.Config
	template.PreviewImageURL = input.PreviewImageURL
	template.IsActive = input.IsActive

	err = s.repo.UpdateTemplate(ctx, template)
	if err != nil {
		logger.Error("Failed to update template", zap.String("id", id.String()), zap.Error(err))
		return nil, fmt.Errorf("更新模板失败: %w", err)
	}

	logger.Info("Updated template", zap.String("id", id.String()), zap.String("name", input.Name))
	return template, nil
}

// DeleteTemplate 删除模板
func (s *cashierService) DeleteTemplate(ctx context.Context, id uuid.UUID) error {
	err := s.repo.DeleteTemplate(ctx, id)
	if err != nil {
		logger.Error("Failed to delete template", zap.String("id", id.String()), zap.Error(err))
		return fmt.Errorf("删除模板失败: %w", err)
	}

	logger.Info("Deleted template", zap.String("id", id.String()))
	return nil
}

// GetPlatformStats 获取平台统计
func (s *cashierService) GetPlatformStats(ctx context.Context) (*PlatformStats, error) {
	stats := &PlatformStats{}

	// 获取活跃商户数
	merchantCount, err := s.repo.GetActiveMerchantCount(ctx)
	if err != nil {
		logger.Error("Failed to get merchant count", zap.Error(err))
	} else {
		stats.TotalMerchants = merchantCount
		stats.ActiveCashiers = merchantCount
	}

	// 获取今日会话数
	startOfDay := time.Now().Truncate(24 * time.Hour)
	endOfDay := startOfDay.Add(24 * time.Hour)

	todaySessions, err := s.repo.GetSessionCount(ctx, startOfDay, endOfDay)
	if err != nil {
		logger.Error("Failed to get today sessions", zap.Error(err))
	} else {
		stats.TotalSessionsToday = todaySessions
	}

	// 获取今日完成会话数
	completedToday, err := s.repo.GetCompletedSessionCount(ctx, startOfDay, endOfDay)
	if err != nil {
		logger.Error("Failed to get completed sessions", zap.Error(err))
	} else {
		stats.CompletedSessionsToday = completedToday
	}

	// 计算平均转化率 (过去7天)
	weekAgo := time.Now().AddDate(0, 0, -7)
	avgRate, err := s.repo.GetAverageConversionRate(ctx, weekAgo, time.Now())
	if err != nil {
		logger.Error("Failed to get avg conversion rate", zap.Error(err))
	} else {
		stats.AvgConversionRate = avgRate
	}

	return stats, nil
}
