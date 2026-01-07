package services

import (
	"context"
	"fmt"
	"loan-disbursement-service/db/daos"
	"loan-disbursement-service/db/schema"
	"loan-disbursement-service/models"
	"time"

	"github.com/google/uuid"
)

type ReconciliationService struct {
	transactionDAO *daos.TransactionDAO
}

func NewReconciliationService(transactionDAO *daos.TransactionDAO) *ReconciliationService {
	return &ReconciliationService{transactionDAO: transactionDAO}
}

func (s *ReconciliationService) Reconcile(
	ctx context.Context,
	req models.ReconciliationRequest,
) (*models.ReconciliationResponse, error) {
	date, err := time.Parse(time.DateOnly, req.StatementDate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse statement date: %w", err)
	}
	ourTransactions, err := s.transactionDAO.ListByDate(
		ctx,
		date,
		[]string{string(models.TransactionStatusSuccess)},
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
					"Transaction marked as SUCCESS in our records but not found in bank statement",
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

		if bankTxn.Status != "SUCCESS" && bankTxn.Status != "COMPLETED" {
			discrepancies = append(discrepancies, models.Discrepancy{
				Type:           "status_mismatch",
				ReferenceID:    refID,
				ExpectedAmount: ourTxn.Amount,
				ActualAmount:   bankTxn.Amount,
				Message: fmt.Sprintf(
					"Status mismatch: we have SUCCESS, bank has %s",
					bankTxn.Status,
				),
			})
			continue
		}

		matchedCount++
	}

	// Check for ghost transactions (in bank statement but not in our records)
	for refID, bankTxn := range bankTxnMap {
		if _, exists := ourTxnMap[refID]; !exists {
			discrepancies = append(discrepancies, models.Discrepancy{
				Type:           "ghost",
				ReferenceID:    refID,
				ExpectedAmount: 0,
				ActualAmount:   bankTxn.Amount,
				Message: fmt.Sprintf(
					"Transaction in bank statement but not in our records - potential fraud or data loss",
				),
			})
		}
	}

	// Calculate totals
	totalExpected := 0.0
	for _, txn := range ourTransactions {
		totalExpected += txn.Amount
	}

	totalActual := 0.0
	for _, txn := range req.Transactions {
		if txn.Status == "SUCCESS" || txn.Status == "COMPLETED" {
			totalActual += txn.Amount
		}
	}

	response := &models.ReconciliationResponse{
		ReconciliationID: uuid.New().String(),
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
