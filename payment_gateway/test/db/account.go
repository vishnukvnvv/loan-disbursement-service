package db_test

import (
	"context"
	"payment-gateway/db/schema"

	"github.com/stretchr/testify/mock"
)

// Mock AccountRepository
type MockAccountRepository struct {
	mock.Mock
}

func (m *MockAccountRepository) Create(
	ctx context.Context,
	id, name string,
	balance, threshold float64,
) (*schema.Account, error) {
	args := m.Called(ctx, id, name, balance, threshold)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*schema.Account), args.Error(1)
}

func (m *MockAccountRepository) List(ctx context.Context) ([]schema.Account, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]schema.Account), args.Error(1)
}

func (m *MockAccountRepository) Update(
	ctx context.Context,
	id string,
	fields map[string]any,
) (*schema.Account, error) {
	args := m.Called(ctx, id, fields)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*schema.Account), args.Error(1)
}

func (m *MockAccountRepository) Get(ctx context.Context, id string) (*schema.Account, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*schema.Account), args.Error(1)
}
