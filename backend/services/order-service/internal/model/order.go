package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Order 订单表
type Order struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID      uuid.UUID      `gorm:"type:uuid;not null;index;index:idx_merchant_status_created,priority:1" json:"merchant_id"`  // 商户ID
	OrderNo         string         `gorm:"type:varchar(64);unique;not null;index" json:"order_no"`   // 订单号
	PaymentNo       string         `gorm:"type:varchar(64);index" json:"payment_no"`                 // 支付流水号
	TotalAmount     int64          `gorm:"type:bigint;not null" json:"total_amount"`                 // 订单总金额（分）
	PayAmount       int64          `gorm:"type:bigint;not null" json:"pay_amount"`                   // 实付金额（分）
	DiscountAmount  int64          `gorm:"type:bigint;default:0" json:"discount_amount"`             // 优惠金额（分）
	ShippingFee     int64          `gorm:"type:bigint;default:0" json:"shipping_fee"`                // 运费（分）
	Currency        string         `gorm:"type:varchar(10);not null" json:"currency"`                // 货币类型
	Status          string         `gorm:"type:varchar(20);not null;index;index:idx_merchant_status_created,priority:2" json:"status"`  // 订单状态
	PayStatus       string         `gorm:"type:varchar(20);not null;index" json:"pay_status"`        // 支付状态
	ShippingStatus  string         `gorm:"type:varchar(20);default:'pending'" json:"shipping_status"` // 配送状态
	CustomerID      uuid.UUID      `gorm:"type:uuid;index" json:"customer_id"`                       // 客户ID
	CustomerEmail   string         `gorm:"type:varchar(255);index" json:"customer_email"`            // 客户邮箱
	CustomerName    string         `gorm:"type:varchar(100)" json:"customer_name"`                   // 客户姓名
	CustomerPhone   string         `gorm:"type:varchar(20)" json:"customer_phone"`                   // 客户手机
	CustomerIP      string         `gorm:"type:varchar(50)" json:"customer_ip"`                      // 客户IP
	ShippingMethod  string         `gorm:"type:varchar(50)" json:"shipping_method"`                  // 配送方式
	ShippingAddress string         `gorm:"type:jsonb" json:"shipping_address"`                       // 配送地址（JSON）
	BillingAddress  string         `gorm:"type:jsonb" json:"billing_address"`                        // 账单地址（JSON）
	Remark          string         `gorm:"type:text" json:"remark"`                                  // 备注
	Extra           string         `gorm:"type:jsonb" json:"extra"`                                  // 扩展信息（JSON）
	Language        string         `gorm:"type:varchar(10);default:'en'" json:"language"`            // 语言
	PaidAt          *time.Time     `gorm:"type:timestamptz" json:"paid_at"`                          // 支付时间
	ShippedAt       *time.Time     `gorm:"type:timestamptz" json:"shipped_at"`                       // 发货时间
	CompletedAt     *time.Time     `gorm:"type:timestamptz" json:"completed_at"`                     // 完成时间
	CancelledAt     *time.Time     `gorm:"type:timestamptz" json:"cancelled_at"`                     // 取消时间
	ExpiredAt       *time.Time     `gorm:"type:timestamptz" json:"expired_at"`                       // 过期时间
	CreatedAt       time.Time      `gorm:"type:timestamptz;default:now();index:idx_merchant_status_created,priority:3,sort:desc" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"type:timestamptz;default:now()" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	Items []*OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
}

// TableName 指定表名
func (Order) TableName() string {
	return "orders"
}

// OrderItem 订单项表
type OrderItem struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OrderID       uuid.UUID `gorm:"type:uuid;not null;index" json:"order_id"`          // 订单ID
	ProductID     string    `gorm:"type:varchar(64)" json:"product_id"`                // 商品ID
	ProductName   string    `gorm:"type:varchar(200);not null" json:"product_name"`    // 商品名称
	ProductSKU    string    `gorm:"type:varchar(100)" json:"product_sku"`              // 商品SKU
	ProductImage  string    `gorm:"type:varchar(500)" json:"product_image"`            // 商品图片
	UnitPrice     int64     `gorm:"type:bigint;not null" json:"unit_price"`            // 单价（分）
	Quantity      int       `gorm:"type:integer;not null" json:"quantity"`             // 数量
	TotalPrice    int64     `gorm:"type:bigint;not null" json:"total_price"`           // 小计（分）
	DiscountPrice int64     `gorm:"type:bigint;default:0" json:"discount_price"`       // 优惠金额（分）
	Attributes    string    `gorm:"type:jsonb" json:"attributes"`                      // 商品属性（JSON）
	Extra         string    `gorm:"type:jsonb" json:"extra"`                           // 扩展信息（JSON）
	CreatedAt     time.Time `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt     time.Time `gorm:"type:timestamptz;default:now()" json:"updated_at"`
}

// TableName 指定表名
func (OrderItem) TableName() string {
	return "order_items"
}

