package retry

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Retrier 重试器
type Retrier struct {
	config *Config
}

// Config 重试配置
type Config struct {
	MaxAttempts  int           // 最大重试次数
	InitialDelay time.Duration // 初始延迟
	MaxDelay     time.Duration // 最大延迟
	Multiplier   float64       // 延迟倍数（指数退避）
	MaxJitter    time.Duration // 最大抖动时间
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		MaxJitter:    1 * time.Second,
	}
}

// NewRetrier 创建重试器
func NewRetrier(config *Config) *Retrier {
	if config == nil {
		config = DefaultConfig()
	}
	return &Retrier{config: config}
}

// RetryFunc 可重试的函数类型
type RetryFunc func() error

// RetryCondition 重试条件函数
type RetryCondition func(error) bool

// Do 执行重试
func (r *Retrier) Do(fn RetryFunc) error {
	return r.DoWithCondition(fn, nil)
}

// DoWithCondition 根据条件执行重试
func (r *Retrier) DoWithCondition(fn RetryFunc, condition RetryCondition) error {
	var lastErr error
	delay := r.config.InitialDelay

	for attempt := 1; attempt <= r.config.MaxAttempts; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// 如果是最后一次尝试，不再重试
		if attempt == r.config.MaxAttempts {
			break
		}

		// 检查重试条件
		if condition != nil && !condition(err) {
			return fmt.Errorf("重试条件不满足: %w", err)
		}

		// 等待后重试
		time.Sleep(delay)

		// 指数退避
		delay = time.Duration(float64(delay) * r.config.Multiplier)
		if delay > r.config.MaxDelay {
			delay = r.config.MaxDelay
		}

		// 添加抖动
		if r.config.MaxJitter > 0 {
			jitter := time.Duration(time.Now().UnixNano() % int64(r.config.MaxJitter))
			delay += jitter
		}
	}

	return fmt.Errorf("重试%d次后仍然失败: %w", r.config.MaxAttempts, lastErr)
}

// DoWithContext 使用context执行重试
func (r *Retrier) DoWithContext(ctx context.Context, fn RetryFunc) error {
	return r.DoWithContextAndCondition(ctx, fn, nil)
}

// DoWithContextAndCondition 使用context和条件执行重试
func (r *Retrier) DoWithContextAndCondition(ctx context.Context, fn RetryFunc, condition RetryCondition) error {
	var lastErr error
	delay := r.config.InitialDelay

	for attempt := 1; attempt <= r.config.MaxAttempts; attempt++ {
		// 检查context是否已取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// 如果是最后一次尝试，不再重试
		if attempt == r.config.MaxAttempts {
			break
		}

		// 检查重试条件
		if condition != nil && !condition(err) {
			return fmt.Errorf("重试条件不满足: %w", err)
		}

		// 等待后重试（可被context取消）
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}

		// 指数退避
		delay = time.Duration(float64(delay) * r.config.Multiplier)
		if delay > r.config.MaxDelay {
			delay = r.config.MaxDelay
		}

		// 添加抖动
		if r.config.MaxJitter > 0 {
			jitter := time.Duration(time.Now().UnixNano() % int64(r.config.MaxJitter))
			delay += jitter
		}
	}

	return fmt.Errorf("重试%d次后仍然失败: %w", r.config.MaxAttempts, lastErr)
}

// 便捷函数

// Do 使用默认配置执行重试
func Do(fn RetryFunc) error {
	return NewRetrier(nil).Do(fn)
}

// DoWithContext 使用默认配置和context执行重试
func DoWithContext(ctx context.Context, fn RetryFunc) error {
	return NewRetrier(nil).DoWithContext(ctx, fn)
}

// 常用重试条件

// IsNetworkError 判断是否为网络错误
func IsNetworkError(err error) bool {
	// 这里可以根据具体的错误类型判断
	// 简单实现：所有错误都可以重试
	return err != nil
}

// IsTemporaryError 判断是否为临时错误
func IsTemporaryError(err error) bool {
	var temp interface{ Temporary() bool }
	if errors.As(err, &temp) {
		return temp.Temporary()
	}
	return false
}
