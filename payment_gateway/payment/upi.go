package payment

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"payment-gateway/db/daos"
	"payment-gateway/db/schema"
	"payment-gateway/failures"
	"payment-gateway/models"
	"payment-gateway/utils"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type UPIProvider struct {
	processor   chan models.ProcessorMessage
	limit       float64
	successRate float64
	fee         float64
	transaction daos.TransactionRepository
	idGenerator utils.IdGenerator
}

func NewUPIProvider(
	processor chan models.ProcessorMessage,
	limit, successRate, fee float64,
	transaction daos.TransactionRepository,
	idGenerator utils.IdGenerator,
) *UPIProvider {
	return &UPIProvider{
		processor:   processor,
		limit:       limit,
		successRate: successRate,
		fee:         fee,
		transaction: transaction,
		idGenerator: utils.NewIdGenerator(),
	}
}

func (u *UPIProvider) ValidateLimit(amount float64) error {
	if amount > u.limit {
		log.Warn().Msgf("amount exceeds limit: %f > %f", amount, u.limit)
		return failures.LIMIT_EXCEEDED
	}
	return nil
}

func (u *UPIProvider) GetTransaction(
	ctx context.Context,
	transactionID string,
) (*models.Transaction, error) {
	transaction, err := u.transaction.Get(ctx, transactionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn().Msgf("transaction not found: %s: %s", transactionID, err)
			return nil, failures.TRANSACTION_NOT_FOUND
		}
		log.Error().Msgf("failed to get transaction: %s: %s", transactionID, err)
		return nil, errors.New("failed to get transaction")
	}
	return toModelTransaction(transaction), nil
}

func (u *UPIProvider) Transfer(
	ctx context.Context,
	request models.PaymentRequest,
) (*models.Transaction, error) {
	transactionId := u.idGenerator.GenerateUPITransactionId()
	newTransaction := schema.Transaction{
		ID:              transactionId,
		ReferenceID:     request.ReferenceID,
		Amount:          request.Amount,
		Channel:         models.PaymentChannelUPI,
		Fee:             u.fee,
		BeneficiaryName: request.Beneficiary.Name,
		AccountNumber:   request.Beneficiary.Account,
		IFSCCode:        request.Beneficiary.IFSC,
		BankName:        request.Beneficiary.Bank,
		Metadata:        request.Metadata,
		Status:          models.TransactionStatusInitiated,
		Message:         nil,
	}
	transaction, err := u.transaction.Create(ctx, newTransaction)
	if err != nil {
		log.Error().Msgf("failed to create transaction: %s: %s", transactionId, err)
		return nil, fmt.Errorf("failed to create transaction: %s", err)
	}

	delay := time.Duration(rand.Intn(2000)) * time.Millisecond
	u.processor <- models.ProcessorMessage{
		TransactionID: transactionId,
		SuccessRate:   u.successRate,
		Delay:         delay.Milliseconds(),
	}
	return toModelTransaction(transaction), nil
}
