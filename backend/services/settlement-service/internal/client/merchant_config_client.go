package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/httpclient"
)

// MerchantConfigClient Merchant Config Service HTTP客户端
type MerchantConfigClient struct {
	baseURL string
	breaker *httpclient.BreakerClient
}

// NewMerchantConfigClient 创建MerchantConfig客户端实例（带熔断器）
func NewMerchantConfigClient(baseURL string) *MerchantConfigClient {
	config := &httpclient.Config{
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RetryDelay: time.Second,
	}
	breakerConfig := httpclient.DefaultBreakerConfig("merchant-config-service")

	return &MerchantConfigClient{
		baseURL: baseURL,
		breaker: httpclient.NewBreakerClient(config, breakerConfig),
	}
}

// SettlementConfig 结算配置
type SettlementConfig struct {
	MerchantID           uuid.UUID `json:"merchant_id"`
	AutoSettlement       bool      `json:"auto_settlement"`        // 是否启用自动结算
	SettlementCycle      string    `json:"settlement_cycle"`       // daily, weekly, monthly
	MinSettlementAmount  int64     `json:"min_settlement_amount"`  // 最小结算金额(分)
	FeeRate              float64   `json:"fee_rate"`               // 手续费率
	FixedFee             int64     `json:"fixed_fee"`              // 固定手续费(分)
	SettlementDay        *int      `json:"settlement_day"`         // 周/月结算日
	HoldDays             int       `json:"hold_days"`              // 保留天数
	AutoApproveThreshold int64     `json:"auto_approve_threshold"` // 自动审批阈值(分)
}

// GetSettlementConfigResponse 获取结算配置响应
type GetSettlementConfigResponse struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Data    *SettlementConfig `json:"data"`
}

// ListAutoSettlementMerchantsResponse 自动结算商户列表响应
type ListAutoSettlementMerchantsResponse struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Data    []uuid.UUID `json:"data"`
}

// GetSettlementConfig 获取商户结算配置
func (c *MerchantConfigClient) GetSettlementConfig(ctx context.Context, merchantID uuid.UUID) (*SettlementConfig, error) {
	url := fmt.Sprintf("%s/api/v1/merchant-configs/%s/settlement", c.baseURL, merchantID.String())

	req := &httpclient.Request{
		Method: "GET",
		URL:    url,
		Ctx:    ctx,
	}

	resp, err := c.breaker.Do(req)
	if err != nil {
		// 降级：返回默认配置
		return &SettlementConfig{
			MerchantID:           merchantID,
			AutoSettlement:       false, // 默认不启用自动结算
			SettlementCycle:      "daily",
			MinSettlementAmount:  10000,   // 100元
			FeeRate:              0.006,   // 0.6%
			FixedFee:             0,
			HoldDays:             1,
			AutoApproveThreshold: 1000000, // 10000元
		}, fmt.Errorf("获取配置失败（已降级）: %w", err)
	}

	var result GetSettlementConfigResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("业务错误: %s", result.Message)
	}

	if result.Data == nil {
		// 返回默认配置
		return &SettlementConfig{
			MerchantID:           merchantID,
			AutoSettlement:       false,
			SettlementCycle:      "daily",
			MinSettlementAmount:  10000,
			FeeRate:              0.006,
			FixedFee:             0,
			HoldDays:             1,
			AutoApproveThreshold: 1000000,
		}, nil
	}

	return result.Data, nil
}

// ListAutoSettlementMerchants 获取启用自动结算的商户列表
func (c *MerchantConfigClient) ListAutoSettlementMerchants(ctx context.Context) ([]uuid.UUID, error) {
	url := fmt.Sprintf("%s/api/v1/merchant-configs/auto-settlement/list", c.baseURL)

	req := &httpclient.Request{
		Method: "GET",
		URL:    url,
		Ctx:    ctx,
	}

	resp, err := c.breaker.Do(req)
	if err != nil {
		// 降级：返回空列表，不阻塞主流程
		return []uuid.UUID{}, fmt.Errorf("获取自动结算商户列表失败（已降级）: %w", err)
	}

	var result ListAutoSettlementMerchantsResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("业务错误: %s", result.Message)
	}

	return result.Data, nil
}
