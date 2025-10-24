package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/payment-platform/pkg/httpclient"
	"github.com/payment-platform/pkg/logger"
	"github.com/redis/go-redis/v9"
	"github.com/sony/gobreaker"
	"go.uber.org/zap"
	"payment-platform/channel-adapter/internal/model"
	"payment-platform/channel-adapter/internal/repository"
)

// ExchangeRateClient 汇率API客户端
type ExchangeRateClient struct {
	baseURL string
	breaker *httpclient.BreakerClient
	redis   *redis.Client
	repo    repository.ExchangeRateRepository
	cacheTTL time.Duration
}

// ExchangeRateResponse 汇率API响应
type ExchangeRateResponse struct {
	Result             string             `json:"result"`
	Documentation      string             `json:"documentation"`
	TermsOfUse         string             `json:"terms_of_use"`
	TimeLastUpdateUnix int64              `json:"time_last_update_unix"`
	TimeLastUpdateUTC  string             `json:"time_last_update_utc"`
	TimeNextUpdateUnix int64              `json:"time_next_update_unix"`
	TimeNextUpdateUTC  string             `json:"time_next_update_utc"`
	BaseCode           string             `json:"base_code"`
	ConversionRates    map[string]float64 `json:"conversion_rates"`
}

// NewExchangeRateClient 创建汇率API客户端（带熔断器）
// 使用 exchangerate-api.com 免费版（1500次/月）
func NewExchangeRateClient(redis *redis.Client, repo repository.ExchangeRateRepository, cacheTTL time.Duration) *ExchangeRateClient {
	// 创建 httpclient 配置
	config := &httpclient.Config{
		Timeout:    5 * time.Second, // 外部API使用较短超时
		MaxRetries: 2,               // 外部API减少重试次数
		RetryDelay: 500 * time.Millisecond,
	}

	// 创建熔断器配置（外部API更宽容的熔断策略）
	breakerConfig := httpclient.DefaultBreakerConfig("exchangerate-api")
	breakerConfig.ReadyToTrip = func(counts gobreaker.Counts) bool {
		// 外部API: 10次请求中80%失败才熔断（更宽容）
		failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
		return counts.Requests >= 10 && failureRatio >= 0.8
	}

	return &ExchangeRateClient{
		baseURL:  "https://api.exchangerate-api.com/v4/latest",
		breaker:  httpclient.NewBreakerClient(config, breakerConfig),
		redis:    redis,
		repo:     repo,
		cacheTTL: cacheTTL,
	}
}

// GetRate 获取汇率（带缓存）
// from: 源货币代码（如 "USD"）
// to: 目标货币代码（如 "EUR"）
// 返回：汇率（1 from = rate to）
func (c *ExchangeRateClient) GetRate(ctx context.Context, from, to string) (float64, error) {
	// 相同货币，汇率为1
	if from == to {
		return 1.0, nil
	}

	// 1. 尝试从缓存读取
	cacheKey := fmt.Sprintf("exchange_rate:%s:%s", from, to)
	cached, err := c.redis.Get(ctx, cacheKey).Result()
	if err == nil && cached != "" {
		var rate float64
		if err := json.Unmarshal([]byte(cached), &rate); err == nil {
			logger.Debug("汇率缓存命中",
				zap.String("from", from),
				zap.String("to", to),
				zap.Float64("rate", rate))
			return rate, nil
		}
	}

	// 2. 调用汇率API获取最新数据
	rates, err := c.fetchRates(ctx, from)
	if err != nil {
		// API调用失败，尝试获取备用汇率
		logger.Warn("汇率API调用失败，使用备用汇率",
			zap.String("from", from),
			zap.String("to", to),
			zap.Error(err))
		return c.getFallbackRate(from, to), nil
	}

	// 3. 从响应中提取目标货币汇率
	rate, ok := rates[to]
	if !ok {
		return 0, fmt.Errorf("不支持的货币: %s", to)
	}

	// 4. 写入缓存
	if data, err := json.Marshal(rate); err == nil {
		c.redis.Set(ctx, cacheKey, string(data), c.cacheTTL)
	}

	logger.Info("汇率查询成功",
		zap.String("from", from),
		zap.String("to", to),
		zap.Float64("rate", rate))

	return rate, nil
}

