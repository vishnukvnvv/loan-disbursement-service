package payment

import (
	"context"
	"payment-gateway/db/daos"
	"payment-gateway/db/schema"
	"payment-gateway/failures"
	"payment-gateway/models"
	"payment-gateway/utils"
)

type PaymentProvider interface {
	ValidateLimit(amount float64) error
	GetTransaction(ctx context.Context, transactionID string) (*models.Transaction, error)
	Transfer(ctx context.Context, request models.PaymentRequest) (*models.Transaction, error)
}

func NewPaymentProvider(
	channel *schema.PaymentChannel,
	processor chan models.ProcessorMessage,
	transactionRepository daos.TransactionRepository,
	idGenerator utils.IdGenerator,
) (PaymentProvider, error) {
	switch channel.Name {
	case models.PaymentChannelUPI:
		return NewUPIProvider(
			processor,
			channel.Limit,
			channel.SuccessRate,
			channel.Fee,
			transactionRepository,
			idGenerator,
		), nil
	case models.PaymentChannelNEFT:
		return NewNEFTProvider(
			processor,
			channel.Limit,
			channel.SuccessRate,
			channel.Fee,
			transactionRepository,
			idGenerator,
		), nil
	case models.PaymentChannelIMPS:
		return NewIMPSProvider(
			processor,
			channel.Limit,
			channel.SuccessRate,
			channel.Fee,
			transactionRepository,
			idGenerator,
		), nil
	}
	return nil, failures.INVALID_PAYMENT_CHANNEL
}

func toModelTransaction(transaction schema.Transaction) *models.Transaction {
	return &models.Transaction{
		ID:          transaction.ID,
		ReferenceID: transaction.ReferenceID,
		Amount:      transaction.Amount,
		Channel:     transaction.Channel,
		Fee:         transaction.Fee,
		Beneficiary: models.Beneficiary{
			Name:    transaction.BeneficiaryName,
			Account: transaction.AccountNumber,
			IFSC:    transaction.IFSCCode,
			Bank:    transaction.BankName,
		},
		Metadata:    transaction.Metadata,
		Status:      transaction.Status,
		Message:     transaction.Message,
		CreatedAt:   transaction.CreatedAt,
		UpdatedAt:   transaction.UpdatedAt,
		ProcessedAt: transaction.ProcessedAt,
		NotifiedAt:  transaction.NotifiedAt,
	}
}
