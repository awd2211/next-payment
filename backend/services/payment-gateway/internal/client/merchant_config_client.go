package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/httpclient"
	"github.com/payment-platform/pkg/logger"
	"go.uber.org/zap"
)

// MerchantConfigClient 商户配置服务客户端
type MerchantConfigClient interface {
	GetWebhookSecret(ctx context.Context, merchantID uuid.UUID) (string, error)
}

type merchantConfigClient struct {
	baseURL string
	client  *httpclient.BreakerClient
}

// NewMerchantConfigClient 创建商户配置服务客户端
func NewMerchantConfigClient(baseURL string) MerchantConfigClient {
	config := &httpclient.Config{
		Timeout:       5 * time.Second,
		MaxRetries:    2,
		RetryDelay:    500 * time.Millisecond,
		EnableLogging: true,
	}

	breakerConfig := httpclient.DefaultBreakerConfig("merchant-config-service")

	return &merchantConfigClient{
		baseURL: baseURL,
		client:  httpclient.NewBreakerClient(config, breakerConfig),
	}
}

// GetWebhookSecret 获取商户的Webhook密钥
func (c *merchantConfigClient) GetWebhookSecret(ctx context.Context, merchantID uuid.UUID) (string, error) {
	url := fmt.Sprintf("%s/api/v1/merchants/%s/webhook-secret", c.baseURL, merchantID.String())
	
	req := &httpclient.Request{
		Method: "GET",
		URL:    url,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Ctx: ctx,
	}

	resp, err := c.client.Do(req)
	if err != nil {
		logger.Error("Failed to get webhook secret from merchant-config-service",
			zap.Error(err),
			zap.String("url", url),
			zap.String("merchant_id", merchantID.String()))
		return "", fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != 200 {
		var errResp struct {
			Error string `json:"error"`
		}
		json.Unmarshal(resp.Body, &errResp)
		return "", fmt.Errorf("get webhook secret failed: %s (status %d)", errResp.Error, resp.StatusCode)
	}

	var result struct {
		Secret string `json:"secret"`
	}
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Secret, nil
}
