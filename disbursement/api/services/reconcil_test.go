package services

import (
	"context"
	"errors"
	"loan-disbursement-service/db/schema"
	"loan-disbursement-service/models"
	db_test "loan-disbursement-service/test/db"
	utils_test "loan-disbursement-service/test/utils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReconciliationService_Reconcile(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully reconciles with all transactions matched", func(t *testing.T) {
		mockTransaction := new(db_test.MockTransactionRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewReconciliationService(mockIdGenerator, mockTransaction)

		statementDate := "2024-01-15"
		date, _ := time.Parse(time.DateOnly, statementDate)

		ourTransactions := []schema.Transaction{
			{
				Id:          "TXN-1",
				ReferenceId: "REF-1",
				Amount:      1000.0,
				Status:      models.TransactionStatusSuccess,
			},
			{
				Id:          "TXN-2",
				ReferenceId: "REF-2",
				Amount:      2000.0,
				Status:      models.TransactionStatusSuccess,
			},
		}

		req := models.ReconciliationRequest{
			StatementDate: statementDate,
			Transactions: []models.ReconciliationTransaction{
				{
					ReferenceID: "REF-1",
					Amount:      1000.0,
					Status:      models.TransactionStatusSuccess,
				},
				{
					ReferenceID: "REF-2",
					Amount:      2000.0,
					Status:      models.TransactionStatusSuccess,
				},
			},
		}

		mockTransaction.On("ListByDate", ctx, date, []models.TransactionStatus{
			models.TransactionStatusSuccess,
		}).Return(ourTransactions, nil).Once()
		mockIdGenerator.On("GenerateReconciliationId").Return("RECON-123").Once()

		response, err := service.Reconcile(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, statementDate, response.StatementDate)
		assert.Equal(t, 2, response.MatchedCount)
		assert.Equal(t, 3000.0, response.TotalExpected)
		assert.Equal(t, 3000.0, response.TotalActual)
		assert.Empty(t, response.Discrepancies)
		assert.Equal(t, "RECON-123", response.ReconciliationID)
		mockTransaction.AssertExpectations(t)
		mockIdGenerator.AssertExpectations(t)
	})

	t.Run("identifies missing transactions", func(t *testing.T) {
		mockTransaction := new(db_test.MockTransactionRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewReconciliationService(mockIdGenerator, mockTransaction)

		statementDate := "2024-01-15"
		date, _ := time.Parse(time.DateOnly, statementDate)

		ourTransactions := []schema.Transaction{
			{
				Id:          "TXN-1",
				ReferenceId: "REF-1",
				Amount:      1000.0,
				Status:      models.TransactionStatusSuccess,
			},
			{
				Id:          "TXN-2",
				ReferenceId: "REF-2",
				Amount:      2000.0,
				Status:      models.TransactionStatusSuccess,
			},
		}

		req := models.ReconciliationRequest{
			StatementDate: statementDate,
			Transactions: []models.ReconciliationTransaction{
				{
					ReferenceID: "REF-1",
					Amount:      1000.0,
					Status:      models.TransactionStatusSuccess,
				},
			},
		}

		mockTransaction.On("ListByDate", ctx, date, []models.TransactionStatus{
			models.TransactionStatusSuccess,
		}).Return(ourTransactions, nil).Once()
		mockIdGenerator.On("GenerateReconciliationId").Return("RECON-123").Once()

		response, err := service.Reconcile(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 1, response.MatchedCount)
		assert.Equal(t, 3000.0, response.TotalExpected)
		assert.Equal(t, 1000.0, response.TotalActual)
		assert.Len(t, response.Discrepancies, 1)
		assert.Equal(t, "missing", response.Discrepancies[0].Type)
		assert.Equal(t, "REF-2", response.Discrepancies[0].ReferenceID)
		assert.Equal(t, 2000.0, response.Discrepancies[0].ExpectedAmount)
		assert.Equal(t, 0.0, response.Discrepancies[0].ActualAmount)
		mockTransaction.AssertExpectations(t)
	})

	t.Run("identifies amount mismatches", func(t *testing.T) {
		mockTransaction := new(db_test.MockTransactionRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewReconciliationService(mockIdGenerator, mockTransaction)

		statementDate := "2024-01-15"
		date, _ := time.Parse(time.DateOnly, statementDate)

		ourTransactions := []schema.Transaction{
			{
				Id:          "TXN-1",
				ReferenceId: "REF-1",
				Amount:      1000.0,
				Status:      models.TransactionStatusSuccess,
			},
		}

		req := models.ReconciliationRequest{
			StatementDate: statementDate,
			Transactions: []models.ReconciliationTransaction{
				{
					ReferenceID: "REF-1",
					Amount:      1500.0,
					Status:      models.TransactionStatusSuccess,
				},
			},
		}

		mockTransaction.On("ListByDate", ctx, date, []models.TransactionStatus{
			models.TransactionStatusSuccess,
		}).Return(ourTransactions, nil).Once()
		mockIdGenerator.On("GenerateReconciliationId").Return("RECON-123").Once()

		response, err := service.Reconcile(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 0, response.MatchedCount)
		assert.Len(t, response.Discrepancies, 1)
		assert.Equal(t, "amount_mismatch", response.Discrepancies[0].Type)
		assert.Equal(t, "REF-1", response.Discrepancies[0].ReferenceID)
		assert.Equal(t, 1000.0, response.Discrepancies[0].ExpectedAmount)
		assert.Equal(t, 1500.0, response.Discrepancies[0].ActualAmount)
		assert.Contains(t, response.Discrepancies[0].Message, "Amount mismatch")
		mockTransaction.AssertExpectations(t)
	})

	t.Run("identifies status mismatches", func(t *testing.T) {
		mockTransaction := new(db_test.MockTransactionRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewReconciliationService(mockIdGenerator, mockTransaction)

		statementDate := "2024-01-15"
		date, _ := time.Parse(time.DateOnly, statementDate)

		ourTransactions := []schema.Transaction{
			{
				Id:          "TXN-1",
				ReferenceId: "REF-1",
				Amount:      1000.0,
				Status:      models.TransactionStatusSuccess,
			},
		}

		req := models.ReconciliationRequest{
			StatementDate: statementDate,
			Transactions: []models.ReconciliationTransaction{
				{
					ReferenceID: "REF-1",
					Amount:      1000.0,
					Status:      models.TransactionStatusFailed,
				},
			},
		}

		mockTransaction.On("ListByDate", ctx, date, []models.TransactionStatus{
			models.TransactionStatusSuccess,
		}).Return(ourTransactions, nil).Once()
		mockIdGenerator.On("GenerateReconciliationId").Return("RECON-123").Once()

		response, err := service.Reconcile(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 0, response.MatchedCount)
		assert.Len(t, response.Discrepancies, 1)
		assert.Equal(t, "status_mismatch", response.Discrepancies[0].Type)
		assert.Equal(t, "REF-1", response.Discrepancies[0].ReferenceID)
		assert.Contains(t, response.Discrepancies[0].Message, "Status mismatch")
		assert.Contains(
			t,
			response.Discrepancies[0].Message,
			string(models.TransactionStatusFailed),
		)
		mockTransaction.AssertExpectations(t)
	})

	t.Run("identifies ghost transactions", func(t *testing.T) {
		mockTransaction := new(db_test.MockTransactionRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewReconciliationService(mockIdGenerator, mockTransaction)

		statementDate := "2024-01-15"
		date, _ := time.Parse(time.DateOnly, statementDate)

		ourTransactions := []schema.Transaction{
			{
				Id:          "TXN-1",
				ReferenceId: "REF-1",
				Amount:      1000.0,
				Status:      models.TransactionStatusSuccess,
			},
		}

		req := models.ReconciliationRequest{
			StatementDate: statementDate,
			Transactions: []models.ReconciliationTransaction{
				{
					ReferenceID: "REF-1",
					Amount:      1000.0,
					Status:      models.TransactionStatusSuccess,
				},
				{
					ReferenceID: "REF-2",
					Amount:      2000.0,
					Status:      models.TransactionStatusSuccess,
				},
			},
		}

		mockTransaction.On("ListByDate", ctx, date, []models.TransactionStatus{
			models.TransactionStatusSuccess,
		}).Return(ourTransactions, nil).Once()
		mockIdGenerator.On("GenerateReconciliationId").Return("RECON-123").Once()

		response, err := service.Reconcile(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 1, response.MatchedCount)
		assert.Len(t, response.Discrepancies, 1)
		assert.Equal(t, "ghost", response.Discrepancies[0].Type)
		assert.Equal(t, "REF-2", response.Discrepancies[0].ReferenceID)
		assert.Equal(t, 0.0, response.Discrepancies[0].ExpectedAmount)
		assert.Equal(t, 2000.0, response.Discrepancies[0].ActualAmount)
		assert.Contains(
			t,
			response.Discrepancies[0].Message,
			"Transaction in bank statement but not in our records",
		)
		mockTransaction.AssertExpectations(t)
	})

	t.Run("handles multiple types of discrepancies", func(t *testing.T) {
		mockTransaction := new(db_test.MockTransactionRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewReconciliationService(mockIdGenerator, mockTransaction)

		statementDate := "2024-01-15"
		date, _ := time.Parse(time.DateOnly, statementDate)

		ourTransactions := []schema.Transaction{
			{
				Id:          "TXN-1",
				ReferenceId: "REF-1",
				Amount:      1000.0,
				Status:      models.TransactionStatusSuccess,
			},
			{
				Id:          "TXN-2",
				ReferenceId: "REF-2",
				Amount:      2000.0,
				Status:      models.TransactionStatusSuccess,
			},
			{
				Id:          "TXN-3",
				ReferenceId: "REF-3",
				Amount:      3000.0,
				Status:      models.TransactionStatusSuccess,
			},
		}

		req := models.ReconciliationRequest{
			StatementDate: statementDate,
			Transactions: []models.ReconciliationTransaction{
				{
					ReferenceID: "REF-1",
					Amount:      1000.0,
					Status:      models.TransactionStatusSuccess,
				},
				{
					ReferenceID: "REF-2",
					Amount:      2500.0,
					Status:      models.TransactionStatusSuccess,
				},
				{
					ReferenceID: "REF-4",
					Amount:      4000.0,
					Status:      models.TransactionStatusSuccess,
				},
			},
		}

		mockTransaction.On("ListByDate", ctx, date, []models.TransactionStatus{
			models.TransactionStatusSuccess,
		}).Return(ourTransactions, nil).Once()
		mockIdGenerator.On("GenerateReconciliationId").Return("RECON-123").Once()

		response, err := service.Reconcile(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 1, response.MatchedCount)
		assert.Equal(t, 6000.0, response.TotalExpected)
		assert.Equal(t, 7500.0, response.TotalActual)
		assert.Len(t, response.Discrepancies, 3)

		discrepancyTypes := make(map[string]bool)
		for _, d := range response.Discrepancies {
			discrepancyTypes[d.Type] = true
		}
		assert.True(t, discrepancyTypes["missing"])
		assert.True(t, discrepancyTypes["amount_mismatch"])
		assert.True(t, discrepancyTypes["ghost"])
		mockTransaction.AssertExpectations(t)
	})

	t.Run("returns error when statement date is invalid", func(t *testing.T) {
		mockTransaction := new(db_test.MockTransactionRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewReconciliationService(mockIdGenerator, mockTransaction)

		req := models.ReconciliationRequest{
			StatementDate: "invalid-date",
			Transactions:  []models.ReconciliationTransaction{},
		}

		response, err := service.Reconcile(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "failed to parse statement date")
		mockTransaction.AssertNotCalled(t, "ListByDate")
	})

	t.Run("returns error when transaction list fails", func(t *testing.T) {
		mockTransaction := new(db_test.MockTransactionRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewReconciliationService(mockIdGenerator, mockTransaction)

		statementDate := "2024-01-15"
		date, _ := time.Parse(time.DateOnly, statementDate)

		req := models.ReconciliationRequest{
			StatementDate: statementDate,
			Transactions:  []models.ReconciliationTransaction{},
		}

		mockTransaction.On("ListByDate", ctx, date, []models.TransactionStatus{
			models.TransactionStatusSuccess,
		}).Return([]schema.Transaction{}, errors.New("database error")).Once()

		response, err := service.Reconcile(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "failed to get successful transactions")
		mockTransaction.AssertExpectations(t)
	})

	t.Run("handles empty transactions", func(t *testing.T) {
		mockTransaction := new(db_test.MockTransactionRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewReconciliationService(mockIdGenerator, mockTransaction)

		statementDate := "2024-01-15"
		date, _ := time.Parse(time.DateOnly, statementDate)

		req := models.ReconciliationRequest{
			StatementDate: statementDate,
			Transactions:  []models.ReconciliationTransaction{},
		}

		mockTransaction.On("ListByDate", ctx, date, []models.TransactionStatus{
			models.TransactionStatusSuccess,
		}).Return([]schema.Transaction{}, nil).Once()
		mockIdGenerator.On("GenerateReconciliationId").Return("RECON-123").Once()

		response, err := service.Reconcile(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 0, response.MatchedCount)
		assert.Equal(t, 0.0, response.TotalExpected)
		assert.Equal(t, 0.0, response.TotalActual)
		assert.Empty(t, response.Discrepancies)
		mockTransaction.AssertExpectations(t)
	})

	t.Run("accepts COMPLETED status as valid", func(t *testing.T) {
		mockTransaction := new(db_test.MockTransactionRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewReconciliationService(mockIdGenerator, mockTransaction)

		statementDate := "2024-01-15"
		date, _ := time.Parse(time.DateOnly, statementDate)

		ourTransactions := []schema.Transaction{
			{
				Id:          "TXN-1",
				ReferenceId: "REF-1",
				Amount:      1000.0,
				Status:      models.TransactionStatusSuccess,
			},
		}

		req := models.ReconciliationRequest{
			StatementDate: statementDate,
			Transactions: []models.ReconciliationTransaction{
				{
					ReferenceID: "REF-1",
					Amount:      1000.0,
					Status:      models.TransactionStatusCompleted,
				},
			},
		}

		mockTransaction.On("ListByDate", ctx, date, []models.TransactionStatus{
			models.TransactionStatusSuccess,
		}).Return(ourTransactions, nil).Once()
		mockIdGenerator.On("GenerateReconciliationId").Return("RECON-123").Once()

		response, err := service.Reconcile(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 1, response.MatchedCount)
		assert.Empty(t, response.Discrepancies)
		mockTransaction.AssertExpectations(t)
	})

	t.Run("handles amount matching with tolerance", func(t *testing.T) {
		mockTransaction := new(db_test.MockTransactionRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewReconciliationService(mockIdGenerator, mockTransaction)

		statementDate := "2024-01-15"
		date, _ := time.Parse(time.DateOnly, statementDate)

		ourTransactions := []schema.Transaction{
			{
				Id:          "TXN-1",
				ReferenceId: "REF-1",
				Amount:      1000.0,
				Status:      models.TransactionStatusSuccess,
			},
		}

		req := models.ReconciliationRequest{
			StatementDate: statementDate,
			Transactions: []models.ReconciliationTransaction{
				{
					ReferenceID: "REF-1",
					Amount:      1000.005,
					Status:      models.TransactionStatusSuccess,
				},
			},
		}

		mockTransaction.On("ListByDate", ctx, date, []models.TransactionStatus{
			models.TransactionStatusSuccess,
		}).Return(ourTransactions, nil).Once()
		mockIdGenerator.On("GenerateReconciliationId").Return("RECON-123").Once()

		response, err := service.Reconcile(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 1, response.MatchedCount)
		assert.Empty(t, response.Discrepancies)
		mockTransaction.AssertExpectations(t)
	})

	t.Run("identifies amount mismatch when difference exceeds tolerance", func(t *testing.T) {
		mockTransaction := new(db_test.MockTransactionRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewReconciliationService(mockIdGenerator, mockTransaction)

		statementDate := "2024-01-15"
		date, _ := time.Parse(time.DateOnly, statementDate)

		ourTransactions := []schema.Transaction{
			{
				Id:          "TXN-1",
				ReferenceId: "REF-1",
				Amount:      1000.0,
				Status:      models.TransactionStatusSuccess,
			},
		}

		req := models.ReconciliationRequest{
			StatementDate: statementDate,
			Transactions: []models.ReconciliationTransaction{
				{
					ReferenceID: "REF-1",
					Amount:      1000.02,
					Status:      models.TransactionStatusSuccess,
				},
			},
		}

		mockTransaction.On("ListByDate", ctx, date, []models.TransactionStatus{
			models.TransactionStatusSuccess,
		}).Return(ourTransactions, nil).Once()
		mockIdGenerator.On("GenerateReconciliationId").Return("RECON-123").Once()

		response, err := service.Reconcile(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 0, response.MatchedCount)
		assert.Len(t, response.Discrepancies, 1)
		assert.Equal(t, "amount_mismatch", response.Discrepancies[0].Type)
		mockTransaction.AssertExpectations(t)
	})

	t.Run(
		"calculates total actual only for SUCCESS and COMPLETED transactions",
		func(t *testing.T) {
			mockTransaction := new(db_test.MockTransactionRepository)
			mockIdGenerator := new(utils_test.MockIdGenerator)

			service := NewReconciliationService(mockIdGenerator, mockTransaction)

			statementDate := "2024-01-15"
			date, _ := time.Parse(time.DateOnly, statementDate)

			ourTransactions := []schema.Transaction{
				{
					Id:          "TXN-1",
					ReferenceId: "REF-1",
					Amount:      1000.0,
					Status:      models.TransactionStatusSuccess,
				},
			}

			req := models.ReconciliationRequest{
				StatementDate: statementDate,
				Transactions: []models.ReconciliationTransaction{
					{
						ReferenceID: "REF-1",
						Amount:      1000.0,
						Status:      models.TransactionStatusSuccess,
					},
					{
						ReferenceID: "REF-2",
						Amount:      2000.0,
						Status:      models.TransactionStatusFailed,
					},
					{
						ReferenceID: "REF-3",
						Amount:      3000.0,
						Status:      models.TransactionStatusCompleted,
					},
				},
			}

			mockTransaction.On("ListByDate", ctx, date, []models.TransactionStatus{
				models.TransactionStatusSuccess,
			}).Return(ourTransactions, nil).Once()
			mockIdGenerator.On("GenerateReconciliationId").Return("RECON-123").Once()

			response, err := service.Reconcile(ctx, req)

			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, 1000.0, response.TotalExpected)
			assert.Equal(t, 4000.0, response.TotalActual)
			mockTransaction.AssertExpectations(t)
			mockIdGenerator.AssertExpectations(t)
		},
	)

	t.Run("generates unique reconciliation ID", func(t *testing.T) {
		mockTransaction := new(db_test.MockTransactionRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewReconciliationService(mockIdGenerator, mockTransaction)

		statementDate := "2024-01-15"
		date, _ := time.Parse(time.DateOnly, statementDate)

		req := models.ReconciliationRequest{
			StatementDate: statementDate,
			Transactions:  []models.ReconciliationTransaction{},
		}

		mockTransaction.On("ListByDate", ctx, date, []models.TransactionStatus{
			models.TransactionStatusSuccess,
		}).Return([]schema.Transaction{}, nil).Once()
		mockIdGenerator.On("GenerateReconciliationId").Return("RECON-1").Once()

		response1, err1 := service.Reconcile(ctx, req)
		assert.NoError(t, err1)

		mockTransaction.On("ListByDate", ctx, date, []models.TransactionStatus{
			models.TransactionStatusSuccess,
		}).Return([]schema.Transaction{}, nil).Once()
		mockIdGenerator.On("GenerateReconciliationId").Return("RECON-2").Once()

		response2, err2 := service.Reconcile(ctx, req)
		assert.NoError(t, err2)

		assert.NotEqual(t, response1.ReconciliationID, response2.ReconciliationID)
		mockTransaction.AssertExpectations(t)
		mockIdGenerator.AssertExpectations(t)
	})
}
