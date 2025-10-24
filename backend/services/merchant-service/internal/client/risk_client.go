package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// RiskClient Risk Service HTTP客户端
type RiskClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewRiskClient 创建Risk客户端实例
func NewRiskClient(baseURL string) *RiskClient {
	return &RiskClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
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
	url := fmt.Sprintf("%s/api/v1/risk/merchants/%s", c.baseURL, merchantID.String())

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	var result RiskInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("业务错误: %s", result.Message)
	}

	return result.Data, nil
}
