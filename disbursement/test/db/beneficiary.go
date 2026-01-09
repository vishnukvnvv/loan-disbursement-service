package db_test

import (
	"context"
	"loan-disbursement-service/db/schema"

	"github.com/stretchr/testify/mock"
)

// Mock BeneficiaryRepository
type MockBeneficiaryRepository struct {
	mock.Mock
}

func (m *MockBeneficiaryRepository) Create(
	ctx context.Context,
	id, account, ifsc, bank string,
) (*schema.Beneficiary, error) {
	args := m.Called(ctx, id, account, ifsc, bank)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*schema.Beneficiary), args.Error(1)
}

func (m *MockBeneficiaryRepository) CreateOrGet(
	ctx context.Context,
	id, name, account, ifsc, bank string,
) (*schema.Beneficiary, error) {
	args := m.Called(ctx, id, name, account, ifsc, bank)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*schema.Beneficiary), args.Error(1)
}

func (m *MockBeneficiaryRepository) Get(
	ctx context.Context,
	account, ifsc, bank string,
) (*schema.Beneficiary, error) {
	args := m.Called(ctx, account, ifsc, bank)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*schema.Beneficiary), args.Error(1)
}

func (m *MockBeneficiaryRepository) GetById(
	ctx context.Context,
	id string,
) (*schema.Beneficiary, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*schema.Beneficiary), args.Error(1)
}

func (m *MockBeneficiaryRepository) Update(
	ctx context.Context,
	id string,
	fields map[string]any,
) error {
	args := m.Called(ctx, id, fields)
	return args.Error(0)
}
