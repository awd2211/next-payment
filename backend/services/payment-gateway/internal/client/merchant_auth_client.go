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

// MerchantAuthClient 商户认证服务客户端
type MerchantAuthClient interface {
	ValidateSignature(ctx context.Context, apiKey, signature, payload string) (*ValidateSignatureResponse, error)
}

type merchantAuthClient struct {
	baseURL string
	client  *httpclient.BreakerClient // 使用熔断器客户端
}

// NewMerchantAuthClient 创建商户认证服务客户端（带熔断器保护）
func NewMerchantAuthClient(baseURL string) MerchantAuthClient {
	// 创建HTTP客户端配置
	config := &httpclient.Config{
		Timeout:       5 * time.Second,
		MaxRetries:    2,
		RetryDelay:    500 * time.Millisecond,
		EnableLogging: true,
	}

	// 创建熔断器配置
	breakerConfig := httpclient.DefaultBreakerConfig("merchant-auth-service")

	return &merchantAuthClient{
		baseURL: baseURL,
		client:  httpclient.NewBreakerClient(config, breakerConfig),
	}
}

// ValidateSignatureResponse 验证签名响应
type ValidateSignatureResponse struct {
	Valid       bool      `json:"valid"`
	MerchantID  uuid.UUID `json:"merchant_id"`
	Environment string    `json:"environment"`
}

// ValidateSignature 验证API签名
func (c *merchantAuthClient) ValidateSignature(ctx context.Context, apiKey, signature, payload string) (*ValidateSignatureResponse, error) {
	// 构建请求体
	reqBody := map[string]string{
		"api_key":   apiKey,
		"signature": signature,
		"payload":   payload,
	}

	// 构建请求
	url := c.baseURL + "/api/v1/auth/validate-signature"
	req := &httpclient.Request{
		Method: "POST",
		URL:    url,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: reqBody,
		Ctx:  ctx,
	}

	// 通过熔断器发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		logger.Error("Failed to validate signature via merchant-auth-service",
			zap.Error(err),
			zap.String("url", url),
			zap.String("circuit_breaker", "merchant-auth-service"),
			zap.String("breaker_state", c.client.State().String()))
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != 200 {
		var errResp struct {
			Error string `json:"error"`
		}
		json.Unmarshal(resp.Body, &errResp)
		return nil, fmt.Errorf("validation failed: %s (status %d)", errResp.Error, resp.StatusCode)
	}

	// 解析响应
	var result struct {
		Valid       bool   `json:"valid"`
		MerchantID  string `json:"merchant_id"`
		Environment string `json:"environment"`
	}
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// 解析 MerchantID
	merchantID, err := uuid.Parse(result.MerchantID)
	if err != nil {
		return nil, fmt.Errorf("invalid merchant_id format: %w", err)
	}

	return &ValidateSignatureResponse{
		Valid:       result.Valid,
		MerchantID:  merchantID,
		Environment: result.Environment,
	}, nil
}
