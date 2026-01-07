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

var UPITransactions = make(map[string]models.Payment)

type UPIPayment struct {
	limit        float64
	timeout      float64
	failure_rate float64
}

func NewUPI(limit float64, timeout float64, failure_rate float64) *UPIPayment {
	return &UPIPayment{
		limit:        limit,
		timeout:      timeout,
		failure_rate: failure_rate,
	}
}

func (u *UPIPayment) ValidateLimit(amount float64) error {
	if amount > u.limit {
		return failures.LIMIT_EXCEEDED
	}
	return nil
}

func (u *UPIPayment) GetTransaction(transactionID string) (*models.Payment, error) {
	if payment, exists := UPITransactions[transactionID]; exists {
		return &payment, nil
	}
	return nil, errors.New("Invalid transaction ID")
}

func (u *UPIPayment) Transfer(request models.PaymentRequest) (*models.Payment, error) {
	transactionID := fmt.Sprintf("UPI%s", uuid.New().String()[:12])
	acceptedAt := time.Now()

	time.Sleep(time.Duration(rand.Intn(2000)) * time.Millisecond)
	processedAt := time.Now()

	status := types.SUCCESS
	if rand.Float64() < u.failure_rate {
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

	UPITransactions[transactionID] = payment

	return &payment, nil
}
