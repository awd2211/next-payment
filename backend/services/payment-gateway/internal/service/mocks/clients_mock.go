package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"payment-platform/payment-gateway/internal/client"
)

// MockOrderClient is a mock implementation of OrderClient
type MockOrderClient struct {
	mock.Mock
}

func (m *MockOrderClient) CreateOrder(ctx context.Context, req *client.CreateOrderRequest) (*client.CreateOrderResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.CreateOrderResponse), args.Error(1)
}

func (m *MockOrderClient) UpdateOrderStatus(ctx context.Context, req *client.UpdateOrderStatusRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

// MockChannelClient is a mock implementation of ChannelClient
type MockChannelClient struct {
	mock.Mock
}

func (m *MockChannelClient) CreatePayment(ctx context.Context, req *client.CreatePaymentRequest) (*client.CreatePaymentResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.CreatePaymentResponse), args.Error(1)
}

func (m *MockChannelClient) CreateRefund(ctx context.Context, req *client.RefundRequest) (*client.RefundResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.RefundResponse), args.Error(1)
}

// MockRiskClient is a mock implementation of RiskClient
type MockRiskClient struct {
	mock.Mock
}

func (m *MockRiskClient) CheckRisk(ctx context.Context, req *client.RiskCheckRequest) (*client.RiskCheckResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.RiskCheckResponse), args.Error(1)
}
