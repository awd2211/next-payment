package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/httpclient"
)

// MerchantClient 商户服务客户端接口
type MerchantClient interface {
	GetMerchant(ctx context.Context, merchantID uuid.UUID) (*MerchantInfo, error)
	GetMerchantWithPassword(ctx context.Context, merchantID uuid.UUID) (*MerchantWithPassword, error)
	UpdatePassword(ctx context.Context, merchantID uuid.UUID, newPasswordHash string) error
	ValidateMerchantStatus(ctx context.Context, merchantID uuid.UUID) error
}

// MerchantInfo 商户基本信息
type MerchantInfo struct {
	ID           uuid.UUID `json:"id"`
	MerchantNo   string    `json:"merchant_no"`
	MerchantName string    `json:"merchant_name"`
	Status       string    `json:"status"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
}

// MerchantWithPassword 带密码哈希的商户信息（仅用于密码验证）
type MerchantWithPassword struct {
	MerchantInfo
	PasswordHash string `json:"password_hash"`
}

type merchantClient struct {
	baseURL string
	breaker *httpclient.BreakerClient
}

// NewMerchantClient 创建商户服务客户端（带熔断器）
func NewMerchantClient(baseURL string) MerchantClient {
	// 创建 httpclient 配置
	config := &httpclient.Config{
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RetryDelay: time.Second,
	}

	// 创建熔断器配置
	breakerConfig := httpclient.DefaultBreakerConfig("merchant-service")

	return &merchantClient{
		baseURL: baseURL,
		breaker: httpclient.NewBreakerClient(config, breakerConfig),
	}
}

// GetMerchant 获取商户信息（使用熔断器）
func (c *merchantClient) GetMerchant(ctx context.Context, merchantID uuid.UUID) (*MerchantInfo, error) {
	url := fmt.Sprintf("%s/api/v1/merchants/%s", c.baseURL, merchantID.String())

	// 创建请求
	req := &httpclient.Request{
		Method: "GET",
		URL:    url,
		Ctx:    ctx,
	}

	// 通过熔断器发送请求
	resp, err := c.breaker.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call merchant service: %w", err)
	}

	// 解析响应
	var result struct {
		Code    int          `json:"code"`
		Message string       `json:"message"`
		Data    MerchantInfo `json:"data"`
	}

	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("merchant service error: %s", result.Message)
	}

	return &result.Data, nil
}

// GetMerchantWithPassword 获取带密码哈希的商户信息（用于密码验证，使用熔断器）
func (c *merchantClient) GetMerchantWithPassword(ctx context.Context, merchantID uuid.UUID) (*MerchantWithPassword, error) {
	url := fmt.Sprintf("%s/api/v1/merchants/%s/with-password", c.baseURL, merchantID.String())

	// 创建请求
	req := &httpclient.Request{
		Method: "GET",
		URL:    url,
		Ctx:    ctx,
	}

	// 通过熔断器发送请求
	resp, err := c.breaker.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call merchant service: %w", err)
	}

	// 解析响应
	var result struct {
		Code    int                  `json:"code"`
		Message string               `json:"message"`
		Data    MerchantWithPassword `json:"data"`
	}

	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("merchant service error: %s", result.Message)
	}

	return &result.Data, nil
}

// UpdatePassword 更新商户密码（使用熔断器）
func (c *merchantClient) UpdatePassword(ctx context.Context, merchantID uuid.UUID, newPasswordHash string) error {
	url := fmt.Sprintf("%s/api/v1/merchants/%s/password", c.baseURL, merchantID.String())

	payload := map[string]string{
		"password_hash": newPasswordHash,
	}

	// 创建请求
	req := &httpclient.Request{
		Method: "PUT",
		URL:    url,
		Body:   payload,
		Ctx:    ctx,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	// 通过熔断器发送请求
	resp, err := c.breaker.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call merchant service: %w", err)
	}

	// 解析响应
	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Code != 0 {
		return fmt.Errorf("merchant service error: %s", result.Message)
	}

	return nil
}

// ValidateMerchantStatus 验证商户状态
func (c *merchantClient) ValidateMerchantStatus(ctx context.Context, merchantID uuid.UUID) error {
	merchant, err := c.GetMerchant(ctx, merchantID)
	if err != nil {
		return err
	}

	if merchant.Status != "active" {
		return fmt.Errorf("merchant status is %s, not active", merchant.Status)
	}

	return nil
}
