package payment

import (
	"errors"
	"mock-payment-gateway/models"
	"mock-payment-gateway/types"
)

type PaymentConstraints struct {
	Limit       float64 `yaml:"limit"`
	Timeout     float64 `yaml:"timeout"`
	FailureRate float64 `yaml:"failure_rate"`
}

type PaymentMethods struct {
	UPI  PaymentConstraints
	NEFT PaymentConstraints
	IMPS PaymentConstraints
}

var payments PaymentMethods

type PaymentService interface {
	ValidateLimit(amount float64) error
	GetTransaction(transactionID string) (*models.Payment, error)
	Transfer(request models.PaymentRequest) (*models.Payment, error)
}

func Initialize(methods PaymentMethods) {
	payments = methods
}

func New(mode types.PaymentMode) (PaymentService, error) {
	switch mode {
	case types.UPI:
		return NewUPI(payments.UPI.Limit, payments.UPI.Timeout, payments.UPI.FailureRate), nil
	case types.NEFT:
		return NewNEFT(payments.NEFT.Timeout, payments.NEFT.FailureRate), nil
	case types.IMPS:
		return NewIMPS(payments.IMPS.Limit, payments.IMPS.Timeout, payments.IMPS.FailureRate), nil
	}
	return nil, errors.New("invalid payment mode")
}
