package service

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

func TestPaymentService_Process(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully processes payment request", func(t *testing.T) {
		mockPaymentChannel := new(db_test.MockPaymentChannelRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)
		processor := make(chan models.ProcessorMessage, 1)

		service := NewPaymentService(
			processor,
			mockPaymentChannel,
			mockTransaction,
			mockIdGenerator,
		)

		request := models.PaymentRequest{
			ReferenceID: "REF-123",
			Amount:      5000.0,
			Channel:     models.PaymentChannelUPI,
			Beneficiary: models.Beneficiary{
				Name:    "John Doe",
				Account: "1234567890",
				IFSC:    "IFSC0001234",
				Bank:    "Test Bank",
			},
			Metadata: map[string]any{"key": "value"},
		}

		paymentChannel := &schema.PaymentChannel{
			Id:          "CH-001",
			Name:        models.PaymentChannelUPI,
			Limit:       100000.0,
			SuccessRate: 0.95,
			Fee:         5.0,
		}

		expectedTransaction := &models.Transaction{
			ID:          "UPI-TXN-123456789012",
			ReferenceID: request.ReferenceID,
			Amount:      request.Amount,
			Channel:     models.PaymentChannelUPI,
			Fee:         5.0,
			Beneficiary: request.Beneficiary,
			Metadata:    request.Metadata,
			Status:      models.TransactionStatusInitiated,
		}

		mockTransaction.On("GetByReferenceID", ctx, request.ReferenceID).
			Return(nil, nil).Once()
		mockPaymentChannel.On("Get", ctx, request.Channel).
			Return(paymentChannel, nil).Once()
		mockTransaction.On("Create", ctx, mock.AnythingOfType("schema.Transaction")).
			Return(schema.Transaction{
				ID:              expectedTransaction.ID,
				ReferenceID:     expectedTransaction.ReferenceID,
				Amount:          expectedTransaction.Amount,
				Channel:         expectedTransaction.Channel,
				Fee:             expectedTransaction.Fee,
				BeneficiaryName: expectedTransaction.Beneficiary.Name,
				AccountNumber:   expectedTransaction.Beneficiary.Account,
				IFSCCode:        expectedTransaction.Beneficiary.IFSC,
				BankName:        expectedTransaction.Beneficiary.Bank,
				Metadata:        expectedTransaction.Metadata,
				Status:          expectedTransaction.Status,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}, nil).Once()

		result, err := service.Process(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, request.ReferenceID, result.ReferenceID)
		assert.Equal(t, request.Amount, result.Amount)

		mockTransaction.AssertExpectations(t)
		mockPaymentChannel.AssertExpectations(t)
	})

	t.Run("returns error when reference ID already processed", func(t *testing.T) {
		mockPaymentChannel := new(db_test.MockPaymentChannelRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)
		processor := make(chan models.ProcessorMessage, 1)

		service := NewPaymentService(
			processor,
			mockPaymentChannel,
			mockTransaction,
			mockIdGenerator,
		)

		request := models.PaymentRequest{
			ReferenceID: "REF-123",
			Amount:      5000.0,
			Channel:     models.PaymentChannelUPI,
			Beneficiary: models.Beneficiary{
				Name:    "John Doe",
				Account: "1234567890",
				IFSC:    "IFSC0001234",
				Bank:    "Test Bank",
			},
		}

		existingTransaction := &schema.Transaction{
			ID:          "UPI-TXN-EXISTING",
			ReferenceID: request.ReferenceID,
		}

		mockTransaction.On("GetByReferenceID", ctx, request.ReferenceID).
			Return(existingTransaction, nil).Once()

		result, err := service.Process(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, failures.REFERENCE_ID_ALREADY_PROCESSED, err)

		mockTransaction.AssertExpectations(t)
		mockPaymentChannel.AssertNotCalled(t, "Get")
	})

	t.Run("returns error when GetByReferenceID fails", func(t *testing.T) {
		mockPaymentChannel := new(db_test.MockPaymentChannelRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)
		processor := make(chan models.ProcessorMessage, 1)

		service := NewPaymentService(
			processor,
			mockPaymentChannel,
			mockTransaction,
			mockIdGenerator,
		)

		request := models.PaymentRequest{
			ReferenceID: "REF-123",
			Amount:      5000.0,
			Channel:     models.PaymentChannelUPI,
			Beneficiary: models.Beneficiary{
				Name:    "John Doe",
				Account: "1234567890",
				IFSC:    "IFSC0001234",
				Bank:    "Test Bank",
			},
		}

		repoError := errors.New("database error")

		mockTransaction.On("GetByReferenceID", ctx, request.ReferenceID).
			Return(nil, repoError).Once()

		result, err := service.Process(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get transaction")

		mockTransaction.AssertExpectations(t)
		mockPaymentChannel.AssertNotCalled(t, "Get")
	})

	t.Run("returns error for invalid IFSC code", func(t *testing.T) {
		mockPaymentChannel := new(db_test.MockPaymentChannelRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)
		processor := make(chan models.ProcessorMessage, 1)

		service := NewPaymentService(
			processor,
			mockPaymentChannel,
			mockTransaction,
			mockIdGenerator,
		)

		request := models.PaymentRequest{
			ReferenceID: "REF-123",
			Amount:      5000.0,
			Channel:     models.PaymentChannelUPI,
			Beneficiary: models.Beneficiary{
				Name:    "John Doe",
				Account: "1234567890",
				IFSC:    "INVALID_IFSC_CODE_1",
				Bank:    "Test Bank",
			},
		}

		mockTransaction.On("GetByReferenceID", ctx, request.ReferenceID).
			Return(nil, nil).Once()

		result, err := service.Process(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, failures.INVALID_IFSC, err)

		mockTransaction.AssertExpectations(t)
		mockPaymentChannel.AssertNotCalled(t, "Get")
	})

	t.Run("returns error for inactive account", func(t *testing.T) {
		mockPaymentChannel := new(db_test.MockPaymentChannelRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)
		processor := make(chan models.ProcessorMessage, 1)

		service := NewPaymentService(
			processor,
			mockPaymentChannel,
			mockTransaction,
			mockIdGenerator,
		)

		request := models.PaymentRequest{
			ReferenceID: "REF-123",
			Amount:      5000.0,
			Channel:     models.PaymentChannelUPI,
			Beneficiary: models.Beneficiary{
				Name:    "John Doe",
				Account: "0000000000000000",
				IFSC:    "IFSC0001234",
				Bank:    "Test Bank",
			},
		}

		mockTransaction.On("GetByReferenceID", ctx, request.ReferenceID).
			Return(nil, nil).Once()

		result, err := service.Process(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, failures.INACTIVE_ACCOUNT, err)

		mockTransaction.AssertExpectations(t)
		mockPaymentChannel.AssertNotCalled(t, "Get")
	})

	t.Run("returns error for invalid payment channel", func(t *testing.T) {
		mockPaymentChannel := new(db_test.MockPaymentChannelRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)
		processor := make(chan models.ProcessorMessage, 1)

		service := NewPaymentService(
			processor,
			mockPaymentChannel,
			mockTransaction,
			mockIdGenerator,
		)

		request := models.PaymentRequest{
			ReferenceID: "REF-123",
			Amount:      5000.0,
			Channel:     models.PaymentChannelUPI,
			Beneficiary: models.Beneficiary{
				Name:    "John Doe",
				Account: "1234567890",
				IFSC:    "IFSC0001234",
				Bank:    "Test Bank",
			},
		}

		mockTransaction.On("GetByReferenceID", ctx, request.ReferenceID).
			Return(nil, nil).Once()
		mockPaymentChannel.On("Get", ctx, request.Channel).
			Return(nil, gorm.ErrRecordNotFound).Once()

		result, err := service.Process(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, failures.INVALID_PAYMENT_CHANNEL, err)

		mockTransaction.AssertExpectations(t)
		mockPaymentChannel.AssertExpectations(t)
	})

	t.Run("returns error when payment channel repository fails", func(t *testing.T) {
		mockPaymentChannel := new(db_test.MockPaymentChannelRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)
		processor := make(chan models.ProcessorMessage, 1)

		service := NewPaymentService(
			processor,
			mockPaymentChannel,
			mockTransaction,
			mockIdGenerator,
		)

		request := models.PaymentRequest{
			ReferenceID: "REF-123",
			Amount:      5000.0,
			Channel:     models.PaymentChannelUPI,
			Beneficiary: models.Beneficiary{
				Name:    "John Doe",
				Account: "1234567890",
				IFSC:    "IFSC0001234",
				Bank:    "Test Bank",
			},
		}

		repoError := errors.New("database connection error")

		mockTransaction.On("GetByReferenceID", ctx, request.ReferenceID).
			Return(nil, nil).Once()
		mockPaymentChannel.On("Get", ctx, request.Channel).
			Return(nil, repoError).Once()

		result, err := service.Process(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get payment channel")

		mockTransaction.AssertExpectations(t)
		mockPaymentChannel.AssertExpectations(t)
	})

	t.Run("returns error when limit exceeded", func(t *testing.T) {
		mockPaymentChannel := new(db_test.MockPaymentChannelRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)
		processor := make(chan models.ProcessorMessage, 1)

		service := NewPaymentService(
			processor,
			mockPaymentChannel,
			mockTransaction,
			mockIdGenerator,
		)

		request := models.PaymentRequest{
			ReferenceID: "REF-123",
			Amount:      150000.0, // Exceeds limit
			Channel:     models.PaymentChannelUPI,
			Beneficiary: models.Beneficiary{
				Name:    "John Doe",
				Account: "1234567890",
				IFSC:    "IFSC0001234",
				Bank:    "Test Bank",
			},
		}

		paymentChannel := &schema.PaymentChannel{
			Id:          "CH-001",
			Name:        models.PaymentChannelUPI,
			Limit:       100000.0,
			SuccessRate: 0.95,
			Fee:         5.0,
		}

		mockTransaction.On("GetByReferenceID", ctx, request.ReferenceID).
			Return(nil, nil).Once()
		mockPaymentChannel.On("Get", ctx, request.Channel).
			Return(paymentChannel, nil).Once()

		result, err := service.Process(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, failures.LIMIT_EXCEEDED, err)

		mockTransaction.AssertExpectations(t)
		mockPaymentChannel.AssertExpectations(t)
	})

	t.Run("processes payment for different channels", func(t *testing.T) {
		testCases := []struct {
			name    string
			channel models.PaymentChannel
		}{
			{"UPI", models.PaymentChannelUPI},
			{"NEFT", models.PaymentChannelNEFT},
			{"IMPS", models.PaymentChannelIMPS},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				mockPaymentChannel := new(db_test.MockPaymentChannelRepository)
				mockTransaction := new(db_test.MockTransactionRepository)
				mockIdGenerator := new(utils_test.MockIdGenerator)
				processor := make(chan models.ProcessorMessage, 1)

				service := NewPaymentService(
					processor,
					mockPaymentChannel,
					mockTransaction,
					mockIdGenerator,
				)

				request := models.PaymentRequest{
					ReferenceID: "REF-123",
					Amount:      5000.0,
					Channel:     tc.channel,
					Beneficiary: models.Beneficiary{
						Name:    "John Doe",
						Account: "1234567890",
						IFSC:    "IFSC0001234",
						Bank:    "Test Bank",
					},
				}

				paymentChannel := &schema.PaymentChannel{
					Id:          "CH-001",
					Name:        tc.channel,
					Limit:       1000000.0,
					SuccessRate: 0.95,
					Fee:         5.0,
				}

				mockTransaction.On("GetByReferenceID", ctx, request.ReferenceID).
					Return(nil, nil).Once()
				mockPaymentChannel.On("Get", ctx, request.Channel).
					Return(paymentChannel, nil).Once()
				mockTransaction.On("Create", ctx, mock.AnythingOfType("schema.Transaction")).
					Return(schema.Transaction{
						ID:              "TXN-123",
						ReferenceID:     request.ReferenceID,
						Amount:          request.Amount,
						Channel:         tc.channel,
						Fee:             5.0,
						BeneficiaryName: request.Beneficiary.Name,
						AccountNumber:   request.Beneficiary.Account,
						IFSCCode:        request.Beneficiary.IFSC,
						BankName:        request.Beneficiary.Bank,
						Status:          models.TransactionStatusInitiated,
						CreatedAt:       time.Now(),
						UpdatedAt:       time.Now(),
					}, nil).Once()

				result, err := service.Process(ctx, request)

				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.channel, result.Channel)

				mockTransaction.AssertExpectations(t)
				mockPaymentChannel.AssertExpectations(t)
			})
		}
	})
}

