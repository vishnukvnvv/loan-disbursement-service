package worker

import (
	"context"
	"errors"
	"net/http"
	"payment-gateway/db/schema"
	"payment-gateway/models"
	db_test "payment-gateway/test/db"
	http_test "payment-gateway/test/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWorker_Notify(t *testing.T) {
	ctx := context.Background()
	transactionID := "TXN-123456789012"

	t.Run("successful notification", func(t *testing.T) {
		mockTransactionRepo := new(db_test.MockTransactionRepository)
		mockHTTPClient := new(http_test.MockHTTPClient)

		worker := &Worker{
			transaction: mockTransactionRepo,
			httpClient:  mockHTTPClient,
		}

		notificationURL := "https://example.com/webhook"
		now := time.Now()
		processedAt := now.Add(5 * time.Minute)
		message := "Transaction successful"

		transaction := schema.Transaction{
			ID:          transactionID,
			ReferenceID: "REF-123",
			Amount:      1000.0,
			Fee:         10.0,
			Channel:     models.PaymentChannelUPI,
			Status:      models.TransactionStatusSuccess,
			Message:     &message,
			CreatedAt:   now,
			UpdatedAt:   now,
			ProcessedAt: &processedAt,
			Metadata: map[string]any{
				"notification_url": notificationURL,
			},
		}

		// Create a mock HTTP response
		mockResponse := &http.Response{
			StatusCode: http.StatusOK,
		}

		updatedTransaction := transaction
		notifiedAt := time.Now()
		updatedTransaction.NotifiedAt = &notifiedAt

		mockTransactionRepo.On("Get", ctx, transactionID).Return(transaction, nil).Once()
		mockHTTPClient.On("POST", ctx, notificationURL, mock.Anything, mock.Anything).
			Return(mockResponse, nil).Once()
		mockTransactionRepo.On("Update", ctx, transactionID, mock.MatchedBy(func(fields map[string]any) bool {
			_, hasNotifiedAt := fields["notified_at"]
			_, hasUpdatedAt := fields["updated_at"]
			return hasNotifiedAt && hasUpdatedAt
		})).
			Return(updatedTransaction, nil).
			Once()

		worker.Notify(ctx, transactionID)

		mockTransactionRepo.AssertExpectations(t)
		mockHTTPClient.AssertExpectations(t)
	})

	t.Run("transaction not found", func(t *testing.T) {
		mockTransactionRepo := new(db_test.MockTransactionRepository)
		mockHTTPClient := new(http_test.MockHTTPClient)

		worker := &Worker{
			transaction: mockTransactionRepo,
			httpClient:  mockHTTPClient,
		}

		repoError := errors.New("transaction not found")
		mockTransactionRepo.On("Get", ctx, transactionID).
			Return(schema.Transaction{}, repoError).
			Once()

		worker.Notify(ctx, transactionID)

		mockTransactionRepo.AssertExpectations(t)
		mockHTTPClient.AssertNotCalled(t, "POST")
		mockTransactionRepo.AssertNotCalled(t, "Update")
	})

	t.Run("notification URL not found in metadata", func(t *testing.T) {
		mockTransactionRepo := new(db_test.MockTransactionRepository)
		mockHTTPClient := new(http_test.MockHTTPClient)

		worker := &Worker{
			transaction: mockTransactionRepo,
			httpClient:  mockHTTPClient,
		}

		transaction := schema.Transaction{
			ID:        transactionID,
			Metadata:  map[string]any{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockTransactionRepo.On("Get", ctx, transactionID).Return(transaction, nil).Once()

		worker.Notify(ctx, transactionID)

		mockTransactionRepo.AssertExpectations(t)
		mockHTTPClient.AssertNotCalled(t, "POST")
		mockTransactionRepo.AssertNotCalled(t, "Update")
	})

	t.Run("notification URL is not a string", func(t *testing.T) {
		mockTransactionRepo := new(db_test.MockTransactionRepository)
		mockHTTPClient := new(http_test.MockHTTPClient)

		worker := &Worker{
			transaction: mockTransactionRepo,
			httpClient:  mockHTTPClient,
		}

		transaction := schema.Transaction{
			ID: transactionID,
			Metadata: map[string]any{
				"notification_url": 123, // Not a string
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockTransactionRepo.On("Get", ctx, transactionID).Return(transaction, nil).Once()

		worker.Notify(ctx, transactionID)

		mockTransactionRepo.AssertExpectations(t)
		mockHTTPClient.AssertNotCalled(t, "POST")
		mockTransactionRepo.AssertNotCalled(t, "Update")
	})

	t.Run("HTTP POST request fails", func(t *testing.T) {
		mockTransactionRepo := new(db_test.MockTransactionRepository)
		mockHTTPClient := new(http_test.MockHTTPClient)

		worker := &Worker{
			transaction: mockTransactionRepo,
			httpClient:  mockHTTPClient,
		}

		notificationURL := "https://example.com/webhook"
		transaction := schema.Transaction{
			ID:        transactionID,
			Metadata:  map[string]any{"notification_url": notificationURL},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		httpError := errors.New("network error")
		mockTransactionRepo.On("Get", ctx, transactionID).Return(transaction, nil).Once()
		mockHTTPClient.On("POST", ctx, notificationURL, mock.Anything, mock.Anything).
			Return(nil, httpError).Once()

		worker.Notify(ctx, transactionID)

		mockTransactionRepo.AssertExpectations(t)
		mockHTTPClient.AssertExpectations(t)
		mockTransactionRepo.AssertNotCalled(t, "Update")
	})

	t.Run("HTTP response status code indicates failure", func(t *testing.T) {
		mockTransactionRepo := new(db_test.MockTransactionRepository)
		mockHTTPClient := new(http_test.MockHTTPClient)

		worker := &Worker{
			transaction: mockTransactionRepo,
			httpClient:  mockHTTPClient,
		}

		notificationURL := "https://example.com/webhook"
		transaction := schema.Transaction{
			ID:        transactionID,
			Metadata:  map[string]any{"notification_url": notificationURL},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Create a mock HTTP response with error status
		mockResponse := &http.Response{
			StatusCode: http.StatusInternalServerError,
		}

		mockTransactionRepo.On("Get", ctx, transactionID).Return(transaction, nil).Once()
		mockHTTPClient.On("POST", ctx, notificationURL, mock.Anything, mock.Anything).
			Return(mockResponse, nil).Once()

		worker.Notify(ctx, transactionID)

		mockTransactionRepo.AssertExpectations(t)
		mockHTTPClient.AssertExpectations(t)
		mockTransactionRepo.AssertNotCalled(t, "Update")
	})

	t.Run("update notified_at fails", func(t *testing.T) {
		mockTransactionRepo := new(db_test.MockTransactionRepository)
		mockHTTPClient := new(http_test.MockHTTPClient)

		worker := &Worker{
			transaction: mockTransactionRepo,
			httpClient:  mockHTTPClient,
		}

		notificationURL := "https://example.com/webhook"
		transaction := schema.Transaction{
			ID:        transactionID,
			Metadata:  map[string]any{"notification_url": notificationURL},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockResponse := &http.Response{
			StatusCode: http.StatusOK,
		}

		updateError := errors.New("database error")
		mockTransactionRepo.On("Get", ctx, transactionID).Return(transaction, nil).Once()
		mockHTTPClient.On("POST", ctx, notificationURL, mock.Anything, mock.Anything).
			Return(mockResponse, nil).Once()
		mockTransactionRepo.On("Update", ctx, transactionID, mock.Anything).
			Return(schema.Transaction{}, updateError).Once()

		worker.Notify(ctx, transactionID)

		mockTransactionRepo.AssertExpectations(t)
		mockHTTPClient.AssertExpectations(t)
	})

	t.Run("payload contains all transaction fields", func(t *testing.T) {
		mockTransactionRepo := new(db_test.MockTransactionRepository)
		mockHTTPClient := new(http_test.MockHTTPClient)

		worker := &Worker{
			transaction: mockTransactionRepo,
			httpClient:  mockHTTPClient,
		}

		notificationURL := "https://example.com/webhook"
		now := time.Now()
		processedAt := now.Add(5 * time.Minute)
		message := "Transaction successful"

		transaction := schema.Transaction{
			ID:          transactionID,
			ReferenceID: "REF-123",
			Amount:      1000.0,
			Fee:         10.0,
			Channel:     models.PaymentChannelUPI,
			Status:      models.TransactionStatusSuccess,
			Message:     &message,
			CreatedAt:   now,
			UpdatedAt:   now,
			ProcessedAt: &processedAt,
			Metadata: map[string]any{
				"notification_url": notificationURL,
			},
		}

		mockResponse := &http.Response{
			StatusCode: http.StatusOK,
		}

		updatedTransaction := transaction
		notifiedAt := time.Now()
		updatedTransaction.NotifiedAt = &notifiedAt

		mockTransactionRepo.On("Get", ctx, transactionID).Return(transaction, nil).Once()
		mockHTTPClient.On("POST", ctx, notificationURL, mock.MatchedBy(func(payload map[string]any) bool {
			return assert.Equal(t, transactionID, payload["transaction_id"]) &&
				assert.Equal(t, "REF-123", payload["reference_id"]) &&
				assert.Equal(t, float64(1000.0), payload["amount"]) &&
				assert.Equal(t, float64(10.0), payload["fee"]) &&
				assert.Equal(t, models.PaymentChannelUPI, payload["channel"]) &&
				assert.Equal(t, models.TransactionStatusSuccess, payload["status"]) &&
				assert.Equal(t, &message, payload["message"])
		}), mock.Anything).
			Return(mockResponse, nil).
			Once()
		mockTransactionRepo.On("Update", ctx, transactionID, mock.Anything).
			Return(updatedTransaction, nil).Once()

		worker.Notify(ctx, transactionID)

		mockTransactionRepo.AssertExpectations(t)
		mockHTTPClient.AssertExpectations(t)
	})
}
