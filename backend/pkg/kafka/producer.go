package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
)

// Producer Kafka生产者
type Producer struct {
	producer sarama.SyncProducer
	config   *ProducerConfig
}

// ProducerConfig 生产者配置
type ProducerConfig struct {
	Brokers          []string
	ClientID         string
	Compression      string // none, gzip, snappy, lz4, zstd
	MaxMessageBytes  int
	RequiredAcks     int16 // 0, 1, -1
	RetryMax         int
	RetryBackoff     time.Duration
	EnableIdempotent bool
}

// DefaultProducerConfig 默认生产者配置
func DefaultProducerConfig() *ProducerConfig {
	return &ProducerConfig{
		Brokers:          []string{"localhost:9092"},
		ClientID:         "payment-producer",
		Compression:      "snappy",
		MaxMessageBytes:  1000000,
		RequiredAcks:     -1, // 等待所有ISR副本确认
		RetryMax:         3,
		RetryBackoff:     100 * time.Millisecond,
		EnableIdempotent: true,
	}
}

// NewProducer 创建Kafka生产者
func NewProducer(config *ProducerConfig) (*Producer, error) {
	if config == nil {
		config = DefaultProducerConfig()
	}

	saramaConfig := sarama.NewConfig()
	saramaConfig.ClientID = config.ClientID
	saramaConfig.Producer.RequiredAcks = sarama.RequiredAcks(config.RequiredAcks)
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Return.Errors = true
	saramaConfig.Producer.Retry.Max = config.RetryMax
	saramaConfig.Producer.Retry.Backoff = config.RetryBackoff
	saramaConfig.Producer.MaxMessageBytes = config.MaxMessageBytes
	saramaConfig.Producer.Idempotent = config.EnableIdempotent

	// 设置压缩
	switch config.Compression {
	case "gzip":
		saramaConfig.Producer.Compression = sarama.CompressionGZIP
	case "snappy":
		saramaConfig.Producer.Compression = sarama.CompressionSnappy
	case "lz4":
		saramaConfig.Producer.Compression = sarama.CompressionLZ4
	case "zstd":
		saramaConfig.Producer.Compression = sarama.CompressionZSTD
	default:
		saramaConfig.Producer.Compression = sarama.CompressionNone
	}

	producer, err := sarama.NewSyncProducer(config.Brokers, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("创建Kafka生产者失败: %w", err)
	}

	return &Producer{
		producer: producer,
		config:   config,
	}, nil
}

// Message 消息
type Message struct {
	Topic string
	Key   string
	Value interface{}
}

// SendMessage 发送消息
func (p *Producer) SendMessage(msg *Message) (partition int32, offset int64, err error) {
	// 序列化Value
	valueBytes, err := json.Marshal(msg.Value)
	if err != nil {
		return 0, 0, fmt.Errorf("序列化消息失败: %w", err)
	}

	producerMsg := &sarama.ProducerMessage{
		Topic: msg.Topic,
		Value: sarama.ByteEncoder(valueBytes),
	}

	// 如果有Key，设置Key
	if msg.Key != "" {
		producerMsg.Key = sarama.StringEncoder(msg.Key)
	}

	partition, offset, err = p.producer.SendMessage(producerMsg)
	if err != nil {
		return 0, 0, fmt.Errorf("发送消息失败: %w", err)
	}

	return partition, offset, nil
}

// SendMessages 批量发送消息
func (p *Producer) SendMessages(msgs []*Message) error {
	producerMsgs := make([]*sarama.ProducerMessage, 0, len(msgs))

	for _, msg := range msgs {
		valueBytes, err := json.Marshal(msg.Value)
		if err != nil {
			return fmt.Errorf("序列化消息失败: %w", err)
		}

		producerMsg := &sarama.ProducerMessage{
			Topic: msg.Topic,
			Value: sarama.ByteEncoder(valueBytes),
		}

		if msg.Key != "" {
			producerMsg.Key = sarama.StringEncoder(msg.Key)
		}

		producerMsgs = append(producerMsgs, producerMsg)
	}

	return p.producer.SendMessages(producerMsgs)
}

// Close 关闭生产者
func (p *Producer) Close() error {
	return p.producer.Close()
}

// SendEvent 发送事件（便捷方法）
func (p *Producer) SendEvent(topic, eventType string, data interface{}) error {
	event := map[string]interface{}{
		"event_type": eventType,
		"data":       data,
		"timestamp":  time.Now().Unix(),
	}

	_, _, err := p.SendMessage(&Message{
		Topic: topic,
		Key:   eventType,
		Value: event,
	})

	return err
}

// SendWithContext 使用context发送消息
func (p *Producer) SendWithContext(ctx context.Context, msg *Message) error {
	// 创建channel接收结果
	resultChan := make(chan error, 1)

	go func() {
		_, _, err := p.SendMessage(msg)
		resultChan <- err
	}()

	// 等待结果或context取消
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-resultChan:
		return err
	}
}
