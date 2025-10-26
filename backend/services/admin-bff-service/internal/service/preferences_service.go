package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"payment-platform/admin-service/internal/model"
	"payment-platform/admin-service/internal/repository"
)

// PreferencesService 用户偏好设置服务接口
type PreferencesService interface {
	GetPreferences(ctx context.Context, userID uuid.UUID, userType string) (*model.UserPreferences, error)
	UpdatePreferences(ctx context.Context, userID uuid.UUID, userType string, input *UpdatePreferencesInput) error
	InitDefaultPreferences(ctx context.Context, userID uuid.UUID, userType string) error
}

type preferencesService struct {
	prefsRepo repository.PreferencesRepository
}

// NewPreferencesService 创建用户偏好设置服务实例
func NewPreferencesService(prefsRepo repository.PreferencesRepository) PreferencesService {
	return &preferencesService{
		prefsRepo: prefsRepo,
	}
}

// UpdatePreferencesInput 更新偏好设置输入
type UpdatePreferencesInput struct {
	Language          *string                `json:"language"`           // 语言
	Currency          *string                `json:"currency"`           // 货币
	Timezone          *string                `json:"timezone"`           // 时区
	DateFormat        *string                `json:"date_format"`        // 日期格式
	TimeFormat        *string                `json:"time_format"`        // 时间格式
	NumberFormat      *string                `json:"number_format"`      // 数字格式
	Theme             *string                `json:"theme"`              // 主题
	DashboardLayout   map[string]interface{} `json:"dashboard_layout"`   // 仪表板布局
	NotificationPrefs map[string]interface{} `json:"notification_prefs"` // 通知偏好
}

// GetPreferences 获取用户偏好设置
func (s *preferencesService) GetPreferences(ctx context.Context, userID uuid.UUID, userType string) (*model.UserPreferences, error) {
	prefs, err := s.prefsRepo.GetByUserID(ctx, userID, userType)
	if err != nil {
		return nil, fmt.Errorf("获取偏好设置失败: %w", err)
	}

	// 如果不存在，创建默认设置
	if prefs == nil {
		if err := s.InitDefaultPreferences(ctx, userID, userType); err != nil {
			return nil, err
		}
		return s.prefsRepo.GetByUserID(ctx, userID, userType)
	}

	return prefs, nil
}

// UpdatePreferences 更新用户偏好设置
func (s *preferencesService) UpdatePreferences(ctx context.Context, userID uuid.UUID, userType string, input *UpdatePreferencesInput) error {
	prefs, err := s.GetPreferences(ctx, userID, userType)
	if err != nil {
		return err
	}

	// 更新语言
	if input.Language != nil {
		if !s.isValidLanguage(*input.Language) {
			return fmt.Errorf("不支持的语言: %s", *input.Language)
		}
		prefs.Language = *input.Language
	}

	// 更新货币
	if input.Currency != nil {
		if !s.isValidCurrency(*input.Currency) {
			return fmt.Errorf("不支持的货币: %s", *input.Currency)
		}
		prefs.Currency = *input.Currency
	}

	// 更新时区
	if input.Timezone != nil {
		if !s.isValidTimezone(*input.Timezone) {
			return fmt.Errorf("不支持的时区: %s", *input.Timezone)
		}
		prefs.Timezone = *input.Timezone
	}

	// 更新日期格式
	if input.DateFormat != nil {
		if !s.isValidDateFormat(*input.DateFormat) {
			return fmt.Errorf("不支持的日期格式: %s", *input.DateFormat)
		}
		prefs.DateFormat = *input.DateFormat
	}

	// 更新时间格式
	if input.TimeFormat != nil {
		if !s.isValidTimeFormat(*input.TimeFormat) {
			return fmt.Errorf("不支持的时间格式: %s", *input.TimeFormat)
		}
		prefs.TimeFormat = *input.TimeFormat
	}

	// 更新数字格式
	if input.NumberFormat != nil {
		if !s.isValidNumberFormat(*input.NumberFormat) {
			return fmt.Errorf("不支持的数字格式: %s", *input.NumberFormat)
		}
		prefs.NumberFormat = *input.NumberFormat
	}

	// 更新主题
	if input.Theme != nil {
		if !s.isValidTheme(*input.Theme) {
			return fmt.Errorf("不支持的主题: %s", *input.Theme)
		}
		prefs.Theme = *input.Theme
	}

	// 更新仪表板布局
	if input.DashboardLayout != nil {
		layoutJSON, err := json.Marshal(input.DashboardLayout)
		if err != nil {
			return fmt.Errorf("仪表板布局格式错误: %w", err)
		}
		prefs.DashboardLayout = string(layoutJSON)
	}

	// 更新通知偏好
	if input.NotificationPrefs != nil {
		prefsJSON, err := json.Marshal(input.NotificationPrefs)
		if err != nil {
			return fmt.Errorf("通知偏好格式错误: %w", err)
		}
		prefs.NotificationPrefs = string(prefsJSON)
	}

	return s.prefsRepo.Update(ctx, prefs)
}

