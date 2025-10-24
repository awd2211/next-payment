package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// StringArray 字符串数组类型,用于 GORM JSON 字段
type StringArray []string

// Scan 实现 sql.Scanner 接口
func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = []string{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, s)
}

// Value 实现 driver.Valuer 接口
func (s StringArray) Value() (driver.Value, error) {
	if len(s) == 0 {
		return nil, nil
	}
	return json.Marshal(s)
}

// CashierConfig 收银台配置模型
type CashierConfig struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"merchant_id"`
	TenantID   uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`

	// 外观配置
	ThemeColor         string `gorm:"type:varchar(20);default:'#1890ff'" json:"theme_color"`
	LogoURL            string `gorm:"type:varchar(500)" json:"logo_url"`
	BackgroundImageURL string `gorm:"type:varchar(500)" json:"background_image_url"`
	CustomCSS          string `gorm:"type:text" json:"custom_css"`

	// 功能配置
	EnabledChannels  StringArray `gorm:"type:jsonb" json:"enabled_channels"`  // ["stripe", "paypal", "alipay"]
	DefaultChannel   string      `gorm:"type:varchar(50)" json:"default_channel"`
	EnabledLanguages StringArray `gorm:"type:jsonb" json:"enabled_languages"` // ["en", "zh-CN", "ja"]
	DefaultLanguage  string      `gorm:"type:varchar(10);default:'en'" json:"default_language"`

	// 支付配置
	AutoSubmit           bool `json:"auto_submit"`
	ShowAmountBreakdown  bool `gorm:"default:true" json:"show_amount_breakdown"`
	AllowChannelSwitch   bool `gorm:"default:true" json:"allow_channel_switch"`
	SessionTimeoutMinutes int  `gorm:"default:30" json:"session_timeout_minutes"`

	// 安全配置
	RequireCVV      bool        `gorm:"default:true" json:"require_cvv"`
	Enable3DSecure  bool        `gorm:"default:true" json:"enable_3d_secure"`
	AllowedCountries StringArray `gorm:"type:jsonb" json:"allowed_countries"` // ["US", "CN", "JP"]

	// 回调配置
	SuccessRedirectURL string `gorm:"type:varchar(500)" json:"success_redirect_url"`
	CancelRedirectURL  string `gorm:"type:varchar(500)" json:"cancel_redirect_url"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (CashierConfig) TableName() string {
	return "cashier_configs"
}

// CashierSession 收银台会话模型
type CashierSession struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SessionToken string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"session_token"`
	MerchantID   uuid.UUID `gorm:"type:uuid;not null;index" json:"merchant_id"`

	// 订单信息
	OrderNo     string `gorm:"type:varchar(100);not null" json:"order_no"`
	Amount      int64  `gorm:"not null" json:"amount"`      // 金额(分)
	Currency    string `gorm:"type:varchar(10);not null" json:"currency"`
	Description string `gorm:"type:text" json:"description"`

	// 客户信息
	CustomerEmail string `gorm:"type:varchar(255)" json:"customer_email"`
	CustomerName  string `gorm:"type:varchar(100)" json:"customer_name"`
	CustomerIP    string `gorm:"type:varchar(50)" json:"customer_ip"`

	// 会话配置
	AllowedChannels StringArray            `gorm:"type:jsonb" json:"allowed_channels"`
	AllowedMethods  StringArray            `gorm:"type:jsonb" json:"allowed_methods"`
	Metadata        map[string]interface{} `gorm:"type:jsonb" json:"metadata"`

	// 状态管理
	Status    string `gorm:"type:varchar(20);default:'pending';index" json:"status"` // pending/active/completed/expired
	PaymentNo string `gorm:"type:varchar(100)" json:"payment_no"`                    // 关联的支付单号

	// 时间管理
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   time.Time  `gorm:"index" json:"expires_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (CashierSession) TableName() string {
	return "cashier_sessions"
}

// CashierLog 收银台访问日志模型
type CashierLog struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SessionID  uuid.UUID `gorm:"type:uuid;index" json:"session_id"`
	MerchantID uuid.UUID `gorm:"type:uuid;not null;index" json:"merchant_id"`

	// 访问信息
	UserIP     string `gorm:"type:varchar(50)" json:"user_ip"`
	UserAgent  string `gorm:"type:text" json:"user_agent"`
	DeviceType string `gorm:"type:varchar(50)" json:"device_type"` // desktop/mobile/tablet
	Browser    string `gorm:"type:varchar(100)" json:"browser"`

	// 用户行为
	SelectedChannel  string `gorm:"type:varchar(50)" json:"selected_channel"`
	SelectedMethod   string `gorm:"type:varchar(50)" json:"selected_method"`
	FormFilled       bool   `json:"form_filled"`
	PaymentSubmitted bool   `json:"payment_submitted"`

	// 时间统计
	PageLoadTime  int `json:"page_load_time"`  // 页面加载时间(ms)
	TimeToSubmit  int `json:"time_to_submit"`  // 用户填写时间(秒)

	// 转化分析
	DroppedAtStep string `gorm:"type:varchar(50)" json:"dropped_at_step"` // channel_select/form_fill/submit
	ErrorMessage  string `gorm:"type:text" json:"error_message"`

	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (CashierLog) TableName() string {
	return "cashier_logs"
}

// CashierTemplate 收银台模板模型(平台级别)
type CashierTemplate struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`

	// 模板配置
	Config map[string]interface{} `gorm:"type:jsonb;not null" json:"config"`

	// 模板类型
	TemplateType string `gorm:"type:varchar(50)" json:"template_type"` // default/ecommerce/subscription/donation

	// 预览
	PreviewImageURL string `gorm:"type:varchar(500)" json:"preview_image_url"`

	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (CashierTemplate) TableName() string {
	return "cashier_templates"
}
