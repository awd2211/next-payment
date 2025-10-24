package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/events"
	"github.com/payment-platform/pkg/logger"
	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// EventPublisher 事件发布器 (统一事件发布接口)
type EventPublisher struct {
	brokers      []string
	writers      map[string]*kafkago.Writer // Topic -> Writer
	fallbackMode bool                       // 降级模式 (Kafka不可用时只打印日志)
}

// NewEventPublisher 创建事件发布器
// brokers: Kafka broker地址列表,如果为nil或空则进入降级模式
func NewEventPublisher(brokers []string) *EventPublisher {
	publisher := &EventPublisher{
		brokers:      brokers,
		writers:      make(map[string]*kafkago.Writer),
		fallbackMode: len(brokers) == 0,
	}

	if publisher.fallbackMode {
		logger.Warn("Kafka brokers not configured, EventPublisher running in fallback mode (log only)")
	}

	return publisher
}

// getOrCreateWriter 获取或创建指定Topic的Writer
func (p *EventPublisher) getOrCreateWriter(topic string) *kafkago.Writer {
	if writer, exists := p.writers[topic]; exists {
		return writer
	}

	writer := &kafkago.Writer{
		Addr:         kafkago.TCP(p.brokers...),
		Topic:        topic,
		Balancer:     &kafkago.LeastBytes{}, // 负载均衡策略
		MaxAttempts:  3,                      // 最多重试3次
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
		RequiredAcks: kafkago.RequireOne, // 至少1个副本确认
		Async:        false,               // 同步发送(保证可靠性)
	}

	p.writers[topic] = writer
	return writer
}

// Publish 发布事件到Kafka
// 如果Kafka不可用,会降级到日志输出,不阻塞业务流程
func (p *EventPublisher) Publish(ctx context.Context, topic string, event events.Event) error {
	// 序列化事件
	data, err := event.ToJSON()
	if err != nil {
		logger.Error("failed to serialize event",
			zap.String("event_type", event.GetEventType()),
			zap.Error(err))
		return fmt.Errorf("序列化事件失败: %w", err)
	}

	// 降级模式: 只打印日志
	if p.fallbackMode {
		logger.Warn("event published in fallback mode (log only)",
			zap.String("topic", topic),
			zap.String("event_type", event.GetEventType()),
			zap.String("aggregate_type", event.GetAggregateType()),
			zap.String("aggregate_id", event.GetAggregateID()),
			zap.ByteString("payload", data))
		return nil
	}

	// 获取Writer
	writer := p.getOrCreateWriter(topic)

	// 构造Kafka消息
	message := kafkago.Message{
		Key:   []byte(event.GetAggregateID()), // 使用aggregate_id作为Key (保证同一实体有序)
		Value: data,
		Headers: []kafkago.Header{
			{Key: "event_type", Value: []byte(event.GetEventType())},
			{Key: "aggregate_type", Value: []byte(event.GetAggregateType())},
			{Key: "timestamp", Value: []byte(time.Now().Format(time.RFC3339))},
		},
	}

	// 发送到Kafka
	err = writer.WriteMessages(ctx, message)
	if err != nil {
		logger.Error("failed to publish event to kafka",
			zap.String("topic", topic),
			zap.String("event_type", event.GetEventType()),
			zap.Error(err))
		return fmt.Errorf("发送事件到Kafka失败: %w", err)
	}

	logger.Info("event published successfully",
		zap.String("topic", topic),
		zap.String("event_type", event.GetEventType()),
		zap.String("aggregate_id", event.GetAggregateID()))

	return nil
}

// PublishAsync 异步发布事件 (不阻塞主流程,失败只记录日志)
// 推荐用于非关键事件 (如通知、统计)
func (p *EventPublisher) PublishAsync(ctx context.Context, topic string, event events.Event) {
	go func() {
		// 使用独立context,避免父context取消影响
		publishCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := p.Publish(publishCtx, topic, event); err != nil {
			logger.Warn("async event publish failed (non-fatal)",
				zap.String("topic", topic),
				zap.String("event_type", event.GetEventType()),
				zap.Error(err))
		}
	}()
}

// PublishBatch 批量发布事件 (用于批处理场景)
func (p *EventPublisher) PublishBatch(ctx context.Context, topic string, events []events.Event) error {
	if p.fallbackMode {
		for _, event := range events {
			_ = p.Publish(ctx, topic, event) // 降级模式逐个打印
		}
		return nil
	}

	writer := p.getOrCreateWriter(topic)
	messages := make([]kafkago.Message, 0, len(events))

	for _, event := range events {
		data, err := event.ToJSON()
		if err != nil {
			logger.Error("failed to serialize event in batch",
				zap.String("event_type", event.GetEventType()),
				zap.Error(err))
			continue
		}

		messages = append(messages, kafkago.Message{
			Key:   []byte(event.GetAggregateID()),
			Value: data,
			Headers: []kafkago.Header{
				{Key: "event_type", Value: []byte(event.GetEventType())},
				{Key: "aggregate_type", Value: []byte(event.GetAggregateType())},
			},
		})
	}

	if len(messages) == 0 {
		return fmt.Errorf("no valid messages to publish")
	}

	err := writer.WriteMessages(ctx, messages...)
	if err != nil {
		logger.Error("failed to publish batch events",
			zap.String("topic", topic),
			zap.Int("count", len(messages)),
			zap.Error(err))
		return err
	}

	logger.Info("batch events published successfully",
		zap.String("topic", topic),
		zap.Int("count", len(messages)))

	return nil
}

// Close 关闭所有Writer
func (p *EventPublisher) Close() error {
	for topic, writer := range p.writers {
		if err := writer.Close(); err != nil {
			logger.Error("failed to close kafka writer",
				zap.String("topic", topic),
				zap.Error(err))
		}
	}
	return nil
}

// PublishPaymentEvent 发布支付事件的便捷方法
func (p *EventPublisher) PublishPaymentEvent(ctx context.Context, event *events.PaymentEvent) error {
	return p.Publish(ctx, events.TopicPaymentEvents, event)
}

// PublishPaymentEventAsync 异步发布支付事件
func (p *EventPublisher) PublishPaymentEventAsync(ctx context.Context, event *events.PaymentEvent) {
	p.PublishAsync(ctx, events.TopicPaymentEvents, event)
}

// PublishOrderEvent 发布订单事件的便捷方法
func (p *EventPublisher) PublishOrderEvent(ctx context.Context, event *events.OrderEvent) error {
	return p.Publish(ctx, events.TopicOrderEvents, event)
}

// PublishOrderEventAsync 异步发布订单事件
func (p *EventPublisher) PublishOrderEventAsync(ctx context.Context, event *events.OrderEvent) {
	p.PublishAsync(ctx, events.TopicOrderEvents, event)
}

// GenerateCorrelationID 生成关联ID (用于追踪事件链路)
func GenerateCorrelationID() string {
	return uuid.New().String()
}
