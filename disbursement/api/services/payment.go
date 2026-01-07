package services

import (
	"context"
	"fmt"
	"loan-disbursement-service/db/daos"
	"loan-disbursement-service/db/schema"
	"loan-disbursement-service/models"
	"loan-disbursement-service/providers"
	"strings"
	"time"

	"github.com/google/uuid"
)

type PaymentService struct {
	disbursementDAO *daos.DisbursementDAO
	transactionDAO  *daos.TransactionDAO
	loanDAO         *daos.LoanDAO
	beneficiaryDAO  *daos.BeneficiaryDAO
	retryPolicy     *RetryPolicy
	gatewayProvider providers.PaymentProvider
}

func NewPaymentService(
	loanDAO *daos.LoanDAO,
	retryPolicy *RetryPolicy,
	beneficiaryDAO *daos.BeneficiaryDAO,
	disbursementDAO *daos.DisbursementDAO,
	transactionDAO *daos.TransactionDAO,
	gatewayProvider providers.PaymentProvider,
) *PaymentService {
	return &PaymentService{
		disbursementDAO: disbursementDAO,
		loanDAO:         loanDAO,
		beneficiaryDAO:  beneficiaryDAO,
		retryPolicy:     retryPolicy,
		transactionDAO:  transactionDAO,
		gatewayProvider: gatewayProvider,
	}
}

func (p PaymentService) Process(
	ctx context.Context,
	disbursement *schema.Disbursement,
) error {
	if !p.shouldProcess(disbursement) {
		return nil
	}
	if err := p.transitionToProcessing(ctx, disbursement); err != nil {
		return fmt.Errorf("failed to transition to processing: %w", err)
	}

	loan, err := p.loanDAO.Get(ctx, disbursement.LoanId)
	if err != nil {
		return fmt.Errorf("failed to get loan: %w", err)
	}

	beneficiary, err := p.beneficiaryDAO.GetById(ctx, *loan.BeneficiaryId)
	if err != nil {
		return fmt.Errorf("failed to get beneficiary: %w", err)
	}

	channel := p.selectChannel(loan, disbursement)

	return p.execute(ctx, disbursement, loan, beneficiary, channel)
}

func (p PaymentService) execute(
	ctx context.Context,
	disbursement *schema.Disbursement,
	loan *schema.Loan,
	beneficiary *schema.Beneficiary,
	channel string,
) error {
	transactionId := fmt.Sprintf("TXN%s", uuid.New().String()[:12])
	referenceId := fmt.Sprintf("REF%s", uuid.New().String()[:12])
	transaction, err := p.transactionDAO.Create(ctx,
		transactionId,
		disbursement.Id,
		referenceId,
		channel,
		loan.Amount,
		string(models.TransactionStatusInitiated),
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}
	request := models.PaymentRequest{
		ReferenceID: referenceId,
		Amount:      loan.Amount,
		Mode:        channel,
		Beneficiary: models.Beneficiary{
			Name:    beneficiary.Name,
			Account: beneficiary.Account,
			IFSC:    beneficiary.IFSC,
			Bank:    beneficiary.Bank,
		},
		Metadata: map[string]any{
			"disbursement_id": disbursement.Id,
			"loan_id":         loan.Id,
		},
	}

	response, err := p.gatewayProvider.Transfer(ctx, request)
	if err != nil {
		return p.handleFailure(ctx, disbursement, transaction, err)
	}

	return p.handleResponse(ctx, disbursement, transaction, response)
}

func (p PaymentService) handleFailure(
	ctx context.Context,
	disbursement *schema.Disbursement,
	transaction *schema.Transaction,
	err error,
) error {
	if err.Error() == "Reference ID already processed" {
		payment, err := p.gatewayProvider.Fetch(ctx, transaction.ReferenceId)
		if err != nil {
			return fmt.Errorf("failed to get payment: %w", err)
		}
		if payment.Status == string(models.TransactionStatusSuccess) {
			return p.handleResponse(ctx, disbursement, transaction, payment)
		}
	}
	_ = p.transactionDAO.Update(ctx, transaction.Id, map[string]any{
		"status":     string(models.TransactionStatusFailed),
		"message":    err.Error(),
		"updated_at": time.Now(),
	})
	status, retryCount := p.evaluateFailure(disbursement.RetryCount, err)
	return p.disbursementDAO.Update(ctx, disbursement.Id, map[string]any{
		"status":      status,
		"retry_count": retryCount,
		"last_error":  err.Error(),
		"updated_at":  time.Now(),
	})
}

func (p PaymentService) evaluateFailure(retryCount int, err error) (string, int) {
	if retryCount >= MaxRetries {
		return string(models.DisbursementStatusFailed), retryCount
	}

	newRetryCount := retryCount + 1
	errMsg := err.Error()
	if strings.Contains(errMsg, "gateway error") {
		return string(models.DisbursementStatusSuspended), newRetryCount
	}

	retriableErrors := []string{
		"Limit Exceeded",
		"Inactive Beneficiary Account",
		"Beneficiary Bank is Down",
	}
	for _, retriableErr := range retriableErrors {
		if strings.Contains(errMsg, retriableErr) {
			return string(models.DisbursementStatusSuspended), newRetryCount
		}
	}

	return string(models.DisbursementStatusFailed), newRetryCount
}

func (p PaymentService) handleResponse(
	ctx context.Context,
	disbursement *schema.Disbursement,
	transaction *schema.Transaction,
	_ models.PaymentResponse,
) error {
	err := p.transactionDAO.Update(ctx,
		transaction.Id,
		map[string]any{
			"status":     string(models.TransactionStatusSuccess),
			"message":    nil,
			"updated_at": time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	_ = p.disbursementDAO.Update(ctx, disbursement.Id, map[string]any{
		"status":     models.DisbursementStatusSuccess,
		"last_error": nil,
		"updated_at": time.Now(),
	})

	return nil
}

func (p PaymentService) shouldProcess(
	disbursement *schema.Disbursement,
) bool {
	switch models.DisbursementStatus(disbursement.Status) {
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

func (p PaymentService) transitionToProcessing(
	ctx context.Context,
	disbursement *schema.Disbursement,
) error {
	disbursement.Status = string(models.DisbursementStatusProcessing)
	disbursement.UpdatedAt = time.Now()
	return p.disbursementDAO.Update(ctx, disbursement.Id, map[string]any{
		"status":     disbursement.Status,
		"updated_at": disbursement.UpdatedAt,
	})
}

func (p PaymentService) selectChannel(
	loan *schema.Loan,
	disbursement *schema.Disbursement,
) string {
	if disbursement.RetryCount != 0 {
		return p.switchChannel(loan, disbursement)
	}
	if loan.Amount <= 100000 {
		return "UPI"
	} else if loan.Amount <= 500000 {
		return "IMPS"
	}
	return "NEFT"
}

func (p PaymentService) switchChannel(
	loan *schema.Loan,
	disbursement *schema.Disbursement,
) string {
	if disbursement.RetryCount == 2 && loan.Amount <= 100000 {
		return "IMPS"
	}
	return "NEFT"
}
