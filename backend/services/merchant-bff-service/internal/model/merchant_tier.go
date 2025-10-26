package model

import (
	"time"

	"github.com/google/uuid"
)

// MerchantTier 商户等级
type MerchantTier string

const (
	TierStarter    MerchantTier = "starter"     // 入门版：个人/小微商户
	TierBusiness   MerchantTier = "business"    // 商业版：中小企业
	TierEnterprise MerchantTier = "enterprise"  // 企业版：大型企业
	TierPremium    MerchantTier = "premium"     // 尊享版：超大型客户
)

// MerchantTierConfig 商户等级配置
type MerchantTierConfig struct {
	ID        uuid.UUID    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Tier      MerchantTier `json:"tier" gorm:"type:varchar(20);not null;unique;index"`
	Name      string       `json:"name" gorm:"type:varchar(100);not null"` // 等级名称
	NameEn    string       `json:"name_en" gorm:"type:varchar(100)"`        // 英文名称

	// 交易限额
	DailyLimit   int64 `json:"daily_limit" gorm:"type:bigint;not null"`   // 日交易限额（分）
	MonthlyLimit int64 `json:"monthly_limit" gorm:"type:bigint;not null"` // 月交易限额（分）
	SingleLimit  int64 `json:"single_limit" gorm:"type:bigint;not null"`  // 单笔限额（分）

	// 费率配置
	FeeRate         float64 `json:"fee_rate" gorm:"type:decimal(5,4);not null"`          // 交易费率（例如 0.006 = 0.6%）
	MinFee          int64   `json:"min_fee" gorm:"type:bigint;default:0"`                // 最低手续费（分）
	RefundFeeRate   float64 `json:"refund_fee_rate" gorm:"type:decimal(5,4);default:0"`  // 退款费率
	WithdrawalFee   int64   `json:"withdrawal_fee" gorm:"type:bigint;default:0"`         // 提现手续费（分）
	WithdrawalFeeRate float64 `json:"withdrawal_fee_rate" gorm:"type:decimal(5,4);default:0"` // 提现费率

	// 结算周期
	SettlementCycle   string `json:"settlement_cycle" gorm:"type:varchar(20);not null"`   // T+1, T+0, D+0
	AutoSettlement    bool   `json:"auto_settlement" gorm:"default:false"`                // 是否自动结算
	MinSettlementAmount int64 `json:"min_settlement_amount" gorm:"type:bigint;default:0"` // 最低结算金额

	// 功能权限
	EnableMultiCurrency  bool `json:"enable_multi_currency" gorm:"default:false"`  // 多币种支持
	EnableRefund         bool `json:"enable_refund" gorm:"default:true"`           // 退款功能
	EnablePartialRefund  bool `json:"enable_partial_refund" gorm:"default:false"`  // 部分退款
	EnablePreAuth        bool `json:"enable_pre_auth" gorm:"default:false"`        // 预授权
	EnableRecurring      bool `json:"enable_recurring" gorm:"default:false"`       // 循环扣款
	EnableSplit          bool `json:"enable_split" gorm:"default:false"`           // 分账功能
	EnableWebhook        bool `json:"enable_webhook" gorm:"default:true"`          // Webhook通知
	MaxWebhookRetry      int  `json:"max_webhook_retry" gorm:"default:3"`          // Webhook重试次数

	// API限制
	APIRateLimit      int `json:"api_rate_limit" gorm:"default:100"`       // API请求限制（次/分钟）
	MaxAPIKeys        int `json:"max_api_keys" gorm:"default:2"`           // 最大API密钥数量
	EnableAPICallback bool `json:"enable_api_callback" gorm:"default:true"` // API回调

	// 风控配置
	RiskLevel         string `json:"risk_level" gorm:"type:varchar(20);default:'medium'"` // low, medium, high
	EnableRiskControl bool   `json:"enable_risk_control" gorm:"default:true"`             // 风控开关
	MaxDailyFailures  int    `json:"max_daily_failures" gorm:"default:100"`               // 每日最大失败次数

	// 技术支持
	SupportLevel    string `json:"support_level" gorm:"type:varchar(20);default:'standard'"` // standard, priority, vip
	SLAResponseTime int    `json:"sla_response_time" gorm:"default:24"`                      // SLA响应时间（小时）
	DedicatedSupport bool  `json:"dedicated_support" gorm:"default:false"`                   // 专属客服

	// 其他限制
	MaxSubAccounts    int    `json:"max_sub_accounts" gorm:"default:1"`           // 最大子账户数
	DataRetention     int    `json:"data_retention" gorm:"default:90"`            // 数据保留天数
	CustomBranding    bool   `json:"custom_branding" gorm:"default:false"`        // 自定义品牌
	Priority          int    `json:"priority" gorm:"default:0"`                   // 优先级（数字越大优先级越高）
	Description       string `json:"description" gorm:"type:text"`                // 等级描述
	DescriptionEn     string `json:"description_en" gorm:"type:text"`             // 英文描述

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetTierConfig 获取默认等级配置
func GetDefaultTierConfig(tier MerchantTier) *MerchantTierConfig {
	configs := map[MerchantTier]*MerchantTierConfig{
		TierStarter: {
			Tier:                TierStarter,
			Name:                "入门版",
			NameEn:              "Starter",
			DailyLimit:          100_00_00,    // 10万/日
			MonthlyLimit:        300_00_00,    // 30万/月
			SingleLimit:         10_00_00,     // 1万/笔
			FeeRate:             0.0080,       // 0.8%
			MinFee:              100,          // 1元
			RefundFeeRate:       0.0000,       // 免费
			WithdrawalFee:       200,          // 2元/笔
			WithdrawalFeeRate:   0.0010,       // 0.1%
			SettlementCycle:     "T+1",
			AutoSettlement:      false,
			MinSettlementAmount: 100_00,       // 100元
			EnableMultiCurrency: false,
			EnableRefund:        true,
			EnablePartialRefund: false,
			EnablePreAuth:       false,
			EnableRecurring:     false,
			EnableSplit:         false,
			EnableWebhook:       true,
			MaxWebhookRetry:     3,
			APIRateLimit:        100,          // 100次/分钟
			MaxAPIKeys:          2,
			EnableAPICallback:   true,
			RiskLevel:           "medium",
			EnableRiskControl:   true,
			MaxDailyFailures:    50,
			SupportLevel:        "standard",
			SLAResponseTime:     24,           // 24小时
			DedicatedSupport:    false,
			MaxSubAccounts:      1,
			DataRetention:       90,           // 90天
			CustomBranding:      false,
			Priority:            1,
			Description:         "适合个人和小微商户，快速接入，基础功能齐全",
			DescriptionEn:       "Ideal for individuals and small businesses with essential features",
		},
		TierBusiness: {
			Tier:                TierBusiness,
			Name:                "商业版",
			NameEn:              "Business",
			DailyLimit:          500_00_00,    // 50万/日
			MonthlyLimit:        1500_00_00,   // 150万/月
			SingleLimit:         50_00_00,     // 5万/笔
			FeeRate:             0.0060,       // 0.6%
			MinFee:              50,           // 0.5元
			RefundFeeRate:       0.0000,
			WithdrawalFee:       100,          // 1元/笔
			WithdrawalFeeRate:   0.0005,       // 0.05%
			SettlementCycle:     "T+1",
			AutoSettlement:      true,
			MinSettlementAmount: 50_00,        // 50元
			EnableMultiCurrency: true,
			EnableRefund:        true,
			EnablePartialRefund: true,
			EnablePreAuth:       false,
			EnableRecurring:     false,
			EnableSplit:         false,
			EnableWebhook:       true,
			MaxWebhookRetry:     5,
			APIRateLimit:        500,          // 500次/分钟
			MaxAPIKeys:          5,
			EnableAPICallback:   true,
			RiskLevel:           "medium",
			EnableRiskControl:   true,
			MaxDailyFailures:    200,
			SupportLevel:        "priority",
			SLAResponseTime:     12,           // 12小时
			DedicatedSupport:    false,
			MaxSubAccounts:      5,
			DataRetention:       180,          // 180天
			CustomBranding:      false,
			Priority:            2,
			Description:         "适合中小企业，支持多币种，更低费率，优先技术支持",
			DescriptionEn:       "Perfect for SMEs with multi-currency support and lower fees",
		},
		TierEnterprise: {
			Tier:                TierEnterprise,
			Name:                "企业版",
			NameEn:              "Enterprise",
			DailyLimit:          2000_00_00,   // 200万/日
			MonthlyLimit:        6000_00_00,   // 600万/月
			SingleLimit:         200_00_00,    // 20万/笔
			FeeRate:             0.0045,       // 0.45%
			MinFee:              20,           // 0.2元
			RefundFeeRate:       0.0000,
			WithdrawalFee:       0,            // 免费
			WithdrawalFeeRate:   0.0000,
			SettlementCycle:     "T+0",
			AutoSettlement:      true,
			MinSettlementAmount: 10_00,        // 10元
			EnableMultiCurrency: true,
			EnableRefund:        true,
			EnablePartialRefund: true,
			EnablePreAuth:       true,
			EnableRecurring:     true,
			EnableSplit:         true,
			EnableWebhook:       true,
			MaxWebhookRetry:     10,
			APIRateLimit:        2000,         // 2000次/分钟
			MaxAPIKeys:          10,
			EnableAPICallback:   true,
			RiskLevel:           "low",
			EnableRiskControl:   true,
			MaxDailyFailures:    500,
			SupportLevel:        "vip",
			SLAResponseTime:     4,            // 4小时
			DedicatedSupport:    true,
			MaxSubAccounts:      20,
			DataRetention:       365,          // 365天
			CustomBranding:      true,
			Priority:            3,
			Description:         "适合大型企业，T+0结算，预授权/分账/循环扣款，专属客服",
			DescriptionEn:       "Designed for large enterprises with T+0 settlement and advanced features",
		},
		TierPremium: {
			Tier:                TierPremium,
			Name:                "尊享版",
			NameEn:              "Premium",
			DailyLimit:          10000_00_00,  // 1000万/日
			MonthlyLimit:        30000_00_00,  // 3000万/月
			SingleLimit:         1000_00_00,   // 100万/笔
			FeeRate:             0.0030,       // 0.3%
			MinFee:              0,            // 无最低费用
			RefundFeeRate:       0.0000,
			WithdrawalFee:       0,
			WithdrawalFeeRate:   0.0000,
			SettlementCycle:     "D+0",        // 当日结算
			AutoSettlement:      true,
			MinSettlementAmount: 0,
			EnableMultiCurrency: true,
			EnableRefund:        true,
			EnablePartialRefund: true,
			EnablePreAuth:       true,
			EnableRecurring:     true,
			EnableSplit:         true,
			EnableWebhook:       true,
			MaxWebhookRetry:     20,
			APIRateLimit:        10000,        // 10000次/分钟
			MaxAPIKeys:          50,
			EnableAPICallback:   true,
			RiskLevel:           "low",
			EnableRiskControl:   true,
			MaxDailyFailures:    2000,
			SupportLevel:        "vip",
			SLAResponseTime:     1,            // 1小时
			DedicatedSupport:    true,
			MaxSubAccounts:      100,
			DataRetention:       730,          // 2年
			CustomBranding:      true,
			Priority:            4,
			Description:         "超大型客户专享，D+0极速结算，最低费率，7x24专属服务",
			DescriptionEn:       "Premium tier for enterprise clients with D+0 settlement and VIP support",
		},
	}

	return configs[tier]
}

// CalculateFee 计算手续费
func (c *MerchantTierConfig) CalculateFee(amount int64) int64 {
	fee := int64(float64(amount) * c.FeeRate)
	if fee < c.MinFee {
		fee = c.MinFee
	}
	return fee
}

// CalculateWithdrawalFee 计算提现手续费
func (c *MerchantTierConfig) CalculateWithdrawalFee(amount int64) int64 {
	fee := int64(float64(amount)*c.WithdrawalFeeRate) + c.WithdrawalFee
	return fee
}

// CanProcess 检查是否可以处理该金额
func (c *MerchantTierConfig) CanProcess(amount int64, dailyUsed, monthlyUsed int64) (bool, string) {
	if amount > c.SingleLimit {
		return false, "超过单笔交易限额"
	}
	if dailyUsed+amount > c.DailyLimit {
		return false, "超过日交易限额"
	}
	if monthlyUsed+amount > c.MonthlyLimit {
		return false, "超过月交易限额"
	}
	return true, ""
}

// CanUpgradeTo 检查是否可以升级到目标等级
func (c *MerchantTierConfig) CanUpgradeTo(target MerchantTier) bool {
	tierLevels := map[MerchantTier]int{
		TierStarter:    1,
		TierBusiness:   2,
		TierEnterprise: 3,
		TierPremium:    4,
	}

	currentLevel := tierLevels[c.Tier]
	targetLevel := tierLevels[target]

	return targetLevel > currentLevel
}
