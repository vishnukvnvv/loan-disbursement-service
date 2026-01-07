package models

import "time"

type DisbursementStatus string
type TransactionStatus string

const (
	TransactionStatusInitiated TransactionStatus = "initiated"
	TransactionStatusSuccess   TransactionStatus = "success"
	TransactionStatusFailed    TransactionStatus = "failed"
)

const (
	DisbursementStatusInitiated  DisbursementStatus = "initiated"
	DisbursementStatusProcessing DisbursementStatus = "processing"
	DisbursementStatusSuccess    DisbursementStatus = "success"
	DisbursementStatusFailed     DisbursementStatus = "failed"
	DisbursementStatusSuspended  DisbursementStatus = "suspended"
)

type DisburseRequest struct {
	LoanId          string  `json:"loan_id"`
	Amount          float64 `json:"amount"`
	AccountNumber   string  `json:"account_number"`
	IFSCCode        string  `json:"ifsc_code"`
	BeneficiaryName string  `json:"beneficiary_name"`
	BeneficiaryBank string  `json:"beneficiary_bank"`
}

type TransactionResponse struct {
	TransactionId string    `json:"transaction_id"`
	Status        string    `json:"status"`
	Mode          string    `json:"mode"`
	Message       *string   `json:"message"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Disbursement struct {
	DisbursementId string                `json:"disbursement_id"`
	Status         string                `json:"status"`
	LoanId         string                `json:"loan_id"`
	Amount         float64               `json:"amount"`
	Transaction    []TransactionResponse `json:"transaction"`
	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at"`
}

type DisbursementResponse struct {
	DisbursementId string `json:"disbursement_id"`
	Status         string `json:"status"`
	Message        string `json:"message"`
}
