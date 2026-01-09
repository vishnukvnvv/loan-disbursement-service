package models

import (
	"errors"
	"time"
)

type PaymentChannel string

const (
	PaymentChannelUPI  PaymentChannel = "UPI"
	PaymentChannelNEFT PaymentChannel = "NEFT"
	PaymentChannelIMPS PaymentChannel = "IMPS"
)

var (
	INVALID_TRANSACTION_ID         = errors.New("invalid transaction ID")
	TRANSACTION_NOT_FOUND          = errors.New("transaction not found")
	TRANSACTION_ID_REQUIRED        = errors.New("transactionId is required")
	NETWORK_ERROR                  = errors.New("network error")
	UNKNOWN_ERROR                  = errors.New("unknown error")
	INVALID_PAYMENT_CHANNEL        = errors.New("Invalid Payment Channel")
	SERVICE_UNAVAILABLE            = errors.New("Service Unavailable")
	BENEFICIARY_BANK_DOWN          = errors.New("Beneficiary Bank is Down")
	LIMIT_EXCEEDED                 = errors.New("Limit Exceeded")
	REFERENCE_ID_ALREADY_PROCESSED = errors.New("Reference ID already processed")
	INVALID_IFSC                   = errors.New("Invalid IFSC code")
	INACTIVE_ACCOUNT               = errors.New("Inactive Beneficiary Account")
	INSUFFICIENT_BALANCE           = errors.New("Insufficient Balance")
)

var TRANSIANT_FAILURES = []error{
	NETWORK_ERROR,
	INVALID_PAYMENT_CHANNEL,
	UNKNOWN_ERROR,
	SERVICE_UNAVAILABLE,
	LIMIT_EXCEEDED,
	BENEFICIARY_BANK_DOWN,
	INSUFFICIENT_BALANCE,
}

var PERMANENT_FAILURES = []error{
	INVALID_IFSC,
	INACTIVE_ACCOUNT,
}

type PaymentMetadata struct {
	NotificationURL string `json:"notification_url"`
	LoanID          string `json:"loan_id"`
	DisbursementID  string `json:"disbursement_id"`
}

type PaymentRequest struct {
	ReferenceID string          `json:"reference_id"`
	Amount      float64         `json:"amount"`
	Channel     PaymentChannel  `json:"channel"`
	Beneficiary Beneficiary     `json:"beneficiary"`
	Metadata    PaymentMetadata `json:"metadata"`
}

type PaymentResponse struct {
	TransactionID string            `json:"transaction_id"`
	ReferenceID   string            `json:"reference_id"`
	Amount        float64           `json:"amount"`
	Status        TransactionStatus `json:"status"`
	Error         error             `json:"error"`
	Channel       PaymentChannel    `json:"channel"`
	Beneficiary   Beneficiary       `json:"beneficiary"`
	AcceptedAT    time.Time         `json:"accepted_at"`
	ProcessedAT   time.Time         `json:"processed_at"`
	Metadata      PaymentMetadata   `json:"metadata"`
}

type PaymentNotificationRequest struct {
	TransactionID string            `json:"transaction_id"`
	ReferenceID   string            `json:"reference_id"`
	Status        TransactionStatus `json:"status"`
	Message       string            `json:"message"`
	Amount        float64           `json:"amount"`
	Fee           float64           `json:"fee"`
	Channel       PaymentChannel    `json:"channel"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	ProcessedAt   time.Time         `json:"processed_at"`
}

type PaymentError struct {
	Error string `json:"error"`
}

type PaymentChannelResponse struct {
	Active bool `json:"active"`
}
