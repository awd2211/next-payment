package idempotent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Service 幂等性服务接口
type Service interface {
	// Check 检查请求是否已处理，如果已处理则返回缓存的结果
	Check(ctx context.Context, key string, result interface{}) (exists bool, err error)

	// Store 存储请求处理结果（用于幂等性）
	Store(ctx context.Context, key string, result interface{}, ttl time.Duration) error

	// Delete 删除幂等性记录（用于失败重试）
	Delete(ctx context.Context, key string) error

	// Try 尝试获取分布式锁（防止并发重复请求）
	Try(ctx context.Context, key string, ttl time.Duration) (acquired bool, err error)

	// Release 释放分布式锁
	Release(ctx context.Context, key string) error
}

type service struct {
	redis *redis.Client
}

// NewService 创建幂等性服务实例
func NewService(redisClient *redis.Client) Service {
	return &service{
		redis: redisClient,
	}
}

// Check 检查请求是否已处理
func (s *service) Check(ctx context.Context, key string, result interface{}) (bool, error) {
	// 构造幂等性键
	idempotentKey := fmt.Sprintf("idempotent:%s", key)

	// 尝试从 Redis 获取缓存结果
	data, err := s.redis.Get(ctx, idempotentKey).Result()
	if err == redis.Nil {
		// 未找到缓存，表示第一次请求
		return false, nil
	}
	if err != nil {
		// Redis 错误，不阻塞请求（降级处理）
		return false, err
	}

	// 找到缓存，反序列化结果
	if result != nil {
		if err := json.Unmarshal([]byte(data), result); err != nil {
			return true, fmt.Errorf("反序列化缓存结果失败: %w", err)
		}
	}

	return true, nil
}

// Store 存储请求处理结果
func (s *service) Store(ctx context.Context, key string, result interface{}, ttl time.Duration) error {
	idempotentKey := fmt.Sprintf("idempotent:%s", key)

	// 序列化结果
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("序列化结果失败: %w", err)
	}

	// 存储到 Redis
	return s.redis.Set(ctx, idempotentKey, data, ttl).Err()
}

// Delete 删除幂等性记录
func (s *service) Delete(ctx context.Context, key string) error {
	idempotentKey := fmt.Sprintf("idempotent:%s", key)
	return s.redis.Del(ctx, idempotentKey).Err()
}

// Try 尝试获取分布式锁
func (s *service) Try(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	lockKey := fmt.Sprintf("lock:%s", key)

	// 使用 SetNX 实现分布式锁
	acquired, err := s.redis.SetNX(ctx, lockKey, "locked", ttl).Result()
	if err != nil {
		return false, err
	}

	return acquired, nil
}

// Release 释放分布式锁
func (s *service) Release(ctx context.Context, key string) error {
	lockKey := fmt.Sprintf("lock:%s", key)
	return s.redis.Del(ctx, lockKey).Err()
}

// GenerateKey 生成幂等性键的辅助函数
func GenerateKey(parts ...string) string {
	result := ""
	for i, part := range parts {
		if i > 0 {
			result += ":"
		}
		result += part
	}
	return result
}
