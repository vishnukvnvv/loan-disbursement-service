package services

import (
	"errors"
	"fmt"
	"mock-payment-gateway/config"
	"mock-payment-gateway/failures"
	"mock-payment-gateway/models"
	"mock-payment-gateway/payment"
	"mock-payment-gateway/types"
	"slices"
	"strings"
	"time"
)

type PaymentService struct {
	config       *config.Configuration
	referenceIds map[string]time.Time
}

func NewPaymentService(config *config.Configuration) *PaymentService {
	return &PaymentService{
		config: config,
	}
}

func (p *PaymentService) Process(request models.PaymentRequest) (*models.Payment, error) {
	if _, exists := p.referenceIds[request.ReferenceID]; exists {
		return nil, errors.New("Reference ID already processed")
	}

	p.referenceIds[request.ReferenceID] = time.Now()

	err := p.validateBeneficiary(request.Beneficiary)
	if err != nil {
		return nil, err
	}

	paymentChannel, err := payment.New(request.Mode)
	if err != nil {
		return nil, err
	}

	err = paymentChannel.ValidateLimit(request.Amount)
	if err != nil {
		return nil, err
	}

	return paymentChannel.Transfer(request)
}

func (p *PaymentService) GetTransaction(transactionID string) (*models.Payment, error) {
	paymentMode, err := p.getPaymentMode(transactionID)
	if err != nil {
		return nil, err
	}

	paymentChannel, err := payment.New(paymentMode)
	if err != nil {
		return nil, err
	}

	return paymentChannel.GetTransaction(transactionID)
}

func (p *PaymentService) validateBeneficiary(beneficiary models.Beneficiary) error {
	if slices.Contains(p.config.InvalidIFSC, beneficiary.IFSC) {
		return failures.INVALID_IFSC
	}

	if slices.Contains(p.config.InvalidAccountNumber, beneficiary.Account) {
		return failures.INACTIVE_ACCOUNT
	}

	if slices.Contains(p.config.BeneficiaryBankDown, beneficiary.Bank) {
		return failures.BENEFICIARY_BANK_DOWN
	}

	return nil
}

func (p *PaymentService) getPaymentMode(transactionID string) (types.PaymentMode, error) {
	if transactionID == "" {
		return "", errors.New("transaction ID cannot be empty")
	}

	switch {
	case strings.HasPrefix(transactionID, "UPI"):
		return types.UPI, nil
	case strings.HasPrefix(transactionID, "IMPS"):
		return types.IMPS, nil
	case strings.HasPrefix(transactionID, "NEFT"):
		return types.NEFT, nil
	default:
		return "", fmt.Errorf(
			"invalid transaction ID format: expected prefix UPI, IMPS, or NEFT, got %q",
			transactionID,
		)
	}
}
