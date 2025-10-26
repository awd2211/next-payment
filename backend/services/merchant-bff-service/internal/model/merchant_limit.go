package model

import (
	"time"

	"github.com/google/uuid"
)

// MerchantLimit 商户交易额度
type MerchantLimit struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID   uuid.UUID `gorm:"type:uuid;not null;unique;index" json:"merchant_id"`

	// 限额配置
	DailyLimit   int64     `gorm:"type:bigint;not null;default:100000000" json:"daily_limit"`    // 日限额（分）默认100万
	MonthlyLimit int64     `gorm:"type:bigint;not null;default:3000000000" json:"monthly_limit"` // 月限额（分）默认3000万
	SingleLimit  int64     `gorm:"type:bigint;not null;default:10000000" json:"single_limit"`    // 单笔限额（分）默认10万

	// 已使用额度（通过Redis实时更新，DB定期同步）
	UsedToday    int64     `gorm:"type:bigint;default:0" json:"used_today"`  // 今日已用
	UsedMonth    int64     `gorm:"type:bigint;default:0" json:"used_month"`  // 本月已用

	// 统计信息
	TodayCount   int       `gorm:"type:integer;default:0" json:"today_count"`  // 今日交易笔数
	MonthCount   int       `gorm:"type:integer;default:0" json:"month_count"`  // 本月交易笔数
	LastResetDay time.Time `gorm:"type:date" json:"last_reset_day"`            // 上次日限额重置日期

	// 风控状态
	IsLimited    bool      `gorm:"default:false" json:"is_limited"`            // 是否被限额
	LimitReason  string    `gorm:"type:varchar(200)" json:"limit_reason"`      // 限额原因

	CreatedAt    time.Time `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt    time.Time `gorm:"type:timestamptz;default:now()" json:"updated_at"`
}

// TableName 指定表名
func (MerchantLimit) TableName() string {
	return "merchant_limits"
}

// GetRemainingDaily 获取剩余日限额
func (m *MerchantLimit) GetRemainingDaily() int64 {
	remaining := m.DailyLimit - m.UsedToday
	if remaining < 0 {
		return 0
	}
	return remaining
}

// GetRemainingMonthly 获取剩余月限额
func (m *MerchantLimit) GetRemainingMonthly() int64 {
	remaining := m.MonthlyLimit - m.UsedMonth
	if remaining < 0 {
		return 0
	}
	return remaining
}

// CanProcess 检查是否可以处理指定金额的交易
func (m *MerchantLimit) CanProcess(amount int64) (bool, string) {
	if m.IsLimited {
		return false, m.LimitReason
	}

	if amount > m.SingleLimit {
		return false, "超过单笔限额"
	}

	if m.UsedToday+amount > m.DailyLimit {
		return false, "超过日限额"
	}

	if m.UsedMonth+amount > m.MonthlyLimit {
		return false, "超过月限额"
	}

	return true, ""
}
