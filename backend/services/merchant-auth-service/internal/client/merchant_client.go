package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
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
	baseURL    string
	httpClient *http.Client
}

// NewMerchantClient 创建商户服务客户端
func NewMerchantClient(baseURL string) MerchantClient {
	return &merchantClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetMerchant 获取商户信息
func (c *merchantClient) GetMerchant(ctx context.Context, merchantID uuid.UUID) (*MerchantInfo, error) {
	url := fmt.Sprintf("%s/api/v1/merchants/%s", c.baseURL, merchantID.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call merchant service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("merchant service returned status %d", resp.StatusCode)
	}

	var result struct {
		Code    int          `json:"code"`
		Message string       `json:"message"`
		Data    MerchantInfo `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("merchant service error: %s", result.Message)
	}

	return &result.Data, nil
}

// GetMerchantWithPassword 获取带密码哈希的商户信息（用于密码验证）
func (c *merchantClient) GetMerchantWithPassword(ctx context.Context, merchantID uuid.UUID) (*MerchantWithPassword, error) {
	url := fmt.Sprintf("%s/api/v1/merchants/%s/with-password", c.baseURL, merchantID.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call merchant service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("merchant service returned status %d", resp.StatusCode)
	}

	var result struct {
		Code    int                  `json:"code"`
		Message string               `json:"message"`
		Data    MerchantWithPassword `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("merchant service error: %s", result.Message)
	}

	return &result.Data, nil
}

// UpdatePassword 更新商户密码
func (c *merchantClient) UpdatePassword(ctx context.Context, merchantID uuid.UUID, newPasswordHash string) error {
	url := fmt.Sprintf("%s/api/v1/merchants/%s/password", c.baseURL, merchantID.String())

	payload := map[string]string{
		"password_hash": newPasswordHash,
	}
	jsonData, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call merchant service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("merchant service returned status %d", resp.StatusCode)
	}

	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
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
