package client

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// RiskClient Risk服务客户端
type RiskClient struct {
	*ServiceClient
}

// NewRiskClient 创建Risk服务客户端（带熔断器）
func NewRiskClient(baseURL string) *RiskClient {
	return &RiskClient{
		ServiceClient: NewServiceClientWithBreaker(baseURL, "risk-service"),
	}
}

// RiskCheckRequest 风控检查请求
type RiskCheckRequest struct {
	MerchantID    uuid.UUID `json:"merchant_id"`
	PaymentNo     string    `json:"payment_no"`
	Amount        int64     `json:"amount"`
	Currency      string    `json:"currency"`
	Channel       string    `json:"channel"`
	PayMethod     string    `json:"pay_method"`
	CustomerEmail string    `json:"customer_email"`
	CustomerName  string    `json:"customer_name"`
	CustomerPhone string    `json:"customer_phone"`
	CustomerIP    string    `json:"customer_ip"`
	DeviceID      string    `json:"device_id"`
	UserAgent     string    `json:"user_agent"`
}

// RiskCheckResponse 风控检查响应
type RiskCheckResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    *RiskResult `json:"data"`
}

// RiskResult 风控结果
type RiskResult struct {
	Decision    string                 `json:"decision"`     // pass, review, reject
	Score       int                    `json:"score"`        // 风险评分 0-100
	Reasons     []string               `json:"reasons"`      // 风险原因
	RiskLevel   string                 `json:"risk_level"`   // low, medium, high
	Suggestions []string               `json:"suggestions"`  // 建议
	Extra       map[string]interface{} `json:"extra"`        // 扩展信息
}

// CheckRisk 风控检查
func (c *RiskClient) CheckRisk(ctx context.Context, req *RiskCheckRequest) (*RiskResult, error) {
	resp, err := c.http.Post(ctx, "/api/v1/risk/check", req, nil)
	if err != nil {
		return nil, fmt.Errorf("调用Risk服务失败: %w", err)
	}

	var result RiskCheckResponse
	if err := resp.ParseResponse(&result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("风控检查失败: %s", result.Message)
	}

	return result.Data, nil
}

// ReportPaymentResult 上报支付结果（用于风控模型训练）
func (c *RiskClient) ReportPaymentResult(ctx context.Context, paymentNo string, success bool, fraudulent bool) error {
	req := map[string]interface{}{
		"payment_no":  paymentNo,
		"success":     success,
		"fraudulent":  fraudulent,
	}

	resp, err := c.http.Post(ctx, "/api/v1/risk/report", req, nil)
	if err != nil {
		return fmt.Errorf("调用Risk服务失败: %w", err)
	}

	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := resp.ParseResponse(&result); err != nil {
		return err
	}

	if result.Code != 0 {
		return fmt.Errorf("上报支付结果失败: %s", result.Message)
	}

	return nil
}
