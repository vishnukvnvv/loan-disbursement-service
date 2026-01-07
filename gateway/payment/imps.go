package payment

import (
	"errors"
	"fmt"
	"math/rand"
	"mock-payment-gateway/failures"
	"mock-payment-gateway/models"
	"mock-payment-gateway/types"
	"time"

	"github.com/google/uuid"
)

var IMPSTransactions map[string]models.Payment

type IMPSPayment struct {
	limit        float64
	timeout      float64
	failure_rate float64
}

func NewIMPS(limit float64, timeout float64, failure_rate float64) *IMPSPayment {
	return &IMPSPayment{
		limit:        limit,
		timeout:      timeout,
		failure_rate: failure_rate,
	}
}

func (i *IMPSPayment) ValidateLimit(amount float64) error {
	if amount > i.limit {
		return failures.LIMIT_EXCEEDED
	}
	return nil
}

func (i *IMPSPayment) GetTransaction(transactionID string) (*models.Payment, error) {
	if payment, exists := IMPSTransactions[transactionID]; exists {
		return &payment, nil
	}
	return nil, errors.New("Invalid transaction ID")
}

func (i *IMPSPayment) Transfer(request models.PaymentRequest) (*models.Payment, error) {
	transactionID := fmt.Sprintf("IMPS%s", uuid.New().String()[:12])
	acceptedAt := time.Now()

	time.Sleep(time.Duration(rand.Intn(1500)) * time.Millisecond)
	processedAt := time.Now()

	status := types.SUCCESS
	if rand.Float64() < i.failure_rate {
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
