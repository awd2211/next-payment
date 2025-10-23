package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"payment-platform/payment-gateway/internal/model"
	"payment-platform/payment-gateway/internal/repository"
)

// MockPaymentRepository is a mock implementation of PaymentRepository
type MockPaymentRepository struct {
	mock.Mock
}

func (m *MockPaymentRepository) Create(ctx context.Context, payment *model.Payment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) Update(ctx context.Context, payment *model.Payment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Payment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Payment), args.Error(1)
}

func (m *MockPaymentRepository) GetByPaymentNo(ctx context.Context, paymentNo string) (*model.Payment, error) {
	args := m.Called(ctx, paymentNo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Payment), args.Error(1)
}

func (m *MockPaymentRepository) GetByOrderNo(ctx context.Context, merchantID uuid.UUID, orderNo string) (*model.Payment, error) {
	args := m.Called(ctx, merchantID, orderNo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Payment), args.Error(1)
}

func (m *MockPaymentRepository) List(ctx context.Context, query *repository.PaymentQuery) ([]*model.Payment, int64, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.Payment), int64(args.Int(1)), args.Error(2)
}

func (m *MockPaymentRepository) CreateRefund(ctx context.Context, refund *model.Refund) error {
	args := m.Called(ctx, refund)
	return args.Error(0)
}

func (m *MockPaymentRepository) UpdateRefund(ctx context.Context, refund *model.Refund) error {
	args := m.Called(ctx, refund)
	return args.Error(0)
}

func (m *MockPaymentRepository) GetRefund(ctx context.Context, refundNo string) (*model.Refund, error) {
	args := m.Called(ctx, refundNo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Refund), args.Error(1)
}

func (m *MockPaymentRepository) ListRefunds(ctx context.Context, query *repository.RefundQuery) ([]*model.Refund, int64, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.Refund), int64(args.Int(1)), args.Error(2)
}

func (m *MockPaymentRepository) SaveCallback(ctx context.Context, callback *model.PaymentCallback) error {
	args := m.Called(ctx, callback)
	return args.Error(0)
}

func (m *MockPaymentRepository) GetCallbacks(ctx context.Context, paymentNo string) ([]*model.PaymentCallback, error) {
	args := m.Called(ctx, paymentNo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.PaymentCallback), args.Error(1)
}
