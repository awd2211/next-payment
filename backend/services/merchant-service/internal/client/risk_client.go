package client

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// RiskClient Risk Service HTTP客户端
type RiskClient struct {
	*ServiceClient
}

// NewRiskClient 创建Risk客户端实例（带熔断器）
func NewRiskClient(baseURL string) *RiskClient {
	return &RiskClient{
		ServiceClient: NewServiceClientWithBreaker(baseURL, "risk-service"),
	}
}

// RiskInfoResponse 风控信息响应
type RiskInfoResponse struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Data    *RiskInfo    `json:"data"`
}

// RiskInfo 风控信息
type RiskInfo struct {
	MerchantID     string `json:"merchant_id"`
	RiskLevel      string `json:"risk_level"`       // low, medium, high
	RiskScore      int    `json:"risk_score"`       // 0-100
	PendingReviews int    `json:"pending_reviews"`  // 待审核交易数
	BlockedCount   int    `json:"blocked_count"`    // 被拦截交易数
	LastUpdatedAt  string `json:"last_updated_at"`
}

// GetRiskInfo 获取风控信息
func (c *RiskClient) GetRiskInfo(ctx context.Context, merchantID uuid.UUID) (*RiskInfo, error) {
	path := fmt.Sprintf("/api/v1/risk/merchants/%s", merchantID.String())

	resp, err := c.http.Get(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	var result RiskInfoResponse
	if err := resp.ParseResponse(&result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("业务错误: %s", result.Message)
	}

	return result.Data, nil
}
