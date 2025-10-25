package router

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/payment-platform/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// ConfigManager 路由配置管理器
type ConfigManager struct {
	channels      []*ChannelConfig
	mu            sync.RWMutex
	redisClient   *redis.Client
	cacheKey      string
	cacheDuration time.Duration
}

// NewConfigManager 创建配置管理器
func NewConfigManager(redisClient *redis.Client) *ConfigManager {
	return &ConfigManager{
		channels:      make([]*ChannelConfig, 0),
		redisClient:   redisClient,
		cacheKey:      "payment:router:channel_configs",
		cacheDuration: 5 * time.Minute,
	}
}

// LoadChannels 加载渠道配置
func (m *ConfigManager) LoadChannels(ctx context.Context) error {
	// 先尝试从Redis缓存加载
	if m.redisClient != nil {
		cached, err := m.redisClient.Get(ctx, m.cacheKey).Result()
		if err == nil && cached != "" {
			var channels []*ChannelConfig
			if err := json.Unmarshal([]byte(cached), &channels); err == nil {
				m.mu.Lock()
				m.channels = channels
				m.mu.Unlock()
				logger.Info("从Redis缓存加载渠道配置成功", zap.Int("count", len(channels)))
				return nil
			}
		}
	}

	// 缓存未命中，使用默认配置
	defaultChannels := m.getDefaultChannels()

	m.mu.Lock()
	m.channels = defaultChannels
	m.mu.Unlock()

	// 保存到Redis缓存
	if m.redisClient != nil {
		data, _ := json.Marshal(defaultChannels)
		m.redisClient.Set(ctx, m.cacheKey, data, m.cacheDuration)
	}

	logger.Info("加载默认渠道配置", zap.Int("count", len(defaultChannels)))
	return nil
}

// GetChannels 获取渠道配置
func (m *ConfigManager) GetChannels() []*ChannelConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 返回副本，避免并发修改
	result := make([]*ChannelConfig, len(m.channels))
	copy(result, m.channels)
	return result
}

// UpdateChannel 更新渠道配置
func (m *ConfigManager) UpdateChannel(ctx context.Context, channel *ChannelConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 查找并更新
	found := false
	for i, ch := range m.channels {
		if ch.Channel == channel.Channel {
			m.channels[i] = channel
			found = true
			break
		}
	}

	if !found {
		m.channels = append(m.channels, channel)
	}

	// 更新Redis缓存
	if m.redisClient != nil {
		data, _ := json.Marshal(m.channels)
		m.redisClient.Set(ctx, m.cacheKey, data, m.cacheDuration)
	}

	logger.Info("渠道配置已更新", zap.String("channel", channel.Channel))
	return nil
}

// GetChannel 获取指定渠道配置
func (m *ConfigManager) GetChannel(channel string) (*ChannelConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, ch := range m.channels {
		if ch.Channel == channel {
			return ch, nil
		}
	}

	return nil, fmt.Errorf("渠道不存在: %s", channel)
}

