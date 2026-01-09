package db_test

import (
	"context"
	"loan-disbursement-service/db/schema"
	"loan-disbursement-service/models"

	"github.com/stretchr/testify/mock"
)

// Mock DisbursementRepository
type MockDisbursementRepository struct {
	mock.Mock
}

func (m *MockDisbursementRepository) Create(
	ctx context.Context,
	id, loanId string,
	channel models.PaymentChannel,
	status models.DisbursementStatus,
	amount float64,
) (*schema.Disbursement, error) {
	args := m.Called(ctx, id, loanId, status, amount)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*schema.Disbursement), args.Error(1)
}

func (m *MockDisbursementRepository) Update(
	ctx context.Context,
	id string,
	fields map[string]any,
) error {
	args := m.Called(ctx, id, fields)
	return args.Error(0)
}

func (m *MockDisbursementRepository) Get(
	ctx context.Context,
	id string,
) (*schema.Disbursement, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*schema.Disbursement), args.Error(1)
}

func (m *MockDisbursementRepository) List(
	ctx context.Context,
	offset, limit int,
	status []models.DisbursementStatus,
	channels []models.PaymentChannel,
) ([]schema.Disbursement, error) {
	args := m.Called(ctx, offset, limit, status, channels)
	if args.Get(0) == nil {
		return []schema.Disbursement{}, args.Error(1)
	}
	return args.Get(0).([]schema.Disbursement), args.Error(1)
}

func (m *MockDisbursementRepository) ListByLoan(
	ctx context.Context,
	loanId string,
) ([]schema.Disbursement, error) {
	args := m.Called(ctx, loanId)
	return args.Get(0).([]schema.Disbursement), args.Error(1)
}
