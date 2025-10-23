package model

import (
	"time"

	"github.com/google/uuid"
)

// UserPreferences 用户偏好设置表
type UserPreferences struct {
	ID               uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID           uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`           // 用户ID
	UserType         string    `gorm:"type:varchar(20);not null" json:"user_type"`              // admin, merchant
	Language         string    `gorm:"type:varchar(10);default:'en'" json:"language"`           // 语言：en, zh-CN, zh-TW, ja, ko等
	Currency         string    `gorm:"type:varchar(10);default:'USD'" json:"currency"`          // 货币：USD, EUR, GBP, CNY, JPY等
	Timezone         string    `gorm:"type:varchar(50);default:'UTC'" json:"timezone"`          // 时区：UTC, Asia/Shanghai, America/New_York等
	DateFormat       string    `gorm:"type:varchar(20);default:'YYYY-MM-DD'" json:"date_format"` // 日期格式
	TimeFormat       string    `gorm:"type:varchar(20);default:'24h'" json:"time_format"`       // 时间格式：12h, 24h
	NumberFormat     string    `gorm:"type:varchar(20);default:'1,234.56'" json:"number_format"` // 数字格式
	Theme            string    `gorm:"type:varchar(20);default:'light'" json:"theme"`           // 主题：light, dark, auto
	DashboardLayout  string    `gorm:"type:jsonb" json:"dashboard_layout"`                      // 仪表板布局配置（JSON）
	NotificationPrefs string   `gorm:"type:jsonb" json:"notification_prefs"`                    // 通知偏好（JSON）
	CreatedAt        time.Time `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt        time.Time `gorm:"type:timestamptz;default:now()" json:"updated_at"`
}

// TableName 指定表名
func (UserPreferences) TableName() string {
	return "user_preferences"
}

// 支持的语言常量
const (
	LanguageEnglish            = "en"    // English
	LanguageChineseSimplified  = "zh-CN" // 简体中文
	LanguageChineseTraditional = "zh-TW" // 繁体中文
	LanguageJapanese           = "ja"    // 日本語
	LanguageKorean             = "ko"    // 한국어
	LanguageSpanish            = "es"    // Español
	LanguageFrench             = "fr"    // Français
	LanguageGerman             = "de"    // Deutsch
	LanguagePortuguese         = "pt"    // Português
	LanguageRussian            = "ru"    // Русский
	LanguageArabic             = "ar"    // العربية
	LanguageHindi              = "hi"    // हिन्दी
)

// 支持的货币常量
const (
	CurrencyUSD = "USD" // 美元
	CurrencyEUR = "EUR" // 欧元
	CurrencyGBP = "GBP" // 英镑
	CurrencyCNY = "CNY" // 人民币
	CurrencyJPY = "JPY" // 日元
	CurrencyKRW = "KRW" // 韩元
	CurrencyHKD = "HKD" // 港币
	CurrencySGD = "SGD" // 新加坡元
	CurrencyAUD = "AUD" // 澳元
	CurrencyCAD = "CAD" // 加元
	CurrencyINR = "INR" // 印度卢比
	CurrencyBRL = "BRL" // 巴西雷亚尔
	CurrencyMXN = "MXN" // 墨西哥比索
	CurrencyRUB = "RUB" // 俄罗斯卢布
	CurrencyTRY = "TRY" // 土耳其里拉
	CurrencyZAR = "ZAR" // 南非兰特
	CurrencyCHF = "CHF" // 瑞士法郎
	CurrencySEK = "SEK" // 瑞典克朗
	CurrencyNOK = "NOK" // 挪威克朗
	CurrencyDKK = "DKK" // 丹麦克朗
)

// 常用时区常量
const (
	TimezoneUTC          = "UTC"                // UTC
	TimezoneNewYork      = "America/New_York"   // 纽约（EST/EDT）
	TimezoneLosAngeles   = "America/Los_Angeles" // 洛杉矶（PST/PDT）
	TimezoneChicago      = "America/Chicago"    // 芝加哥（CST/CDT）
	TimezoneDenver       = "America/Denver"     // 丹佛（MST/MDT）
	TimezoneLondon       = "Europe/London"      // 伦敦（GMT/BST）
	TimezoneParis        = "Europe/Paris"       // 巴黎（CET/CEST）
	TimezoneBerlin       = "Europe/Berlin"      // 柏林（CET/CEST）
	TimezoneMoscow       = "Europe/Moscow"      // 莫斯科（MSK）
	TimezoneShanghai     = "Asia/Shanghai"      // 上海（CST）
	TimezoneHongKong     = "Asia/Hong_Kong"     // 香港（HKT）
	TimezoneTokyo        = "Asia/Tokyo"         // 东京（JST）
	TimezoneSeoul        = "Asia/Seoul"         // 首尔（KST）
	TimezoneSingapore    = "Asia/Singapore"     // 新加坡（SGT）
	TimezoneDubai        = "Asia/Dubai"         // 迪拜（GST）
	TimezoneSydney       = "Australia/Sydney"   // 悉尼（AEDT/AEST）
	TimezoneMelbourne    = "Australia/Melbourne" // 墨尔本（AEDT/AEST）
	TimezoneToronto      = "America/Toronto"    // 多伦多（EST/EDT）
	TimezoneSaoPaulo     = "America/Sao_Paulo"  // 圣保罗（BRT）
)

// 日期格式常量
const (
	DateFormatYYYYMMDD = "YYYY-MM-DD" // 2024-01-15
	DateFormatDDMMYYYY = "DD/MM/YYYY" // 15/01/2024
	DateFormatMMDDYYYY = "MM/DD/YYYY" // 01/15/2024
	DateFormatDDMonYYYY = "DD-Mon-YYYY" // 15-Jan-2024
)

// 时间格式常量
const (
	TimeFormat12Hour = "12h" // 12小时制（AM/PM）
	TimeFormat24Hour = "24h" // 24小时制
)

// 数字格式常量
const (
	NumberFormat1234Dot56    = "1,234.56"   // 英语（美国、英国等）
	NumberFormat1234Comma56  = "1.234,56"   // 欧洲（德国、西班牙等）
	NumberFormat1234Space56  = "1 234,56"   // 法语
	NumberFormat1234Apos56   = "1'234.56"   // 瑞士
)

// 主题常量
const (
	ThemeLight = "light" // 浅色主题
	ThemeDark  = "dark"  // 深色主题
	ThemeAuto  = "auto"  // 自动（跟随系统）
)
