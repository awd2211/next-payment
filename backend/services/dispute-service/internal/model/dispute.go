package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Dispute 拒付/争议表
type Dispute struct {
	ID               uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	DisputeNo        string    `gorm:"type:varchar(64);unique;not null;index" json:"dispute_no"`
	Channel          string    `gorm:"type:varchar(50);not null;index" json:"channel"`
	ChannelDisputeID string    `gorm:"type:varchar(128);unique;index" json:"channel_dispute_id"`

	// 关联订单/支付信息
	PaymentNo      string     `gorm:"type:varchar(64);index" json:"payment_no"`
	OrderNo        string     `gorm:"type:varchar(64);index" json:"order_no"`
	MerchantID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"merchant_id"`
	ChannelTradeNo string     `gorm:"type:varchar(128)" json:"channel_trade_no"`

	// 拒付金额信息
	Amount   int64  `gorm:"type:bigint;not null" json:"amount"`
	Currency string `gorm:"type:varchar(10);not null" json:"currency"`

	// 拒付原因和类型
	Reason     string `gorm:"type:varchar(100)" json:"reason"`           // fraudulent, product_not_received, etc.
	ReasonCode string `gorm:"type:varchar(50)" json:"reason_code"`
	Status     string `gorm:"type:varchar(30);not null;index" json:"status"` // warning_needs_response, needs_response, under_review, won, lost, charge_refunded

	// 证据和响应
	EvidenceDueBy      *time.Time `gorm:"type:timestamptz" json:"evidence_due_by,omitempty"`
	EvidenceSubmitted  bool       `gorm:"default:false" json:"evidence_submitted"`
	EvidenceSubmitTime *time.Time `gorm:"type:timestamptz" json:"evidence_submit_time,omitempty"`

	// 处理人员
	AssignedTo *uuid.UUID `gorm:"type:uuid;index" json:"assigned_to,omitempty"`
	AssignedAt *time.Time `gorm:"type:timestamptz" json:"assigned_at,omitempty"`

	// 结果信息
	Result       string     `gorm:"type:varchar(20)" json:"result,omitempty"`       // won, lost
	ResolvedAt   *time.Time `gorm:"type:timestamptz" json:"resolved_at,omitempty"`
	IsRefunded   bool       `gorm:"default:false" json:"is_refunded"`
	RefundAmount int64      `gorm:"type:bigint;default:0" json:"refund_amount"`

	// 扩展信息
	Extra string `gorm:"type:jsonb" json:"extra,omitempty"`

	// 时间戳
	CreatedAt time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Dispute) TableName() string {
	return "disputes"
}

// 拒付状态常量
const (
	DisputeStatusWarningNeedsResponse = "warning_needs_response" // 警告需要响应
	DisputeStatusNeedsResponse        = "needs_response"         // 需要响应
	DisputeStatusUnderReview          = "under_review"           // 审核中
	DisputeStatusWon                  = "won"                    // 胜诉
	DisputeStatusLost                 = "lost"                   // 败诉
	DisputeStatusChargeRefunded       = "charge_refunded"        // 已退款
)

// 拒付原因常量
const (
	DisputeReasonFraudulent            = "fraudulent"              // 欺诈
	DisputeReasonProductNotReceived    = "product_not_received"    // 未收到商品
	DisputeReasonProductUnacceptable   = "product_unacceptable"    // 商品不符
	DisputeReasonDuplicate             = "duplicate"               // 重复扣款
	DisputeReasonCreditNotProcessed    = "credit_not_processed"    // 退款未处理
	DisputeReasonSubscriptionCanceled  = "subscription_canceled"   // 订阅已取消
	DisputeReasonGeneralServiceFailure = "general_service_failure" // 服务失败
)

// 拒付结果常量
const (
	DisputeResultWon  = "won"  // 商户胜诉
	DisputeResultLost = "lost" // 商户败诉
)

// DisputeEvidence 拒付证据表
type DisputeEvidence struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	DisputeID  uuid.UUID `gorm:"type:uuid;not null;index" json:"dispute_id"`
	DisputeNo  string    `gorm:"type:varchar(64);not null;index" json:"dispute_no"`

	// 证据类型
	EvidenceType string `gorm:"type:varchar(50);not null" json:"evidence_type"` // receipt, shipping_proof, communication, etc.

	// 证据内容
	Title       string `gorm:"type:varchar(200)" json:"title"`
	Description string `gorm:"type:text" json:"description"`
	FileURL     string `gorm:"type:varchar(500)" json:"file_url,omitempty"`
	FileName    string `gorm:"type:varchar(255)" json:"file_name,omitempty"`
	FileSize    int64  `gorm:"type:bigint" json:"file_size,omitempty"`
	FileHash    string `gorm:"type:varchar(64)" json:"file_hash,omitempty"`

	// 上传信息
	UploadedBy uuid.UUID  `gorm:"type:uuid;not null" json:"uploaded_by"`
	UploadedAt time.Time  `gorm:"type:timestamptz;default:now()" json:"uploaded_at"`

	// 提交状态
	IsSubmitted    bool       `gorm:"default:false" json:"is_submitted"`
	SubmittedAt    *time.Time `gorm:"type:timestamptz" json:"submitted_at,omitempty"`
	SubmitResponse string     `gorm:"type:text" json:"submit_response,omitempty"`

	// 时间戳
	CreatedAt time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (DisputeEvidence) TableName() string {
	return "dispute_evidence"
}

// 证据类型常量
const (
	EvidenceTypeReceipt             = "receipt"               // 收据
	EvidenceTypeShippingProof       = "shipping_proof"        // 物流证明
	EvidenceTypeCommunication       = "communication"         // 沟通记录
	EvidenceTypeRefundPolicy        = "refund_policy"         // 退款政策
	EvidenceTypeCancellationPolicy  = "cancellation_policy"   // 取消政策
	EvidenceTypeCustomerSignature   = "customer_signature"    // 客户签名
	EvidenceTypeServiceDocumentation = "service_documentation" // 服务文档
	EvidenceTypeOther               = "other"                 // 其他
)

// DisputeTimeline 拒付时间线表
type DisputeTimeline struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	DisputeID uuid.UUID `gorm:"type:uuid;not null;index" json:"dispute_id"`
	DisputeNo string    `gorm:"type:varchar(64);not null;index" json:"dispute_no"`

	// 事件信息
	EventType   string `gorm:"type:varchar(50);not null;index" json:"event_type"` // created, updated, evidence_uploaded, submitted, won, lost
	EventStatus string `gorm:"type:varchar(30)" json:"event_status"`
	Description string `gorm:"type:text" json:"description"`

	// 操作人员
	OperatorID   *uuid.UUID `gorm:"type:uuid" json:"operator_id,omitempty"`
	OperatorType string     `gorm:"type:varchar(20)" json:"operator_type,omitempty"` // admin, merchant, system, stripe

	// 扩展信息
	Metadata string `gorm:"type:jsonb" json:"metadata,omitempty"`

	// 时间戳
	CreatedAt time.Time      `gorm:"type:timestamptz;default:now();index" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (DisputeTimeline) TableName() string {
	return "dispute_timeline"
}

// 时间线事件类型常量
const (
	TimelineEventCreated          = "created"           // 拒付创建
	TimelineEventUpdated          = "updated"           // 状态更新
	TimelineEventAssigned         = "assigned"          // 分配处理人
	TimelineEventEvidenceUploaded = "evidence_uploaded" // 证据上传
	TimelineEventEvidenceSubmitted = "evidence_submitted" // 证据提交
	TimelineEventWon              = "won"               // 胜诉
	TimelineEventLost             = "lost"              // 败诉
	TimelineEventRefunded         = "refunded"          // 已退款
)

// 操作人员类型常量
const (
	OperatorTypeAdmin    = "admin"    // 管理员
	OperatorTypeMerchant = "merchant" // 商户
	OperatorTypeSystem   = "system"   // 系统
	OperatorTypeStripe   = "stripe"   // Stripe
)
