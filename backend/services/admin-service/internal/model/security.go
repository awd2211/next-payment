package model

import (
	"time"

	"github.com/google/uuid"
)

// TwoFactorAuth 双因素认证表
type TwoFactorAuth struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID     uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`      // 用户ID（Admin或Merchant）
	UserType   string     `gorm:"type:varchar(20);not null" json:"user_type"`         // admin, merchant
	Secret     string     `gorm:"type:varchar(256);not null" json:"-"`                // TOTP密钥（加密存储）
	IsEnabled  bool       `gorm:"default:false" json:"is_enabled"`                    // 是否启用
	IsVerified bool       `gorm:"default:false" json:"is_verified"`                   // 是否已验证
	BackupCodes string    `gorm:"type:jsonb" json:"-"`                                // 备用恢复代码（加密存储，JSON数组）
	CreatedAt  time.Time  `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	VerifiedAt *time.Time `gorm:"type:timestamptz" json:"verified_at"` // 验证时间
}

// TableName 指定表名
func (TwoFactorAuth) TableName() string {
	return "two_factor_auth"
}

// LoginActivity 登录活动记录表
type LoginActivity struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID        uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`                // 用户ID
	UserType      string     `gorm:"type:varchar(20);not null;index" json:"user_type"`       // admin, merchant
	LoginType     string     `gorm:"type:varchar(20);not null" json:"login_type"`            // password, 2fa, api_key
	Status        string     `gorm:"type:varchar(20);not null;index" json:"status"`          // success, failed, blocked
	IP            string     `gorm:"type:varchar(50);not null;index" json:"ip"`              // IP地址
	UserAgent     string     `gorm:"type:text" json:"user_agent"`                            // User-Agent
	DeviceType    string     `gorm:"type:varchar(20)" json:"device_type"`                    // desktop, mobile, tablet, unknown
	Browser       string     `gorm:"type:varchar(50)" json:"browser"`                        // Chrome, Firefox, Safari等
	OS            string     `gorm:"type:varchar(50)" json:"os"`                             // Windows, macOS, Linux, iOS, Android
	Country       string     `gorm:"type:varchar(50)" json:"country"`                        // 国家
	City          string     `gorm:"type:varchar(100)" json:"city"`                          // 城市
	Location      string     `gorm:"type:varchar(200)" json:"location"`                      // 完整位置信息
	IsAbnormal    bool       `gorm:"default:false;index" json:"is_abnormal"`                 // 是否异常登录
	AbnormalReason string    `gorm:"type:text" json:"abnormal_reason"`                       // 异常原因（多个原因用逗号分隔）
	FailedReason  string     `gorm:"type:varchar(200)" json:"failed_reason"`                 // 失败原因
	SessionID     string     `gorm:"type:varchar(128);index" json:"session_id"`              // 会话ID
	LoginAt       time.Time  `gorm:"type:timestamptz;default:now();index" json:"login_at"`
	LogoutAt      *time.Time `gorm:"type:timestamptz" json:"logout_at"` // 登出时间
}

// TableName 指定表名
func (LoginActivity) TableName() string {
	return "login_activities"
}

// SecuritySettings 安全设置表
type SecuritySettings struct {
	ID                    uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID                uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`             // 用户ID
	UserType              string     `gorm:"type:varchar(20);not null" json:"user_type"`                // admin, merchant
	PasswordChangedAt     *time.Time `gorm:"type:timestamptz" json:"password_changed_at"`               // 最后修改密码时间
	RequirePasswordChange bool       `gorm:"default:false" json:"require_password_change"`              // 是否要求修改密码
	PasswordExpiryDays    int        `gorm:"type:integer;default:90" json:"password_expiry_days"`       // 密码过期天数（0表示永不过期）
	SessionTimeoutMinutes int        `gorm:"type:integer;default:60" json:"session_timeout_minutes"`    // 会话超时分钟数
	MaxConcurrentSessions int        `gorm:"type:integer;default:5" json:"max_concurrent_sessions"`     // 最大并发会话数
	IPWhitelist           string     `gorm:"type:jsonb" json:"ip_whitelist"`                            // IP白名单（JSON数组）
	AllowedCountries      string     `gorm:"type:jsonb" json:"allowed_countries"`                       // 允许的国家列表（JSON数组）
	BlockedCountries      string     `gorm:"type:jsonb" json:"blocked_countries"`                       // 禁止的国家列表（JSON数组）
	LoginNotification     bool       `gorm:"default:true" json:"login_notification"`                    // 登录通知（新设备）
	AbnormalNotification  bool       `gorm:"default:true" json:"abnormal_notification"`                 // 异常活动通知
	CreatedAt             time.Time  `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt             time.Time  `gorm:"type:timestamptz;default:now()" json:"updated_at"`
}

// TableName 指定表名
func (SecuritySettings) TableName() string {
	return "security_settings"
}

// PasswordHistory 密码历史记录表（防止重复使用旧密码）
type PasswordHistory struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`              // 用户ID
	UserType     string    `gorm:"type:varchar(20);not null" json:"user_type"`           // admin, merchant
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`                  // 密码哈希
	CreatedAt    time.Time `gorm:"type:timestamptz;default:now();index" json:"created_at"`
}

// TableName 指定表名
func (PasswordHistory) TableName() string {
	return "password_history"
}

// Session 会话表
type Session struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SessionID string     `gorm:"type:varchar(128);unique;not null;index" json:"session_id"` // 会话ID
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`                   // 用户ID
	UserType  string     `gorm:"type:varchar(20);not null" json:"user_type"`                // admin, merchant
	IP        string     `gorm:"type:varchar(50);not null" json:"ip"`                       // IP地址
	UserAgent string     `gorm:"type:text" json:"user_agent"`                               // User-Agent
	Data      string     `gorm:"type:jsonb" json:"data"`                                    // 会话数据（JSON）
	ExpiresAt time.Time  `gorm:"type:timestamptz;not null;index" json:"expires_at"`         // 过期时间
	IsActive  bool       `gorm:"default:true;index" json:"is_active"`                       // 是否活跃
	CreatedAt time.Time  `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt time.Time  `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	LastSeenAt time.Time `gorm:"type:timestamptz;default:now()" json:"last_seen_at"` // 最后活跃时间
}

