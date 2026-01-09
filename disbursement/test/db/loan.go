package db_test

import (
	"context"
	"loan-disbursement-service/db/schema"

	"github.com/stretchr/testify/mock"
)

// Mock LoanRepository
type MockLoanRepository struct {
	mock.Mock
}

func (m *MockLoanRepository) Create(
	ctx context.Context,
	id string,
	amount float64,
) (*schema.Loan, error) {
	args := m.Called(ctx, id, amount)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*schema.Loan), args.Error(1)
}

func (m *MockLoanRepository) Update(
	ctx context.Context,
	id string,
	fields map[string]any,
) (*schema.Loan, error) {
	args := m.Called(ctx, id, fields)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*schema.Loan), args.Error(1)
}

func (m *MockLoanRepository) Get(ctx context.Context, id string) (*schema.Loan, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*schema.Loan), args.Error(1)
}

func (m *MockLoanRepository) List(ctx context.Context) ([]schema.Loan, error) {
	args := m.Called(ctx)
	return args.Get(0).([]schema.Loan), args.Error(1)
}
