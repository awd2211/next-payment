package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimitConfig 速率限制配置
type RateLimitConfig struct {
	// 每分钟请求数
	RequestsPerMinute int
	// 每小时请求数
	RequestsPerHour int
	// 突发容量（令牌桶）
	BurstCapacity int
	// 是否按用户限流
	PerUser bool
	// 是否按IP限流
	PerIP bool
	// 自定义key提取器
	KeyExtractor func(*gin.Context) string
}

// rateLimitEntry 速率限制条目
type rateLimitEntry struct {
	tokens         float64
	lastRefillTime time.Time
	requestCount   int
	hourlyCount    int
	hourStartTime  time.Time
	mu             sync.Mutex
}

// AdvancedRateLimiter 高级速率限制器
type AdvancedRateLimiter struct {
	config  *RateLimitConfig
	buckets map[string]*rateLimitEntry
	mu      sync.RWMutex
	cleanup *time.Ticker
}

// NewAdvancedRateLimiter 创建高级速率限制器
func NewAdvancedRateLimiter(config *RateLimitConfig) *AdvancedRateLimiter {
	limiter := &AdvancedRateLimiter{
		config:  config,
		buckets: make(map[string]*rateLimitEntry),
		cleanup: time.NewTicker(5 * time.Minute),
	}

	// 定期清理过期条目
	go limiter.cleanupRoutine()

	return limiter
}

