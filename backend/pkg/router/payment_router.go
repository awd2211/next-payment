package router

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/payment-platform/pkg/logger"
	"go.uber.org/zap"
)

// PaymentRouter 支付路由器
type PaymentRouter struct {
	strategies []RoutingStrategy
}

// RoutingStrategy 路由策略接口
type RoutingStrategy interface {
	// SelectChannel 选择最优支付渠道
	SelectChannel(ctx context.Context, req *RoutingRequest) (*RoutingResult, error)

	// Name 策略名称
	Name() string

	// Priority 优先级（数字越大优先级越高）
	Priority() int
}

// RoutingRequest 路由请求
type RoutingRequest struct {
	MerchantID    string                 `json:"merchant_id"`
	Amount        int64                  `json:"amount"`
	Currency      string                 `json:"currency"`
	Country       string                 `json:"country"`        // 国家代码
	PayMethod     string                 `json:"pay_method"`     // 支付方式：card, wallet, bank_transfer
	PreferChannel string                 `json:"prefer_channel"` // 优先渠道（可选）
	Extra         map[string]interface{} `json:"extra"`
}

// RoutingResult 路由结果
type RoutingResult struct {
	Channel      string  `json:"channel"`       // 选择的渠道
	Reason       string  `json:"reason"`        // 选择原因
	EstimatedFee int64   `json:"estimated_fee"` // 估算手续费
	FeeRate      float64 `json:"fee_rate"`      // 费率
	Priority     int     `json:"priority"`      // 优先级评分
}

// ChannelConfig 渠道配置
type ChannelConfig struct {
	Channel        string   `json:"channel"`
	IsEnabled      bool     `json:"is_enabled"`
	SupportedCurrencies []string `json:"supported_currencies"`
	SupportedCountries  []string `json:"supported_countries"`
	SupportedPayMethods []string `json:"supported_pay_methods"`
	FeeRate        float64  `json:"fee_rate"`        // 基础费率
	MinFee         int64    `json:"min_fee"`         // 最低手续费
	MaxAmount      int64    `json:"max_amount"`      // 最大交易金额
	MinAmount      int64    `json:"min_amount"`      // 最小交易金额
	SuccessRate    float64  `json:"success_rate"`    // 历史成功率
	AvgResponseTime int64   `json:"avg_response_time"` // 平均响应时间（毫秒）
	Weight         int      `json:"weight"`          // 权重（用于负载均衡）
}

// NewPaymentRouter 创建支付路由器
func NewPaymentRouter() *PaymentRouter {
	return &PaymentRouter{
		strategies: make([]RoutingStrategy, 0),
	}
}

// RegisterStrategy 注册路由策略
func (r *PaymentRouter) RegisterStrategy(strategy RoutingStrategy) {
	r.strategies = append(r.strategies, strategy)

	// 按优先级排序（冒泡排序）
	for i := 0; i < len(r.strategies); i++ {
		for j := i + 1; j < len(r.strategies); j++ {
			if r.strategies[i].Priority() < r.strategies[j].Priority() {
				r.strategies[i], r.strategies[j] = r.strategies[j], r.strategies[i]
			}
		}
	}

	logger.Info("路由策略已注册",
		zap.String("strategy", strategy.Name()),
		zap.Int("priority", strategy.Priority()))
}

// Route 执行路由选择
func (r *PaymentRouter) Route(ctx context.Context, req *RoutingRequest) (*RoutingResult, error) {
	// 如果指定了优先渠道，先尝试使用
	if req.PreferChannel != "" {
		logger.Info("使用指定的优先渠道",
			zap.String("channel", req.PreferChannel),
			zap.String("merchant_id", req.MerchantID))

		return &RoutingResult{
			Channel:  req.PreferChannel,
			Reason:   "商户指定优先渠道",
			Priority: 100,
		}, nil
	}

	// 按优先级依次尝试各个策略
	for _, strategy := range r.strategies {
		result, err := strategy.SelectChannel(ctx, req)
		if err != nil {
			logger.Warn("路由策略执行失败",
				zap.String("strategy", strategy.Name()),
				zap.Error(err))
			continue
		}

		if result != nil && result.Channel != "" {
			logger.Info("路由策略选择成功",
				zap.String("strategy", strategy.Name()),
				zap.String("channel", result.Channel),
				zap.String("reason", result.Reason))
			return result, nil
		}
	}

	return nil, fmt.Errorf("没有可用的支付渠道")
}

