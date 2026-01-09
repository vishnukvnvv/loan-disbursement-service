package service

import (
	"context"
	"errors"
	"fmt"
	"payment-gateway/db/daos"
	"payment-gateway/db/schema"
	"payment-gateway/failures"
	"payment-gateway/models"
	"payment-gateway/payment"
	"payment-gateway/utils"
	"slices"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type PaymentService interface {
	Process(ctx context.Context, request models.PaymentRequest) (*models.Transaction, error)
	GetTransaction(
		ctx context.Context,
		channel models.PaymentChannel,
		transactionID string,
	) (*models.Transaction, error)
}

type PaymentServiceImpl struct {
	paymentProvider payment.PaymentProvider
	paymentChannel  daos.PaymentChannelRepository
	transaction     daos.TransactionRepository
	idGenerator     utils.IdGenerator
	processor       chan models.ProcessorMessage
}

func NewPaymentService(
	processor chan models.ProcessorMessage,
	paymentChannel daos.PaymentChannelRepository,
	transactionRepository daos.TransactionRepository,
	idGenerator utils.IdGenerator,
) PaymentService {
	return &PaymentServiceImpl{
		processor:      processor,
		paymentChannel: paymentChannel,
		transaction:    transactionRepository,
		idGenerator:    idGenerator,
	}
}

func (s *PaymentServiceImpl) Process(
	ctx context.Context,
	request models.PaymentRequest,
) (*models.Transaction, error) {
	existingTransaction, err := s.transaction.GetByReferenceID(ctx, request.ReferenceID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error().Msgf("failed to get transaction: %s", err)
		return nil, fmt.Errorf("failed to get transaction: %s", err)
	}

	if existingTransaction != nil {
		log.Warn().Msgf("transaction already exists: %s", request.ReferenceID)
		return nil, failures.REFERENCE_ID_ALREADY_PROCESSED
	}

	err = s.validateBeneficiary(request.Beneficiary)
	if err != nil {
		log.Error().Msgf("beneficiary validation failed: %s", err)
		return nil, err
	}

	paymentChannel, err := s.getPaymentChannel(ctx, request.Channel)
	if err != nil {
		return nil, err
	}

	paymentProvider, err := s.newPaymentProvider(paymentChannel)
	if err != nil {
		return nil, err
	}

	err = paymentProvider.ValidateLimit(request.Amount)
	if err != nil {
		return nil, err
	}

	return paymentProvider.Transfer(ctx, request)
}

func (p *PaymentServiceImpl) GetTransaction(
	ctx context.Context,
	channel models.PaymentChannel,
	transactionID string,
) (*models.Transaction, error) {
	paymentChannel, err := p.getPaymentChannel(ctx, channel)
	if err != nil {
		return nil, err
	}

	paymentProvider, err := p.newPaymentProvider(paymentChannel)
	if err != nil {
		return nil, err
	}

	return paymentProvider.GetTransaction(ctx, transactionID)
}

func (p *PaymentServiceImpl) validateBeneficiary(beneficiary models.Beneficiary) error {
	if slices.Contains(failures.INVALID_IFSC_CODES, beneficiary.IFSC) {
		return failures.INVALID_IFSC
	}

	if slices.Contains(failures.INVALID_ACCOUNT_NUMBERS, beneficiary.Account) {
		return failures.INACTIVE_ACCOUNT
	}

	return nil
}

func (p *PaymentServiceImpl) getPaymentChannel(
	ctx context.Context,
	channel models.PaymentChannel,
) (*schema.PaymentChannel, error) {
	paymentChannel, err := p.paymentChannel.Get(ctx, channel)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Msgf("invalid payment channel: %s", channel)
			return nil, failures.INVALID_PAYMENT_CHANNEL
		}
		log.Error().Msgf("failed to get payment channel: %s", err)
		return nil, fmt.Errorf("failed to get payment channel: %s", err)
	}
	return paymentChannel, nil
}

func (p *PaymentServiceImpl) newPaymentProvider(
	paymentChannel *schema.PaymentChannel,
) (payment.PaymentProvider, error) {
	return payment.NewPaymentProvider(
		paymentChannel,
		p.processor,
		p.transaction,
		p.idGenerator,
	)
}
