package models

type ProcessorMessage struct {
	TransactionID string  `json:"transaction_id"`
	SuccessRate   float64 `json:"success_rate"`
	Delay         int64   `json:"delay"`
}
