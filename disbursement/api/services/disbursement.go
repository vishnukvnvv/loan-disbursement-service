package services

import (
	"context"
	"errors"
	"fmt"
	"loan-disbursement-service/db/daos"
	"loan-disbursement-service/models"
	"loan-disbursement-service/utils"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type DisbursementService interface {
	Disburse(ctx context.Context, req *models.DisburseRequest) (*models.DisbursementResponse, error)
	Fetch(ctx context.Context, disbursementId string) (any, error)
	Retry(ctx context.Context, disbursementId string) (any, error)
}

type DisbursementServiceImpl struct {
	idGenerator  utils.IdGenerator
	loan         daos.LoanRepository
	disbursement daos.DisbursementRepository
	transaction  daos.TransactionRepository
	beneficiary  daos.BeneficiaryRepository
	paymentChan  chan string
}

func NewDisbursementService(
	idGenerator utils.IdGenerator,
	loan daos.LoanRepository,
	disbursement daos.DisbursementRepository,
	transaction daos.TransactionRepository,
	beneficiary daos.BeneficiaryRepository,
	paymentChan chan string,
) DisbursementService {
	return &DisbursementServiceImpl{
		idGenerator:  idGenerator,
		loan:         loan,
		disbursement: disbursement,
		transaction:  transaction,
		beneficiary:  beneficiary,
		paymentChan:  paymentChan,
	}
}

func (d *DisbursementServiceImpl) Disburse(
	ctx context.Context,
	req *models.DisburseRequest,
) (*models.DisbursementResponse, error) {
	existing, err := d.disbursement.GetByLoanId(ctx, req.LoanId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error().Err(err).Str("loan_id", req.LoanId).Msg("failed to check existing disbursement")
		return nil, fmt.Errorf("failed to check existing disbursement: %w", err)
	}
	log.Info().Msgf("existing disbursement: %+v", existing)
	if existing != nil {
		return &models.DisbursementResponse{
			DisbursementId: existing.Id,
			Status:         existing.Status,
			Message:        "Disbursement already exists",
		}, nil
	}

	loan, err := d.loan.Get(ctx, req.LoanId)
	if err != nil {
		log.Error().Err(err).Str("loan_id", req.LoanId).Msg("failed to get loan")
		return nil, errors.New("invalid loan id")
	}

	if loan.Amount != req.Amount {
		return nil, fmt.Errorf("loan amount does not match disbursement amount")
	}

	if loan.BeneficiaryId == nil {
		beneficiaryId := d.idGenerator.GenerateBeneficiaryId()
		beneficiary, err := d.beneficiary.CreateOrGet(
			ctx,
			beneficiaryId,
			req.BeneficiaryName,
			req.AccountNumber,
			req.IFSCCode,
			req.BeneficiaryBank,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create or get beneficiary: %w", err)
		}
		loan.BeneficiaryId = &beneficiary.Id
		_, err = d.loan.Update(ctx, loan.Id, map[string]any{
			"beneficiary_id": beneficiary.Id,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to update loan: %w", err)
		}
	}

	disbursementId := d.idGenerator.GenerateDisbursementId()
	channel := d.selectChannel(req.Amount)
	_, err = d.disbursement.Create(
		ctx,
		disbursementId,
		loan.Id,
		channel,
		models.DisbursementStatusInitiated,
		req.Amount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create disbursement: %w", err)
	}
	d.paymentChan <- disbursementId

	return &models.DisbursementResponse{
		DisbursementId: disbursementId,
		Status:         models.DisbursementStatusInitiated,
		Message:        "Disbursement created",
	}, nil
}

func (d *DisbursementServiceImpl) Fetch(ctx context.Context, disbursementId string) (any, error) {
	disbursement, err := d.disbursement.Get(ctx, disbursementId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch disbursement: %w", err)
	}

	transactions, err := d.transaction.ListByDisbursement(ctx, disbursementId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions: %w", err)
	}

	return models.Disbursement{
		DisbursementId: disbursement.Id,
		Amount:         disbursement.Amount,
		LoanId:         disbursement.LoanId,
		Status:         disbursement.Status,
		Transaction: func() []models.TransactionResponse {
			txs := make([]models.TransactionResponse, len(transactions))
			for i, transaction := range transactions {
				txs[i] = models.TransactionResponse{
					TransactionId: transaction.Id,
					Status:        transaction.Status,
					Channel:       transaction.Channel,
					Message:       transaction.Message,
					CreatedAt:     transaction.CreatedAt,
					UpdatedAt:     transaction.UpdatedAt,
				}
			}
			return txs
		}(),
		CreatedAt: disbursement.CreatedAt,
		UpdatedAt: disbursement.UpdatedAt,
	}, nil
}

func (d *DisbursementServiceImpl) Retry(ctx context.Context, disbursementId string) (any, error) {
	disbursement, err := d.disbursement.Get(ctx, disbursementId)
	if err != nil {
		return nil, fmt.Errorf("failed to get disbursement: %w", err)
	}

	if disbursement.Status == models.DisbursementStatusProcessing {
		return nil, fmt.Errorf("disbursement is in-progress")
	}

	if disbursement.Status == models.DisbursementStatusSuccess {
		return nil, fmt.Errorf("disbursement is completed")
	}

	err = d.disbursement.Update(ctx, disbursementId, map[string]any{
		"status":     string(models.DisbursementStatusInitiated),
		"last_error": nil,
		"updated_at": time.Now(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update disbursement: %w", err)
	}

	return &models.DisbursementResponse{
		DisbursementId: disbursementId,
		Status:         models.DisbursementStatusInitiated,
		Message:        "Disbursement retried",
	}, nil
}

func (d *DisbursementServiceImpl) selectChannel(
	amount float64,
) models.PaymentChannel {
	if amount <= 100000 {
		return models.PaymentChannelUPI
	}
	if amount <= 500000 {
		return models.PaymentChannelIMPS
	}
	return models.PaymentChannelNEFT
}
