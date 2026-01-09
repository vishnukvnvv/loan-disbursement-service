package db_test

import (
	"context"
	"payment-gateway/db/schema"
	"payment-gateway/models"

	"github.com/stretchr/testify/mock"
)

// Mock PaymentChannelRepository
type MockPaymentChannelRepository struct {
	mock.Mock
}

func (m *MockPaymentChannelRepository) Create(
	ctx context.Context,
	id string, name models.PaymentChannel,
	limit, successRate, fee float64,
) (*schema.PaymentChannel, error) {
	args := m.Called(ctx, id, name, limit, successRate, fee)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*schema.PaymentChannel), args.Error(1)
}

func (m *MockPaymentChannelRepository) List(ctx context.Context) ([]schema.PaymentChannel, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]schema.PaymentChannel), args.Error(1)
}

func (m *MockPaymentChannelRepository) Update(
	ctx context.Context,
	channel models.PaymentChannel,
	fields map[string]any,
) (*schema.PaymentChannel, error) {
	args := m.Called(ctx, channel, fields)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*schema.PaymentChannel), args.Error(1)
}

func (m *MockPaymentChannelRepository) Get(
	ctx context.Context,
	channel models.PaymentChannel,
) (*schema.PaymentChannel, error) {
	args := m.Called(ctx, channel)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*schema.PaymentChannel), args.Error(1)
}
