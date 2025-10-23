package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"payment-platform/payment-gateway/internal/client"
	"payment-platform/payment-gateway/internal/model"
	"payment-platform/payment-gateway/internal/service/mocks"
)

// setupTestService creates a payment service with mocked dependencies for testing
func setupTestService() (*paymentService, *mocks.MockPaymentRepository, *mocks.MockOrderClient, *mocks.MockChannelClient, *mocks.MockRiskClient) {
	mockRepo := new(mocks.MockPaymentRepository)
	mockOrderClient := new(mocks.MockOrderClient)
	mockChannelClient := new(mocks.MockChannelClient)
	mockRiskClient := new(mocks.MockRiskClient)

	// Create a nil DB and Redis for testing (not used in mocked scenarios)
	service := &paymentService{
		db:            nil, // Will be mocked via repository
		paymentRepo:   mockRepo,
		orderClient:   mockOrderClient,
		channelClient: mockChannelClient,
		riskClient:    mockRiskClient,
		redisClient:   nil, // Not needed for these tests
	}

	return service, mockRepo, mockOrderClient, mockChannelClient, mockRiskClient
}

func TestCreatePayment_Success(t *testing.T) {
	// Arrange
	service, mockRepo, mockOrderClient, mockChannelClient, mockRiskClient := setupTestService()
	ctx := context.Background()

	merchantID := uuid.New()
	input := &CreatePaymentInput{
		MerchantID:    merchantID,
		OrderNo:       "ORDER-001",
		Amount:        10000, // $100.00
		Currency:      "USD",
		Channel:       "stripe",
		PayMethod:     "credit_card",
		CustomerEmail: "test@example.com",
		CustomerName:  "Test User",
		Description:   "Test payment",
		NotifyURL:     "https://merchant.com/notify",
		ReturnURL:     "https://merchant.com/return",
	}

	// Mock: No existing order
	mockRepo.On("GetByOrderNo", ctx, merchantID, "ORDER-001").Return(nil, gorm.ErrRecordNotFound)

	// Mock: Risk check passes
	mockRiskClient.On("CheckRisk", ctx, mock.AnythingOfType("*client.RiskCheckRequest")).Return(&client.RiskCheckResponse{
		Decision: "approve",
		Score:    85,
		Reasons:  []string{},
	}, nil)

	// Mock: Create payment record
	mockRepo.On("Create", ctx, mock.AnythingOfType("*model.Payment")).Return(nil)

	// Mock: Create order succeeds
	mockOrderClient.On("CreateOrder", ctx, mock.AnythingOfType("*client.CreateOrderRequest")).Return(&client.CreateOrderResponse{
		OrderID: uuid.New().String(),
		Status:  "pending",
	}, nil)

	// Mock: Channel payment succeeds
	mockChannelClient.On("CreatePayment", ctx, mock.AnythingOfType("*client.CreatePaymentRequest")).Return(&client.CreatePaymentResponse{
		ChannelOrderNo: "stripe_pi_123456",
		PaymentURL:     "https://checkout.stripe.com/pay/123",
		Status:         "pending",
	}, nil)

	// Mock: Update payment with channel info
	mockRepo.On("Update", ctx, mock.AnythingOfType("*model.Payment")).Return(nil)

	// Act
	payment, err := service.CreatePayment(ctx, input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, payment)
	assert.Equal(t, merchantID, payment.MerchantID)
	assert.Equal(t, "ORDER-001", payment.OrderNo)
	assert.Equal(t, int64(10000), payment.Amount)
	assert.Equal(t, "USD", payment.Currency)
	assert.NotEmpty(t, payment.PaymentNo)

	mockRepo.AssertExpectations(t)
	mockOrderClient.AssertExpectations(t)
	mockChannelClient.AssertExpectations(t)
	mockRiskClient.AssertExpectations(t)
}

func TestCreatePayment_InvalidCurrency(t *testing.T) {
	// Arrange
	service, _, _, _, _ := setupTestService()
	ctx := context.Background()

	input := &CreatePaymentInput{
		MerchantID: uuid.New(),
		OrderNo:    "ORDER-002",
		Amount:     10000,
		Currency:   "INVALID", // Invalid currency
		Channel:    "stripe",
	}

	// Act
	payment, err := service.CreatePayment(ctx, input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, payment)
	assert.Contains(t, err.Error(), "不支持的货币类型")
}

