package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache Redis缓存实现
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache 创建Redis缓存
func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{
		client: client,
	}
}

// NewRedisCacheFromConfig 从配置创建Redis缓存
func NewRedisCacheFromConfig(addr, password string, db int) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("连接Redis失败: %w", err)
	}

	return &RedisCache{client: client}, nil
}

// Get 获取缓存
func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, fmt.Errorf("获取缓存失败: %w", err)
	}
	return val, nil
}

// Set 设置缓存
func (c *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	err := c.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		return fmt.Errorf("设置缓存失败: %w", err)
	}
	return nil
}

// Delete 删除缓存
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("删除缓存失败: %w", err)
	}
	return nil
}

// Exists 检查缓存是否存在
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("检查缓存失败: %w", err)
	}
	return n > 0, nil
}

// Expire 设置过期时间
func (c *RedisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	err := c.client.Expire(ctx, key, ttl).Err()
	if err != nil {
		return fmt.Errorf("设置过期时间失败: %w", err)
	}
	return nil
}

// TTL 获取剩余过期时间
func (c *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := c.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("获取过期时间失败: %w", err)
	}
	return ttl, nil
}

// Clear 清空所有缓存
func (c *RedisCache) Clear(ctx context.Context) error {
	err := c.client.FlushDB(ctx).Err()
	if err != nil {
		return fmt.Errorf("清空缓存失败: %w", err)
	}
	return nil
}

// Close 关闭连接
func (c *RedisCache) Close() error {
	return c.client.Close()
}

// GetString 获取字符串缓存
func (c *RedisCache) GetString(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", ErrCacheMiss
	}
	if err != nil {
		return "", fmt.Errorf("获取缓存失败: %w", err)
	}
	return val, nil
}

// SetString 设置字符串缓存
func (c *RedisCache) SetString(ctx context.Context, key, value string, ttl time.Duration) error {
	err := c.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		return fmt.Errorf("设置缓存失败: %w", err)
	}
	return nil
}

// Incr 自增
func (c *RedisCache) Incr(ctx context.Context, key string) (int64, error) {
	val, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("自增失败: %w", err)
	}
	return val, nil
}

// Decr 自减
func (c *RedisCache) Decr(ctx context.Context, key string) (int64, error) {
	val, err := c.client.Decr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("自减失败: %w", err)
	}
	return val, nil
}
