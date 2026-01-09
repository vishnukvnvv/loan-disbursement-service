package db_test

import (
	"payment-gateway/db/daos"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockDatabase is a mock implementation of Database
type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) GetDB() *gorm.DB {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*gorm.DB)
}

func (m *MockDatabase) GetAccountRepository() daos.AccountRepository {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(daos.AccountRepository)
}

func (m *MockDatabase) GetPaymentChannelRepository() daos.PaymentChannelRepository {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(daos.PaymentChannelRepository)
}

func (m *MockDatabase) GetTransactionRepository() daos.TransactionRepository {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(daos.TransactionRepository)
}
