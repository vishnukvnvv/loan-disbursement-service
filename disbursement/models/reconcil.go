package models

import "time"

type ReconciliationRequest struct {
	StatementDate string                      `json:"statement_date"`
	Transactions  []ReconciliationTransaction `json:"transactions" `
}

type ReconciliationTransaction struct {
	ReferenceID string            `json:"reference_id"`
	Amount      float64           `json:"amount"`
	Date        time.Time         `json:"date"`
	Status      TransactionStatus `json:"status"`
}

type Discrepancy struct {
	Type           string  `json:"type"`
	ReferenceID    string  `json:"reference_id"`
	ExpectedAmount float64 `json:"expected_amount"`
	ActualAmount   float64 `json:"actual_amount"`
	Message        string  `json:"message"`
}

type ReconciliationResponse struct {
	ReconciliationID string        `json:"reconciliation_id"`
	StatementDate    string        `json:"statement_date"`
	TotalExpected    float64       `json:"total_expected"`
	TotalActual      float64       `json:"total_actual"`
	MatchedCount     int           `json:"matched_count"`
	Discrepancies    []Discrepancy `json:"discrepancies"`
}