func TestCreatePayment_DuplicateOrder(t *testing.T) {
	// Arrange
	service, mockRepo, _, _, _ := setupTestService()
	ctx := context.Background()

	merchantID := uuid.New()
	input := &CreatePaymentInput{
		MerchantID: merchantID,
		OrderNo:    "ORDER-003",
		Amount:     10000,
		Currency:   "USD",
		Channel:    "stripe",
	}

	// Mock: Existing payment found
	existingPayment := &model.Payment{
		ID:         uuid.New(),
		MerchantID: merchantID,
		OrderNo:    "ORDER-003",
		PaymentNo:  "PAY-EXISTING",
		Status:     model.PaymentStatusSuccess,
	}
	mockRepo.On("GetByOrderNo", ctx, merchantID, "ORDER-003").Return(existingPayment, nil)

	// Act
	payment, err := service.CreatePayment(ctx, input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, payment)
	assert.Contains(t, err.Error(), "订单号已存在")

	mockRepo.AssertExpectations(t)
}

func TestCreatePayment_RiskRejected(t *testing.T) {
	// Arrange
	service, mockRepo, _, _, mockRiskClient := setupTestService()
	ctx := context.Background()

	merchantID := uuid.New()
	input := &CreatePaymentInput{
		MerchantID: merchantID,
		OrderNo:    "ORDER-004",
		Amount:     10000,
		Currency:   "USD",
		Channel:    "stripe",
	}

	// Mock: No existing order
	mockRepo.On("GetByOrderNo", ctx, merchantID, "ORDER-004").Return(nil, gorm.ErrRecordNotFound)

	// Mock: Risk check rejects
	mockRiskClient.On("CheckRisk", ctx, mock.AnythingOfType("*client.RiskCheckRequest")).Return(&client.RiskCheckResponse{
		Decision: "reject",
		Score:    30,
		Reasons:  []string{"高风险交易", "异常IP地址"},
	}, nil)

	// Act
	payment, err := service.CreatePayment(ctx, input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, payment)
	assert.Contains(t, err.Error(), "风控拒绝")

	mockRepo.AssertExpectations(t)
	mockRiskClient.AssertExpectations(t)
}

func TestCreatePayment_OrderCreationFailed(t *testing.T) {
	// Arrange
	service, mockRepo, mockOrderClient, _, mockRiskClient := setupTestService()
	ctx := context.Background()

	merchantID := uuid.New()
	input := &CreatePaymentInput{
		MerchantID: merchantID,
		OrderNo:    "ORDER-005",
		Amount:     10000,
		Currency:   "USD",
		Channel:    "stripe",
	}

	// Mock: No existing order
	mockRepo.On("GetByOrderNo", ctx, merchantID, "ORDER-005").Return(nil, gorm.ErrRecordNotFound)

	// Mock: Risk check passes
	mockRiskClient.On("CheckRisk", ctx, mock.AnythingOfType("*client.RiskCheckRequest")).Return(&client.RiskCheckResponse{
		Decision: "approve",
		Score:    85,
	}, nil)

	// Mock: Create payment record
	mockRepo.On("Create", ctx, mock.AnythingOfType("*model.Payment")).Return(nil)

	// Mock: Order creation fails
	mockOrderClient.On("CreateOrder", ctx, mock.AnythingOfType("*client.CreateOrderRequest")).Return(nil, fmt.Errorf("order service unavailable"))

	// Mock: Update payment to failed status
	mockRepo.On("Update", ctx, mock.MatchedBy(func(p *model.Payment) bool {
		return p.Status == model.PaymentStatusFailed
	})).Return(nil)

	// Act
	payment, err := service.CreatePayment(ctx, input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, payment)
	assert.Contains(t, err.Error(), "创建订单失败")

	mockRepo.AssertExpectations(t)
	mockOrderClient.AssertExpectations(t)
	mockRiskClient.AssertExpectations(t)
}

