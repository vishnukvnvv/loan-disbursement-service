package utils_test

import "github.com/stretchr/testify/mock"

// Mock IdGenerator
type MockIdGenerator struct {
	mock.Mock
}

func (m *MockIdGenerator) GenerateLoanId() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockIdGenerator) GenerateTransactionId() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockIdGenerator) GenerateBeneficiaryId() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockIdGenerator) GenerateDisbursementId() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockIdGenerator) GenerateReferenceId() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockIdGenerator) GenerateReconciliationId() string {
	args := m.Called()
	return args.String(0)
}
