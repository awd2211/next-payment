# Circuit Breaker Implementation Examples

## Overview

This document shows code examples from the actual codebase for each implementation pattern.

---

## Pattern A: ServiceClient with Fallback (Backward Compatible)

### Example: Payment Gateway's OrderClient

**File**: `/home/eric/payment/backend/services/payment-gateway/internal/client/order_client.go`

```go
// OrderClient Order服务客户端
type OrderClient struct {
	*ServiceClient
}

// NewOrderClient 创建Order服务客户端（带熔断器）
func NewOrderClient(baseURL string) *OrderClient {
	return &OrderClient{
		ServiceClient: NewServiceClientWithBreaker(baseURL, "order-service"),
	}
}

// CreateOrder 创建订单 (automatically uses breaker)
func (c *OrderClient) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
	resp, err := c.http.Post(ctx, "/api/v1/orders", req, nil)
	if err != nil {
		return nil, fmt.Errorf("调用Order服务失败: %w", err)
	}
	// ...
}
```

### Base ServiceClient Implementation

**File**: `/home/eric/payment/backend/services/payment-gateway/internal/client/http_client.go`

```go
// ServiceClient 微服务客户端基类
type ServiceClient struct {
	http    *HTTPClient
	breaker *httpclient.BreakerClient
	baseURL string
}

// NewServiceClientWithBreaker 创建带熔断器的微服务客户端
func NewServiceClientWithBreaker(baseURL string, breakerName string) *ServiceClient {
	// 创建 pkg/httpclient 配置
	config := &httpclient.Config{
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RetryDelay: time.Second,
	}

	// 创建熔断器配置
	breakerConfig := httpclient.DefaultBreakerConfig(breakerName)

	// 创建带熔断器的客户端
	breakerClient := httpclient.NewBreakerClient(config, breakerConfig)

	return &ServiceClient{
		http:    NewHTTPClient(baseURL, 30*time.Second),
		breaker: breakerClient,
		baseURL: baseURL,
	}
}

// Post 执行POST请求（自动使用熔断器）
func (sc *ServiceClient) Post(ctx context.Context, path string, body interface{}, headers map[string]string) (*Response, error) {
	if sc.breaker != nil {
		return sc.doWithBreaker(ctx, http.MethodPost, path, body, headers)
	}
	return sc.http.Post(ctx, path, body, headers)
}

// doWithBreaker 通过熔断器执行请求
func (sc *ServiceClient) doWithBreaker(ctx context.Context, method, path string, body interface{}, headers map[string]string) (*Response, error) {
	fullURL := sc.baseURL + path

	req := &httpclient.Request{
		Method:  method,
		URL:     fullURL,
		Body:    body,
		Headers: headers,
		Ctx:     ctx,
	}

	resp, err := sc.breaker.Do(req)
	if err != nil {
		return nil, err
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       resp.Body,
		Headers:    resp.Headers,
	}, nil
}
```

### Usage Pattern A Summary

**Advantages**:
- Backward compatible - can use without breaker initially
- Reusable across multiple clients
- Smooth migration path

**When to use**:
- Migrating existing code
- Multiple clients share infrastructure
- Gradual rollout required

---

## Pattern B: Direct BreakerClient (Recommended)

### Example: Settlement Service's AccountingClient

**File**: `/home/eric/payment/backend/services/settlement-service/internal/client/accounting_client.go`

```go
// AccountingClient Accounting Service HTTP客户端
type AccountingClient struct {
	baseURL string
	breaker *httpclient.BreakerClient
}

// NewAccountingClient 创建Accounting客户端实例（带熔断器）
func NewAccountingClient(baseURL string) *AccountingClient {
	// 创建 httpclient 配置
	config := &httpclient.Config{
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RetryDelay: time.Second,
	}

	// 创建熔断器配置
	breakerConfig := httpclient.DefaultBreakerConfig("accounting-service")

	return &AccountingClient{
		baseURL: baseURL,
		breaker: httpclient.NewBreakerClient(config, breakerConfig),
	}
}

// GetTransactions 获取交易列表用于结算（使用熔断器）
func (c *AccountingClient) GetTransactions(ctx context.Context, merchantID uuid.UUID, startDate, endDate time.Time) ([]TransactionInfo, error) {
	// 构建URL
	url := fmt.Sprintf("%s/api/v1/transactions?merchant_id=%s&start_date=%s&end_date=%s&status=success&page_size=10000",
		c.baseURL,
		merchantID.String(),
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"))

	// 创建请求
	req := &httpclient.Request{
		Method: "GET",
		URL:    url,
		Ctx:    ctx,
	}

	// 通过熔断器发送请求
	resp, err := c.breaker.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	// 解析响应
	var result TransactionListResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("业务错误: %s", result.Message)
	}

	return result.Data.List, nil
}
```

### Pattern B Summary

**Advantages**:
- Enforces circuit breaker usage - no bypass possible
- Cleaner, more direct code path
- Easier to understand and maintain

**When to use**:
- New code (recommended)
- Critical service paths
- No need for backward compatibility

---

## Pattern C: Custom Breaker Configuration (External APIs)

### Example: Channel Adapter's ExchangeRateClient

**File**: `/home/eric/payment/backend/services/channel-adapter/internal/client/exchange_rate_client.go`

