package db_test

import (
	"context"
	"payment-gateway/db/schema"
	"payment-gateway/models"

	"github.com/stretchr/testify/mock"
)

// Mock TransactionRepository
type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) Create(
	ctx context.Context,
	transaction schema.Transaction,
) (schema.Transaction, error) {
	args := m.Called(ctx, transaction)
	return args.Get(0).(schema.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) Update(
	ctx context.Context,
	id string,
	fields map[string]any,
) (schema.Transaction, error) {
	args := m.Called(ctx, id, fields)
	return args.Get(0).(schema.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) Get(
	ctx context.Context,
	id string,
) (schema.Transaction, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return schema.Transaction{}, args.Error(1)
	}
	return args.Get(0).(schema.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) List(
	ctx context.Context,
	offset, limit int,
	status []models.TransactionStatus,
) ([]schema.Transaction, error) {
	args := m.Called(ctx, offset, limit, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]schema.Transaction), args.Error(1)
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
