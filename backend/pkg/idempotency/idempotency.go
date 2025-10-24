package idempotency

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// IdempotencyManager 幂等性管理器
type IdempotencyManager struct {
	redis  *redis.Client
	prefix string
	ttl    time.Duration
}

// Response 幂等性响应缓存
type Response struct {
	StatusCode int         `json:"status_code"`
	Body       interface{} `json:"body"`
	Error      string      `json:"error,omitempty"`
	CreatedAt  time.Time   `json:"created_at"`
}

// NewIdempotencyManager 创建幂等性管理器
func NewIdempotencyManager(redis *redis.Client, prefix string, ttl time.Duration) *IdempotencyManager {
	if ttl == 0 {
		ttl = 24 * time.Hour // 默认24小时
	}
	return &IdempotencyManager{
		redis:  redis,
		prefix: prefix,
		ttl:    ttl,
	}
}

// GetKey 构建Redis键
func (m *IdempotencyManager) GetKey(idempotencyKey string) string {
	return fmt.Sprintf("%s:idempotency:%s", m.prefix, idempotencyKey)
}

// Check 检查幂等性Key是否已存在
// 返回 (isProcessing, cachedResponse, error)
func (m *IdempotencyManager) Check(ctx context.Context, idempotencyKey string) (bool, *Response, error) {
	if idempotencyKey == "" {
		return false, nil, nil // 未提供幂等性Key，允许继续
	}

	key := m.GetKey(idempotencyKey)

	// 使用 Redis SETNX 实现分布式锁
	// 如果key不存在，设置为 "processing"，返回true（未在处理中）
	// 如果key存在，返回false（正在处理或已完成）
	lockKey := key + ":lock"
	locked, err := m.redis.SetNX(ctx, lockKey, "processing", 10*time.Second).Result()
	if err != nil {
		return false, nil, fmt.Errorf("检查幂等性失败: %w", err)
	}

	if locked {
		// 成功获取锁，说明这是第一次处理该请求
		return false, nil, nil
	}

	// 未能获取锁，检查是否有缓存的响应
	val, err := m.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		// 正在处理中，没有响应缓存
		return true, nil, nil
	}
	if err != nil {
		return false, nil, fmt.Errorf("获取缓存响应失败: %w", err)
	}

	// 解析缓存的响应
	var resp Response
	if err := json.Unmarshal([]byte(val), &resp); err != nil {
		return false, nil, fmt.Errorf("解析缓存响应失败: %w", err)
	}

	return false, &resp, nil
}

// Store 存储响应到缓存
func (m *IdempotencyManager) Store(ctx context.Context, idempotencyKey string, statusCode int, body interface{}, errorMsg string) error {
	if idempotencyKey == "" {
		return nil // 未提供幂等性Key，不缓存
	}

	key := m.GetKey(idempotencyKey)

	resp := Response{
		StatusCode: statusCode,
		Body:       body,
		Error:      errorMsg,
		CreatedAt:  time.Now(),
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("序列化响应失败: %w", err)
	}

	// 存储响应，设置TTL
	if err := m.redis.Set(ctx, key, data, m.ttl).Err(); err != nil {
		return fmt.Errorf("存储响应缓存失败: %w", err)
	}

	// 删除处理锁
	lockKey := key + ":lock"
	m.redis.Del(ctx, lockKey)

	return nil
}

// Delete 删除幂等性Key（用于测试或清理）
func (m *IdempotencyManager) Delete(ctx context.Context, idempotencyKey string) error {
	if idempotencyKey == "" {
		return nil
	}

	key := m.GetKey(idempotencyKey)
	lockKey := key + ":lock"

	pipe := m.redis.Pipeline()
	pipe.Del(ctx, key)
	pipe.Del(ctx, lockKey)
	_, err := pipe.Exec(ctx)

	return err
}

// Cleanup 清理过期的幂等性记录（可以定时调用）
func (m *IdempotencyManager) Cleanup(ctx context.Context) error {
	// Redis会自动清理过期的key，这里主要是清理孤儿锁
	pattern := m.prefix + ":idempotency:*:lock"
	iter := m.redis.Scan(ctx, 0, pattern, 100).Iterator()

	count := 0
	for iter.Next(ctx) {
		key := iter.Val()
		// 检查对应的响应key是否存在
		respKey := key[:len(key)-5] // 去掉 ":lock" 后缀
		exists, _ := m.redis.Exists(ctx, respKey).Result()
		if exists == 0 {
			// 响应不存在但锁存在，说明是孤儿锁，删除
			m.redis.Del(ctx, key)
			count++
		}
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("清理幂等性锁失败: %w", err)
	}

	return nil
}
