package services

import (
	"context"
	"errors"
	"fmt"
	"loan-disbursement-service/db"
	"loan-disbursement-service/db/daos"
	"loan-disbursement-service/db/schema"
	"loan-disbursement-service/models"
	"loan-disbursement-service/providers"
	"loan-disbursement-service/utils"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type PaymentService interface {
	Process(ctx context.Context, disbursement *schema.Disbursement) error
	HandleNotification(ctx context.Context, notification models.PaymentNotificationRequest) error
	HandleFailure(
		ctx context.Context,
		disbursement *schema.Disbursement,
		transaction *schema.Transaction,
		channel models.PaymentChannel,
		err error,
	) error
	HandleSuccess(
		ctx context.Context,
		disbursementId, transactionId string,
		channel models.PaymentChannel,
	) error
}

type PaymentServiceImpl struct {
	db              *db.Database
	disbursement    daos.DisbursementRepository
	transaction     daos.TransactionRepository
	loan            daos.LoanRepository
	beneficiary     daos.BeneficiaryRepository
	retryPolicy     RetryPolicy
	gatewayProvider providers.PaymentProvider
	idGenerator     utils.IdGenerator
	notificationURL string
}

func NewPaymentService(
	database *db.Database,
	disbursement daos.DisbursementRepository,
	transaction daos.TransactionRepository,
	loan daos.LoanRepository,
	beneficiary daos.BeneficiaryRepository,
	retryPolicy RetryPolicy,
	gatewayProvider providers.PaymentProvider,
	idGenerator utils.IdGenerator,
	notificationURL string,
) PaymentService {
	return &PaymentServiceImpl{
		db:              database,
		disbursement:    disbursement,
		transaction:     transaction,
		loan:            loan,
		beneficiary:     beneficiary,
		retryPolicy:     retryPolicy,
		gatewayProvider: gatewayProvider,
		idGenerator:     idGenerator,
		notificationURL: notificationURL,
	}
}

func (p PaymentServiceImpl) Process(
	ctx context.Context,
	disbursement *schema.Disbursement,
) error {
	if !p.shouldProcess(disbursement) {
		log.Info().Msgf("disbursement %v is not eligible for processing", disbursement.Id)
		return nil
	}

	loan, err := p.loan.Get(ctx, disbursement.LoanId)
	if err != nil {
		return fmt.Errorf("failed to get loan: %w", err)
	}

	beneficiary, err := p.beneficiary.GetById(ctx, *loan.BeneficiaryId)
	if err != nil {
		return fmt.Errorf("failed to get beneficiary: %w", err)
	}

	channel := p.selectChannel(loan, disbursement)
	if err := p.transitionToProcessing(ctx, disbursement, channel); err != nil {
		return fmt.Errorf("failed to transition to processing: %w", err)
	}

	return p.execute(ctx, disbursement, loan, beneficiary, channel)
}

func (p PaymentServiceImpl) execute(
	ctx context.Context,
	disbursement *schema.Disbursement,
	loan *schema.Loan,
	beneficiary *schema.Beneficiary,
	channel models.PaymentChannel,
) error {
	activeChannel := channel
	var err error

	isChannelActive := p.isChannelActive(ctx, channel)
	if !isChannelActive {
		activeChannel, err = p.channelFallback(channel, isChannelActive)
		log.Info().
			Msgf("Channel %s is not active, falling back to %s", channel, activeChannel)
		if err != nil {
			return fmt.Errorf("failed to fallback channel: %w", err)
		}
	}
	transactionId := p.idGenerator.GenerateTransactionId()
	referenceId := p.idGenerator.GenerateReferenceId()
	transaction, err := p.transaction.Create(ctx,
		schema.Transaction{
			Id:             transactionId,
			DisbursementId: disbursement.Id,
			ReferenceId:    referenceId,
			Channel:        activeChannel,
			Amount:         loan.Amount,
			Status:         models.TransactionStatusInitiated,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	response, err := p.transfer(ctx, referenceId, disbursement, loan, beneficiary, activeChannel)
	if err != nil {
		return p.HandleFailure(ctx, disbursement, transaction, activeChannel, err)
	}

	return p.handleResponse(ctx, transaction.Id, response)
}

func (p PaymentServiceImpl) HandleNotification(
	ctx context.Context,
	notification models.PaymentNotificationRequest,
) error {
	transaction, err := p.transaction.GetByReferenceID(ctx, notification.ReferenceID)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %w", err)
	}
	disbursement, err := p.disbursement.Get(ctx, transaction.DisbursementId)
	if err != nil {
		return fmt.Errorf("failed to get disbursement: %w", err)
	}
	if notification.Status == models.TransactionStatusSuccess {
		return p.HandleSuccess(ctx, disbursement.Id, transaction.Id, notification.Channel)
	}
	return p.HandleFailure(
		ctx,
		disbursement,
		transaction,
		notification.Channel,
		errors.New(notification.Message),
	)
}

func (p PaymentServiceImpl) HandleFailure(
	ctx context.Context,
	disbursement *schema.Disbursement,
	transaction *schema.Transaction,
	channel models.PaymentChannel,
	err error,
) error {
	if errors.Is(err, models.REFERENCE_ID_ALREADY_PROCESSED) {
		payment, gatewayErr := p.gatewayProvider.Fetch(
			ctx,
			transaction.Channel,
			transaction.ReferenceId,
		)
		if gatewayErr != nil {
			return fmt.Errorf("failed to get payment: %w", gatewayErr)
		}
		if payment.Status == models.TransactionStatusSuccess {
			return p.handleResponse(ctx, transaction.Id, payment)
		}
		err = payment.Error
	}
	status, retryCount := p.evaluateFailure(disbursement.RetryCount, err)
	return p.db.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		dbErr := p.transaction.Update(ctx, transaction.Id, map[string]any{
			"status":     models.TransactionStatusFailed,
			"message":    err.Error(),
			"updated_at": time.Now(),
		})
		if dbErr != nil {
			return dbErr
		}
		return p.disbursement.Update(ctx, disbursement.Id, map[string]any{
			"status":      status,
			"channel":     channel,
			"retry_count": retryCount,
			"last_error":  err.Error(),
			"updated_at":  time.Now(),
		})
	})
}

