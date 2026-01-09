package utils_test

import "github.com/stretchr/testify/mock"

// Mock IdGenerator
type MockIdGenerator struct {
	mock.Mock
}

func (m *MockIdGenerator) GenerateAccountId() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockIdGenerator) GeneratePaymentChannelId() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockIdGenerator) GenerateUPITransactionId() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockIdGenerator) GenerateNEFTTransactionId() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockIdGenerator) GenerateIMPSTransactionId() string {
	args := m.Called()
	return args.String(0)
}
