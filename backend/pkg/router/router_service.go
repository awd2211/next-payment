package router

import (
	"context"
	"fmt"

	"github.com/payment-platform/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RouterService 路由服务
type RouterService struct {
	router        *PaymentRouter
	configManager *ConfigManager
}

// NewRouterService 创建路由服务
func NewRouterService(redisClient *redis.Client) *RouterService {
	configManager := NewConfigManager(redisClient)
	router := NewPaymentRouter()

	return &RouterService{
		router:        router,
		configManager: configManager,
	}
}

// Initialize 初始化路由服务
func (s *RouterService) Initialize(ctx context.Context, strategyMode string) error {
	// 加载渠道配置
	if err := s.configManager.LoadChannels(ctx); err != nil {
		return fmt.Errorf("加载渠道配置失败: %w", err)
	}

	channels := s.configManager.GetChannels()

	// 根据模式注册策略
	switch strategyMode {
	case "cost":
		// 成本优先模式
		s.router.RegisterStrategy(NewCostOptimizationStrategy(channels))
		s.router.RegisterStrategy(NewLoadBalanceStrategy(channels))
		logger.Info("路由模式: 成本优先")

	case "success":
		// 成功率优先模式
		s.router.RegisterStrategy(NewSuccessRateStrategy(channels))
		s.router.RegisterStrategy(NewCostOptimizationStrategy(channels))
		s.router.RegisterStrategy(NewLoadBalanceStrategy(channels))
		logger.Info("路由模式: 成功率优先")

	case "geographic":
		// 地域优化模式
		s.router.RegisterStrategy(NewGeographicStrategy(channels))
		s.router.RegisterStrategy(NewSuccessRateStrategy(channels))
		s.router.RegisterStrategy(NewCostOptimizationStrategy(channels))
		s.router.RegisterStrategy(NewLoadBalanceStrategy(channels))
		logger.Info("路由模式: 地域优化")

	case "balanced":
		fallthrough
	default:
		// 平衡模式（默认）
		s.router.RegisterStrategy(NewGeographicStrategy(channels))
		s.router.RegisterStrategy(NewSuccessRateStrategy(channels))
		s.router.RegisterStrategy(NewCostOptimizationStrategy(channels))
		s.router.RegisterStrategy(NewLoadBalanceStrategy(channels))
		logger.Info("路由模式: 平衡模式（默认）")
	}

	logger.Info("路由服务初始化完成", zap.Int("channels", len(channels)))
	return nil
}

// SelectChannel 选择支付渠道
func (s *RouterService) SelectChannel(ctx context.Context, req *RoutingRequest) (*RoutingResult, error) {
	return s.router.Route(ctx, req)
}

// GetChannelConfig 获取渠道配置
func (s *RouterService) GetChannelConfig(channel string) (*ChannelConfig, error) {
	return s.configManager.GetChannel(channel)
}

// GetAllChannels 获取所有渠道配置
func (s *RouterService) GetAllChannels() []*ChannelConfig {
	return s.configManager.GetChannels()
}

// UpdateChannelStatus 更新渠道状态
func (s *RouterService) UpdateChannelStatus(ctx context.Context, channel string, enabled bool) error {
	return s.configManager.UpdateChannelStatus(ctx, channel, enabled)
}

// UpdateChannelMetrics 更新渠道指标
func (s *RouterService) UpdateChannelMetrics(ctx context.Context, channel string, successRate float64, avgResponseTime int64) error {
	return s.configManager.UpdateChannelMetrics(ctx, channel, successRate, avgResponseTime)
}

// ReloadChannels 重新加载渠道配置
func (s *RouterService) ReloadChannels(ctx context.Context) error {
	if err := s.configManager.LoadChannels(ctx); err != nil {
		return err
	}

	// 重新注册策略
	channels := s.configManager.GetChannels()
	s.router = NewPaymentRouter()

	s.router.RegisterStrategy(NewGeographicStrategy(channels))
	s.router.RegisterStrategy(NewSuccessRateStrategy(channels))
	s.router.RegisterStrategy(NewCostOptimizationStrategy(channels))
	s.router.RegisterStrategy(NewLoadBalanceStrategy(channels))

	logger.Info("渠道配置已重新加载", zap.Int("channels", len(channels)))
	return nil
}

// EstimateFee 估算手续费
func (s *RouterService) EstimateFee(ctx context.Context, channel string, amount int64) (int64, error) {
	config, err := s.configManager.GetChannel(channel)
	if err != nil {
		return 0, err
	}

	fee := int64(float64(amount) * config.FeeRate)
	if fee < config.MinFee {
		fee = config.MinFee
	}

	return fee, nil
}

// GetChannelsByCountry 根据国家获取支持的渠道
func (s *RouterService) GetChannelsByCountry(country string) []*ChannelConfig {
	allChannels := s.configManager.GetChannels()
	var result []*ChannelConfig

	for _, ch := range allChannels {
		if ch.IsEnabled && (len(ch.SupportedCountries) == 0 || contains(ch.SupportedCountries, country)) {
			result = append(result, ch)
		}
	}

	return result
}

// GetChannelsByCurrency 根据币种获取支持的渠道
func (s *RouterService) GetChannelsByCurrency(currency string) []*ChannelConfig {
	allChannels := s.configManager.GetChannels()
	var result []*ChannelConfig

	for _, ch := range allChannels {
		if ch.IsEnabled && contains(ch.SupportedCurrencies, currency) {
			result = append(result, ch)
		}
	}

	return result
}