func TestPaymentService_GetTransaction(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully retrieves transaction", func(t *testing.T) {
		mockPaymentChannel := new(db_test.MockPaymentChannelRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)
		processor := make(chan models.ProcessorMessage, 1)

		service := NewPaymentService(
			processor,
			mockPaymentChannel,
			mockTransaction,
			mockIdGenerator,
		)

		channel := models.PaymentChannelUPI
		transactionID := "UPI-TXN-123456789012"

		paymentChannel := &schema.PaymentChannel{
			Id:          "CH-001",
			Name:        channel,
			Limit:       100000.0,
			SuccessRate: 0.95,
			Fee:         5.0,
		}

		expectedTransaction := &models.Transaction{
			ID:          transactionID,
			ReferenceID: "REF-123",
			Amount:      5000.0,
			Channel:     channel,
			Fee:         5.0,
			Status:      models.TransactionStatusSuccess,
		}

		mockPaymentChannel.On("Get", ctx, channel).
			Return(paymentChannel, nil).Once()
		mockTransaction.On("Get", ctx, transactionID).
			Return(schema.Transaction{
				ID:              transactionID,
				ReferenceID:     expectedTransaction.ReferenceID,
				Amount:          expectedTransaction.Amount,
				Channel:         channel,
				Fee:             expectedTransaction.Fee,
				Status:          expectedTransaction.Status,
				BeneficiaryName: "John Doe",
				AccountNumber:   "1234567890",
				IFSCCode:        "IFSC0001234",
				BankName:        "Test Bank",
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}, nil).Once()

		result, err := service.GetTransaction(ctx, channel, transactionID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, transactionID, result.ID)
		assert.Equal(t, expectedTransaction.ReferenceID, result.ReferenceID)

		mockPaymentChannel.AssertExpectations(t)
		mockTransaction.AssertExpectations(t)
	})

	t.Run("returns error for invalid payment channel", func(t *testing.T) {
		mockPaymentChannel := new(db_test.MockPaymentChannelRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)
		processor := make(chan models.ProcessorMessage, 1)

		service := NewPaymentService(
			processor,
			mockPaymentChannel,
			mockTransaction,
			mockIdGenerator,
		)

		channel := models.PaymentChannelUPI
		transactionID := "UPI-TXN-123456789012"

		mockPaymentChannel.On("Get", ctx, channel).
			Return(nil, gorm.ErrRecordNotFound).Once()

		result, err := service.GetTransaction(ctx, channel, transactionID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, failures.INVALID_PAYMENT_CHANNEL, err)

		mockPaymentChannel.AssertExpectations(t)
		mockTransaction.AssertNotCalled(t, "Get")
	})

	t.Run("returns error when payment channel repository fails", func(t *testing.T) {
		mockPaymentChannel := new(db_test.MockPaymentChannelRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)
		processor := make(chan models.ProcessorMessage, 1)

		service := NewPaymentService(
			processor,
			mockPaymentChannel,
			mockTransaction,
			mockIdGenerator,
		)

		channel := models.PaymentChannelUPI
		transactionID := "UPI-TXN-123456789012"

		repoError := errors.New("database connection error")

		mockPaymentChannel.On("Get", ctx, channel).
			Return(nil, repoError).Once()

		result, err := service.GetTransaction(ctx, channel, transactionID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get payment channel")

		mockPaymentChannel.AssertExpectations(t)
		mockTransaction.AssertNotCalled(t, "Get")
	})

	t.Run("returns error when transaction not found", func(t *testing.T) {
		mockPaymentChannel := new(db_test.MockPaymentChannelRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)
		processor := make(chan models.ProcessorMessage, 1)

		service := NewPaymentService(
			processor,
			mockPaymentChannel,
			mockTransaction,
			mockIdGenerator,
		)

		channel := models.PaymentChannelUPI
		transactionID := "UPI-TXN-NONEXISTENT"

		paymentChannel := &schema.PaymentChannel{
			Id:          "CH-001",
			Name:        channel,
			Limit:       100000.0,
			SuccessRate: 0.95,
			Fee:         5.0,
		}

		mockPaymentChannel.On("Get", ctx, channel).
			Return(paymentChannel, nil).Once()
		mockTransaction.On("Get", ctx, transactionID).
			Return(schema.Transaction{}, gorm.ErrRecordNotFound).Once()

		result, err := service.GetTransaction(ctx, channel, transactionID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, failures.TRANSACTION_NOT_FOUND, err)

		mockPaymentChannel.AssertExpectations(t)
		mockTransaction.AssertExpectations(t)
	})
}
