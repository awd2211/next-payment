package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RiskRule 风控规则
type RiskRule struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`
	RuleType    string         `gorm:"type:varchar(50);not null" json:"rule_type"`    // 规则类型：amount_limit, frequency_limit, blacklist等
	Priority    int            `gorm:"type:integer;default:0" json:"priority"`
	Conditions  string         `gorm:"type:jsonb" json:"conditions"`
	Action      string         `gorm:"type:varchar(50)" json:"action"`               // 动作：block, review, alert
	IsEnabled   bool           `gorm:"default:true" json:"is_enabled"`
	Description string         `gorm:"type:text" json:"description"`
	CreatedAt   time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (RiskRule) TableName() string {
	return "risk_rules"
}

// RiskCheck 风控检查记录
type RiskCheck struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PaymentID   uuid.UUID      `gorm:"type:uuid;index" json:"payment_id"`
	MerchantID  uuid.UUID      `gorm:"type:uuid;index" json:"merchant_id"`
	CheckType   string         `gorm:"type:varchar(50)" json:"check_type"`
	RiskScore   int            `gorm:"type:integer" json:"risk_score"`               // 风险评分 0-100
	RiskLevel   string         `gorm:"type:varchar(20)" json:"risk_level"`           // low, medium, high, critical
	MatchedRules string        `gorm:"type:jsonb" json:"matched_rules"`
	Decision    string         `gorm:"type:varchar(20)" json:"decision"`             // pass, reject, review
	Reason      string         `gorm:"type:text" json:"reason"`
	CreatedAt   time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (RiskCheck) TableName() string {
	return "risk_checks"
}

// Blacklist 黑名单
type Blacklist struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ListType   string         `gorm:"type:varchar(50);not null;index" json:"list_type"` // email, ip, card等
	Value      string         `gorm:"type:varchar(255);not null;index" json:"value"`
	Reason     string         `gorm:"type:text" json:"reason"`
	ExpiredAt  *time.Time     `gorm:"type:timestamptz" json:"expired_at"`
	CreatedAt  time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Blacklist) TableName() string {
	return "blacklist"
}

// 风控级别常量
const (
	RiskLevelLow      = "low"
	RiskLevelMedium   = "medium"
	RiskLevelHigh     = "high"
	RiskLevelCritical = "critical"
)

// 决策常量
const (
	DecisionPass   = "pass"
	DecisionReject = "reject"
	DecisionReview = "review"
)