func (p PaymentServiceImpl) HandleSuccess(
	ctx context.Context,
	disbursementId, transactionId string,
	channel models.PaymentChannel,
) error {
	return p.db.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		dbErr := p.transaction.Update(ctx, transactionId, map[string]any{
			"status":     models.TransactionStatusSuccess,
			"updated_at": time.Now(),
		})
		if dbErr != nil {
			return dbErr
		}
		return p.disbursement.Update(ctx, disbursementId, map[string]any{
			"status":     models.DisbursementStatusSuccess,
			"channel":    channel,
			"updated_at": time.Now(),
		})
	})
}

func (p PaymentServiceImpl) evaluateFailure(
	retryCount int,
	err error,
) (models.DisbursementStatus, int) {
	log.Info().Msgf("evaluating failure: %v", err)
	if retryCount >= MaxRetries {
		return models.DisbursementStatusFailed, retryCount
	}

	newRetryCount := retryCount + 1

	for _, retriableErr := range models.TRANSIANT_FAILURES {
		if strings.Contains(err.Error(), retriableErr.Error()) {
			return models.DisbursementStatusSuspended, newRetryCount
		}
	}

	log.Info().Msgf("error is not transient, returning failed")
	return models.DisbursementStatusFailed, newRetryCount
}

func (p PaymentServiceImpl) handleResponse(
	ctx context.Context,
	transactionId string,
	response models.PaymentResponse,
) error {
	return p.transaction.Update(ctx,
		transactionId,
		map[string]any{
			"status":     response.Status,
			"updated_at": time.Now(),
		},
	)
}

func (p PaymentServiceImpl) shouldProcess(
	disbursement *schema.Disbursement,
) bool {
	switch disbursement.Status {
	case models.DisbursementStatusInitiated:
		return true
	case models.DisbursementStatusProcessing:
		return false
	case models.DisbursementStatusSuccess:
		return false
	case models.DisbursementStatusSuspended:
		if p.retryPolicy.IsRetryEligible(disbursement.UpdatedAt, disbursement.RetryCount) {
			return true
		}
		return false
	default:
		return false
	}
}

func (p PaymentServiceImpl) transitionToProcessing(
	ctx context.Context,
	disbursement *schema.Disbursement,
	channel models.PaymentChannel,
) error {
	return p.disbursement.Update(ctx, disbursement.Id, map[string]any{
		"status":     models.DisbursementStatusProcessing,
		"channel":    channel,
		"updated_at": time.Now(),
	})
}

func (p PaymentServiceImpl) selectChannel(
	loan *schema.Loan,
	disbursement *schema.Disbursement,
) models.PaymentChannel {
	if disbursement.RetryCount != 0 {
		return p.switchChannel(loan, disbursement)
	}
	if loan.Amount <= 100000 {
		return models.PaymentChannelUPI
	} else if loan.Amount <= 500000 {
		return models.PaymentChannelIMPS
	}
	return models.PaymentChannelNEFT
}

func (p PaymentServiceImpl) switchChannel(
	loan *schema.Loan,
	disbursement *schema.Disbursement,
) models.PaymentChannel {
	if disbursement.RetryCount == 2 && loan.Amount <= 100000 {
		return models.PaymentChannelIMPS
	}
	if loan.Amount <= 500000 {
		return models.PaymentChannelIMPS
	}
	return models.PaymentChannelNEFT
}

func (p PaymentServiceImpl) isChannelActive(
	ctx context.Context,
	channel models.PaymentChannel,
) bool {
	active, err := p.gatewayProvider.IsActive(ctx, channel)
	if err != nil {
		return false
	}
	return active
}

func (p PaymentServiceImpl) channelFallback(
	channel models.PaymentChannel,
	isChannelActive bool,
) (models.PaymentChannel, error) {
	if isChannelActive {
		return channel, nil
	}
	if channel == models.PaymentChannelUPI {
		return models.PaymentChannelIMPS, nil
	}
	return channel, nil
}

func (p PaymentServiceImpl) transfer(
	ctx context.Context,
	referenceId string,
	disbursement *schema.Disbursement,
	loan *schema.Loan,
	beneficiary *schema.Beneficiary,
	channel models.PaymentChannel,
) (models.PaymentResponse, error) {
	request := models.PaymentRequest{
		ReferenceID: referenceId,
		Amount:      loan.Amount,
		Channel:     channel,
		Beneficiary: models.Beneficiary{
			Name:    beneficiary.Name,
			Account: beneficiary.Account,
			IFSC:    beneficiary.IFSC,
			Bank:    beneficiary.Bank,
		},
		Metadata: models.PaymentMetadata{
			LoanID:          loan.Id,
			DisbursementID:  disbursement.Id,
			NotificationURL: p.notificationURL,
		},
	}

	return p.gatewayProvider.Transfer(ctx, request)
}
