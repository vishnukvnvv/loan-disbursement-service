package payment

import (
	"context"
	"errors"
	"payment-gateway/db/schema"
	"payment-gateway/failures"
	"payment-gateway/models"
	db_test "payment-gateway/test/db"
	utils_test "payment-gateway/test/utils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestNEFTProvider_ValidateLimit(t *testing.T) {
	t.Run("always returns nil (no limit validation)", func(t *testing.T) {
		mockTransaction := new(db_test.MockTransactionRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		processor := make(chan models.ProcessorMessage, 1)
		provider := NewNEFTProvider(processor, 0.0, 0.995, 0.0, mockTransaction, mockIdGenerator)

		err := provider.ValidateLimit(1000000.0)
		assert.NoError(t, err)
	})

	t.Run("returns nil for zero amount", func(t *testing.T) {
		mockTransaction := new(db_test.MockTransactionRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		processor := make(chan models.ProcessorMessage, 1)
		provider := NewNEFTProvider(processor, 0.0, 0.995, 0.0, mockTransaction, mockIdGenerator)

		err := provider.ValidateLimit(0.0)
		assert.NoError(t, err)
	})

	t.Run("returns nil for very large amount", func(t *testing.T) {
		mockTransaction := new(db_test.MockTransactionRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		processor := make(chan models.ProcessorMessage, 1)
		provider := NewNEFTProvider(processor, 0.0, 0.995, 0.0, mockTransaction, mockIdGenerator)

		err := provider.ValidateLimit(1000000000.0)
		assert.NoError(t, err)
	})
}

func TestNEFTProvider_GetTransaction(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully retrieves transaction", func(t *testing.T) {
		mockTransaction := new(db_test.MockTransactionRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		processor := make(chan models.ProcessorMessage, 1)
		provider := NewNEFTProvider(processor, 0.0, 0.995, 0.0, mockTransaction, mockIdGenerator)

		transactionID := "NEFT-TXN-123456789012"
		schemaTransaction := schema.Transaction{
			ID:              transactionID,
			ReferenceID:     "REF-123",
			Amount:          1000000.0,
			Channel:         models.PaymentChannelNEFT,
			Fee:             0.0,
			BeneficiaryName: "John Doe",
			AccountNumber:   "1234567890",
			IFSCCode:        "IFSC0001234",
			BankName:        "Test Bank",
			Status:          models.TransactionStatusSuccess,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		mockTransaction.On("Get", ctx, transactionID).Return(schemaTransaction, nil).Once()

		result, err := provider.GetTransaction(ctx, transactionID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, transactionID, result.ID)
		assert.Equal(t, "REF-123", result.ReferenceID)
		assert.Equal(t, 1000000.0, result.Amount)
		assert.Equal(t, models.PaymentChannelNEFT, result.Channel)
		assert.Equal(t, 0.0, result.Fee)
		assert.Equal(t, "John Doe", result.Beneficiary.Name)
		assert.Equal(t, "1234567890", result.Beneficiary.Account)
		assert.Equal(t, "IFSC0001234", result.Beneficiary.IFSC)
		assert.Equal(t, "Test Bank", result.Beneficiary.Bank)
		assert.Equal(t, models.TransactionStatusSuccess, result.Status)

		mockTransaction.AssertExpectations(t)
	})

	t.Run("returns TRANSACTION_NOT_FOUND when transaction does not exist", func(t *testing.T) {
		mockTransaction := new(db_test.MockTransactionRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		processor := make(chan models.ProcessorMessage, 1)
		provider := NewNEFTProvider(processor, 0.0, 0.995, 0.0, mockTransaction, mockIdGenerator)

		transactionID := "NEFT-TXN-NONEXISTENT"

		mockTransaction.On("Get", ctx, transactionID).
			Return(schema.Transaction{}, gorm.ErrRecordNotFound).Once()

		result, err := provider.GetTransaction(ctx, transactionID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, failures.TRANSACTION_NOT_FOUND, err)

		mockTransaction.AssertExpectations(t)
	})

	t.Run(
		"returns error when repository get fails with non-record-not-found error",
		func(t *testing.T) {
			mockTransaction := new(db_test.MockTransactionRepository)
			mockIdGenerator := new(utils_test.MockIdGenerator)

			processor := make(chan models.ProcessorMessage, 1)
			provider := NewNEFTProvider(
				processor,
				0.0,
				0.995,
				0.0,
				mockTransaction,
				mockIdGenerator,
			)

			transactionID := "NEFT-TXN-123456789012"
			repoError := errors.New("database connection error")

			mockTransaction.On("Get", ctx, transactionID).
				Return(schema.Transaction{}, repoError).Once()

			result, err := provider.GetTransaction(ctx, transactionID)

			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), "failed to get transaction")

			mockTransaction.AssertExpectations(t)
		},
	)
}

func TestNEFTProvider_Transfer(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully creates and returns transaction", func(t *testing.T) {
		mockTransaction := new(db_test.MockTransactionRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		processor := make(chan models.ProcessorMessage, 1)
		provider := NewNEFTProvider(processor, 0.0, 0.995, 0.0, mockTransaction, mockIdGenerator)

		transactionID := "NEFT-TXN-123456789012"
		request := models.PaymentRequest{
			ReferenceID: "REF-123",
			Amount:      1000000.0,
			Channel:     models.PaymentChannelNEFT,
			Beneficiary: models.Beneficiary{
				Name:    "John Doe",
				Account: "1234567890",
				IFSC:    "IFSC0001234",
				Bank:    "Test Bank",
			},
			Metadata: map[string]any{
				"key": "value",
			},
		}

		expectedSchemaTransaction := schema.Transaction{
			ID:              transactionID,
			ReferenceID:     request.ReferenceID,
			Amount:          request.Amount,
			Channel:         models.PaymentChannelNEFT,
			Fee:             0.0,
			BeneficiaryName: request.Beneficiary.Name,
			AccountNumber:   request.Beneficiary.Account,
			IFSCCode:        request.Beneficiary.IFSC,
			BankName:        request.Beneficiary.Bank,
			Metadata:        request.Metadata,
			Status:          models.TransactionStatusInitiated,
			Message:         nil,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		mockTransaction.On("Create", ctx, mock.AnythingOfType("schema.Transaction")).
			Return(expectedSchemaTransaction, nil).Once()

		result, err := provider.Transfer(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, request.ReferenceID, result.ReferenceID)
		assert.Equal(t, request.Amount, result.Amount)
		assert.Equal(t, models.PaymentChannelNEFT, result.Channel)
		assert.Equal(t, 0.0, result.Fee)
		assert.Equal(t, request.Beneficiary.Name, result.Beneficiary.Name)
		assert.Equal(t, request.Beneficiary.Account, result.Beneficiary.Account)
		assert.Equal(t, request.Beneficiary.IFSC, result.Beneficiary.IFSC)
		assert.Equal(t, request.Beneficiary.Bank, result.Beneficiary.Bank)
		assert.Equal(t, request.Metadata, result.Metadata)
		assert.Equal(t, models.TransactionStatusInitiated, result.Status)
		assert.Nil(t, result.Message)

		mockTransaction.AssertExpectations(t)
	})

	t.Run("successfully creates transaction with nil metadata", func(t *testing.T) {
		mockTransaction := new(db_test.MockTransactionRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		processor := make(chan models.ProcessorMessage, 1)
		provider := NewNEFTProvider(processor, 0.0, 0.995, 0.0, mockTransaction, mockIdGenerator)

		transactionID := "NEFT-TXN-123456789012"
		request := models.PaymentRequest{
			ReferenceID: "REF-123",
			Amount:      1000000.0,
			Channel:     models.PaymentChannelNEFT,
			Beneficiary: models.Beneficiary{
				Name:    "John Doe",
				Account: "1234567890",
				IFSC:    "IFSC0001234",
				Bank:    "Test Bank",
			},
			Metadata: nil,
		}

		expectedSchemaTransaction := schema.Transaction{
			ID:              transactionID,
			ReferenceID:     request.ReferenceID,
			Amount:          request.Amount,
			Channel:         models.PaymentChannelNEFT,
			Fee:             0.0,
			BeneficiaryName: request.Beneficiary.Name,
			AccountNumber:   request.Beneficiary.Account,
			IFSCCode:        request.Beneficiary.IFSC,
			BankName:        request.Beneficiary.Bank,
			Metadata:        nil,
			Status:          models.TransactionStatusInitiated,
			Message:         nil,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		mockTransaction.On("Create", ctx, mock.AnythingOfType("schema.Transaction")).
			Return(expectedSchemaTransaction, nil).Once()

		result, err := provider.Transfer(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.Metadata)

		mockTransaction.AssertExpectations(t)
	})

	t.Run("returns error when repository create fails", func(t *testing.T) {
		mockTransaction := new(db_test.MockTransactionRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		processor := make(chan models.ProcessorMessage, 1)
		provider := NewNEFTProvider(processor, 0.0, 0.995, 0.0, mockTransaction, mockIdGenerator)

		request := models.PaymentRequest{
			ReferenceID: "REF-123",
			Amount:      1000000.0,
			Channel:     models.PaymentChannelNEFT,
			Beneficiary: models.Beneficiary{
				Name:    "John Doe",
				Account: "1234567890",
				IFSC:    "IFSC0001234",
				Bank:    "Test Bank",
			},
			Metadata: nil,
		}

		repoError := errors.New("database connection error")

		mockTransaction.On("Create", ctx, mock.AnythingOfType("schema.Transaction")).
			Return(schema.Transaction{}, repoError).Once()

		result, err := provider.Transfer(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to create transaction")
		assert.Contains(t, err.Error(), "database connection error")

		mockTransaction.AssertExpectations(t)
	})

	t.Run("creates transaction with correct fee from provider", func(t *testing.T) {
		mockTransaction := new(db_test.MockTransactionRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		fee := 2.0
		processor := make(chan models.ProcessorMessage, 1)
		provider := NewNEFTProvider(processor, 0.0, 0.995, fee, mockTransaction, mockIdGenerator)

		request := models.PaymentRequest{
			ReferenceID: "REF-123",
			Amount:      1000000.0,
			Channel:     models.PaymentChannelNEFT,
			Beneficiary: models.Beneficiary{
				Name:    "John Doe",
				Account: "1234567890",
				IFSC:    "IFSC0001234",
				Bank:    "Test Bank",
			},
			Metadata: nil,
		}

		expectedSchemaTransaction := schema.Transaction{
			ID:              "NEFT-TXN-123456789012",
			ReferenceID:     request.ReferenceID,
			Amount:          request.Amount,
			Channel:         models.PaymentChannelNEFT,
			Fee:             fee,
			BeneficiaryName: request.Beneficiary.Name,
			AccountNumber:   request.Beneficiary.Account,
			IFSCCode:        request.Beneficiary.IFSC,
			BankName:        request.Beneficiary.Bank,
			Metadata:        nil,
			Status:          models.TransactionStatusInitiated,
			Message:         nil,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		mockTransaction.On("Create", ctx, mock.MatchedBy(func(tx schema.Transaction) bool {
			return tx.Fee == fee &&
				tx.ReferenceID == request.ReferenceID &&
				tx.Amount == request.Amount &&
				tx.Status == models.TransactionStatusInitiated
		})).Return(expectedSchemaTransaction, nil).Once()

		result, err := provider.Transfer(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, fee, result.Fee)

		mockTransaction.AssertExpectations(t)
	})

	t.Run("creates transaction with zero amount", func(t *testing.T) {
		mockTransaction := new(db_test.MockTransactionRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		processor := make(chan models.ProcessorMessage, 1)
		provider := NewNEFTProvider(processor, 0.0, 0.995, 0.0, mockTransaction, mockIdGenerator)

		request := models.PaymentRequest{
			ReferenceID: "REF-123",
			Amount:      0.0,
			Channel:     models.PaymentChannelNEFT,
			Beneficiary: models.Beneficiary{
				Name:    "John Doe",
				Account: "1234567890",
				IFSC:    "IFSC0001234",
				Bank:    "Test Bank",
			},
			Metadata: nil,
		}

		expectedSchemaTransaction := schema.Transaction{
			ID:              "NEFT-TXN-123456789012",
			ReferenceID:     request.ReferenceID,
			Amount:          0.0,
			Channel:         models.PaymentChannelNEFT,
			Fee:             0.0,
			BeneficiaryName: request.Beneficiary.Name,
			AccountNumber:   request.Beneficiary.Account,
			IFSCCode:        request.Beneficiary.IFSC,
			BankName:        request.Beneficiary.Bank,
			Metadata:        nil,
			Status:          models.TransactionStatusInitiated,
			Message:         nil,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		mockTransaction.On("Create", ctx, mock.MatchedBy(func(tx schema.Transaction) bool {
			return tx.Amount == 0.0
		})).Return(expectedSchemaTransaction, nil).Once()

		result, err := provider.Transfer(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 0.0, result.Amount)

		mockTransaction.AssertExpectations(t)
	})
}
