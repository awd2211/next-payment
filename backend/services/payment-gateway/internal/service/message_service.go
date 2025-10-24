package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

// MessageService 消息服务接口
type MessageService interface {
	// 发送补偿消息
	SendCompensationMessage(ctx context.Context, msg *CompensationMessage) error
	// 发送通知消息
	SendNotificationMessage(ctx context.Context, msg *NotificationMessage) error
	// 发送退款通知消息
	SendRefundNotificationMessage(ctx context.Context, msg *RefundNotificationMessage) error
}

type messageService struct {
	brokers []string
}

// NewMessageService 创建消息服务实例
// brokers: Kafka broker地址列表，如果为nil则使用降级模式（打印日志）
func NewMessageService(brokers []string) MessageService {
	return &messageService{
		brokers: brokers,
	}
}

// CompensationType 补偿类型
type CompensationType string

const (
	CompensationTypeCancelOrder        CompensationType = "cancel_order"         // 取消订单
	CompensationTypeUpdateRefundStatus CompensationType = "update_refund_status" // 更新退款状态
	CompensationTypeRetryNotification  CompensationType = "retry_notification"   // 重试通知
)

// CompensationMessage 补偿消息
type CompensationMessage struct {
	MessageID     string                 `json:"message_id"`      // 消息ID
	Type          CompensationType       `json:"type"`            // 补偿类型
	PaymentNo     string                 `json:"payment_no"`      // 支付流水号
	RefundNo      string                 `json:"refund_no"`       // 退款流水号（可选）
	OrderNo       string                 `json:"order_no"`        // 订单号（可选）
	MerchantID    string                 `json:"merchant_id"`     // 商户ID
	Reason        string                 `json:"reason"`          // 补偿原因
	RetryCount    int                    `json:"retry_count"`     // 重试次数
	MaxRetries    int                    `json:"max_retries"`     // 最大重试次数
	NextRetryTime time.Time              `json:"next_retry_time"` // 下次重试时间
	Extra         map[string]interface{} `json:"extra"`           // 扩展信息
	CreatedAt     time.Time              `json:"created_at"`      // 创建时间
}

// NotificationMessage 通知消息（支付结果通知）
type NotificationMessage struct {
	MessageID       string                 `json:"message_id"`        // 消息ID
	PaymentNo       string                 `json:"payment_no"`        // 支付流水号
	MerchantID      string                 `json:"merchant_id"`       // 商户ID
	NotifyURL       string                 `json:"notify_url"`        // 通知URL
	NotifyData      map[string]interface{} `json:"notify_data"`       // 通知数据
	Signature       string                 `json:"signature"`         // 签名
	RetryCount      int                    `json:"retry_count"`       // 重试次数
	MaxRetries      int                    `json:"max_retries"`       // 最大重试次数（默认5次）
	NextRetryTime   time.Time              `json:"next_retry_time"`   // 下次重试时间
	LastRetryTime   *time.Time             `json:"last_retry_time"`   // 上次重试时间
	LastRetryResult string                 `json:"last_retry_result"` // 上次重试结果
	CreatedAt       time.Time              `json:"created_at"`        // 创建时间
}

// RefundNotificationMessage 退款通知消息
type RefundNotificationMessage struct {
	MessageID       string                 `json:"message_id"`        // 消息ID
	PaymentNo       string                 `json:"payment_no"`        // 支付流水号
	RefundNo        string                 `json:"refund_no"`         // 退款流水号
	MerchantID      string                 `json:"merchant_id"`       // 商户ID
	NotifyURL       string                 `json:"notify_url"`        // 通知URL
	NotifyData      map[string]interface{} `json:"notify_data"`       // 通知数据
	Signature       string                 `json:"signature"`         // 签名
	RetryCount      int                    `json:"retry_count"`       // 重试次数
	MaxRetries      int                    `json:"max_retries"`       // 最大重试次数
	NextRetryTime   time.Time              `json:"next_retry_time"`   // 下次重试时间
	LastRetryTime   *time.Time             `json:"last_retry_time"`   // 上次重试时间
	LastRetryResult string                 `json:"last_retry_result"` // 上次重试结果
	CreatedAt       time.Time              `json:"created_at"`        // 创建时间
}

// Kafka Topic 定义
const (
	TopicCompensation         = "payment.compensation"          // 补偿消息主题
	TopicNotification         = "payment.notification"          // 支付通知主题
	TopicRefundNotification   = "payment.refund.notification"   // 退款通知主题
)

