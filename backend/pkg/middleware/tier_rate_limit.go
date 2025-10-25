package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/payment-platform/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// TierRateLimitConfig 商户等级限流配置
type TierRateLimitConfig struct {
	Tier               string // starter, business, enterprise, premium
	RequestsPerSecond  int    // 每秒请求数
	RequestsPerMinute  int    // 每分钟请求数
	RequestsPerHour    int    // 每小时请求数
	BurstSize          int    // 突发请求数（令牌桶大小）
	ConcurrentRequests int    // 最大并发请求数
}

// DefaultTierConfigs 默认的等级限流配置
var DefaultTierConfigs = map[string]*TierRateLimitConfig{
	"starter": {
		Tier:               "starter",
		RequestsPerSecond:  10,
		RequestsPerMinute:  500,
		RequestsPerHour:    10000,
		BurstSize:          20,
		ConcurrentRequests: 10,
	},
	"business": {
		Tier:               "business",
		RequestsPerSecond:  50,
		RequestsPerMinute:  2500,
		RequestsPerHour:    50000,
		BurstSize:          100,
		ConcurrentRequests: 50,
	},
	"enterprise": {
		Tier:               "enterprise",
		RequestsPerSecond:  200,
		RequestsPerMinute:  10000,
		RequestsPerHour:    200000,
		BurstSize:          400,
		ConcurrentRequests: 200,
	},
	"premium": {
		Tier:               "premium",
		RequestsPerSecond:  500,
		RequestsPerMinute:  25000,
		RequestsPerHour:    500000,
		BurstSize:          1000,
		ConcurrentRequests: 500,
	},
}

// TierRateLimiter 基于商户等级的限流器
type TierRateLimiter struct {
	redisClient    *redis.Client
	tierConfigs    map[string]*TierRateLimitConfig
	getTierFunc    func(merchantID uuid.UUID) (string, error) // 获取商户等级的函数
	enableMetrics  bool
	enableRedis    bool
}

// TierRateLimiterOption 配置选项
type TierRateLimiterOption func(*TierRateLimiter)

// WithTierConfigs 自定义等级配置
func WithTierConfigs(configs map[string]*TierRateLimitConfig) TierRateLimiterOption {
	return func(r *TierRateLimiter) {
		r.tierConfigs = configs
	}
}

// WithMetrics 启用指标收集
func WithMetrics(enable bool) TierRateLimiterOption {
	return func(r *TierRateLimiter) {
		r.enableMetrics = enable
	}
}

// NewTierRateLimiter 创建等级限流器
func NewTierRateLimiter(
	redisClient *redis.Client,
	getTierFunc func(merchantID uuid.UUID) (string, error),
	opts ...TierRateLimiterOption,
) *TierRateLimiter {
	limiter := &TierRateLimiter{
		redisClient:  redisClient,
		tierConfigs:  DefaultTierConfigs,
		getTierFunc:  getTierFunc,
		enableMetrics: true,
		enableRedis:  redisClient != nil,
	}

	for _, opt := range opts {
		opt(limiter)
	}

	return limiter
}

// Middleware 限流中间件
func (r *TierRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文获取商户 ID（由 AuthMiddleware 或 SignatureMiddleware 设置）
		merchantIDValue, exists := c.Get("merchant_id")
		if !exists {
			// 没有商户 ID，跳过限流（可能是公开接口）
			c.Next()
			return
		}

		var merchantID uuid.UUID
		switch v := merchantIDValue.(type) {
		case uuid.UUID:
			merchantID = v
		case string:
			var err error
			merchantID, err = uuid.Parse(v)
			if err != nil {
				logger.Warn("无效的商户 ID 格式", zap.String("merchant_id", v))
				c.Next()
				return
			}
		default:
			logger.Warn("未知的商户 ID 类型", zap.Any("type", fmt.Sprintf("%T", v)))
			c.Next()
			return
		}

		// 获取商户等级
		tier, err := r.getTierFunc(merchantID)
		if err != nil {
			logger.Warn("获取商户等级失败",
				zap.Error(err),
				zap.String("merchant_id", merchantID.String()))
			// 降级：使用 starter 等级
			tier = "starter"
		}

		// 获取限流配置
		config, ok := r.tierConfigs[tier]
		if !ok {
			logger.Warn("未知的商户等级",
				zap.String("tier", tier),
				zap.String("merchant_id", merchantID.String()))
			// 降级：使用 starter 等级
			config = r.tierConfigs["starter"]
		}

		// 执行限流检查
		allowed, reason := r.checkRateLimit(c.Request.Context(), merchantID, config)
		if !allowed {
			// 记录限流事件
			logger.Warn("请求被限流",
				zap.String("merchant_id", merchantID.String()),
				zap.String("tier", tier),
				zap.String("reason", reason),
				zap.String("path", c.Request.URL.Path))

			// 设置响应头
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.RequestsPerSecond))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Second).Unix()))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    "RATE_LIMIT_EXCEEDED",
				"message": fmt.Sprintf("请求频率超过限制: %s", reason),
				"tier":    tier,
			})
			c.Abort()
			return
		}

		// 设置响应头
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.RequestsPerSecond))
		c.Header("X-RateLimit-Tier", tier)

		c.Next()
	}
}

