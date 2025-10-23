package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Payment 支付记录表
type Payment struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"merchant_id"`             // 商户ID
	OrderNo         string         `gorm:"type:varchar(64);unique;not null;index" json:"order_no"`  // 订单号（商户侧）
	PaymentNo       string         `gorm:"type:varchar(64);unique;not null;index" json:"payment_no"` // 支付流水号（平台侧）
	Channel         string         `gorm:"type:varchar(50);not null;index" json:"channel"`          // 支付渠道：stripe, paypal, crypto
	ChannelOrderNo  string         `gorm:"type:varchar(128);index" json:"channel_order_no"`         // 渠道订单号
	Amount          int64          `gorm:"type:bigint;not null" json:"amount"`                      // 支付金额（分）
	Currency        string         `gorm:"type:varchar(10);not null" json:"currency"`               // 货币类型：USD, EUR, CNY等
	Status          string         `gorm:"type:varchar(20);not null;index" json:"status"`           // 状态：pending, processing, success, failed, cancelled, expired
	PayMethod       string         `gorm:"type:varchar(50)" json:"pay_method"`                      // 支付方式：card, bank_transfer, wallet, crypto
	CustomerEmail   string         `gorm:"type:varchar(255)" json:"customer_email"`                 // 客户邮箱
	CustomerName    string         `gorm:"type:varchar(100)" json:"customer_name"`                  // 客户姓名
	CustomerPhone   string         `gorm:"type:varchar(20)" json:"customer_phone"`                  // 客户手机
	CustomerIP      string         `gorm:"type:varchar(50)" json:"customer_ip"`                     // 客户IP
	Description     string         `gorm:"type:text" json:"description"`                            // 商品描述
	NotifyURL       string         `gorm:"type:varchar(500)" json:"notify_url"`                     // 异步通知URL
	ReturnURL       string         `gorm:"type:varchar(500)" json:"return_url"`                     // 同步跳转URL
	Extra           string         `gorm:"type:jsonb" json:"extra"`                                 // 扩展信息（JSON）
	ErrorCode       string         `gorm:"type:varchar(50)" json:"error_code"`                      // 错误码
	ErrorMsg        string         `gorm:"type:text" json:"error_msg"`                              // 错误信息
	NotifyStatus    string         `gorm:"type:varchar(20);default:'pending'" json:"notify_status"` // 通知状态：pending, notified, failed
	NotifyTimes     int            `gorm:"type:integer;default:0" json:"notify_times"`              // 通知次数
	LastNotifyAt    *time.Time     `gorm:"type:timestamptz" json:"last_notify_at"`                  // 最后通知时间
	PaidAt          *time.Time     `gorm:"type:timestamptz" json:"paid_at"`                         // 支付完成时间
	ExpiredAt       *time.Time     `gorm:"type:timestamptz" json:"expired_at"`                      // 过期时间
	CreatedAt       time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Payment) TableName() string {
	return "payments"
}

// Refund 退款记录表
type Refund struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PaymentID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"payment_id"`           // 关联的支付ID
	MerchantID     uuid.UUID      `gorm:"type:uuid;not null;index" json:"merchant_id"`          // 商户ID
	RefundNo       string         `gorm:"type:varchar(64);unique;not null;index" json:"refund_no"` // 退款单号（平台）
	ChannelRefundNo string        `gorm:"type:varchar(128);index" json:"channel_refund_no"`     // 渠道退款单号
	Amount         int64          `gorm:"type:bigint;not null" json:"amount"`                   // 退款金额（分）
	Currency       string         `gorm:"type:varchar(10);not null" json:"currency"`            // 货币类型
	Status         string         `gorm:"type:varchar(20);not null;index" json:"status"`        // 状态：pending, processing, success, failed
	Reason         string         `gorm:"type:varchar(200)" json:"reason"`                      // 退款原因
	Description    string         `gorm:"type:text" json:"description"`                         // 退款说明
	OperatorID     uuid.UUID      `gorm:"type:uuid" json:"operator_id"`                         // 操作人ID
	OperatorType   string         `gorm:"type:varchar(20)" json:"operator_type"`                // 操作人类型：merchant, admin, system
	ErrorCode      string         `gorm:"type:varchar(50)" json:"error_code"`                   // 错误码
	ErrorMsg       string         `gorm:"type:text" json:"error_msg"`                           // 错误信息
	RefundedAt     *time.Time     `gorm:"type:timestamptz" json:"refunded_at"`                  // 退款完成时间
	CreatedAt      time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	Payment *Payment `gorm:"foreignKey:PaymentID" json:"payment,omitempty"`
}

