package models

import "time"

type TransactionStatus string

const (
	TransactionStatusInitiated  TransactionStatus = "initiated"
	TransactionStatusProcessing TransactionStatus = "processing"
	TransactionStatusSuccess    TransactionStatus = "success"
	TransactionStatusFailed     TransactionStatus = "failed"
)

type Transaction struct {
	ID          string            `json:"id"`
	ReferenceID string            `json:"reference_id"`
	Amount      float64           `json:"amount"`
	Channel     PaymentChannel    `json:"channel"`
	Fee         float64           `json:"fee"`
	Beneficiary Beneficiary       `json:"beneficiary"`
	Metadata    map[string]any    `json:"metadata"`
	Status      TransactionStatus `json:"status"`
	Message     *string           `json:"message"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	ProcessedAt *time.Time        `json:"processed_at"`
	NotifiedAt  *time.Time        `json:"notified_at"`
}
