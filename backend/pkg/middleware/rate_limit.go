package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimiter implements token bucket algorithm using Redis
type RateLimiter struct {
	redis  *redis.Client
	limit  int           // max requests
	window time.Duration // time window
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(redis *redis.Client, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		redis:  redis,
		limit:  limit,
		window: window,
	}
}

// RateLimit middleware limits requests based on IP or user ID
func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use IP as key, or user ID if authenticated
		key := fmt.Sprintf("rate_limit:%s", c.ClientIP())

		// Try to get claims for user-based rate limiting
		if claims, err := GetClaims(c); err == nil {
			key = fmt.Sprintf("rate_limit:user:%s", claims.UserID.String())
		}

		ctx := context.Background()

		// Increment counter
		pipe := rl.redis.Pipeline()
		incr := pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, rl.window)
		_, err := pipe.Exec(ctx)

		if err != nil {
			c.JSON(500, gin.H{"error": "rate limit check failed"})
			c.Abort()
			return
		}

		count := incr.Val()
		if count > int64(rl.limit) {
			c.JSON(429, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": rl.window.Seconds(),
			})
			c.Abort()
			return
		}

		// Add rate limit headers
		c.Writer.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.limit))
		c.Writer.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", rl.limit-int(count)))
		c.Writer.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(rl.window).Unix()))

		c.Next()
	}
}
