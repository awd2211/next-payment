package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	// ErrCacheMiss 缓存未命中
	ErrCacheMiss = errors.New("cache miss")
)

// MemoryCache 内存缓存实现
type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]*Entry
}

// NewMemoryCache 创建内存缓存
func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		items: make(map[string]*Entry),
	}

	// 启动过期清理协程
	go cache.cleanupExpired()

	return cache
}

// Get 获取缓存
func (c *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.items[key]
	if !exists {
		return nil, ErrCacheMiss
	}

	if entry.IsExpired() {
		return nil, ErrCacheMiss
	}

	return entry.Value, nil
}

// Set 设置缓存
func (c *MemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := &Entry{
		Key:   key,
		Value: value,
	}

	if ttl > 0 {
		entry.ExpiresAt = time.Now().Add(ttl)
	}

	c.items[key] = entry
	return nil
}

// Delete 删除缓存
func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
	return nil
}

// Exists 检查缓存是否存在
func (c *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.items[key]
	if !exists {
		return false, nil
	}

	if entry.IsExpired() {
		return false, nil
	}

	return true, nil
}

// Expire 设置过期时间
func (c *MemoryCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.items[key]
	if !exists {
		return ErrCacheMiss
	}

	entry.ExpiresAt = time.Now().Add(ttl)
	return nil
}

// TTL 获取剩余过期时间
func (c *MemoryCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.items[key]
	if !exists {
		return 0, ErrCacheMiss
	}

	if entry.ExpiresAt.IsZero() {
		return -1, nil // 永不过期
	}

	ttl := time.Until(entry.ExpiresAt)
	if ttl < 0 {
		return 0, ErrCacheMiss
	}

	return ttl, nil
}

// Clear 清空所有缓存
func (c *MemoryCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*Entry)
	return nil
}

// Close 关闭缓存（内存缓存无需关闭）
func (c *MemoryCache) Close() error {
	return nil
}

// cleanupExpired 清理过期缓存
func (c *MemoryCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		for key, entry := range c.items {
			if entry.IsExpired() {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

// Size 获取缓存项数量
func (c *MemoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.items)
}
