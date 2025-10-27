package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/httpclient"
)

// PolicyClient 策略服务客户端接口
type PolicyClient interface {
	GetEffectiveLimitPolicy(ctx context.Context, merchantID uuid.UUID, channel, currency string) (*LimitPolicy, error)
	GetEffectiveFeePolicy(ctx context.Context, merchantID uuid.UUID, channel, paymentMethod, currency string) (*FeePolicy, error)
	CheckLimit(ctx context.Context, merchantID uuid.UUID, channel, currency string, amount, dailyUsed, monthlyUsed int64) (*LimitCheckResult, error)
}

// LimitPolicy 限额策略
type LimitPolicy struct {
	SingleTransMin int64  `json:"single_trans_min"` // 单笔最小金额
	SingleTransMax int64  `json:"single_trans_max"` // 单笔最大金额
	DailyLimit     int64  `json:"daily_limit"`      // 日限额
	MonthlyLimit   int64  `json:"monthly_limit"`    // 月限额
	YearlyLimit    int64  `json:"yearly_limit"`     // 年限额
	Channel        string `json:"channel"`
	Currency       string `json:"currency"`
}

// FeePolicy 费率策略
type FeePolicy struct {
	FeeType       string  `json:"fee_type"`       // percentage, fixed, tiered
	FeePercentage float64 `json:"fee_percentage"` // 费率百分比 (0.025 = 2.5%)
	FeeFixed      int64   `json:"fee_fixed"`      // 固定费用（分）
	FeeMin        int64   `json:"fee_min"`        // 最小费用（分）
	FeeMax        int64   `json:"fee_max"`        // 最大费用（分）
}

// LimitCheckResult 限额检查结果
type LimitCheckResult struct {
	IsAllowed        bool   `json:"is_allowed"`
	RejectionReason  string `json:"rejection_reason"`
	SingleTransMax   int64  `json:"single_trans_max"`
	DailyLimit       int64  `json:"daily_limit"`
	DailyRemaining   int64  `json:"daily_remaining"`
	MonthlyLimit     int64  `json:"monthly_limit"`
	MonthlyRemaining int64  `json:"monthly_remaining"`
}

type policyClient struct {
	baseURL    string
	httpClient *httpclient.Client
}

// NewPolicyClient 创建策略服务客户端
func NewPolicyClient(baseURL string) PolicyClient {
	return &policyClient{
		baseURL: baseURL,
		httpClient: httpclient.NewClient(&httpclient.Config{
			Timeout:       10 * time.Second,
			MaxRetries:    3,
			RetryDelay:    time.Second,
			EnableLogging: false,
		}),
	}
}

// GetEffectiveLimitPolicy 获取商户的有效限额策略
func (c *policyClient) GetEffectiveLimitPolicy(ctx context.Context, merchantID uuid.UUID, channel, currency string) (*LimitPolicy, error) {
	url := fmt.Sprintf("%s/api/v1/policy-engine/limits/%s?channel=%s&currency=%s",
		c.baseURL, merchantID.String(), channel, currency)

	resp, err := c.httpClient.Get(url, nil)
	if err != nil {
		return nil, fmt.Errorf("request limit policy failed: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		// Return default limits if no custom policy exists
		return &LimitPolicy{
			SingleTransMin: 100,        // $1.00
			SingleTransMax: 100000000,  // $1,000,000
			DailyLimit:     500000000,  // $5,000,000
			MonthlyLimit:   10000000000, // $100,000,000
			Channel:        channel,
			Currency:       currency,
		}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("policy service returned status: %d", resp.StatusCode)
	}

	var result struct {
		Code int          `json:"code"`
		Data *LimitPolicy `json:"data"`
	}

	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("policy service returned error code: %d", result.Code)
	}

	if result.Data == nil {
		// Return default limits as fallback
		return &LimitPolicy{
			SingleTransMin: 100,
			SingleTransMax: 100000000,
			DailyLimit:     500000000,
			MonthlyLimit:   10000000000,
			Channel:        channel,
			Currency:       currency,
		}, nil
	}

	return result.Data, nil
}

// GetEffectiveFeePolicy 获取商户的有效费率策略
func (c *policyClient) GetEffectiveFeePolicy(ctx context.Context, merchantID uuid.UUID, channel, paymentMethod, currency string) (*FeePolicy, error) {
	url := fmt.Sprintf("%s/api/v1/policy-engine/fees/%s?channel=%s&payment_method=%s&currency=%s",
		c.baseURL, merchantID.String(), channel, paymentMethod, currency)

	resp, err := c.httpClient.Get(url, nil)
	if err != nil {
		return nil, fmt.Errorf("request fee policy failed: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		// Return default fee policy
		return &FeePolicy{
			FeeType:       "percentage",
			FeePercentage: 0.029, // 2.9%
			FeeFixed:      30,    // $0.30
			FeeMin:        10,
			FeeMax:        0, // No max
		}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("policy service returned status: %d", resp.StatusCode)
	}

	var result struct {
		Code int        `json:"code"`
		Data *FeePolicy `json:"data"`
	}

	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("policy service returned error code: %d", result.Code)
	}

	if result.Data == nil {
		return &FeePolicy{
			FeeType:       "percentage",
			FeePercentage: 0.029,
			FeeFixed:      30,
			FeeMin:        10,
			FeeMax:        0,
		}, nil
	}

	return result.Data, nil
}

// CheckLimit 检查交易是否超限
func (c *policyClient) CheckLimit(ctx context.Context, merchantID uuid.UUID, channel, currency string, amount, dailyUsed, monthlyUsed int64) (*LimitCheckResult, error) {
	url := fmt.Sprintf("%s/api/v1/policy-engine/limits/%s/check?channel=%s&currency=%s&amount=%d&daily_used=%d&monthly_used=%d",
		c.baseURL, merchantID.String(), channel, currency, amount, dailyUsed, monthlyUsed)

	resp, err := c.httpClient.Get(url, nil)
	if err != nil {
		return nil, fmt.Errorf("check limit failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("policy service returned status: %d", resp.StatusCode)
	}

	var result struct {
		Code int               `json:"code"`
		Data *LimitCheckResult `json:"data"`
	}

	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("policy service returned error code: %d", result.Code)
	}

	return result.Data, nil
}
