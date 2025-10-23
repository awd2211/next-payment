package cache

import (
	"context"
	"time"
)

// Cache 缓存接口
type Cache interface {
	// Get 获取缓存
	Get(ctx context.Context, key string) ([]byte, error)

	// Set 设置缓存
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete 删除缓存
	Delete(ctx context.Context, key string) error

	// Exists 检查缓存是否存在
	Exists(ctx context.Context, key string) (bool, error)

	// Expire 设置过期时间
	Expire(ctx context.Context, key string, ttl time.Duration) error

	// TTL 获取剩余过期时间
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Clear 清空所有缓存
	Clear(ctx context.Context) error

	// Close 关闭缓存连接
	Close() error
}

// Entry 缓存项
type Entry struct {
	Key       string
	Value     []byte
	ExpiresAt time.Time
}

// IsExpired 检查是否过期
func (e *Entry) IsExpired() bool {
	return !e.ExpiresAt.IsZero() && time.Now().After(e.ExpiresAt)
}
