package payment

import (
	"errors"
	"fmt"
	"math/rand"
	"mock-payment-gateway/models"
	"mock-payment-gateway/types"
	"time"

	"github.com/google/uuid"
)

var NEFTTransactions map[string]models.Payment

type NEFTPayment struct {
	timeout      float64
	failure_rate float64
}

func NewNEFT(timeout float64, failure_rate float64) *NEFTPayment {
	return &NEFTPayment{
		timeout:      timeout,
		failure_rate: failure_rate,
	}
}

func (n *NEFTPayment) ValidateLimit(amount float64) error {
	return nil
}

func (n *NEFTPayment) GetTransaction(transactionID string) (*models.Payment, error) {
	if payment, exists := NEFTTransactions[transactionID]; exists {
		return &payment, nil
	}
	return nil, errors.New("Invalid transaction ID")
}

func (n *NEFTPayment) Transfer(request models.PaymentRequest) (*models.Payment, error) {
	transactionID := fmt.Sprintf("NEFT%s", uuid.New().String()[:12])
	acceptedAt := time.Now()

	time.Sleep(time.Duration(rand.Intn(1500)) * time.Millisecond)
	processedAt := time.Now()

	status := types.SUCCESS
	if rand.Float64() < n.failure_rate {
		status = types.FAILED
	}

	payment := models.Payment{
		TransactionID: transactionID,
		Status:        status,
		Amount:        request.Amount,
		Error:         nil,
		Mode:          types.UPI,
		Beneficiary:   request.Beneficiary,
		AcceptedAT:    acceptedAt,
		ProcessedAT:   processedAt,
		Metadata:      request.Metadata,
	}

	IMPSTransactions[transactionID] = payment

	return &payment, nil
}