// InitDefaultPreferences 初始化默认偏好设置
func (s *preferencesService) InitDefaultPreferences(ctx context.Context, userID uuid.UUID, userType string) error {
	prefs := &model.UserPreferences{
		UserID:       userID,
		UserType:     userType,
		Language:     model.LanguageEnglish,
		Currency:     model.CurrencyUSD,
		Timezone:     model.TimezoneUTC,
		DateFormat:   model.DateFormatYYYYMMDD,
		TimeFormat:   model.TimeFormat24Hour,
		NumberFormat: model.NumberFormat1234Dot56,
		Theme:        model.ThemeLight,
	}

	return s.prefsRepo.Create(ctx, prefs)
}

// 验证函数

func (s *preferencesService) isValidLanguage(lang string) bool {
	validLanguages := []string{
		model.LanguageEnglish,
		model.LanguageChineseSimplified,
		model.LanguageChineseTraditional,
		model.LanguageJapanese,
		model.LanguageKorean,
		model.LanguageSpanish,
		model.LanguageFrench,
		model.LanguageGerman,
		model.LanguagePortuguese,
		model.LanguageRussian,
		model.LanguageArabic,
		model.LanguageHindi,
	}

	for _, v := range validLanguages {
		if lang == v {
			return true
		}
	}
	return false
}

func (s *preferencesService) isValidCurrency(currency string) bool {
	validCurrencies := []string{
		model.CurrencyUSD, model.CurrencyEUR, model.CurrencyGBP, model.CurrencyCNY,
		model.CurrencyJPY, model.CurrencyKRW, model.CurrencyHKD, model.CurrencySGD,
		model.CurrencyAUD, model.CurrencyCAD, model.CurrencyINR, model.CurrencyBRL,
		model.CurrencyMXN, model.CurrencyRUB, model.CurrencyTRY, model.CurrencyZAR,
		model.CurrencyCHF, model.CurrencySEK, model.CurrencyNOK, model.CurrencyDKK,
	}

	for _, v := range validCurrencies {
		if currency == v {
			return true
		}
	}
	return false
}

func (s *preferencesService) isValidTimezone(tz string) bool {
	// 这里列出常用时区，实际应该验证更全面
	validTimezones := []string{
		model.TimezoneUTC, model.TimezoneNewYork, model.TimezoneLosAngeles,
		model.TimezoneChicago, model.TimezoneDenver, model.TimezoneLondon,
		model.TimezoneParis, model.TimezoneBerlin, model.TimezoneMoscow,
		model.TimezoneShanghai, model.TimezoneHongKong, model.TimezoneTokyo,
		model.TimezoneSeoul, model.TimezoneSingapore, model.TimezoneDubai,
		model.TimezoneSydney, model.TimezoneMelbourne, model.TimezoneToronto,
		model.TimezoneSaoPaulo,
	}

	for _, v := range validTimezones {
		if tz == v {
			return true
		}
	}
	return false
}

func (s *preferencesService) isValidDateFormat(format string) bool {
	validFormats := []string{
		model.DateFormatYYYYMMDD,
		model.DateFormatDDMMYYYY,
		model.DateFormatMMDDYYYY,
		model.DateFormatDDMonYYYY,
	}

	for _, v := range validFormats {
		if format == v {
			return true
		}
	}
	return false
}

func (s *preferencesService) isValidTimeFormat(format string) bool {
	return format == model.TimeFormat12Hour || format == model.TimeFormat24Hour
}

func (s *preferencesService) isValidNumberFormat(format string) bool {
	validFormats := []string{
		model.NumberFormat1234Dot56,
		model.NumberFormat1234Comma56,
		model.NumberFormat1234Space56,
		model.NumberFormat1234Apos56,
	}

	for _, v := range validFormats {
		if format == v {
			return true
		}
	}
	return false
}

func (s *preferencesService) isValidTheme(theme string) bool {
	return theme == model.ThemeLight || theme == model.ThemeDark || theme == model.ThemeAuto
}
