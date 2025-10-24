package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Merchant 商户表（核心领域模型）
type Merchant struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name         string         `gorm:"type:varchar(100);not null" json:"name"`                      // 商户名称
	Email        string         `gorm:"type:varchar(255);unique;not null;index" json:"email"`        // 邮箱
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"`                         // 密码哈希（不返回给前端）
	Phone        string         `gorm:"type:varchar(20)" json:"phone"`                               // 手机号
	CompanyName  string         `gorm:"type:varchar(200)" json:"company_name"`                       // 公司名称
	BusinessType string         `gorm:"type:varchar(50)" json:"business_type"`                       // 业务类型：individual, company
	Country      string         `gorm:"type:varchar(50)" json:"country"`                             // 国家
	Website      string         `gorm:"type:varchar(255)" json:"website"`                            // 网站
	Status       string         `gorm:"type:varchar(20);default:'pending';index" json:"status"`      // pending, active, suspended, rejected
	KYCStatus    string         `gorm:"type:varchar(20);default:'pending';index" json:"kyc_status"`  // pending, verified, rejected
	IsTestMode   bool           `gorm:"default:true" json:"is_test_mode"`                            // 是否测试模式
	Metadata     *string        `gorm:"type:jsonb" json:"metadata"`                                  // 扩展元数据（JSON）- 使用指针以支持 NULL
	CreatedAt    time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Merchant) TableName() string {
	return "merchants"
}

// 商户状态常量
const (
	MerchantStatusPending   = "pending"   // 待审核
	MerchantStatusActive    = "active"    // 活跃
	MerchantStatusSuspended = "suspended" // 暂停
	MerchantStatusRejected  = "rejected"  // 已拒绝
)

// KYC状态常量
const (
	KYCStatusPending  = "pending"  // 待审核
	KYCStatusVerified = "verified" // 已验证
	KYCStatusRejected = "rejected" // 已拒绝
)

// 业务类型常量
const (
	BusinessTypeIndividual = "individual" // 个人
	BusinessTypeCompany    = "company"    // 企业
)

// ============================================================================
// 以下模型已迁移到其他服务（Phase 1）
// ============================================================================
// - APIKey → merchant-auth-service (port 40011)
// - ChannelConfig → merchant-config-service (port 40012)
// ============================================================================