func TestCreatePayment_ChannelPaymentFailed(t *testing.T) {
	// Arrange
	service, mockRepo, mockOrderClient, mockChannelClient, mockRiskClient := setupTestService()
	ctx := context.Background()

	merchantID := uuid.New()
	input := &CreatePaymentInput{
		MerchantID: merchantID,
		OrderNo:    "ORDER-006",
		Amount:     10000,
		Currency:   "USD",
		Channel:    "stripe",
	}

	// Mock: No existing order
	mockRepo.On("GetByOrderNo", ctx, merchantID, "ORDER-006").Return(nil, gorm.ErrRecordNotFound)

	// Mock: Risk check passes
	mockRiskClient.On("CheckRisk", ctx, mock.AnythingOfType("*client.RiskCheckRequest")).Return(&client.RiskCheckResponse{
		Decision: "approve",
		Score:    85,
	}, nil)

	// Mock: Create payment record
	mockRepo.On("Create", ctx, mock.AnythingOfType("*model.Payment")).Return(nil)

	// Mock: Order creation succeeds
	mockOrderClient.On("CreateOrder", ctx, mock.AnythingOfType("*client.CreateOrderRequest")).Return(&client.CreateOrderResponse{
		OrderID: uuid.New().String(),
		Status:  "pending",
	}, nil)

	// Mock: Channel payment fails
	mockChannelClient.On("CreatePayment", ctx, mock.AnythingOfType("*client.CreatePaymentRequest")).Return(nil, fmt.Errorf("stripe api error: card declined"))

	// Mock: Update payment to failed status
	mockRepo.On("Update", ctx, mock.MatchedBy(func(p *model.Payment) bool {
		return p.Status == model.PaymentStatusFailed
	})).Return(nil)

	// Act
	payment, err := service.CreatePayment(ctx, input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, payment)
	assert.Contains(t, err.Error(), "发起支付失败")

	mockRepo.AssertExpectations(t)
	mockOrderClient.AssertExpectations(t)
	mockChannelClient.AssertExpectations(t)
	mockRiskClient.AssertExpectations(t)
}

func TestCreatePayment_WithManualReview(t *testing.T) {
	// Arrange
	service, mockRepo, mockOrderClient, mockChannelClient, mockRiskClient := setupTestService()
	ctx := context.Background()

	merchantID := uuid.New()
	input := &CreatePaymentInput{
		MerchantID: merchantID,
		OrderNo:    "ORDER-007",
		Amount:     50000, // $500.00 - higher amount
		Currency:   "USD",
		Channel:    "stripe",
	}

	// Mock: No existing order
	mockRepo.On("GetByOrderNo", ctx, merchantID, "ORDER-007").Return(nil, gorm.ErrRecordNotFound)

	// Mock: Risk check requires manual review
	mockRiskClient.On("CheckRisk", ctx, mock.AnythingOfType("*client.RiskCheckRequest")).Return(&client.RiskCheckResponse{
		Decision: "review",
		Score:    55,
		Reasons:  []string{"金额较大", "新商户"},
	}, nil)

	// Mock: Create payment record
	mockRepo.On("Create", ctx, mock.AnythingOfType("*model.Payment")).Return(nil)

	// Mock: Order creation succeeds
	mockOrderClient.On("CreateOrder", ctx, mock.AnythingOfType("*client.CreateOrderRequest")).Return(&client.CreateOrderResponse{
		OrderID: uuid.New().String(),
		Status:  "pending",
	}, nil)

	// Mock: Channel payment succeeds
	mockChannelClient.On("CreatePayment", ctx, mock.AnythingOfType("*client.CreatePaymentRequest")).Return(&client.CreatePaymentResponse{
		ChannelOrderNo: "stripe_pi_review",
		PaymentURL:     "https://checkout.stripe.com/pay/review",
		Status:         "pending",
	}, nil)

	// Mock: Update payment with channel info
	mockRepo.On("Update", ctx, mock.AnythingOfType("*model.Payment")).Return(nil)

	// Act
	payment, err := service.CreatePayment(ctx, input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, payment)
	// Payment should succeed even with manual review flag

	mockRepo.AssertExpectations(t)
	mockOrderClient.AssertExpectations(t)
	mockChannelClient.AssertExpectations(t)
	mockRiskClient.AssertExpectations(t)
}