// checkRateLimit 检查限流（多维度）
func (r *TierRateLimiter) checkRateLimit(
	ctx context.Context,
	merchantID uuid.UUID,
	config *TierRateLimitConfig,
) (bool, string) {
	if !r.enableRedis {
		// Redis 不可用，跳过限流
		return true, ""
	}

	now := time.Now()
	merchantKey := merchantID.String()

	// 1. 检查每秒限流（滑动窗口）
	if !r.checkSlidingWindow(ctx, merchantKey, "second", config.RequestsPerSecond, time.Second) {
		return false, fmt.Sprintf("超过每秒 %d 次请求限制", config.RequestsPerSecond)
	}

	// 2. 检查每分钟限流
	if !r.checkSlidingWindow(ctx, merchantKey, "minute", config.RequestsPerMinute, time.Minute) {
		return false, fmt.Sprintf("超过每分钟 %d 次请求限制", config.RequestsPerMinute)
	}

	// 3. 检查每小时限流
	if !r.checkSlidingWindow(ctx, merchantKey, "hour", config.RequestsPerHour, time.Hour) {
		return false, fmt.Sprintf("超过每小时 %d 次请求限制", config.RequestsPerHour)
	}

	// 4. 检查并发请求数（计数器）
	if !r.checkConcurrency(ctx, merchantKey, config.ConcurrentRequests) {
		return false, fmt.Sprintf("超过最大并发 %d 个请求限制", config.ConcurrentRequests)
	}

	// 记录本次请求
	r.recordRequest(ctx, merchantKey, now)

	return true, ""
}

// checkSlidingWindow 滑动窗口算法检查限流
func (r *TierRateLimiter) checkSlidingWindow(
	ctx context.Context,
	merchantKey string,
	window string,
	limit int,
	duration time.Duration,
) bool {
	key := fmt.Sprintf("ratelimit:%s:%s", merchantKey, window)
	now := time.Now()
	windowStart := now.Add(-duration).UnixNano()

	// Lua 脚本：原子操作
	// 1. 删除窗口外的记录
	// 2. 统计窗口内的请求数
	// 3. 如果未超限，添加当前请求
	script := `
		local key = KEYS[1]
		local window_start = ARGV[1]
		local now = ARGV[2]
		local limit = tonumber(ARGV[3])

		-- 删除窗口外的记录
		redis.call('ZREMRANGEBYSCORE', key, '-inf', window_start)

		-- 统计窗口内的请求数
		local count = redis.call('ZCARD', key)

		if count < limit then
			-- 未超限，添加当前请求
			redis.call('ZADD', key, now, now)
			redis.call('EXPIRE', key, 3600)  -- 1小时过期
			return 1
		else
			return 0
		end
	`

	result, err := r.redisClient.Eval(
		ctx,
		script,
		[]string{key},
		windowStart,
		now.UnixNano(),
		limit,
	).Int()

	if err != nil {
		logger.Error("限流检查失败",
			zap.Error(err),
			zap.String("key", key))
		// Redis 错误时，降级为允许通过
		return true
	}

	return result == 1
}

// checkConcurrency 检查并发请求数
func (r *TierRateLimiter) checkConcurrency(
	ctx context.Context,
	merchantKey string,
	limit int,
) bool {
	key := fmt.Sprintf("ratelimit:%s:concurrent", merchantKey)

	// 获取当前并发数
	count, err := r.redisClient.Get(ctx, key).Int()
	if err != nil && err != redis.Nil {
		logger.Error("获取并发数失败",
			zap.Error(err),
			zap.String("key", key))
		return true // 降级
	}

	if count >= limit {
		return false
	}

	// 增加并发数（请求结束后需要调用 releaseConcurrency 减少）
	r.redisClient.Incr(ctx, key)
	r.redisClient.Expire(ctx, key, 60*time.Second) // 60秒过期（防止泄漏）

	return true
}

// releaseConcurrency 释放并发计数（在请求完成后调用）
func (r *TierRateLimiter) ReleaseConcurrency(ctx context.Context, merchantID uuid.UUID) {
	if !r.enableRedis {
		return
	}

	key := fmt.Sprintf("ratelimit:%s:concurrent", merchantID.String())
	r.redisClient.Decr(ctx, key)
}

// recordRequest 记录请求（用于统计）
func (r *TierRateLimiter) recordRequest(ctx context.Context, merchantKey string, now time.Time) {
	// 可选：记录到 Redis 用于后续分析
	statsKey := fmt.Sprintf("ratelimit:stats:%s:%s", merchantKey, now.Format("2006-01-02"))
	r.redisClient.Incr(ctx, statsKey)
	r.redisClient.Expire(ctx, statsKey, 30*24*time.Hour) // 保留 30 天
}

// GetMerchantStats 获取商户请求统计
func (r *TierRateLimiter) GetMerchantStats(
	ctx context.Context,
	merchantID uuid.UUID,
	date time.Time,
) (int64, error) {
	if !r.enableRedis {
		return 0, fmt.Errorf("Redis 未启用")
	}

	key := fmt.Sprintf("ratelimit:stats:%s:%s", merchantID.String(), date.Format("2006-01-02"))
	return r.redisClient.Get(ctx, key).Int64()
}

// MiddlewareWithRelease 限流中间件（自动释放并发计数）
func (r *TierRateLimiter) MiddlewareWithRelease() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 执行限流检查
		r.Middleware()(c)

		if c.IsAborted() {
			return
		}

		// 请求完成后释放并发计数
		defer func() {
			merchantIDValue, exists := c.Get("merchant_id")
			if !exists {
				return
			}

			var merchantID uuid.UUID
			switch v := merchantIDValue.(type) {
			case uuid.UUID:
				merchantID = v
			case string:
				merchantID, _ = uuid.Parse(v)
			default:
				return
			}

			r.ReleaseConcurrency(c.Request.Context(), merchantID)
		}()

		c.Next()
	}
}
