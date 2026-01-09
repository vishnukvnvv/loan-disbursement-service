package services

import (
	"context"
	"fmt"
	"loan-disbursement-service/db/daos"
	"loan-disbursement-service/db/schema"
	"loan-disbursement-service/models"
	"loan-disbursement-service/utils"
	"time"
)

type ReconciliationService interface {
	Reconcile(
		ctx context.Context,
		req models.ReconciliationRequest,
	) (*models.ReconciliationResponse, error)
}

type ReconciliationServiceImpl struct {
	idGenerator utils.IdGenerator
	transaction daos.TransactionRepository
}

func NewReconciliationService(
	idGenerator utils.IdGenerator,
	transaction daos.TransactionRepository,
) ReconciliationService {
	return &ReconciliationServiceImpl{idGenerator: idGenerator, transaction: transaction}
}

func (s *ReconciliationServiceImpl) Reconcile(
	ctx context.Context,
	req models.ReconciliationRequest,
) (*models.ReconciliationResponse, error) {
	date, err := time.Parse(time.DateOnly, req.StatementDate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse statement date: %w", err)
	}
	ourTransactions, err := s.transaction.ListByDate(
		ctx,
		date,
		[]models.TransactionStatus{models.TransactionStatusSuccess},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get successful transactions: %w", err)
	}

	ourTxnMap := make(map[string]*schema.Transaction)
	for _, txn := range ourTransactions {
		ourTxnMap[txn.ReferenceId] = &txn
	}

	bankTxnMap := make(map[string]*models.ReconciliationTransaction)
	for i := range req.Transactions {
		bankTxnMap[req.Transactions[i].ReferenceID] = &req.Transactions[i]
	}

	var discrepancies []models.Discrepancy
	matchedCount := 0

	for refID, ourTxn := range ourTxnMap {
		bankTxn, exists := bankTxnMap[refID]

		if !exists {
			discrepancies = append(discrepancies, models.Discrepancy{
				Type:           "missing",
				ReferenceID:    refID,
				ExpectedAmount: ourTxn.Amount,
				ActualAmount:   0,
				Message: fmt.Sprintf(
					"Transaction marked as %s in our records but not found in bank statement",
					models.TransactionStatusSuccess,
				),
			})
			continue
		}

		if !amountsMatch(ourTxn.Amount, bankTxn.Amount) {
			discrepancies = append(discrepancies, models.Discrepancy{
				Type:           "amount_mismatch",
				ReferenceID:    refID,
				ExpectedAmount: ourTxn.Amount,
				ActualAmount:   bankTxn.Amount,
				Message: fmt.Sprintf(
					"Amount mismatch: expected %.2f, got %.2f",
					ourTxn.Amount,
					bankTxn.Amount,
				),
			})
			continue
		}

		if bankTxn.Status != models.TransactionStatusSuccess &&
			bankTxn.Status != models.TransactionStatusCompleted {
			discrepancies = append(discrepancies, models.Discrepancy{
				Type:           "status_mismatch",
				ReferenceID:    refID,
				ExpectedAmount: ourTxn.Amount,
				ActualAmount:   bankTxn.Amount,
				Message: fmt.Sprintf(
					"Status mismatch: we have %s, bank has %s",
					models.TransactionStatusSuccess,
					bankTxn.Status,
				),
			})
			continue
		}

		matchedCount++
	}

	for refID, bankTxn := range bankTxnMap {
		if _, exists := ourTxnMap[refID]; !exists {
			if bankTxn.Status != models.TransactionStatusSuccess &&
				bankTxn.Status != models.TransactionStatusCompleted {
				continue
			}
			discrepancies = append(discrepancies, models.Discrepancy{
				Type:           "ghost",
				ReferenceID:    refID,
				ExpectedAmount: 0,
				ActualAmount:   bankTxn.Amount,
				Message: fmt.Sprintf(
					"Transaction in bank statement but not in our records - potential fraud or data loss: %s",
					bankTxn.Status,
				),
			})
		}
	}

	totalExpected := 0.0
	for _, txn := range ourTransactions {
		totalExpected += txn.Amount
	}

	totalActual := 0.0
	for _, txn := range req.Transactions {
		if txn.Status == models.TransactionStatusSuccess ||
			txn.Status == models.TransactionStatusCompleted {
			totalActual += txn.Amount
		}
	}

	response := &models.ReconciliationResponse{
		ReconciliationID: s.idGenerator.GenerateReconciliationId(),
		StatementDate:    req.StatementDate,
		TotalExpected:    totalExpected,
		TotalActual:      totalActual,
		MatchedCount:     matchedCount,
		Discrepancies:    discrepancies,
	}

	return response, nil
}

func amountsMatch(a, b float64) bool {
	const tolerance = 0.01
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < tolerance
}