// TableName 指定表名
func (Refund) TableName() string {
	return "refunds"
}

// PaymentCallback 支付回调记录表
type PaymentCallback struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PaymentID  uuid.UUID `gorm:"type:uuid;not null;index" json:"payment_id"`       // 支付ID
	Channel    string    `gorm:"type:varchar(50);not null" json:"channel"`         // 支付渠道
	Event      string    `gorm:"type:varchar(50);not null" json:"event"`           // 事件类型
	RawData    string    `gorm:"type:text;not null" json:"raw_data"`               // 原始回调数据
	Signature  string    `gorm:"type:varchar(500)" json:"signature"`               // 签名
	IsVerified bool      `gorm:"default:false" json:"is_verified"`                 // 是否验证通过
	IsProcessed bool     `gorm:"default:false" json:"is_processed"`                // 是否已处理
	ErrorMsg   string    `gorm:"type:text" json:"error_msg"`                       // 错误信息
	CreatedAt  time.Time `gorm:"type:timestamptz;default:now()" json:"created_at"`
}

// TableName 指定表名
func (PaymentCallback) TableName() string {
	return "payment_callbacks"
}

// PaymentRoute 支付路由规则表
type PaymentRoute struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`              // 规则名称
	Priority    int            `gorm:"type:integer;not null;default:0" json:"priority"`     // 优先级（越大越优先）
	Channel     string         `gorm:"type:varchar(50);not null" json:"channel"`            // 目标渠道
	Conditions  string         `gorm:"type:jsonb;not null" json:"conditions"`               // 路由条件（JSON）
	IsEnabled   bool           `gorm:"default:true" json:"is_enabled"`                      // 是否启用
	Description string         `gorm:"type:text" json:"description"`                        // 规则描述
	CreatedAt   time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (PaymentRoute) TableName() string {
	return "payment_routes"
}

// 支付状态常量
const (
	PaymentStatusPending    = "pending"    // 待支付
	PaymentStatusProcessing = "processing" // 处理中
	PaymentStatusSuccess    = "success"    // 支付成功
	PaymentStatusFailed     = "failed"     // 支付失败
	PaymentStatusCancelled  = "cancelled"  // 已取消
	PaymentStatusExpired    = "expired"    // 已过期
)

// 退款状态常量
const (
	RefundStatusPending    = "pending"    // 待退款
	RefundStatusProcessing = "processing" // 处理中
	RefundStatusSuccess    = "success"    // 退款成功
	RefundStatusFailed     = "failed"     // 退款失败
)

// 通知状态常量
const (
	NotifyStatusPending  = "pending"  // 待通知
	NotifyStatusNotified = "notified" // 已通知
	NotifyStatusFailed   = "failed"   // 通知失败
)

// 支付渠道常量
const (
	ChannelStripe  = "stripe"  // Stripe
	ChannelPayPal  = "paypal"  // PayPal
	ChannelCrypto  = "crypto"  // 加密货币
	ChannelAdyen   = "adyen"   // Adyen
	ChannelSquare  = "square"  // Square
)

// 支付方式常量
const (
	PayMethodCard         = "card"          // 信用卡/借记卡
	PayMethodBankTransfer = "bank_transfer" // 银行转账
	PayMethodWallet       = "wallet"        // 电子钱包
	PayMethodCrypto       = "crypto"        // 加密货币
)

// 操作人类型常量
const (
	OperatorTypeMerchant = "merchant" // 商户
	OperatorTypeAdmin    = "admin"    // 管理员
	OperatorTypeSystem   = "system"   // 系统
)
