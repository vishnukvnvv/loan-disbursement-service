package models

type PaymentRequest struct {
	ReferenceID string         `json:"reference_id"`
	Amount      float64        `json:"amount"`
	Channel     PaymentChannel `json:"channel"`
	Beneficiary Beneficiary    `json:"beneficiary"`
	Metadata    map[string]any `json:"metadata"`
}
