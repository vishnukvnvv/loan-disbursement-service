package models

import (
	"mock-payment-gateway/types"
	"time"
)

type Payment struct {
	TransactionID string              `json:"transaction_id"`
	Amount        float64             `json:"amount"`
	Status        types.PaymentStatus `json:"status"`
	Error         error               `json:"error"`
	Mode          types.PaymentMode   `json:"mode"`
	Beneficiary   Beneficiary         `json:"beneficiary"`
	AcceptedAT    time.Time           `json:"accepted_at"`
	ProcessedAT   time.Time           `json:"processed_at"`
	Metadata      map[string]any      `json:"metadata"`
}
