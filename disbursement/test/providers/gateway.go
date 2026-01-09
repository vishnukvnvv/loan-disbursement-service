package provider_test

import (
	"context"
	"loan-disbursement-service/models"

	"github.com/stretchr/testify/mock"
)

type MockGatewayProvider struct {
	mock.Mock
}

func (m *MockGatewayProvider) Transfer(
	ctx context.Context,
	req models.PaymentRequest,
) (models.PaymentResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(models.PaymentResponse), args.Error(1)
}

func (m *MockGatewayProvider) Fetch(
	ctx context.Context,
	channel models.PaymentChannel,
	transactionID string,
) (models.PaymentResponse, error) {
	args := m.Called(ctx, channel, transactionID)
	return args.Get(0).(models.PaymentResponse), args.Error(1)
}

func (m *MockGatewayProvider) IsActive(
	ctx context.Context,
	channel models.PaymentChannel,
) (bool, error) {
	args := m.Called(ctx, channel)
	return args.Bool(0), args.Error(1)
}
