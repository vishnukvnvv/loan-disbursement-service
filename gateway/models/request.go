package models

import "mock-payment-gateway/types"

type PaymentRequest struct {
	ReferenceID string            `json:"reference_id"`
	Amount      float64           `json:"amount"`
	Mode        types.PaymentMode `json:"mode"`
	Beneficiary Beneficiary       `json:"beneficiary"`
	Metadata    map[string]any    `json:"metadata"`
}