// Middleware 返回Gin中间件
func (rl *AdvancedRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 提取限流key
		key := rl.extractKey(c)

		// 2. 获取或创建bucket
		entry := rl.getOrCreateEntry(key)

		// 3. 检查限流
		entry.mu.Lock()
		allowed, retryAfter := rl.checkLimit(entry)
		entry.mu.Unlock()

		// 4. 设置响应头
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rl.config.RequestsPerMinute))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%.0f", entry.tokens))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", entry.lastRefillTime.Add(time.Minute).Unix()))

		if !allowed {
			c.Header("Retry-After", fmt.Sprintf("%d", int(retryAfter.Seconds())))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "请求过于频繁",
				"code":    "RATE_LIMIT_EXCEEDED",
				"message": fmt.Sprintf("请在 %d 秒后重试", int(retryAfter.Seconds())),
				"details": map[string]interface{}{
					"limit":     rl.config.RequestsPerMinute,
					"remaining": int(entry.tokens),
					"reset_at":  entry.lastRefillTime.Add(time.Minute).Unix(),
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// extractKey 提取限流key
func (rl *AdvancedRateLimiter) extractKey(c *gin.Context) string {
	// 自定义提取器
	if rl.config.KeyExtractor != nil {
		return rl.config.KeyExtractor(c)
	}

	// 按用户 + IP
	if rl.config.PerUser && rl.config.PerIP {
		userID := c.GetString("user_id")
		ip := c.ClientIP()
		return fmt.Sprintf("user:%s:ip:%s", userID, ip)
	}

	// 按用户
	if rl.config.PerUser {
		userID := c.GetString("user_id")
		if userID != "" {
			return fmt.Sprintf("user:%s", userID)
		}
	}

	// 按IP
	return fmt.Sprintf("ip:%s", c.ClientIP())
}

// getOrCreateEntry 获取或创建条目
func (rl *AdvancedRateLimiter) getOrCreateEntry(key string) *rateLimitEntry {
	rl.mu.RLock()
	entry, exists := rl.buckets[key]
	rl.mu.RUnlock()

	if exists {
		return entry
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// 双重检查
	if entry, exists := rl.buckets[key]; exists {
		return entry
	}

	entry = &rateLimitEntry{
		tokens:         float64(rl.config.BurstCapacity),
		lastRefillTime: time.Now(),
		hourStartTime:  time.Now(),
	}
	rl.buckets[key] = entry

	return entry
}

// checkLimit 检查是否允许请求（令牌桶算法）
func (rl *AdvancedRateLimiter) checkLimit(entry *rateLimitEntry) (allowed bool, retryAfter time.Duration) {
	now := time.Now()

	// 1. 重新填充令牌
	elapsed := now.Sub(entry.lastRefillTime)
	tokensToAdd := elapsed.Seconds() * float64(rl.config.RequestsPerMinute) / 60.0
	entry.tokens += tokensToAdd
	if entry.tokens > float64(rl.config.BurstCapacity) {
		entry.tokens = float64(rl.config.BurstCapacity)
	}
	entry.lastRefillTime = now

	// 2. 检查小时限制
	if rl.config.RequestsPerHour > 0 {
		if now.Sub(entry.hourStartTime) >= time.Hour {
			// 重置小时计数
			entry.hourlyCount = 0
			entry.hourStartTime = now
		}

		if entry.hourlyCount >= rl.config.RequestsPerHour {
			retryAfter = entry.hourStartTime.Add(time.Hour).Sub(now)
			return false, retryAfter
		}
	}

	// 3. 检查令牌
	if entry.tokens < 1.0 {
		// 计算需要等待的时间
		tokensNeeded := 1.0 - entry.tokens
		retryAfter = time.Duration(tokensNeeded*60.0/float64(rl.config.RequestsPerMinute)) * time.Second
		return false, retryAfter
	}

	// 4. 消耗令牌
	entry.tokens -= 1.0
	entry.requestCount++
	entry.hourlyCount++

	return true, 0
}

// cleanupRoutine 清理过期条目
func (rl *AdvancedRateLimiter) cleanupRoutine() {
	for range rl.cleanup.C {
		rl.mu.Lock()
		now := time.Now()
		for key, entry := range rl.buckets {
			entry.mu.Lock()
			// 删除10分钟未使用的条目
			if now.Sub(entry.lastRefillTime) > 10*time.Minute {
				delete(rl.buckets, key)
			}
			entry.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

// PresetConfigs 预设配置

// StrictRateLimit 严格限流（防止暴力破解）
var StrictRateLimit = &RateLimitConfig{
	RequestsPerMinute: 10,
	RequestsPerHour:   100,
	BurstCapacity:     5,
	PerUser:           true,
	PerIP:             true,
}

// NormalRateLimit 正常限流（一般API）
var NormalRateLimit = &RateLimitConfig{
	RequestsPerMinute: 60,
	RequestsPerHour:   1000,
	BurstCapacity:     30,
	PerUser:           true,
	PerIP:             false,
}

// RelaxedRateLimit 宽松限流（只读API）
var RelaxedRateLimit = &RateLimitConfig{
	RequestsPerMinute: 300,
	RequestsPerHour:   5000,
	BurstCapacity:     100,
	PerUser:           true,
	PerIP:             false,
}

// SensitiveOperationLimit 敏感操作限流
var SensitiveOperationLimit = &RateLimitConfig{
	RequestsPerMinute: 5,
	RequestsPerHour:   20,
	BurstCapacity:     2,
	PerUser:           true,
	PerIP:             true,
}

// RateLimitByEndpoint 根据端点类型自动选择限流策略
func RateLimitByEndpoint() gin.HandlerFunc {
	strictLimiter := NewAdvancedRateLimiter(StrictRateLimit)
	normalLimiter := NewAdvancedRateLimiter(NormalRateLimit)
	relaxedLimiter := NewAdvancedRateLimiter(RelaxedRateLimit)
	sensitiveLimiter := NewAdvancedRateLimiter(SensitiveOperationLimit)

	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method

		// 敏感操作
		if method == "DELETE" ||
			containsAny(path, []string{"/approve", "/reject", "/freeze", "/unfreeze"}) {
			sensitiveLimiter.Middleware()(c)
			return
		}

		// 写操作
		if method == "POST" || method == "PUT" || method == "PATCH" {
			strictLimiter.Middleware()(c)
			return
		}

		// 登录/认证
		if containsAny(path, []string{"/login", "/auth", "/token"}) {
			strictLimiter.Middleware()(c)
			return
		}

		// 只读操作
		if method == "GET" {
			relaxedLimiter.Middleware()(c)
			return
		}

		// 默认
		normalLimiter.Middleware()(c)
	}
}

// containsAny 检查字符串是否包含任一子串
func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if contains(s, substr) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || len(s) > len(substr) && contains(s[1:], substr)
}
