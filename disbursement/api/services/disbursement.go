package services

import (
	"context"
	"errors"
	"fmt"
	"loan-disbursement-service/db/daos"
	"loan-disbursement-service/models"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type DisbursementService struct {
	loanDAO         *daos.LoanDAO
	disbursementDAO *daos.DisbursementDAO
	transactionDAO  *daos.TransactionDAO
	beneficiaryDAO  *daos.BeneficiaryDAO
}

func NewDisbursementService(
	loanDAO *daos.LoanDAO,
	disbursementDAO *daos.DisbursementDAO,
	transactionDAO *daos.TransactionDAO,
	beneficiaryDAO *daos.BeneficiaryDAO,
) *DisbursementService {
	return &DisbursementService{
		loanDAO:         loanDAO,
		disbursementDAO: disbursementDAO,
		transactionDAO:  transactionDAO,
		beneficiaryDAO:  beneficiaryDAO,
	}
}

func (d *DisbursementService) Disburse(
	ctx context.Context,
	req *models.DisburseRequest,
) (*models.DisbursementResponse, error) {
	existing, err := d.disbursementDAO.Get(ctx, req.LoanId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error().Err(err).Str("loan_id", req.LoanId).Msg("failed to check existing disbursement")
		return nil, fmt.Errorf("failed to check existing disbursement: %w", err)
	}

	if existing != nil {
		return &models.DisbursementResponse{
			DisbursementId: existing.Id,
			Status:         existing.Status,
			Message:        "Disbursement already exists",
		}, nil
	}

	loan, err := d.loanDAO.Get(ctx, req.LoanId)
	if err != nil {
		log.Error().Err(err).Str("loan_id", req.LoanId).Msg("failed to get loan")
		return nil, errors.New("invalid loan id")
	}

	if loan.Amount != req.Amount {
		return nil, fmt.Errorf("loan amount does not match disbursement amount")
	}

	if loan.BeneficiaryId == nil {
		beneficiary, err := d.beneficiaryDAO.CreateOrGet(
			ctx,
			req.BeneficiaryName,
			req.AccountNumber,
			req.IFSCCode,
			req.BeneficiaryBank,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create or get beneficiary: %w", err)
		}
		loan.BeneficiaryId = &beneficiary.Id
		_, err = d.loanDAO.Update(ctx, loan.Id, map[string]any{
			"beneficiary_id": beneficiary.Id,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to update loan: %w", err)
		}
	}

	disbursementId := fmt.Sprintf("DIS%s", uuid.New().String()[:12])
	_, err = d.disbursementDAO.Create(
		ctx,
		disbursementId,
		loan.Id,
		string(models.DisbursementStatusInitiated),
		req.Amount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create disbursement: %w", err)
	}

	return &models.DisbursementResponse{
		DisbursementId: disbursementId,
		Status:         string(models.DisbursementStatusInitiated),
		Message:        "Disbursement created",
	}, nil
}

func (d *DisbursementService) Fetch(ctx context.Context, disbursementId string) (any, error) {
	disbursement, err := d.disbursementDAO.Get(ctx, disbursementId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch disbursement: %w", err)
	}

	transactions, err := d.transactionDAO.ListByDisbursement(ctx, disbursementId)
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
					Mode:          transaction.Mode,
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

func (d *DisbursementService) Retry(ctx context.Context, disbursementId string) (any, error) {
	disbursement, err := d.disbursementDAO.Get(ctx, disbursementId)
	if err != nil {
		return nil, fmt.Errorf("failed to get disbursement: %w", err)
	}

	if disbursement.Status == string(models.DisbursementStatusProcessing) {
		return nil, fmt.Errorf("disbursement is in-progress")
	}

	if disbursement.Status == string(models.DisbursementStatusSuccess) {
		return nil, fmt.Errorf("disbursement is completed")
	}

	err = d.disbursementDAO.Update(ctx, disbursementId, map[string]any{
		"status":     string(models.DisbursementStatusInitiated),
		"last_error": nil,
		"updated_at": time.Now(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update disbursement: %w", err)
	}

	return &models.DisbursementResponse{
		DisbursementId: disbursementId,
		Status:         string(models.DisbursementStatusInitiated),
		Message:        "Disbursement retried",
	}, nil
}
