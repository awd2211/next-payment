package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
)

// Producer Kafka生产者
type Producer struct {
	writer *kafka.Writer
}

// ProducerConfig 生产者配置
type ProducerConfig struct {
	Brokers []string
	Topic   string
}

// NewProducer 创建Kafka生产者
func NewProducer(config ProducerConfig) *Producer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(config.Brokers...),
		Topic:    config.Topic,
		Balancer: &kafka.LeastBytes{},
	}

	return &Producer{
		writer: writer,
	}
}

// Publish 发布消息
func (p *Producer) Publish(ctx context.Context, key string, value interface{}) error {
	// 序列化消息
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	// 发送消息
	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: valueBytes,
	})
	if err != nil {
		return fmt.Errorf("发送Kafka消息失败: %w", err)
	}

	return nil
}

// PublishBatch 批量发布消息
func (p *Producer) PublishBatch(ctx context.Context, messages []kafka.Message) error {
	err := p.writer.WriteMessages(ctx, messages...)
	if err != nil {
		return fmt.Errorf("批量发送Kafka消息失败: %w", err)
	}
	return nil
}

// Close 关闭生产者
func (p *Producer) Close() error {
	return p.writer.Close()
}
