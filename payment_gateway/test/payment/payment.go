package payment_test

import (
	"context"
	"payment-gateway/models"

	"github.com/stretchr/testify/mock"
)

// Mock PaymentProvider
type MockPaymentProvider struct {
	mock.Mock
}

func (m *MockPaymentProvider) ValidateLimit(amount float64) error {
	args := m.Called(amount)
	return args.Error(0)
}

func (m *MockPaymentProvider) GetTransaction(
	ctx context.Context,
	transactionID string,
) (*models.Transaction, error) {
	args := m.Called(ctx, transactionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transaction), args.Error(1)
}

func (m *MockPaymentProvider) Transfer(
	ctx context.Context,
	request models.PaymentRequest,
) (*models.Transaction, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transaction), args.Error(1)
}
