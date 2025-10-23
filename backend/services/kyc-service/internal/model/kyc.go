package model

import (
	"time"

	"github.com/google/uuid"
)

// KYCStatus KYC状态
type KYCStatus string

const (
	KYCStatusPending   KYCStatus = "pending"    // 待审核
	KYCStatusApproved  KYCStatus = "approved"   // 已通过
	KYCStatusRejected  KYCStatus = "rejected"   // 已拒绝
	KYCStatusExpired   KYCStatus = "expired"    // 已过期
	KYCStatusSuspended KYCStatus = "suspended"  // 已暂停
)

// DocumentType 文档类型
type DocumentType string

const (
	DocumentTypeIDCard         DocumentType = "id_card"          // 身份证
	DocumentTypePassport       DocumentType = "passport"         // 护照
	DocumentTypeDriverLicense  DocumentType = "driver_license"   // 驾照
	DocumentTypeBusinessLicense DocumentType = "business_license" // 营业执照
	DocumentTypeBankStatement  DocumentType = "bank_statement"   // 银行对账单
	DocumentTypeTaxCert        DocumentType = "tax_cert"         // 税务证明
	DocumentTypeOther          DocumentType = "other"            // 其他
)

// KYCLevel KYC级别
type KYCLevel string

const (
	KYCLevelBasic      KYCLevel = "basic"       // 基础级别 (个人信息)
	KYCLevelIntermediate KYCLevel = "intermediate" // 中级 (身份证明)
	KYCLevelAdvanced   KYCLevel = "advanced"    // 高级 (地址证明、银行验证)
	KYCLevelEnterprise KYCLevel = "enterprise"  // 企业级 (营业执照、法人信息)
)