// GetRates 批量获取汇率
// baseCurrency: 基准货币（如 "USD"）
// 返回：所有货币相对于基准货币的汇率
func (c *ExchangeRateClient) GetRates(ctx context.Context, baseCurrency string) (map[string]float64, error) {
	// 尝试从缓存读取
	cacheKey := fmt.Sprintf("exchange_rates:%s:all", baseCurrency)
	cached, err := c.redis.Get(ctx, cacheKey).Result()
	if err == nil && cached != "" {
		var rates map[string]float64
		if err := json.Unmarshal([]byte(cached), &rates); err == nil {
			logger.Debug("批量汇率缓存命中", zap.String("base", baseCurrency))
			return rates, nil
		}
	}

	// 调用API获取
	rates, err := c.fetchRates(ctx, baseCurrency)
	if err != nil {
		return nil, err
	}

	// 写入缓存
	if data, err := json.Marshal(rates); err == nil {
		c.redis.Set(ctx, cacheKey, string(data), c.cacheTTL)
	}

	// 保存历史快照到数据库（如果配置了repository）
	if c.repo != nil {
		snapshot := &model.ExchangeRateSnapshot{
			BaseCurrency: baseCurrency,
			Rates:        rates,
			Source:       "exchangerate-api",
			SnapshotTime: time.Now(),
		}
		if err := c.repo.SaveSnapshot(ctx, snapshot); err != nil {
			logger.Warn("保存汇率快照失败",
				zap.String("base", baseCurrency),
				zap.Error(err))
			// 不影响正常返回
		}
	}

	return rates, nil
}

// fetchRates 从API获取汇率数据（使用熔断器）
func (c *ExchangeRateClient) fetchRates(ctx context.Context, baseCurrency string) (map[string]float64, error) {
	url := fmt.Sprintf("%s/%s", c.baseURL, baseCurrency)

	// 创建请求
	req := &httpclient.Request{
		Method: "GET",
		URL:    url,
		Ctx:    ctx,
	}

	// 通过熔断器发送请求
	resp, err := c.breaker.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API调用失败: %w", err)
	}

	// 解析响应
	var apiResp ExchangeRateResponse
	if err := json.Unmarshal(resp.Body, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if apiResp.Result != "success" {
		return nil, fmt.Errorf("API返回失败: %s", apiResp.Result)
	}

	return apiResp.ConversionRates, nil
}

// getFallbackRate 获取备用汇率（当API失败时使用）
func (c *ExchangeRateClient) getFallbackRate(from, to string) float64 {
	// 常用货币的近似汇率（用于降级）
	// 实际生产环境应该从数据库加载最近一次成功的汇率
	fallbackRates := map[string]map[string]float64{
		"USD": {
			"EUR": 0.92,
			"GBP": 0.79,
			"CNY": 7.24,
			"JPY": 149.50,
			"KRW": 1320.00,
			"HKD": 7.82,
			"SGD": 1.35,
			"AUD": 1.52,
			"CAD": 1.36,
			"CHF": 0.88,
			"INR": 83.10,
			"BRL": 4.98,
		},
		"EUR": {
			"USD": 1.09,
			"GBP": 0.86,
			"CNY": 7.87,
		},
		"CNY": {
			"USD": 0.14,
			"EUR": 0.13,
		},
	}

	if rates, ok := fallbackRates[from]; ok {
		if rate, ok := rates[to]; ok {
			logger.Warn("使用备用汇率",
				zap.String("from", from),
				zap.String("to", to),
				zap.Float64("rate", rate))
			return rate
		}
	}

	// 如果没有备用汇率，返回1（需要人工处理）
	logger.Error("无法获取汇率，返回默认值1.0",
		zap.String("from", from),
		zap.String("to", to))
	return 1.0
}

// Convert 货币转换
// amount: 金额（以最小单位，如分）
// from: 源货币
// to: 目标货币
// 返回：转换后的金额（以最小单位）
func (c *ExchangeRateClient) Convert(ctx context.Context, amount int64, from, to string) (int64, error) {
	rate, err := c.GetRate(ctx, from, to)
	if err != nil {
		return 0, err
	}

	// 转换计算（保持精度）
	result := float64(amount) * rate

	return int64(result), nil
}

