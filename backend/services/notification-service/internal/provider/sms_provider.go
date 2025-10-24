package provider

import (
	"context"
	"fmt"

	"github.com/payment-platform/pkg/logger"
	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
	"go.uber.org/zap"
)

// SMSProvider 短信提供商接口
type SMSProvider interface {
	Send(ctx context.Context, req *SMSRequest) (*SMSResponse, error)
	GetProviderName() string
}

// SMSRequest 短信发送请求
type SMSRequest struct {
	To      string `json:"to"`      // 收件人手机号
	Content string `json:"content"` // 短信内容
	SignName string `json:"sign_name"` // 签名
}

// SMSResponse 短信发送响应
type SMSResponse struct {
	MessageID string `json:"message_id"` // 消息ID
	Status    string `json:"status"`     // 状态
}

// TwilioProvider Twilio 短信提供商
type TwilioProvider struct {
	accountSID string
	authToken  string
	from       string
	client     *twilio.RestClient
}

// NewTwilioProvider 创建 Twilio 提供商实例
func NewTwilioProvider(accountSID, authToken, from string) *TwilioProvider {
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSID,
		Password: authToken,
	})

	return &TwilioProvider{
		accountSID: accountSID,
		authToken:  authToken,
		from:       from,
		client:     client,
	}
}

// GetProviderName 获取提供商名称
func (p *TwilioProvider) GetProviderName() string {
	return "twilio"
}

// Send 发送短信
func (p *TwilioProvider) Send(ctx context.Context, req *SMSRequest) (*SMSResponse, error) {
	params := &twilioApi.CreateMessageParams{}
	params.SetTo(req.To)
	params.SetFrom(p.from)
	params.SetBody(req.Content)

	resp, err := p.client.Api.CreateMessage(params)
	if err != nil {
		return nil, fmt.Errorf("Twilio 发送短信失败: %w", err)
	}

	return &SMSResponse{
		MessageID: *resp.Sid,
		Status:    "sent",
	}, nil
}

// MockSMSProvider 模拟短信提供商（用于测试）
type MockSMSProvider struct{}

// NewMockSMSProvider 创建模拟短信提供商实例
func NewMockSMSProvider() *MockSMSProvider {
	return &MockSMSProvider{}
}

// GetProviderName 获取提供商名称
func (p *MockSMSProvider) GetProviderName() string {
	return "mock"
}

// Send 发送短信（模拟）
func (p *MockSMSProvider) Send(ctx context.Context, req *SMSRequest) (*SMSResponse, error) {
	logger.Info("mock sms sent",
		zap.String("to", req.To),
		zap.String("content", req.Content))
	return &SMSResponse{
		MessageID: "mock-" + req.To,
		Status:    "sent",
	}, nil
}

// SMSProviderFactory 短信提供商工厂
type SMSProviderFactory struct {
	providers map[string]SMSProvider
}

// NewSMSProviderFactory 创建短信提供商工厂
func NewSMSProviderFactory() *SMSProviderFactory {
	return &SMSProviderFactory{
		providers: make(map[string]SMSProvider),
	}
}

// Register 注册提供商
func (f *SMSProviderFactory) Register(name string, provider SMSProvider) {
	f.providers[name] = provider
}

// GetProvider 获取提供商
func (f *SMSProviderFactory) GetProvider(name string) (SMSProvider, bool) {
	provider, ok := f.providers[name]
	return provider, ok
}
