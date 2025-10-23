package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

// Consumer Kafka消费者
type Consumer struct {
	reader *kafka.Reader
}

// ConsumerConfig 消费者配置
type ConsumerConfig struct {
	Brokers  []string
	Topic    string
	GroupID  string
	MinBytes int
	MaxBytes int
}

// MessageHandler 消息处理函数
type MessageHandler func(ctx context.Context, message []byte) error

// NewConsumer 创建Kafka消费者
func NewConsumer(config ConsumerConfig) *Consumer {
	// 默认值
	if config.MinBytes == 0 {
		config.MinBytes = 10e3 // 10KB
	}
	if config.MaxBytes == 0 {
		config.MaxBytes = 10e6 // 10MB
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  config.Brokers,
		Topic:    config.Topic,
		GroupID:  config.GroupID,
		MinBytes: config.MinBytes,
		MaxBytes: config.MaxBytes,
	})

	return &Consumer{
		reader: reader,
	}
}

// Consume 开始消费消息
func (c *Consumer) Consume(ctx context.Context, handler MessageHandler) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// 读取消息
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				return fmt.Errorf("读取Kafka消息失败: %w", err)
			}

			// 处理消息
			if err := handler(ctx, msg.Value); err != nil {
				// 处理失败，记录错误但继续消费
				fmt.Printf("处理Kafka消息失败: %v\n", err)
				// 这里可以选择是否提交offset
				// 如果不提交，下次会重新消费
				continue
			}

			// 提交offset
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				fmt.Printf("提交Kafka offset失败: %v\n", err)
			}
		}
	}
}

// ConsumeWithRetry 消费消息并支持重试
func (c *Consumer) ConsumeWithRetry(ctx context.Context, handler MessageHandler, maxRetries int) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				return fmt.Errorf("读取Kafka消息失败: %w", err)
			}

			// 重试逻辑
			var lastErr error
			for i := 0; i <= maxRetries; i++ {
				if err := handler(ctx, msg.Value); err != nil {
					lastErr = err
					// 指数退避
					if i < maxRetries {
						backoff := time.Duration(1<<uint(i)) * time.Second
						fmt.Printf("处理失败，%d秒后重试 (尝试 %d/%d): %v\n", backoff/time.Second, i+1, maxRetries+1, err)
						time.Sleep(backoff)
						continue
					}
				} else {
					// 成功处理
					lastErr = nil
					break
				}
			}

			if lastErr != nil {
				// 所有重试都失败，记录错误
				fmt.Printf("处理Kafka消息失败（已达最大重试次数）: %v\n", lastErr)
				// 可以发送到死信队列
			}

			// 提交offset（即使处理失败，避免无限重试阻塞队列）
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				fmt.Printf("提交Kafka offset失败: %v\n", err)
			}
		}
	}
}

// Close 关闭消费者
func (c *Consumer) Close() error {
	return c.reader.Close()
}

// UnmarshalMessage 辅助函数：反序列化消息
func UnmarshalMessage(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