// getDefaultChannels 获取默认渠道配置
func (m *ConfigManager) getDefaultChannels() []*ChannelConfig {
	return []*ChannelConfig{
		{
			Channel:             "stripe",
			IsEnabled:           true,
			SupportedCurrencies: []string{"USD", "EUR", "GBP", "JPY", "CNY", "SGD"},
			SupportedCountries:  []string{"US", "EU", "GB", "JP", "SG", "CN"},
			SupportedPayMethods: []string{"card", "wallet"},
			FeeRate:             0.029,  // 2.9%
			MinFee:              30,     // $0.30
			MaxAmount:           999999900, // $999,999.00
			MinAmount:           50,     // $0.50
			SuccessRate:         0.95,   // 95%
			AvgResponseTime:     500,    // 500ms
			Weight:              100,
		},
		{
			Channel:             "paypal",
			IsEnabled:           true,
			SupportedCurrencies: []string{"USD", "EUR", "GBP", "JPY", "CNY"},
			SupportedCountries:  []string{"US", "EU", "GB", "JP", "CN"},
			SupportedPayMethods: []string{"wallet", "card"},
			FeeRate:             0.034,  // 3.4%
			MinFee:              30,     // $0.30
			MaxAmount:           1000000000, // $1,000,000.00
			MinAmount:           100,    // $1.00
			SuccessRate:         0.92,   // 92%
			AvgResponseTime:     800,    // 800ms
			Weight:              80,
		},
		{
			Channel:             "alipay",
			IsEnabled:           true,
			SupportedCurrencies: []string{"CNY", "USD"},
			SupportedCountries:  []string{"CN"},
			SupportedPayMethods: []string{"wallet"},
			FeeRate:             0.006,  // 0.6%
			MinFee:              10,     // ¥0.10
			MaxAmount:           5000000000, // ¥50,000.00
			MinAmount:           1,      // ¥0.01
			SuccessRate:         0.98,   // 98%
			AvgResponseTime:     300,    // 300ms
			Weight:              120,
		},
		{
			Channel:             "wechat",
			IsEnabled:           true,
			SupportedCurrencies: []string{"CNY"},
			SupportedCountries:  []string{"CN"},
			SupportedPayMethods: []string{"wallet"},
			FeeRate:             0.006,  // 0.6%
			MinFee:              10,     // ¥0.10
			MaxAmount:           5000000000, // ¥50,000.00
			MinAmount:           1,      // ¥0.01
			SuccessRate:         0.97,   // 97%
			AvgResponseTime:     400,    // 400ms
			Weight:              110,
		},
		{
			Channel:             "crypto",
			IsEnabled:           false, // 默认禁用
			SupportedCurrencies: []string{"BTC", "ETH", "USDT"},
			SupportedCountries:  []string{}, // 全球支持
			SupportedPayMethods: []string{"crypto"},
			FeeRate:             0.010,  // 1.0%
			MinFee:              0,
			MaxAmount:           10000000000, // $100,000.00
			MinAmount:           100,    // $1.00
			SuccessRate:         0.88,   // 88%
			AvgResponseTime:     2000,   // 2000ms
			Weight:              50,
		},
	}
}

// RefreshCache 刷新缓存
func (m *ConfigManager) RefreshCache(ctx context.Context) error {
	m.mu.RLock()
	channels := make([]*ChannelConfig, len(m.channels))
	copy(channels, m.channels)
	m.mu.RUnlock()

	if m.redisClient != nil {
		data, _ := json.Marshal(channels)
		return m.redisClient.Set(ctx, m.cacheKey, data, m.cacheDuration).Err()
	}

	return nil
}

// UpdateChannelStatus 更新渠道启用状态
func (m *ConfigManager) UpdateChannelStatus(ctx context.Context, channel string, enabled bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, ch := range m.channels {
		if ch.Channel == channel {
			m.channels[i].IsEnabled = enabled

			// 更新Redis缓存
			if m.redisClient != nil {
				data, _ := json.Marshal(m.channels)
				m.redisClient.Set(ctx, m.cacheKey, data, m.cacheDuration)
			}

			logger.Info("渠道状态已更新",
				zap.String("channel", channel),
				zap.Bool("enabled", enabled))
			return nil
		}
	}

	return fmt.Errorf("渠道不存在: %s", channel)
}

// UpdateChannelMetrics 更新渠道指标（成功率、响应时间）
func (m *ConfigManager) UpdateChannelMetrics(ctx context.Context, channel string, successRate float64, avgResponseTime int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, ch := range m.channels {
		if ch.Channel == channel {
			m.channels[i].SuccessRate = successRate
			m.channels[i].AvgResponseTime = avgResponseTime

			// 更新Redis缓存
			if m.redisClient != nil {
				data, _ := json.Marshal(m.channels)
				m.redisClient.Set(ctx, m.cacheKey, data, m.cacheDuration)
			}

			logger.Debug("渠道指标已更新",
				zap.String("channel", channel),
				zap.Float64("success_rate", successRate),
				zap.Int64("avg_response_time", avgResponseTime))
			return nil
		}
	}

	return fmt.Errorf("渠道不存在: %s", channel)
}