```go
// ExchangeRateClient 汇率API客户端
type ExchangeRateClient struct {
	baseURL string
	breaker *httpclient.BreakerClient
	redis   *redis.Client
	repo    repository.ExchangeRateRepository
	cacheTTL time.Duration
}

// NewExchangeRateClient 创建汇率API客户端（带熔断器）
func NewExchangeRateClient(redis *redis.Client, repo repository.ExchangeRateRepository, cacheTTL time.Duration) *ExchangeRateClient {
	// 创建 httpclient 配置
	config := &httpclient.Config{
		Timeout:    5 * time.Second,     // 外部API使用较短超时
		MaxRetries: 2,                   // 外部API减少重试次数
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

	// ... rest of implementation
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
	fallbackRates := map[string]map[string]float64{
		"USD": {
			"EUR": 0.92,
			"GBP": 0.79,
			"CNY": 7.24,
		},
		// ... more rates
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

	// Default to 1.0 if no fallback available
	logger.Error("无法获取汇率，返回默认值1.0",
		zap.String("from", from),
		zap.String("to", to))
	return 1.0
}
```

### Pattern C Summary

**Key Features**:
- **Lower timeout**: 5s instead of 30s (faster failure detection)
- **Fewer retries**: 2 instead of 3 (reduce hanging requests)
- **Higher failure threshold**: 80% instead of 60% (more forgiving for external APIs)
- **Graceful degradation**: Fallback rates when API unavailable
- **Caching strategy**: Redis caching to reduce API calls

**When to use**:
- Third-party external APIs
- Services with unreliable connectivity
- When fallback data is available

---

## Anti-Pattern: What NOT to Do

### DON'T: Raw http.Client (Payment Gateway's Merchant Auth Client)

**File**: `/home/eric/payment/backend/services/payment-gateway/internal/client/merchant_auth_client.go`

```go
// ❌ WRONG - This is the current problematic implementation
type merchantAuthClient struct {
	baseURL string
	client  *http.Client  // ❌ NO CIRCUIT BREAKER
}

func NewMerchantAuthClient(baseURL string) MerchantAuthClient {
	return &merchantAuthClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 5 * time.Second,  // ❌ ONLY TIMEOUT, NO PROTECTION
		},
	}
}

func (c *merchantAuthClient) ValidateSignature(ctx context.Context, apiKey, signature, payload string) (*ValidateSignatureResponse, error) {
	// ...
	resp, err := c.client.Do(req)  // ❌ NO CIRCUIT BREAKER PROTECTION
	// ...
}
```

**Problems**:
1. No circuit breaker - no automatic failure detection
2. Only 5s timeout - will block payment processing
3. No retry logic - immediate failure
4. Can cascade failures across entire payment system
5. Resource exhaustion risk with hanging connections

**Fix Required**:
Convert to Pattern B (Direct BreakerClient) as shown above.

---

## Configuration Patterns

### Default Configuration (Service-to-Service)

```go
config := &httpclient.Config{
	Timeout:    30 * time.Second,
	MaxRetries: 3,
	RetryDelay: time.Second,
}

breakerConfig := httpclient.DefaultBreakerConfig("service-name")
// Default settings:
// - Max 3 requests in half-open state
// - Reset every 1 minute
// - Trip timeout 60 seconds
// - Trip on: 5+ requests with 60%+ failure rate
```

### External API Configuration

```go
config := &httpclient.Config{
	Timeout:    5 * time.Second,       // Much shorter
	MaxRetries: 2,                     // Fewer retries
	RetryDelay: 500 * time.Millisecond,
}

breakerConfig := httpclient.DefaultBreakerConfig("external-api")
breakerConfig.ReadyToTrip = func(counts gobreaker.Counts) bool {
	failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
	return counts.Requests >= 10 && failureRatio >= 0.8  // 80% failure threshold
}
```

### Critical Service Configuration (Optional - Stricter)

```go
config := &httpclient.Config{
	Timeout:    10 * time.Second,      // Lower timeout
	MaxRetries: 2,                     // Fewer retries
	RetryDelay: 500 * time.Millisecond,
}

breakerConfig := httpclient.DefaultBreakerConfig("critical-service")
breakerConfig.ReadyToTrip = func(counts gobreaker.Counts) bool {
	failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
	return counts.Requests >= 3 && failureRatio >= 0.5  // Trip sooner
}
```

---

## Monitoring & Observability

### Circuit Breaker Events to Monitor

```
CircuitBreakerStateChangeEvent:
- Service name
- From state (Closed/Open/HalfOpen)
- To state
- Timestamp
- Reason

CircuitBreakerTrippedEvent:
- Service name
- Failure count
- Failure ratio
- Threshold that was exceeded
```

### Prometheus Metrics (if implemented)

```
circuit_breaker_state{service="order-service"}
circuit_breaker_trips_total{service="order-service"}
circuit_breaker_requests_total{service="order-service",state="closed|open|half-open"}
circuit_breaker_failures_total{service="order-service"}
```

### Log Examples

```
// Circuit breaker opened
WARN "Circuit breaker tripped" service=channel-adapter reason=high_failure_rate
     requests=15 failures=10 failure_ratio=0.67 threshold=0.60

// Circuit breaker recovering
INFO "Circuit breaker testing recovery" service=order-service
     max_test_requests=3 elapsed=65s

// Circuit breaker closed (recovered)
INFO "Circuit breaker recovered" service=risk-service
     state=closed downtime=125s
```

