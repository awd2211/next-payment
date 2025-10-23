package currency

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// ExchangeRateProvider 汇率提供商接口
type ExchangeRateProvider interface {
	GetRates(ctx context.Context, baseCurrency string) (map[string]float64, error)
	GetRate(ctx context.Context, from, to string) (float64, error)
}

// ExchangeRateService 汇率服务
type ExchangeRateService struct {
	provider      ExchangeRateProvider
	redis         *redis.Client
	cacheTTL      time.Duration
	mu            sync.RWMutex
	ratesCache    map[string]*CachedRates
	enableCache   bool
}

// CachedRates 缓存的汇率数据
type CachedRates struct {
	Base      string             `json:"base"`
	Rates     map[string]float64 `json:"rates"`
	Timestamp time.Time          `json:"timestamp"`
}

// NewExchangeRateService 创建汇率服务实例
func NewExchangeRateService(provider ExchangeRateProvider, redisClient *redis.Client) *ExchangeRateService {
	return &ExchangeRateService{
		provider:    provider,
		redis:       redisClient,
		cacheTTL:    10 * time.Minute, // 缓存10分钟
		ratesCache:  make(map[string]*CachedRates),
		enableCache: true,
	}
}

// GetRate 获取汇率（from -> to）
func (s *ExchangeRateService) GetRate(ctx context.Context, from, to string) (float64, error) {
	// 相同货币，汇率为1
	if from == to {
		return 1.0, nil
	}

	// 先尝试从Redis缓存获取
	if s.enableCache && s.redis != nil {
		cacheKey := fmt.Sprintf("exchange_rate:%s:%s", from, to)
		cached, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			rate, err := strconv.ParseFloat(cached, 64)
			if err == nil {
				return rate, nil
			}
		}
	}

	// 从提供商获取
	rate, err := s.provider.GetRate(ctx, from, to)
	if err != nil {
		return 0, err
	}

	// 存入Redis缓存
	if s.enableCache && s.redis != nil {
		cacheKey := fmt.Sprintf("exchange_rate:%s:%s", from, to)
		s.redis.Set(ctx, cacheKey, fmt.Sprintf("%.6f", rate), s.cacheTTL)
	}

	return rate, nil
}

// GetRates 获取基于某个货币的所有汇率
func (s *ExchangeRateService) GetRates(ctx context.Context, baseCurrency string) (map[string]float64, error) {
	// 先尝试从内存缓存获取
	if s.enableCache {
		s.mu.RLock()
		cached, exists := s.ratesCache[baseCurrency]
		s.mu.RUnlock()

		if exists && time.Since(cached.Timestamp) < s.cacheTTL {
			return cached.Rates, nil
		}
	}

	// 从提供商获取
	rates, err := s.provider.GetRates(ctx, baseCurrency)
	if err != nil {
		return nil, err
	}

	// 存入内存缓存
	if s.enableCache {
		s.mu.Lock()
		s.ratesCache[baseCurrency] = &CachedRates{
			Base:      baseCurrency,
			Rates:     rates,
			Timestamp: time.Now(),
		}
		s.mu.Unlock()
	}

	return rates, nil
}

// Convert 货币转换
func (s *ExchangeRateService) Convert(ctx context.Context, amount float64, from, to string) (float64, error) {
	rate, err := s.GetRate(ctx, from, to)
	if err != nil {
		return 0, err
	}

	return amount * rate, nil
}

// ConvertAmount 货币转换（金额以分为单位）
func (s *ExchangeRateService) ConvertAmount(ctx context.Context, amountCents int64, from, to string) (int64, error) {
	rate, err := s.GetRate(ctx, from, to)
	if err != nil {
		return 0, err
	}

	// 转换为浮点数计算，然后四舍五入
	amountFloat := float64(amountCents) * rate
	return int64(amountFloat + 0.5), nil
}

// ClearCache 清除缓存
func (s *ExchangeRateService) ClearCache(ctx context.Context) {
	s.mu.Lock()
	s.ratesCache = make(map[string]*CachedRates)
	s.mu.Unlock()

	if s.redis != nil {
		// 清除Redis中的汇率缓存
		iter := s.redis.Scan(ctx, 0, "exchange_rate:*", 0).Iterator()
		for iter.Next(ctx) {
			s.redis.Del(ctx, iter.Val())
		}
	}
}

