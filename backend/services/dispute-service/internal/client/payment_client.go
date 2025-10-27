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

// PaymentClient 支付网关客户端接口
type PaymentClient interface {
	GetPaymentByChannelTradeNo(ctx context.Context, channelTradeNo string) (*PaymentInfo, error)
}

// PaymentInfo 支付信息
type PaymentInfo struct {
	PaymentNo      string    `json:"payment_no"`
	MerchantID     uuid.UUID `json:"merchant_id"`
	OrderNo        string    `json:"order_no"`
	ChannelTradeNo string    `json:"channel_trade_no"`
	Amount         int64     `json:"amount"`
	Currency       string    `json:"currency"`
	Status         string    `json:"status"`
}

type paymentClient struct {
	baseURL    string
	httpClient *httpclient.Client
}

// NewPaymentClient 创建支付网关客户端
func NewPaymentClient(baseURL string) PaymentClient {
	return &paymentClient{
		baseURL: baseURL,
		httpClient: httpclient.NewClient(&httpclient.Config{
			Timeout:       30 * time.Second,
			MaxRetries:    3,
			RetryDelay:    time.Second,
			EnableLogging: false,
		}),
	}
}

// GetPaymentByChannelTradeNo 根据渠道交易号获取支付信息
func (c *paymentClient) GetPaymentByChannelTradeNo(ctx context.Context, channelTradeNo string) (*PaymentInfo, error) {
	url := fmt.Sprintf("%s/api/v1/internal/payments/by-channel-trade-no/%s", c.baseURL, channelTradeNo)

	resp, err := c.httpClient.Get(url, nil)
	if err != nil {
		return nil, fmt.Errorf("request payment failed: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("payment not found for channel_trade_no: %s", channelTradeNo)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("payment service returned status: %d", resp.StatusCode)
	}

	var result struct {
		Code int          `json:"code"`
		Data *PaymentInfo `json:"data"`
	}

	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("payment service returned error code: %d", result.Code)
	}

	if result.Data == nil {
		return nil, fmt.Errorf("payment data is null")
	}

	return result.Data, nil
}
