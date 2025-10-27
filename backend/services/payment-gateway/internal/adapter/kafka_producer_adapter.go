package adapter

import (
	"context"
	"time"

	"github.com/payment-platform/pkg/logger"
	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// KafkaProducerAdapter 将原始 Kafka writer 适配为 service.KafkaProducer 接口
// 这个适配器用于将字符串消息发布到Kafka topic
type KafkaProducerAdapter struct {
	brokers      []string
	writers      map[string]*kafkago.Writer // topic -> writer
	fallbackMode bool                       // 降级模式（Kafka不可用时只打印日志）
}

// NewKafkaProducerAdapter 创建 Kafka 生产者适配器
func NewKafkaProducerAdapter(brokers []string) *KafkaProducerAdapter {
	return &KafkaProducerAdapter{
		brokers:      brokers,
		writers:      make(map[string]*kafkago.Writer),
		fallbackMode: len(brokers) == 0,
	}
}

// getOrCreateWriter 获取或创建指定topic的writer
func (a *KafkaProducerAdapter) getOrCreateWriter(topic string) *kafkago.Writer {
	if writer, exists := a.writers[topic]; exists {
		return writer
	}

	writer := &kafkago.Writer{
		Addr:         kafkago.TCP(a.brokers...),
		Topic:        topic,
		Balancer:     &kafkago.LeastBytes{},
		MaxAttempts:  3,
		WriteTimeout: 10 * time.Second,
	}

	a.writers[topic] = writer
	return writer
}

// Publish 实现 service.KafkaProducer 接口
// 直接发布字符串消息到指定topic
func (a *KafkaProducerAdapter) Publish(ctx context.Context, topic string, message string) error {
	// 降级模式：静默返回（不发送消息，不记录日志）
	if a.fallbackMode {
		return nil
	}

	writer := a.getOrCreateWriter(topic)
	err := writer.WriteMessages(ctx, kafkago.Message{
		Value: []byte(message),
	})

	if err != nil {
		logger.Error("failed to publish message to Kafka",
			zap.String("topic", topic),
			zap.Error(err))
		return err
	}

	return nil
}

// Close 关闭所有 writers
func (a *KafkaProducerAdapter) Close() error {
	for topic, writer := range a.writers {
		if err := writer.Close(); err != nil {
			logger.Error("failed to close Kafka writer",
				zap.String("topic", topic),
				zap.Error(err))
		}
	}
	return nil
}
