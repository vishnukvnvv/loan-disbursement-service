package db_test

import (
	"context"
	"loan-disbursement-service/db/schema"
	"loan-disbursement-service/models"
	"time"

	"github.com/stretchr/testify/mock"
)

// Mock TransactionRepository
type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) Create(
	ctx context.Context,
	transaction schema.Transaction,
) (*schema.Transaction, error) {
	args := m.Called(ctx, transaction)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*schema.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) Update(
	ctx context.Context,
	id string,
	fields map[string]any,
) error {
	args := m.Called(ctx, id, fields)
	return args.Error(0)
}

func (m *MockTransactionRepository) Get(
	ctx context.Context,
	id string,
) (*schema.Transaction, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*schema.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) GetByReferenceID(
	ctx context.Context,
	referenceID string,
) (*schema.Transaction, error) {
	args := m.Called(ctx, referenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*schema.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) ListByDisbursement(
	ctx context.Context,
	disbursementId string,
) ([]schema.Transaction, error) {
	args := m.Called(ctx, disbursementId)
	return args.Get(0).([]schema.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) ListByDate(
	ctx context.Context,
	date time.Time,
	status []models.TransactionStatus,
) ([]schema.Transaction, error) {
	args := m.Called(ctx, date, status)
	return args.Get(0).([]schema.Transaction), args.Error(1)
}