// =============================================================================
// 内置路由策略
// =============================================================================

// CostOptimizationStrategy 成本优化策略（选择费率最低的渠道）
type CostOptimizationStrategy struct {
	channels []*ChannelConfig
}

// NewCostOptimizationStrategy 创建成本优化策略
func NewCostOptimizationStrategy(channels []*ChannelConfig) *CostOptimizationStrategy {
	return &CostOptimizationStrategy{
		channels: channels,
	}
}

func (s *CostOptimizationStrategy) Name() string {
	return "CostOptimization"
}

func (s *CostOptimizationStrategy) Priority() int {
	return 50 // 中等优先级
}

func (s *CostOptimizationStrategy) SelectChannel(ctx context.Context, req *RoutingRequest) (*RoutingResult, error) {
	var bestChannel *ChannelConfig
	var lowestFee int64 = -1

	for _, ch := range s.channels {
		if !s.isChannelSupported(ch, req) {
			continue
		}

		// 计算手续费
		fee := int64(float64(req.Amount) * ch.FeeRate)
		if fee < ch.MinFee {
			fee = ch.MinFee
		}

		if lowestFee == -1 || fee < lowestFee {
			lowestFee = fee
			bestChannel = ch
		}
	}

	if bestChannel == nil {
		return nil, fmt.Errorf("没有支持的渠道")
	}

	return &RoutingResult{
		Channel:      bestChannel.Channel,
		Reason:       fmt.Sprintf("成本最优（费率: %.2f%%）", bestChannel.FeeRate*100),
		EstimatedFee: lowestFee,
		FeeRate:      bestChannel.FeeRate,
		Priority:     70,
	}, nil
}

func (s *CostOptimizationStrategy) isChannelSupported(ch *ChannelConfig, req *RoutingRequest) bool {
	if !ch.IsEnabled {
		return false
	}

	// 检查金额范围
	if req.Amount < ch.MinAmount || req.Amount > ch.MaxAmount {
		return false
	}

	// 检查币种
	if !contains(ch.SupportedCurrencies, req.Currency) {
		return false
	}

	// 检查国家
	if req.Country != "" && len(ch.SupportedCountries) > 0 && !contains(ch.SupportedCountries, req.Country) {
		return false
	}

	// 检查支付方式
	if req.PayMethod != "" && len(ch.SupportedPayMethods) > 0 && !contains(ch.SupportedPayMethods, req.PayMethod) {
		return false
	}

	return true
}

// SuccessRateStrategy 成功率优先策略（选择成功率最高的渠道）
type SuccessRateStrategy struct {
	channels []*ChannelConfig
}

// NewSuccessRateStrategy 创建成功率优先策略
func NewSuccessRateStrategy(channels []*ChannelConfig) *SuccessRateStrategy {
	return &SuccessRateStrategy{
		channels: channels,
	}
}

func (s *SuccessRateStrategy) Name() string {
	return "SuccessRate"
}

func (s *SuccessRateStrategy) Priority() int {
	return 80 // 高优先级
}

func (s *SuccessRateStrategy) SelectChannel(ctx context.Context, req *RoutingRequest) (*RoutingResult, error) {
	var bestChannel *ChannelConfig
	var highestRate float64 = -1

	for _, ch := range s.channels {
		if !s.isChannelSupported(ch, req) {
			continue
		}

		if ch.SuccessRate > highestRate {
			highestRate = ch.SuccessRate
			bestChannel = ch
		}
	}

	if bestChannel == nil {
		return nil, fmt.Errorf("没有支持的渠道")
	}

	fee := int64(float64(req.Amount) * bestChannel.FeeRate)
	if fee < bestChannel.MinFee {
		fee = bestChannel.MinFee
	}

	return &RoutingResult{
		Channel:      bestChannel.Channel,
		Reason:       fmt.Sprintf("成功率最高（%.1f%%）", bestChannel.SuccessRate*100),
		EstimatedFee: fee,
		FeeRate:      bestChannel.FeeRate,
		Priority:     80,
	}, nil
}

func (s *SuccessRateStrategy) isChannelSupported(ch *ChannelConfig, req *RoutingRequest) bool {
	return (&CostOptimizationStrategy{}).isChannelSupported(ch, req)
}

