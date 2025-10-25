package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"payment-platform/reconciliation-service/internal/service"
)

// PlatformClient 平台数据客户端
type PlatformClient struct {
	paymentGatewayURL string
	httpClient        *http.Client
}

// NewPlatformClient 创建平台客户端
func NewPlatformClient(paymentGatewayURL string) *PlatformClient {
	return &PlatformClient{
		paymentGatewayURL: paymentGatewayURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchPayments 获取平台支付记录
func (c *PlatformClient) FetchPayments(ctx context.Context, date time.Time, channel string) ([]*service.PlatformPayment, error) {
	// Build request URL
	url := fmt.Sprintf("%s/internal/payments/reconciliation?date=%s&channel=%s",
		c.paymentGatewayURL,
		date.Format("2006-01-02"),
		channel,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Name", "reconciliation-service")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	// Parse response
	var response PaymentListResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	if response.Code != "SUCCESS" {
		return nil, fmt.Errorf("api error: %s", response.Message)
	}

	// Convert to service model
	var payments []*service.PlatformPayment
	for _, p := range response.Data.Payments {
		merchantID, _ := uuid.Parse(p.MerchantID)
		payments = append(payments, &service.PlatformPayment{
			PaymentNo:      p.PaymentNo,
			ChannelTradeNo: p.ChannelTradeNo,
			OrderNo:        p.OrderNo,
			MerchantID:     &merchantID,
			Amount:         p.Amount,
			Currency:       p.Currency,
			Status:         p.Status,
			PaymentTime:    p.CreatedAt,
		})
	}

	return payments, nil
}

// Response DTOs

type PaymentListResponse struct {
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Data    PaymentListData `json:"data"`
}

type PaymentListData struct {
	Payments []*PaymentDTO `json:"payments"`
	Total    int64         `json:"total"`
}

type PaymentDTO struct {
	PaymentNo      string    `json:"payment_no"`
	ChannelTradeNo string    `json:"channel_trade_no"`
	OrderNo        string    `json:"order_no"`
	MerchantID     string    `json:"merchant_id"`
	Amount         int64     `json:"amount"`
	Currency       string    `json:"currency"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}
