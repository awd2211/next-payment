package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Admin 管理员表
type Admin struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Username     string         `gorm:"type:varchar(50);unique;not null" json:"username"`
	Email        string         `gorm:"type:varchar(255);unique;not null" json:"email"`
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"` // 不返回给前端
	FullName     string         `gorm:"type:varchar(100)" json:"full_name"`
	Phone        string         `gorm:"type:varchar(20)" json:"phone"`
	Avatar       string         `gorm:"type:text" json:"avatar"`
	Status       string         `gorm:"type:varchar(20);default:'active'" json:"status"` // active/disabled/locked
	IsSuper      bool           `gorm:"default:false" json:"is_super"`                   // 超级管理员
	LastLoginAt  *time.Time     `gorm:"type:timestamptz" json:"last_login_at"`
	LastLoginIP  string         `gorm:"type:varchar(50)" json:"last_login_ip"`
	CreatedAt    time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	Roles []Role `gorm:"many2many:admin_roles;" json:"roles"`
}

// Role 角色表
type Role struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string         `gorm:"type:varchar(50);unique;not null" json:"name"`
	DisplayName string         `gorm:"type:varchar(100);not null" json:"display_name"`
	Description string         `gorm:"type:text" json:"description"`
	IsSystem    bool           `gorm:"default:false" json:"is_system"` // 系统内置角色不可删除
	CreatedAt   time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions"`
}

// Permission 权限表
type Permission struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Code        string         `gorm:"type:varchar(100);unique;not null" json:"code"` // 如：merchant.view, merchant.edit
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`
	Resource    string         `gorm:"type:varchar(50);not null" json:"resource"`     // 资源：merchant, order, payment
	Action      string         `gorm:"type:varchar(50);not null" json:"action"`       // 操作：view, create, edit, delete
	Description string         `gorm:"type:text" json:"description"`
	CreatedAt   time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
}

// AdminRole 管理员角色关联表
type AdminRole struct {
	AdminID   uuid.UUID `gorm:"type:uuid;not null" json:"admin_id"`
	RoleID    uuid.UUID `gorm:"type:uuid;not null" json:"role_id"`
	CreatedAt time.Time `gorm:"type:timestamptz;default:now()" json:"created_at"`
}

// RolePermission 角色权限关联表
type RolePermission struct {
	RoleID       uuid.UUID `gorm:"type:uuid;not null" json:"role_id"`
	PermissionID uuid.UUID `gorm:"type:uuid;not null" json:"permission_id"`
	CreatedAt    time.Time `gorm:"type:timestamptz;default:now()" json:"created_at"`
}

// AuditLog 操作审计日志
type AuditLog struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AdminID    uuid.UUID `gorm:"type:uuid;not null;index" json:"admin_id"`
	AdminName  string    `gorm:"type:varchar(100)" json:"admin_name"`
	Action     string    `gorm:"type:varchar(100);not null" json:"action"`     // create_merchant, approve_kyc
	Resource   string    `gorm:"type:varchar(50);not null" json:"resource"`    // merchant, payment, order
	ResourceID string    `gorm:"type:varchar(100)" json:"resource_id"`         // 被操作资源的ID
	Method     string    `gorm:"type:varchar(10)" json:"method"`               // POST, PUT, DELETE
	Path       string    `gorm:"type:varchar(255)" json:"path"`                // API路径
	IP         string    `gorm:"type:varchar(50)" json:"ip"`
	UserAgent  string    `gorm:"type:text" json:"user_agent"`
	RequestBody  string  `gorm:"type:jsonb" json:"request_body,omitempty"`    // 请求参数
	ResponseCode int     `gorm:"type:integer" json:"response_code"`           // HTTP状态码
	Description  string  `gorm:"type:text" json:"description"`                // 操作描述
	CreatedAt    time.Time `gorm:"type:timestamptz;default:now();index" json:"created_at"`
}

// SystemConfig 系统配置表
type SystemConfig struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Key         string         `gorm:"type:varchar(100);unique;not null;index" json:"key"`
	Value       string         `gorm:"type:text;not null" json:"value"`
	Type        string         `gorm:"type:varchar(20);not null" json:"type"`        // string, number, boolean, json
	Category    string         `gorm:"type:varchar(50);not null;index" json:"category"` // payment, notification, risk
	Description string         `gorm:"type:text" json:"description"`
	IsPublic    bool           `gorm:"default:false" json:"is_public"` // 是否暴露给商户
	UpdatedBy   uuid.UUID      `gorm:"type:uuid" json:"updated_by"`
	CreatedAt   time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// MerchantReview 商户审核记录
type MerchantReview struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"merchant_id"`
	ReviewType   string     `gorm:"type:varchar(50);not null" json:"review_type"` // kyc, contract, qualification
	Status       string     `gorm:"type:varchar(20);not null" json:"status"`      // pending, approved, rejected
	ReviewerID   uuid.UUID  `gorm:"type:uuid" json:"reviewer_id"`                 // 审核人
	ReviewerName string     `gorm:"type:varchar(100)" json:"reviewer_name"`
	Reason       string     `gorm:"type:text" json:"reason"`                      // 拒绝原因
	Documents    string     `gorm:"type:jsonb" json:"documents"`                  // 审核材料
	ReviewedAt   *time.Time `gorm:"type:timestamptz" json:"reviewed_at"`
	CreatedAt    time.Time  `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"type:timestamptz;default:now()" json:"updated_at"`
}

// ApprovalFlow 审批流程
type ApprovalFlow struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	FlowType      string     `gorm:"type:varchar(50);not null" json:"flow_type"` // withdrawal, refund, large_payment
	ResourceID    string     `gorm:"type:varchar(100);not null;index" json:"resource_id"`
	ResourceType  string     `gorm:"type:varchar(50);not null" json:"resource_type"`
	ApplicantID   uuid.UUID  `gorm:"type:uuid;not null" json:"applicant_id"`     // 申请人
	ApplicantName string     `gorm:"type:varchar(100)" json:"applicant_name"`
	Status        string     `gorm:"type:varchar(20);not null" json:"status"`    // pending, approved, rejected
	CurrentStep   int        `gorm:"type:integer;default:1" json:"current_step"`
	TotalSteps    int        `gorm:"type:integer;not null" json:"total_steps"`
	ApproverID    uuid.UUID  `gorm:"type:uuid" json:"approver_id"`               // 当前审批人
	ApproverName  string     `gorm:"type:varchar(100)" json:"approver_name"`
	Reason        string     `gorm:"type:text" json:"reason"`
	Data          string     `gorm:"type:jsonb" json:"data"`                     // 审批相关数据
	ApprovedAt    *time.Time `gorm:"type:timestamptz" json:"approved_at"`
	CreatedAt     time.Time  `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"type:timestamptz;default:now()" json:"updated_at"`
}

// TableName 指定表名
func (Admin) TableName() string           { return "admins" }
func (Role) TableName() string            { return "roles" }
func (Permission) TableName() string      { return "permissions" }
func (AdminRole) TableName() string       { return "admin_roles" }
func (RolePermission) TableName() string  { return "role_permissions" }
func (AuditLog) TableName() string        { return "audit_logs" }
func (SystemConfig) TableName() string    { return "system_configs" }
func (MerchantReview) TableName() string  { return "merchant_reviews" }
func (ApprovalFlow) TableName() string    { return "approval_flows" }