// SendCompensationMessage 发送补偿消息
func (s *messageService) SendCompensationMessage(ctx context.Context, msg *CompensationMessage) error {
	if msg.MessageID == "" {
		msg.MessageID = uuid.New().String()
	}
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now()
	}
	if msg.MaxRetries == 0 {
		msg.MaxRetries = 5 // 默认最多重试5次
	}
	if msg.NextRetryTime.IsZero() {
		msg.NextRetryTime = time.Now().Add(1 * time.Minute) // 1分钟后首次重试
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化补偿消息失败: %w", err)
	}

	if len(s.brokers) > 0 {
		writer := &kafka.Writer{
			Addr:     kafka.TCP(s.brokers...),
			Topic:    TopicCompensation,
			Balancer: &kafka.LeastBytes{},
		}
		defer writer.Close()

		if err := writer.WriteMessages(ctx, kafka.Message{
			Key:   []byte(msg.MessageID),
			Value: data,
		}); err != nil {
			return fmt.Errorf("发送补偿消息到Kafka失败: %w", err)
		}
		fmt.Printf("[MQ] 补偿消息已发送: Type=%s, PaymentNo=%s, MessageID=%s\n",
			msg.Type, msg.PaymentNo, msg.MessageID)
	} else {
		fmt.Printf("[MQ] Kafka未配置，模拟发送补偿消息: %s\n", string(data))
	}

	return nil
}

// SendNotificationMessage 发送通知消息
func (s *messageService) SendNotificationMessage(ctx context.Context, msg *NotificationMessage) error {
	if msg.MessageID == "" {
		msg.MessageID = uuid.New().String()
	}
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now()
	}
	if msg.MaxRetries == 0 {
		msg.MaxRetries = 5 // 默认最多重试5次
	}
	if msg.NextRetryTime.IsZero() {
		msg.NextRetryTime = time.Now().Add(5 * time.Second) // 5秒后首次重试
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化通知消息失败: %w", err)
	}

	if len(s.brokers) > 0 {
		writer := &kafka.Writer{
			Addr:     kafka.TCP(s.brokers...),
			Topic:    TopicNotification,
			Balancer: &kafka.LeastBytes{},
		}
		defer writer.Close()

		if err := writer.WriteMessages(ctx, kafka.Message{
			Key:   []byte(msg.MessageID),
			Value: data,
		}); err != nil {
			return fmt.Errorf("发送通知消息到Kafka失败: %w", err)
		}
		fmt.Printf("[MQ] 通知消息已发送: PaymentNo=%s, MessageID=%s, RetryCount=%d\n",
			msg.PaymentNo, msg.MessageID, msg.RetryCount)
	} else {
		fmt.Printf("[MQ] Kafka未配置，模拟发送通知消息: %s\n", string(data))
	}

	return nil
}

// SendRefundNotificationMessage 发送退款通知消息
func (s *messageService) SendRefundNotificationMessage(ctx context.Context, msg *RefundNotificationMessage) error {
	if msg.MessageID == "" {
		msg.MessageID = uuid.New().String()
	}
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now()
	}
	if msg.MaxRetries == 0 {
		msg.MaxRetries = 5 // 默认最多重试5次
	}
	if msg.NextRetryTime.IsZero() {
		msg.NextRetryTime = time.Now().Add(5 * time.Second) // 5秒后首次重试
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化退款通知消息失败: %w", err)
	}

	if len(s.brokers) > 0 {
		writer := &kafka.Writer{
			Addr:     kafka.TCP(s.brokers...),
			Topic:    TopicRefundNotification,
			Balancer: &kafka.LeastBytes{},
		}
		defer writer.Close()

		if err := writer.WriteMessages(ctx, kafka.Message{
			Key:   []byte(msg.MessageID),
			Value: data,
		}); err != nil {
			return fmt.Errorf("发送退款通知消息到Kafka失败: %w", err)
		}
		fmt.Printf("[MQ] 退款通知消息已发送: RefundNo=%s, MessageID=%s, RetryCount=%d\n",
			msg.RefundNo, msg.MessageID, msg.RetryCount)
	} else {
		fmt.Printf("[MQ] Kafka未配置，模拟发送退款通知消息: %s\n", string(data))
	}

	return nil
}

// CalculateNextRetryTime 计算下次重试时间（指数退避）
func CalculateNextRetryTime(retryCount int) time.Time {
	// 指数退避：1分钟, 2分钟, 4分钟, 8分钟, 16分钟
	delay := time.Duration(1<<uint(retryCount)) * time.Minute
	if delay > 30*time.Minute {
		delay = 30 * time.Minute // 最大延迟30分钟
	}
	return time.Now().Add(delay)
}
