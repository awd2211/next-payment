package httpclient

import (
	"fmt"
	"time"

	"github.com/sony/gobreaker"
)

// BreakerConfig 熔断器配置
type BreakerConfig struct {
	Name          string        // 熔断器名称
	MaxRequests   uint32        // 半开状态允许的最大请求数
	Interval      time.Duration // 统计时间窗口
	Timeout       time.Duration // 熔断器打开后多久尝试半开
	ReadyToTrip   func(counts gobreaker.Counts) bool // 判断是否应该熔断
	OnStateChange func(name string, from gobreaker.State, to gobreaker.State) // 状态变化回调
}

// DefaultBreakerConfig 默认熔断器配置
func DefaultBreakerConfig(name string) *BreakerConfig {
	return &BreakerConfig{
		Name:        name,
		MaxRequests: 3,                // 半开状态允许3个请求
		Interval:    time.Minute,      // 1分钟统计窗口
		Timeout:     30 * time.Second, // 30秒后尝试半开
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 5 && failureRatio >= 0.6 // 5次请求中60%失败则熔断
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			fmt.Printf("[Breaker] %s: %s -> %s\n", name, from.String(), to.String())
		},
	}
}

// CreateBreaker 创建熔断器
func CreateBreaker(config *BreakerConfig) *gobreaker.CircuitBreaker {
	if config == nil {
		config = DefaultBreakerConfig("default")
	}

	settings := gobreaker.Settings{
		Name:          config.Name,
		MaxRequests:   config.MaxRequests,
		Interval:      config.Interval,
		Timeout:       config.Timeout,
		ReadyToTrip:   config.ReadyToTrip,
		OnStateChange: config.OnStateChange,
	}

	return gobreaker.NewCircuitBreaker(settings)
}

// BreakerClient 带熔断器的HTTP客户端
type BreakerClient struct {
	*Client
	breaker *gobreaker.CircuitBreaker
}

// NewBreakerClient 创建带熔断器的HTTP客户端
func NewBreakerClient(config *Config, breakerConfig *BreakerConfig) *BreakerClient {
	client := NewClient(config)
	breaker := CreateBreaker(breakerConfig)

	return &BreakerClient{
		Client:  client,
		breaker: breaker,
	}
}

// Do 通过熔断器发送HTTP请求
func (c *BreakerClient) Do(req *Request) (*Response, error) {
	result, err := c.breaker.Execute(func() (interface{}, error) {
		return c.Client.Do(req)
	})

	if err != nil {
		return nil, fmt.Errorf("熔断器错误: %w", err)
	}

	return result.(*Response), nil
}

// State 获取熔断器状态
func (c *BreakerClient) State() gobreaker.State {
	return c.breaker.State()
}

// Name 获取熔断器名称
func (c *BreakerClient) Name() string {
	return c.breaker.Name()
}

// Counts 获取熔断器统计信息
func (c *BreakerClient) Counts() gobreaker.Counts {
	return c.breaker.Counts()
}
