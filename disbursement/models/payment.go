package models

import "time"

type PaymentRequest struct {
	ReferenceID string         `json:"reference_id"`
	Amount      float64        `json:"amount"`
	Mode        string         `json:"mode"`
	Beneficiary Beneficiary    `json:"beneficiary"`
	Metadata    map[string]any `json:"metadata"`
}

type PaymentResponse struct {
	TransactionID string         `json:"transaction_id"`
	ReferenceID   string         `json:"reference_id"`
	Amount        float64        `json:"amount"`
	Status        string         `json:"status"`
	Error         error          `json:"error"`
	Mode          string         `json:"mode"`
	Beneficiary   Beneficiary    `json:"beneficiary"`
	AcceptedAT    time.Time      `json:"accepted_at"`
	ProcessedAT   time.Time      `json:"processed_at"`
	Metadata      map[string]any `json:"metadata"`
}

type PaymentError struct {
	Error string `json:"error"`
}