// TableName 指定表名
func (Session) TableName() string {
	return "sessions"
}

// 登录状态常量
const (
	LoginStatusSuccess = "success" // 成功
	LoginStatusFailed  = "failed"  // 失败
	LoginStatusBlocked = "blocked" // 被阻止
)

// 登录类型常量
const (
	LoginTypePassword = "password" // 密码登录
	LoginType2FA      = "2fa"      // 双因素认证
	LoginTypeAPIKey   = "api_key"  // API密钥
)

// 用户类型常量
const (
	UserTypeAdmin    = "admin"    // 管理员
	UserTypeMerchant = "merchant" // 商户
)

// 设备类型常量
const (
	DeviceTypeDesktop = "desktop" // 桌面
	DeviceTypeMobile  = "mobile"  // 手机
	DeviceTypeTablet  = "tablet"  // 平板
	DeviceTypeUnknown = "unknown" // 未知
)

// 异常登录原因
const (
	AbnormalReasonNewDevice   = "new_device"     // 新设备
	AbnormalReasonNewLocation = "new_location"   // 新位置
	AbnormalReasonNewIP       = "new_ip"         // 新IP
	AbnormalReasonNewCountry  = "new_country"    // 新国家
	AbnormalReasonHighRisk    = "high_risk"      // 高风险
	AbnormalReasonVPN         = "vpn"            // VPN/代理
	AbnormalReasonRateLimit   = "rate_limit"     // 频率限制
	AbnormalReasonTimezone    = "timezone"       // 时区异常
)