// SupportedCurrencies 返回支持的货币列表
func (c *ExchangeRateClient) SupportedCurrencies() []string {
	return []string{
		"USD", "EUR", "GBP", "CNY", "JPY", "KRW", "HKD", "SGD",
		"AUD", "CAD", "CHF", "INR", "BRL", "MXN", "RUB", "ZAR",
		"THB", "IDR", "MYR", "PHP", "VND", "TRY", "AED", "SAR",
		"ILS", "PLN", "CZK", "HUF", "RON", "DKK", "SEK", "NOK",
		"TWD", "NZD", "ARS", "CLP", "COP", "PEN",
	}
}

// PreloadRates 预加载常用货币对的汇率到缓存
// 这个方法可以在服务启动时或定期调用，以确保缓存中有最新的汇率数据
func (c *ExchangeRateClient) PreloadRates(ctx context.Context) error {
	// 定义需要预加载的基准货币
	baseCurrencies := []string{"USD", "EUR", "CNY", "GBP", "JPY"}

	successCount := 0
	failureCount := 0

	for _, base := range baseCurrencies {
		rates, err := c.GetRates(ctx, base)
		if err != nil {
			logger.Error("预加载汇率失败",
				zap.String("base", base),
				zap.Error(err))
			failureCount++
			continue
		}

		// 统计成功缓存的货币对数量
		successCount += len(rates)
		logger.Info("预加载汇率成功",
			zap.String("base", base),
			zap.Int("pairs", len(rates)))
	}

	logger.Info("汇率预加载完成",
		zap.Int("success_pairs", successCount),
		zap.Int("failed_bases", failureCount))

	if failureCount == len(baseCurrencies) {
		return fmt.Errorf("所有基准货币汇率获取失败")
	}

	return nil
}

// StartPeriodicUpdate 启动定期更新任务
// interval: 更新间隔（建议 1-6 小时）
func (c *ExchangeRateClient) StartPeriodicUpdate(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)

	// 启动时立即执行一次
	if err := c.PreloadRates(ctx); err != nil {
		logger.Error("初始汇率预加载失败", zap.Error(err))
	}

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				logger.Info("开始定期更新汇率...")
				if err := c.PreloadRates(ctx); err != nil {
					logger.Error("定期汇率更新失败", zap.Error(err))
				}
			case <-ctx.Done():
				logger.Info("汇率定期更新任务已停止")
				return
			}
		}
	}()

	logger.Info("汇率定期更新任务已启动",
		zap.Duration("interval", interval))
}

// GetHistoricalRate 获取历史汇率（从数据库查询）
// 如果没有配置repository，返回错误
func (c *ExchangeRateClient) GetHistoricalRate(ctx context.Context, from, to string, timestamp time.Time) (float64, error) {
	if c.repo == nil {
		return 0, fmt.Errorf("未配置汇率历史存储")
	}

	rate, err := c.repo.GetRateAtTime(ctx, from, to, timestamp)
	if err != nil {
		return 0, fmt.Errorf("查询历史汇率失败: %w", err)
	}

	if rate == nil {
		return 0, fmt.Errorf("未找到 %s 的历史汇率", timestamp.Format("2006-01-02 15:04:05"))
	}

	return rate.Rate, nil
}

// GetRateHistory 获取汇率历史记录
func (c *ExchangeRateClient) GetRateHistory(ctx context.Context, from, to string, startTime, endTime time.Time) ([]model.ExchangeRate, error) {
	if c.repo == nil {
		return nil, fmt.Errorf("未配置汇率历史存储")
	}

	rates, err := c.repo.GetRateHistory(ctx, from, to, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("查询汇率历史失败: %w", err)
	}

	// 将指针切片转换为值切片
	result := make([]model.ExchangeRate, len(rates))
	for i, rate := range rates {
		result[i] = *rate
	}

	return result, nil
}

// GetSnapshotHistory 获取汇率快照历史
func (c *ExchangeRateClient) GetSnapshotHistory(ctx context.Context, baseCurrency string, startTime, endTime time.Time) ([]model.ExchangeRateSnapshot, error) {
	if c.repo == nil {
		return nil, fmt.Errorf("未配置汇率历史存储")
	}

	snapshots, err := c.repo.GetSnapshotHistory(ctx, baseCurrency, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("查询快照历史失败: %w", err)
	}

	// 将指针切片转换为值切片
	result := make([]model.ExchangeRateSnapshot, len(snapshots))
	for i, snapshot := range snapshots {
		result[i] = *snapshot
	}

	return result, nil
}