// OrderLog 订单日志表
type OrderLog struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OrderID    uuid.UUID `gorm:"type:uuid;not null;index" json:"order_id"`       // 订单ID
	Action     string    `gorm:"type:varchar(50);not null" json:"action"`        // 操作类型
	OldStatus  string    `gorm:"type:varchar(20)" json:"old_status"`             // 旧状态
	NewStatus  string    `gorm:"type:varchar(20)" json:"new_status"`             // 新状态
	OperatorID uuid.UUID `gorm:"type:uuid" json:"operator_id"`                   // 操作人ID
	OperatorType string  `gorm:"type:varchar(20)" json:"operator_type"`          // 操作人类型
	Remark     string    `gorm:"type:text" json:"remark"`                        // 备注
	Extra      string    `gorm:"type:jsonb" json:"extra"`                        // 扩展信息
	CreatedAt  time.Time `gorm:"type:timestamptz;default:now()" json:"created_at"`
}

// TableName 指定表名
func (OrderLog) TableName() string {
	return "order_logs"
}

// OrderStatistics 订单统计表（按天统计）
type OrderStatistics struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MerchantID     uuid.UUID `gorm:"type:uuid;not null;index:idx_merchant_date" json:"merchant_id"` // 商户ID
	StatDate       time.Time `gorm:"type:date;not null;index:idx_merchant_date" json:"stat_date"`   // 统计日期
	Currency       string    `gorm:"type:varchar(10);not null" json:"currency"`                     // 货币类型
	TotalOrders    int       `gorm:"type:integer;default:0" json:"total_orders"`                    // 订单总数
	PaidOrders     int       `gorm:"type:integer;default:0" json:"paid_orders"`                     // 已支付订单数
	CancelledOrders int      `gorm:"type:integer;default:0" json:"cancelled_orders"`                // 已取消订单数
	TotalAmount    int64     `gorm:"type:bigint;default:0" json:"total_amount"`                     // 订单总金额（分）
	PaidAmount     int64     `gorm:"type:bigint;default:0" json:"paid_amount"`                      // 已支付金额（分）
	RefundAmount   int64     `gorm:"type:bigint;default:0" json:"refund_amount"`                    // 退款金额（分）
	AvgOrderAmount int64     `gorm:"type:bigint;default:0" json:"avg_order_amount"`                 // 平均订单金额（分）
	CreatedAt      time.Time `gorm:"type:timestamptz;default:now()" json:"created_at"`
	UpdatedAt      time.Time `gorm:"type:timestamptz;default:now()" json:"updated_at"`
}

// TableName 指定表名
func (OrderStatistics) TableName() string {
	return "order_statistics"
}

// 订单状态常量
const (
	OrderStatusPending   = "pending"   // 待支付
	OrderStatusPaid      = "paid"      // 已支付
	OrderStatusProcessing = "processing" // 处理中
	OrderStatusShipped   = "shipped"   // 已发货
	OrderStatusCompleted = "completed" // 已完成
	OrderStatusCancelled = "cancelled" // 已取消
	OrderStatusRefunded  = "refunded"  // 已退款
	OrderStatusExpired   = "expired"   // 已过期
)

// 支付状态常量
const (
	PayStatusPending = "pending" // 待支付
	PayStatusPaid    = "paid"    // 已支付
	PayStatusFailed  = "failed"  // 支付失败
	PayStatusRefunded = "refunded" // 已退款
	PayStatusPartialRefunded = "partial_refunded" // 部分退款
)

// 配送状态常量
const (
	ShippingStatusPending   = "pending"   // 待发货
	ShippingStatusPreparing = "preparing" // 备货中
	ShippingStatusShipped   = "shipped"   // 已发货
	ShippingStatusInTransit = "in_transit" // 运输中
	ShippingStatusDelivered = "delivered" // 已送达
	ShippingStatusReturned  = "returned"  // 已退货
)

// 订单操作类型常量
const (
	OrderActionCreate     = "create"      // 创建
	OrderActionPay        = "pay"         // 支付
	OrderActionCancel     = "cancel"      // 取消
	OrderActionShip       = "ship"        // 发货
	OrderActionComplete   = "complete"    // 完成
	OrderActionRefund     = "refund"      // 退款
	OrderActionUpdateStatus = "update_status" // 更新状态
)

// 配送方式常量
const (
	ShippingMethodStandard = "standard" // 标准配送
	ShippingMethodExpress  = "express"  // 快递
	ShippingMethodPickup   = "pickup"   // 自提
	ShippingMethodDigital  = "digital"  // 数字商品（无需配送）
)

// 操作人类型常量
const (
	OperatorTypeSystem   = "system"   // 系统
	OperatorTypeMerchant = "merchant" // 商户
	OperatorTypeAdmin    = "admin"    // 管理员
	OperatorTypeCustomer = "customer" // 客户
)

// Address 地址结构（用于JSON存储）
type Address struct {
	Country    string `json:"country"`     // 国家
	Province   string `json:"province"`    // 省/州
	City       string `json:"city"`        // 城市
	District   string `json:"district"`    // 区/县
	Street     string `json:"street"`      // 街道地址
	PostalCode string `json:"postal_code"` // 邮编
	Phone      string `json:"phone"`       // 联系电话
	Name       string `json:"name"`        // 收件人姓名
}
