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

type NEFTProvider struct {
	processor   chan models.ProcessorMessage
	limit       float64
	successRate float64
	fee         float64
	transaction daos.TransactionRepository
	idGenerator utils.IdGenerator
}

func NewNEFTProvider(
	processor chan models.ProcessorMessage,
	limit, successRate, fee float64,
	transaction daos.TransactionRepository,
	idGenerator utils.IdGenerator,
) *NEFTProvider {
	return &NEFTProvider{
		processor:   processor,
		limit:       limit,
		successRate: successRate,
		fee:         fee,
		transaction: transaction,
		idGenerator: utils.NewIdGenerator(),
	}
}

func (n *NEFTProvider) ValidateLimit(amount float64) error {
	return nil
}

func (n *NEFTProvider) GetTransaction(
	ctx context.Context,
	transactionID string,
) (*models.Transaction, error) {
	transaction, err := n.transaction.Get(ctx, transactionID)
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

func (n *NEFTProvider) Transfer(
	ctx context.Context,
	request models.PaymentRequest,
) (*models.Transaction, error) {
	transactionId := n.idGenerator.GenerateNEFTTransactionId()
	newTransaction := schema.Transaction{
		ID:              transactionId,
		ReferenceID:     request.ReferenceID,
		Amount:          request.Amount,
		Channel:         models.PaymentChannelNEFT,
		Fee:             n.fee,
		BeneficiaryName: request.Beneficiary.Name,
		AccountNumber:   request.Beneficiary.Account,
		IFSCCode:        request.Beneficiary.IFSC,
		BankName:        request.Beneficiary.Bank,
		Metadata:        request.Metadata,
		Status:          models.TransactionStatusInitiated,
		Message:         nil,
	}
	transaction, err := n.transaction.Create(ctx, newTransaction)
	if err != nil {
		log.Error().Msgf("failed to create transaction: %s: %s", transactionId, err)
		return nil, fmt.Errorf("failed to create transaction: %s", err)
	}

	delay := time.Duration(rand.Intn(1500)) * time.Millisecond
	n.processor <- models.ProcessorMessage{
		TransactionID: transactionId,
		SuccessRate:   n.successRate,
		Delay:         delay.Milliseconds(),
	}
	return toModelTransaction(transaction), nil
}
