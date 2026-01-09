package worker

import (
	"context"
	"errors"
	"payment-gateway/db/schema"
	"payment-gateway/failures"
	"payment-gateway/models"
	db_test "payment-gateway/test/db"
	"testing"

	"github.com/stretchr/testify/mock"
)

func TestWorker_Process(t *testing.T) {
	ctx := context.Background()
	transactionID := "TXN-123456789012"

	t.Run("transaction not found", func(t *testing.T) {
		mockTransactionRepo := new(db_test.MockTransactionRepository)
		mockAccountRepo := new(db_test.MockAccountRepository)

		worker := &Worker{
			transaction: mockTransactionRepo,
			account:     mockAccountRepo,
		}

		message := models.ProcessorMessage{
			TransactionID: transactionID,
			SuccessRate:   1.0,
			Delay:         0,
		}

		repoError := errors.New("transaction not found")
		mockTransactionRepo.On("Get", ctx, transactionID).
			Return(schema.Transaction{}, repoError).
			Once()

		worker.Process(ctx, message)

		mockTransactionRepo.AssertExpectations(t)
		mockAccountRepo.AssertNotCalled(t, "List")
	})

	t.Run("transaction already processed", func(t *testing.T) {
		mockTransactionRepo := new(db_test.MockTransactionRepository)
		mockAccountRepo := new(db_test.MockAccountRepository)

		worker := &Worker{
			transaction: mockTransactionRepo,
			account:     mockAccountRepo,
		}

		message := models.ProcessorMessage{
			TransactionID: transactionID,
			SuccessRate:   1.0,
			Delay:         0,
		}

		transaction := schema.Transaction{
			ID:     transactionID,
			Status: models.TransactionStatusSuccess, // Already processed
		}

		mockTransactionRepo.On("Get", ctx, transactionID).Return(transaction, nil).Once()

		worker.Process(ctx, message)

		mockTransactionRepo.AssertExpectations(t)
		mockTransactionRepo.AssertNotCalled(t, "Update")
		mockAccountRepo.AssertNotCalled(t, "List")
	})

	t.Run("failed to update transaction to processing status", func(t *testing.T) {
		mockTransactionRepo := new(db_test.MockTransactionRepository)
		mockAccountRepo := new(db_test.MockAccountRepository)

		worker := &Worker{
			transaction: mockTransactionRepo,
			account:     mockAccountRepo,
		}

		message := models.ProcessorMessage{
			TransactionID: transactionID,
			SuccessRate:   1.0,
			Delay:         0,
		}

		transaction := schema.Transaction{
			ID:     transactionID,
			Status: models.TransactionStatusInitiated,
		}

		updateError := errors.New("database error")
		mockTransactionRepo.On("Get", ctx, transactionID).Return(transaction, nil).Once()
		mockTransactionRepo.On("Update", ctx, transactionID, mock.Anything).
			Return(schema.Transaction{}, updateError).Once()

		worker.Process(ctx, message)

		mockTransactionRepo.AssertExpectations(t)
		mockAccountRepo.AssertNotCalled(t, "List")
	})

	t.Run("transaction fails due to random failure (success rate)", func(t *testing.T) {
		mockTransactionRepo := new(db_test.MockTransactionRepository)
		mockAccountRepo := new(db_test.MockAccountRepository)

		worker := &Worker{
			transaction: mockTransactionRepo,
			account:     mockAccountRepo,
		}

		message := models.ProcessorMessage{
			TransactionID: transactionID,
			SuccessRate:   0.0, // 0% success rate - will always fail
			Delay:         0,
		}

		transaction := schema.Transaction{
			ID:     transactionID,
			Status: models.TransactionStatusInitiated,
		}

		processingTransaction := transaction
		processingTransaction.Status = models.TransactionStatusProcessing

		failedTransaction := transaction
		failedTransaction.Status = models.TransactionStatusFailed

		mockTransactionRepo.On("Get", ctx, transactionID).Return(transaction, nil).Once()
		mockTransactionRepo.On("Update", ctx, transactionID, mock.MatchedBy(func(fields map[string]any) bool {
			return fields["status"] == models.TransactionStatusProcessing
		})).
			Return(processingTransaction, nil).
			Once()
		mockTransactionRepo.On("Update", ctx, transactionID, mock.MatchedBy(func(fields map[string]any) bool {
			return fields["status"] == models.TransactionStatusFailed
		})).
			Return(failedTransaction, nil).
			Once()

		worker.Process(ctx, message)

		mockTransactionRepo.AssertExpectations(t)
		mockAccountRepo.AssertNotCalled(t, "List")
	})

	t.Run("failed to get accounts", func(t *testing.T) {
		mockTransactionRepo := new(db_test.MockTransactionRepository)
		mockAccountRepo := new(db_test.MockAccountRepository)

		worker := &Worker{
			transaction: mockTransactionRepo,
			account:     mockAccountRepo,
		}

		message := models.ProcessorMessage{
			TransactionID: transactionID,
			SuccessRate:   1.0,
			Delay:         0,
		}

		transaction := schema.Transaction{
			ID:     transactionID,
			Status: models.TransactionStatusInitiated,
		}

		processingTransaction := transaction
		processingTransaction.Status = models.TransactionStatusProcessing

		failedTransaction := transaction
		failedTransaction.Status = models.TransactionStatusFailed

		repoError := errors.New("database error")
		mockTransactionRepo.On("Get", ctx, transactionID).Return(transaction, nil).Once()
		mockTransactionRepo.On("Update", ctx, transactionID, mock.MatchedBy(func(fields map[string]any) bool {
			return fields["status"] == models.TransactionStatusProcessing
		})).
			Return(processingTransaction, nil).
			Once()
		mockAccountRepo.On("List", ctx).Return(nil, repoError).Once()
		mockTransactionRepo.On("Update", ctx, transactionID, mock.MatchedBy(func(fields map[string]any) bool {
			return fields["status"] == models.TransactionStatusFailed &&
				fields["message"] == failures.UNKNOWN_ERROR.Error()
		})).
			Return(failedTransaction, nil).
			Once()

		worker.Process(ctx, message)

		mockTransactionRepo.AssertExpectations(t)
		mockAccountRepo.AssertExpectations(t)
	})

	t.Run("no accounts found", func(t *testing.T) {
		mockTransactionRepo := new(db_test.MockTransactionRepository)
		mockAccountRepo := new(db_test.MockAccountRepository)

		worker := &Worker{
			transaction: mockTransactionRepo,
			account:     mockAccountRepo,
		}

		message := models.ProcessorMessage{
			TransactionID: transactionID,
			SuccessRate:   1.0,
			Delay:         0,
		}

		transaction := schema.Transaction{
			ID:     transactionID,
			Status: models.TransactionStatusInitiated,
		}

		processingTransaction := transaction
		processingTransaction.Status = models.TransactionStatusProcessing

		failedTransaction := transaction
		failedTransaction.Status = models.TransactionStatusFailed

		mockTransactionRepo.On("Get", ctx, transactionID).Return(transaction, nil).Once()
		mockTransactionRepo.On("Update", ctx, transactionID, mock.MatchedBy(func(fields map[string]any) bool {
			return fields["status"] == models.TransactionStatusProcessing
		})).
			Return(processingTransaction, nil).
			Once()
		mockAccountRepo.On("List", ctx).Return([]schema.Account{}, nil).Once()
		mockTransactionRepo.On("Update", ctx, transactionID, mock.MatchedBy(func(fields map[string]any) bool {
			return fields["status"] == models.TransactionStatusFailed &&
				fields["message"] == failures.UNKNOWN_ERROR.Error()
		})).
			Return(failedTransaction, nil).
			Once()

		worker.Process(ctx, message)

		mockTransactionRepo.AssertExpectations(t)
		mockAccountRepo.AssertExpectations(t)
	})

	t.Run("insufficient balance", func(t *testing.T) {
		mockTransactionRepo := new(db_test.MockTransactionRepository)
		mockAccountRepo := new(db_test.MockAccountRepository)

		worker := &Worker{
			transaction: mockTransactionRepo,
			account:     mockAccountRepo,
		}

		message := models.ProcessorMessage{
			TransactionID: transactionID,
			SuccessRate:   1.0,
			Delay:         0,
		}

		transaction := schema.Transaction{
			ID:     transactionID,
			Amount: 1000.0,
			Fee:    10.0,
			Status: models.TransactionStatusInitiated,
		}

		account := schema.Account{
			Id:      "ACC-123",
			Balance: 500.0, // Insufficient balance (need 1010.0)
		}

		processingTransaction := transaction
		processingTransaction.Status = models.TransactionStatusProcessing

		failedTransaction := transaction
		failedTransaction.Status = models.TransactionStatusFailed

		mockTransactionRepo.On("Get", ctx, transactionID).Return(transaction, nil).Once()
		mockTransactionRepo.On("Update", ctx, transactionID, mock.MatchedBy(func(fields map[string]any) bool {
			return fields["status"] == models.TransactionStatusProcessing
		})).
			Return(processingTransaction, nil).
			Once()
		mockAccountRepo.On("List", ctx).Return([]schema.Account{account}, nil).Once()
		mockTransactionRepo.On("Update", ctx, transactionID, mock.MatchedBy(func(fields map[string]any) bool {
			return fields["status"] == models.TransactionStatusFailed &&
				fields["message"] == failures.INSUFFICIENT_BALANCE.Error()
		})).
			Return(failedTransaction, nil).
			Once()

		worker.Process(ctx, message)

		mockTransactionRepo.AssertExpectations(t)
		mockAccountRepo.AssertExpectations(t)
	})

}
