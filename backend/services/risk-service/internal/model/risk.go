package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RiskRule 风控规则
type RiskRule struct {
	ID          uuid.UUID              `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RuleName    string                 `gorm:"type:varchar(100);not null" json:"rule_name"`
	RuleType    string                 `gorm:"type:varchar(50);not null" json:"rule_type"`    // 规则类型：amount_limit, frequency_limit, blacklist等
	Priority    int                    `gorm:"type:integer;default:0" json:"priority"`
	Conditions  map[string]interface{} `gorm:"type:jsonb;serializer:json" json:"conditions"`
	Actions     map[string]interface{} `gorm:"type:jsonb;serializer:json" json:"actions"`     // 动作：block, review, alert
	Status      string                 `gorm:"type:varchar(20);default:'active'" json:"status"` // active, inactive
	Description string                 `gorm:"type:text" json:"description"`
	CreatedAt   time.Time              `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt   time.Time              `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt   gorm.DeletedAt         `gorm:"index" json:"-"`
}

func (RiskRule) TableName() string {
	return "risk_rules"
}

// RiskCheck 风控检查记录
type RiskCheck struct {
	ID          uuid.UUID              `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID  uuid.UUID              `gorm:"type:uuid;index" json:"merchant_id"`
	RelatedID   uuid.UUID              `gorm:"type:uuid;index" json:"related_id"`     // 关联的支付/订单ID
	RelatedType string                 `gorm:"type:varchar(50)" json:"related_type"` // payment, order
	CheckData   map[string]interface{} `gorm:"type:jsonb;serializer:json" json:"check_data"`
	RiskScore   int                    `gorm:"type:integer;default:0" json:"risk_score"`               // 风险评分 0-100
	RiskLevel   string                 `gorm:"type:varchar(20)" json:"risk_level"`           // low, medium, high, critical
	CheckResult map[string]interface{} `gorm:"type:jsonb;serializer:json" json:"check_result"`
	Decision    string                 `gorm:"type:varchar(20)" json:"decision"`             // pass, reject, review
	Reason      string                 `gorm:"type:text" json:"reason"`
	CreatedAt   time.Time              `gorm:"type:timestamptz;default:now()" json:"created_at"`
	DeletedAt   gorm.DeletedAt         `gorm:"index" json:"-"`
}

func (RiskCheck) TableName() string {
	return "risk_checks"
}

// Blacklist 黑名单
type Blacklist struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EntityType  string         `gorm:"type:varchar(50);not null;index" json:"entity_type"` // email, ip, card, device等
	EntityValue string         `gorm:"type:varchar(255);not null;index" json:"entity_value"`
	Reason      string         `gorm:"type:text" json:"reason"`
	AddedBy     string         `gorm:"type:varchar(100)" json:"added_by"`
	Status      string         `gorm:"type:varchar(20);default:'active'" json:"status"` // active, removed
	ExpireAt    *time.Time     `gorm:"type:timestamptz" json:"expire_at"`
	RemovedAt   *time.Time     `gorm:"type:timestamptz" json:"removed_at"`
	CreatedAt   time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
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

// 规则状态常量
const (
	RuleStatusActive   = "active"
	RuleStatusInactive = "inactive"
)
