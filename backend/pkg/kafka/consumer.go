package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

// Consumer Kafka消费者
type Consumer struct {
	consumerGroup sarama.ConsumerGroup
	config        *ConsumerConfig
}

// ConsumerConfig 消费者配置
type ConsumerConfig struct {
	Brokers  []string
	GroupID  string
	Topics   []string
	ClientID string
}

// DefaultConsumerConfig 默认消费者配置
func DefaultConsumerConfig() *ConsumerConfig {
	return &ConsumerConfig{
		Brokers:  []string{"localhost:9092"},
		GroupID:  "payment-consumer-group",
		ClientID: "payment-consumer",
	}
}

// NewConsumer 创建Kafka消费者
func NewConsumer(config *ConsumerConfig) (*Consumer, error) {
	if config == nil {
		config = DefaultConsumerConfig()
	}

	saramaConfig := sarama.NewConfig()
	saramaConfig.ClientID = config.ClientID
	saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
	saramaConfig.Consumer.Return.Errors = true

	consumerGroup, err := sarama.NewConsumerGroup(config.Brokers, config.GroupID, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("创建Kafka消费者失败: %w", err)
	}

	return &Consumer{
		consumerGroup: consumerGroup,
		config:        config,
	}, nil
}

// MessageHandler 消息处理器函数
type MessageHandler func(ctx context.Context, message *ConsumedMessage) error

// ConsumedMessage 消费的消息
type ConsumedMessage struct {
	Topic     string
	Partition int32
	Offset    int64
	Key       string
	Value     []byte
	Timestamp int64
}

// ParseJSON 解析JSON消息
func (m *ConsumedMessage) ParseJSON(v interface{}) error {
	return json.Unmarshal(m.Value, v)
}

// Consume 开始消费消息
func (c *Consumer) Consume(ctx context.Context, handler MessageHandler) error {
	consumerHandler := &consumerGroupHandler{
		handler: handler,
	}

	for {
		err := c.consumerGroup.Consume(ctx, c.config.Topics, consumerHandler)
		if err != nil {
			return fmt.Errorf("消费消息失败: %w", err)
		}

		// 检查context是否已取消
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

// Close 关闭消费者
func (c *Consumer) Close() error {
	return c.consumerGroup.Close()
}

// consumerGroupHandler 实现sarama.ConsumerGroupHandler接口
type consumerGroupHandler struct {
	handler MessageHandler
}

// Setup 在开始消费前调用
func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup 在停止消费后调用
func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim 处理消息
func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		msg := &ConsumedMessage{
			Topic:     message.Topic,
			Partition: message.Partition,
			Offset:    message.Offset,
			Key:       string(message.Key),
			Value:     message.Value,
			Timestamp: message.Timestamp.Unix(),
		}

		// 调用用户提供的处理器
		if err := h.handler(session.Context(), msg); err != nil {
			log.Printf("处理消息失败: %v (Topic: %s, Partition: %d, Offset: %d)",
				err, msg.Topic, msg.Partition, msg.Offset)
			// 继续处理下一条消息
			continue
		}

		// 标记消息已处理
		session.MarkMessage(message, "")
	}

	return nil
}
