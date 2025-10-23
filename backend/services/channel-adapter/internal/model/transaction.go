package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Transaction 渠道交易记录表
type Transaction struct {
	ID                uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID        uuid.UUID      `gorm:"type:uuid;not null;index" json:"merchant_id"`           // 商户ID
	OrderNo           string         `gorm:"type:varchar(64);index" json:"order_no"`                // 订单号
	PaymentNo         string         `gorm:"type:varchar(64);unique;not null;index" json:"payment_no"` // 支付流水号
	Channel           string         `gorm:"type:varchar(50);not null;index" json:"channel"`        // 渠道
	ChannelTradeNo    string         `gorm:"type:varchar(200);index" json:"channel_trade_no"`       // 渠道交易号
	TransactionType   string         `gorm:"type:varchar(20);not null" json:"transaction_type"`     // 交易类型：payment, refund
	Amount            int64          `gorm:"type:bigint;not null" json:"amount"`                    // 金额（分）
	Currency          string         `gorm:"type:varchar(10);not null" json:"currency"`             // 货币
	Status            string         `gorm:"type:varchar(20);not null;index" json:"status"`         // 状态
	CustomerEmail     string         `gorm:"type:varchar(255)" json:"customer_email"`               // 客户邮箱
	CustomerName      string         `gorm:"type:varchar(100)" json:"customer_name"`                // 客户姓名
	PaymentMethod     string         `gorm:"type:varchar(50)" json:"payment_method"`                // 支付方式
	PaymentMethodDetails string      `gorm:"type:jsonb" json:"payment_method_details"`              // 支付方式详情
	FeeAmount         int64          `gorm:"type:bigint;default:0" json:"fee_amount"`               // 手续费（分）
	NetAmount         int64          `gorm:"type:bigint" json:"net_amount"`                         // 净额（分）
	ErrorCode         string         `gorm:"type:varchar(50)" json:"error_code"`                    // 错误码
	ErrorMessage      string         `gorm:"type:text" json:"error_message"`                        // 错误信息
	RequestData       string         `gorm:"type:jsonb" json:"request_data"`                        // 请求数据
	ResponseData      string         `gorm:"type:jsonb" json:"response_data"`                       // 响应数据
	WebhookData       string         `gorm:"type:jsonb" json:"webhook_data"`                        // Webhook数据
	Extra             string         `gorm:"type:jsonb" json:"extra"`                               // 扩展信息
	ProcessedAt       *time.Time     `gorm:"type:timestamptz" json:"processed_at"`                  // 处理时间
	CreatedAt         time.Time      `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Transaction) TableName() string {
	return "channel_transactions"
}

// WebhookLog Webhook 日志表
type WebhookLog struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID  uuid.UUID `gorm:"type:uuid;index" json:"merchant_id"`                    // 商户ID
	Channel     string    `gorm:"type:varchar(50);not null;index" json:"channel"`        // 渠道
	EventID     string    `gorm:"type:varchar(200);unique" json:"event_id"`              // 事件ID
	EventType   string    `gorm:"type:varchar(100)" json:"event_type"`                   // 事件类型
	PaymentNo   string    `gorm:"type:varchar(64);index" json:"payment_no"`              // 支付流水号
	Signature   string    `gorm:"type:text" json:"signature"`                            // 签名
	IsVerified  bool      `gorm:"default:false" json:"is_verified"`                      // 是否验证通过
	IsProcessed bool      `gorm:"default:false;index" json:"is_processed"`               // 是否已处理
	RequestBody string    `gorm:"type:jsonb" json:"request_body"`                        // 请求体
	RequestHeaders string `gorm:"type:jsonb" json:"request_headers"`                     // 请求头
	ProcessResult string  `gorm:"type:text" json:"process_result"`                       // 处理结果
	RetryCount  int       `gorm:"default:0" json:"retry_count"`                          // 重试次数
	CreatedAt   time.Time `gorm:"type:timestamptz;default:now();index" json:"created_at"`
	ProcessedAt *time.Time `gorm:"type:timestamptz" json:"processed_at"`                 // 处理时间
}

// TableName 指定表名
func (WebhookLog) TableName() string {
	return "webhook_logs"
}

// 交易类型常量
const (
	TransactionTypePayment = "payment" // 支付
	TransactionTypeRefund  = "refund"  // 退款
)

// 交易状态常量
const (
	TransactionStatusPending    = "pending"    // 待处理
	TransactionStatusProcessing = "processing" // 处理中
	TransactionStatusSuccess    = "success"    // 成功
	TransactionStatusFailed     = "failed"     // 失败
	TransactionStatusCancelled  = "cancelled"  // 已取消
	TransactionStatusRefunded   = "refunded"   // 已退款
)