// ========== 汇率提供商实现 ==========

// ExchangeRateAPIProvider 使用ExchangeRate-API.com的免费API
type ExchangeRateAPIProvider struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewExchangeRateAPIProvider 创建ExchangeRate-API提供商
// 免费API: https://www.exchangerate-api.com/
func NewExchangeRateAPIProvider(apiKey string) *ExchangeRateAPIProvider {
	return &ExchangeRateAPIProvider{
		apiKey:  apiKey,
		baseURL: "https://v6.exchangerate-api.com/v6",
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetRates 获取汇率
func (p *ExchangeRateAPIProvider) GetRates(ctx context.Context, baseCurrency string) (map[string]float64, error) {
	url := fmt.Sprintf("%s/%s/latest/%s", p.baseURL, p.apiKey, baseCurrency)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回错误状态: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var result struct {
		Result           string             `json:"result"`
		BaseCode         string             `json:"base_code"`
		ConversionRates  map[string]float64 `json:"conversion_rates"`
		TimeLastUpdateUnix int64            `json:"time_last_update_unix"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Result != "success" {
		return nil, fmt.Errorf("API返回失败结果")
	}

	return result.ConversionRates, nil
}

// GetRate 获取单个汇率
func (p *ExchangeRateAPIProvider) GetRate(ctx context.Context, from, to string) (float64, error) {
	url := fmt.Sprintf("%s/%s/pair/%s/%s", p.baseURL, p.apiKey, from, to)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API返回错误状态: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("读取响应失败: %w", err)
	}

	var result struct {
		Result        string  `json:"result"`
		BaseCode      string  `json:"base_code"`
		TargetCode    string  `json:"target_code"`
		ConversionRate float64 `json:"conversion_rate"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Result != "success" {
		return 0, fmt.Errorf("API返回失败结果")
	}

	return result.ConversionRate, nil
}

// ========== 其他汇率提供商 ==========

// FixerIOProvider Fixer.io汇率提供商
type FixerIOProvider struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewFixerIOProvider 创建Fixer.io提供商
// 官网：https://fixer.io/
func NewFixerIOProvider(apiKey string) *FixerIOProvider {
	return &FixerIOProvider{
		apiKey:  apiKey,
		baseURL: "https://api.fixer.io",
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetRates 获取汇率
func (p *FixerIOProvider) GetRates(ctx context.Context, baseCurrency string) (map[string]float64, error) {
	url := fmt.Sprintf("%s/latest?access_key=%s&base=%s", p.baseURL, p.apiKey, baseCurrency)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var result struct {
		Success bool               `json:"success"`
		Base    string             `json:"base"`
		Rates   map[string]float64 `json:"rates"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("API返回失败")
	}

	return result.Rates, nil
}

// GetRate 获取单个汇率
func (p *FixerIOProvider) GetRate(ctx context.Context, from, to string) (float64, error) {
	rates, err := p.GetRates(ctx, from)
	if err != nil {
		return 0, err
	}

	rate, ok := rates[to]
	if !ok {
		return 0, fmt.Errorf("未找到汇率: %s -> %s", from, to)
	}

	return rate, nil
}

// ========== 支持的货币列表 ==========

// SupportedCurrencies 支持的货币列表（与用户偏好设置保持一致）
var SupportedCurrencies = []string{
	"USD", "EUR", "GBP", "CNY", "JPY", "KRW", "HKD", "SGD",
	"AUD", "CAD", "INR", "BRL", "MXN", "RUB", "TRY", "ZAR",
	"CHF", "SEK", "NOK", "DKK", "PLN", "CZK", "HUF", "THB",
	"IDR", "MYR", "PHP", "VND", "AED", "SAR", "ILS", "EGP",
}

// IsSupportedCurrency 检查是否为支持的货币
func IsSupportedCurrency(currency string) bool {
	for _, c := range SupportedCurrencies {
		if c == currency {
			return true
		}
	}
	return false
}