// LoadBalanceStrategy 负载均衡策略（基于权重随机选择）
type LoadBalanceStrategy struct {
	channels []*ChannelConfig
	rand     *rand.Rand
}

// NewLoadBalanceStrategy 创建负载均衡策略
func NewLoadBalanceStrategy(channels []*ChannelConfig) *LoadBalanceStrategy {
	return &LoadBalanceStrategy{
		channels: channels,
		rand:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *LoadBalanceStrategy) Name() string {
	return "LoadBalance"
}

func (s *LoadBalanceStrategy) Priority() int {
	return 30 // 低优先级（作为兜底策略）
}

func (s *LoadBalanceStrategy) SelectChannel(ctx context.Context, req *RoutingRequest) (*RoutingResult, error) {
	var eligibleChannels []*ChannelConfig
	var totalWeight int

	for _, ch := range s.channels {
		if !s.isChannelSupported(ch, req) {
			continue
		}

		eligibleChannels = append(eligibleChannels, ch)
		totalWeight += ch.Weight
	}

	if len(eligibleChannels) == 0 {
		return nil, fmt.Errorf("没有支持的渠道")
	}

	// 加权随机选择
	randomWeight := s.rand.Intn(totalWeight)
	currentWeight := 0

	var selectedChannel *ChannelConfig
	for _, ch := range eligibleChannels {
		currentWeight += ch.Weight
		if randomWeight < currentWeight {
			selectedChannel = ch
			break
		}
	}

	if selectedChannel == nil {
		selectedChannel = eligibleChannels[0]
	}

	fee := int64(float64(req.Amount) * selectedChannel.FeeRate)
	if fee < selectedChannel.MinFee {
		fee = selectedChannel.MinFee
	}

	return &RoutingResult{
		Channel:      selectedChannel.Channel,
		Reason:       fmt.Sprintf("负载均衡（权重: %d）", selectedChannel.Weight),
		EstimatedFee: fee,
		FeeRate:      selectedChannel.FeeRate,
		Priority:     30,
	}, nil
}

func (s *LoadBalanceStrategy) isChannelSupported(ch *ChannelConfig, req *RoutingRequest) bool {
	return (&CostOptimizationStrategy{}).isChannelSupported(ch, req)
}

// GeographicStrategy 地域优化策略（根据国家选择本地化渠道）
type GeographicStrategy struct {
	channels []*ChannelConfig
}

// NewGeographicStrategy 创建地域优化策略
func NewGeographicStrategy(channels []*ChannelConfig) *GeographicStrategy {
	return &GeographicStrategy{
		channels: channels,
	}
}

func (s *GeographicStrategy) Name() string {
	return "Geographic"
}

func (s *GeographicStrategy) Priority() int {
	return 90 // 最高优先级
}

func (s *GeographicStrategy) SelectChannel(ctx context.Context, req *RoutingRequest) (*RoutingResult, error) {
	if req.Country == "" {
		return nil, fmt.Errorf("未指定国家代码")
	}

	// 地域到渠道的映射
	countryToChannel := map[string]string{
		"CN": "alipay",  // 中国 → 支付宝
		"US": "stripe",  // 美国 → Stripe
		"EU": "stripe",  // 欧洲 → Stripe
		"JP": "stripe",  // 日本 → Stripe
		"SG": "stripe",  // 新加坡 → Stripe
	}

	preferredChannel := countryToChannel[req.Country]
	if preferredChannel == "" {
		return nil, fmt.Errorf("该地区没有推荐渠道")
	}

	// 检查推荐渠道是否可用
	for _, ch := range s.channels {
		if ch.Channel == preferredChannel && s.isChannelSupported(ch, req) {
			fee := int64(float64(req.Amount) * ch.FeeRate)
			if fee < ch.MinFee {
				fee = ch.MinFee
			}

			return &RoutingResult{
				Channel:      ch.Channel,
				Reason:       fmt.Sprintf("地域优化（%s 本地化渠道）", req.Country),
				EstimatedFee: fee,
				FeeRate:      ch.FeeRate,
				Priority:     90,
			}, nil
		}
	}

	return nil, fmt.Errorf("推荐渠道不可用")
}

func (s *GeographicStrategy) isChannelSupported(ch *ChannelConfig, req *RoutingRequest) bool {
	return (&CostOptimizationStrategy{}).isChannelSupported(ch, req)
}

// 辅助函数
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