// KYCDocument KYC文档
type KYCDocument struct {
	ID              uuid.UUID    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID      uuid.UUID    `gorm:"type:uuid;index;not null" json:"merchant_id"`
	DocumentType    DocumentType `gorm:"type:varchar(50);not null" json:"document_type"`
	DocumentNumber  string       `gorm:"type:varchar(100)" json:"document_number"`        // 文档编号
	DocumentURL     string       `gorm:"type:varchar(500);not null" json:"document_url"`  // 文档URL
	FrontImageURL   string       `gorm:"type:varchar(500)" json:"front_image_url"`        // 正面图片
	BackImageURL    string       `gorm:"type:varchar(500)" json:"back_image_url"`         // 背面图片
	IssueDate       *time.Time   `json:"issue_date"`                                      // 签发日期
	ExpiryDate      *time.Time   `json:"expiry_date"`                                     // 过期日期
	IssuingCountry  string       `gorm:"type:varchar(10)" json:"issuing_country"`         // 签发国家
	Status          KYCStatus    `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	ReviewerID      *uuid.UUID   `gorm:"type:uuid" json:"reviewer_id"`
	ReviewerName    string       `gorm:"type:varchar(100)" json:"reviewer_name"`
	ReviewComments  string       `gorm:"type:text" json:"review_comments"`
	ReviewedAt      *time.Time   `json:"reviewed_at"`
	RejectionReason string       `gorm:"type:text" json:"rejection_reason"`
	OCRData         string       `gorm:"type:jsonb" json:"ocr_data"` // OCR识别数据
	CreatedAt       time.Time    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time    `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (KYCDocument) TableName() string {
	return "kyc_documents"
}

// BusinessQualification 企业资质
type BusinessQualification struct {
	ID                  uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID          uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"merchant_id"`
	CompanyName         string    `gorm:"type:varchar(200);not null" json:"company_name"`
	BusinessLicenseNo   string    `gorm:"type:varchar(100);not null" json:"business_license_no"`
	BusinessLicenseURL  string    `gorm:"type:varchar(500);not null" json:"business_license_url"`
	LegalPersonName     string    `gorm:"type:varchar(100);not null" json:"legal_person_name"`
	LegalPersonIDCard   string    `gorm:"type:varchar(50);not null" json:"legal_person_id_card"`
	LegalPersonIDCardFrontURL string `gorm:"type:varchar(500)" json:"legal_person_id_card_front_url"`
	LegalPersonIDCardBackURL  string `gorm:"type:varchar(500)" json:"legal_person_id_card_back_url"`
	RegisteredAddress   string    `gorm:"type:text" json:"registered_address"`
	RegisteredCapital   int64     `gorm:"not null;default:0" json:"registered_capital"` // 注册资本（分）
	EstablishedDate     *time.Time `json:"established_date"`
	BusinessScope       string    `gorm:"type:text" json:"business_scope"`
	Industry            string    `gorm:"type:varchar(100)" json:"industry"`
	TaxRegistrationNo   string    `gorm:"type:varchar(100)" json:"tax_registration_no"`
	TaxRegistrationURL  string    `gorm:"type:varchar(500)" json:"tax_registration_url"`
	OrganizationCode    string    `gorm:"type:varchar(50)" json:"organization_code"`
	Status              KYCStatus `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	ReviewerID          *uuid.UUID `gorm:"type:uuid" json:"reviewer_id"`
	ReviewerName        string    `gorm:"type:varchar(100)" json:"reviewer_name"`
	ReviewComments      string    `gorm:"type:text" json:"review_comments"`
	ReviewedAt          *time.Time `json:"reviewed_at"`
	RejectionReason     string    `gorm:"type:text" json:"rejection_reason"`
	CreatedAt           time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (BusinessQualification) TableName() string {
	return "business_qualifications"
}

// KYCReview KYC审核记录
type KYCReview struct {
	ID              uuid.UUID    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID      uuid.UUID    `gorm:"type:uuid;index;not null" json:"merchant_id"`
	DocumentID      *uuid.UUID   `gorm:"type:uuid;index" json:"document_id"`
	QualificationID *uuid.UUID   `gorm:"type:uuid;index" json:"qualification_id"`
	ReviewerID      uuid.UUID    `gorm:"type:uuid;not null" json:"reviewer_id"`
	ReviewerName    string       `gorm:"type:varchar(100);not null" json:"reviewer_name"`
	Action          string       `gorm:"type:varchar(20);not null" json:"action"` // approve, reject, request_resubmit
	Status          KYCStatus    `gorm:"type:varchar(20);not null" json:"status"`
	Comments        string       `gorm:"type:text" json:"comments"`
	RejectionReason string       `gorm:"type:text" json:"rejection_reason"`
	ReviewedAt      time.Time    `gorm:"not null" json:"reviewed_at"`
	CreatedAt       time.Time    `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (KYCReview) TableName() string {
	return "kyc_reviews"
}

// MerchantKYCLevel 商户KYC级别
type MerchantKYCLevel struct {
	ID                uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID        uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"merchant_id"`
	CurrentLevel      KYCLevel  `gorm:"type:varchar(20);not null;default:'basic'" json:"current_level"`
	ApprovedLevel     KYCLevel  `gorm:"type:varchar(20)" json:"approved_level"`
	HasBasic          bool      `gorm:"not null;default:false" json:"has_basic"`
	HasIntermediate   bool      `gorm:"not null;default:false" json:"has_intermediate"`
	HasAdvanced       bool      `gorm:"not null;default:false" json:"has_advanced"`
	HasEnterprise     bool      `gorm:"not null;default:false" json:"has_enterprise"`
	BasicApprovedAt   *time.Time `json:"basic_approved_at"`
	IntermediateApprovedAt *time.Time `json:"intermediate_approved_at"`
	AdvancedApprovedAt     *time.Time `json:"advanced_approved_at"`
	EnterpriseApprovedAt   *time.Time `json:"enterprise_approved_at"`
	TransactionLimit  int64     `gorm:"not null;default:0" json:"transaction_limit"`       // 交易限额（分）
	DailyLimit        int64     `gorm:"not null;default:0" json:"daily_limit"`             // 日限额（分）
	MonthlyLimit      int64     `gorm:"not null;default:0" json:"monthly_limit"`           // 月限额（分）
	NextReviewDate    *time.Time `json:"next_review_date"`                                 // 下次审核日期
	CreatedAt         time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (MerchantKYCLevel) TableName() string {
	return "merchant_kyc_levels"
}

// KYCAlert KYC预警
type KYCAlert struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID  uuid.UUID `gorm:"type:uuid;index;not null" json:"merchant_id"`
	AlertType   string    `gorm:"type:varchar(50);not null" json:"alert_type"` // expiry, suspicious, limit_exceeded
	Severity    string    `gorm:"type:varchar(20);not null" json:"severity"`   // low, medium, high, critical
	Title       string    `gorm:"type:varchar(200);not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	Status      string    `gorm:"type:varchar(20);not null;default:'open'" json:"status"` // open, acknowledged, resolved
	ResolvedBy  *uuid.UUID `gorm:"type:uuid" json:"resolved_by"`
	ResolvedAt  *time.Time `json:"resolved_at"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (KYCAlert) TableName() string {
	return "kyc_alerts"
}
