package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/payment-platform/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// ExchangeRateClient 汇率API客户端
type ExchangeRateClient struct {
	baseURL    string
	httpClient *http.Client
	redis      *redis.Client
	cacheTTL   time.Duration
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

// NewExchangeRateClient 创建汇率API客户端
// 使用 exchangerate-api.com 免费版（1500次/月）
func NewExchangeRateClient(redis *redis.Client, cacheTTL time.Duration) *ExchangeRateClient {
	return &ExchangeRateClient{
		baseURL: "https://api.exchangerate-api.com/v4/latest",
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		redis:    redis,
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

	return rates, nil
}

// fetchRates 从API获取汇率数据
func (c *ExchangeRateClient) fetchRates(ctx context.Context, baseCurrency string) (map[string]float64, error) {
	url := fmt.Sprintf("%s/%s", c.baseURL, baseCurrency)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API调用失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API返回错误: status=%d", resp.StatusCode)
	}

	var apiResp ExchangeRateResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
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
