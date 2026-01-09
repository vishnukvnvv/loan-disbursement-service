package services

import (
	"context"
	"errors"
	"loan-disbursement-service/db/schema"
	"loan-disbursement-service/models"
	db_test "loan-disbursement-service/test/db"
	utils_test "loan-disbursement-service/test/utils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestDisbursementService_Disburse(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully creates disbursement when loan has beneficiary", func(t *testing.T) {
		mockIdGenerator := new(utils_test.MockIdGenerator)
		mockLoan := new(db_test.MockLoanRepository)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)
		paymentChan := make(chan string, 1)

		service := NewDisbursementService(
			mockIdGenerator,
			mockLoan,
			mockDisbursement,
			mockTransaction,
			mockBeneficiary,
			paymentChan,
		)

		loanId := "LOAN-123456789012"
		beneficiaryId := "BEN-987654321098"
		disbursementId := "DISB-123456789012"

		request := &models.DisburseRequest{
			LoanId:          loanId,
			Amount:          10000.0,
			AccountNumber:   "1234567890",
			IFSCCode:        "IFSC0001234",
			BeneficiaryName: "John Doe",
			BeneficiaryBank: "Test Bank",
		}

		loan := &schema.Loan{
			Id:            loanId,
			Amount:        request.Amount,
			BeneficiaryId: &beneficiaryId,
			CreatedAt:     time.Now().Add(-24 * time.Hour),
			UpdatedAt:     time.Now().Add(-24 * time.Hour),
		}

		disbursement := schema.Disbursement{
			Id:         disbursementId,
			LoanId:     loanId,
			Amount:     request.Amount,
			Status:     models.DisbursementStatusInitiated,
			RetryCount: 0,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		mockDisbursement.On("Get", ctx, loanId).
			Return(nil, gorm.ErrRecordNotFound).Once()
		mockLoan.On("Get", ctx, loanId).
			Return(loan, nil).Once()
		mockIdGenerator.On("GenerateDisbursementId").Return(disbursementId).Once()
		mockDisbursement.On("Create", ctx, disbursementId, loanId, models.PaymentChannelUPI, models.DisbursementStatusInitiated, request.Amount).
			Return(&disbursement, nil).
			Once()

		result, err := service.Disburse(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, disbursementId, result.DisbursementId)
		assert.Equal(t, models.DisbursementStatusInitiated, result.Status)
		assert.Equal(t, "Disbursement created", result.Message)

		// Drain the payment channel
		select {
		case <-paymentChan:
		default:
		}

		mockDisbursement.AssertExpectations(t)
		mockLoan.AssertExpectations(t)
		mockIdGenerator.AssertExpectations(t)
	})

	t.Run(
		"successfully creates disbursement and creates beneficiary when loan has no beneficiary",
		func(t *testing.T) {
			mockIdGenerator := new(utils_test.MockIdGenerator)
			mockLoan := new(db_test.MockLoanRepository)
			mockDisbursement := new(db_test.MockDisbursementRepository)
			mockTransaction := new(db_test.MockTransactionRepository)
			mockBeneficiary := new(db_test.MockBeneficiaryRepository)
			paymentChan := make(chan string, 1)

			service := NewDisbursementService(
				mockIdGenerator,
				mockLoan,
				mockDisbursement,
				mockTransaction,
				mockBeneficiary,
				paymentChan,
			)

			loanId := "LOAN-123456789012"
			beneficiaryId := "BEN-987654321098"
			disbursementId := "DISB-123456789012"

			request := &models.DisburseRequest{
				LoanId:          loanId,
				Amount:          10000.0,
				AccountNumber:   "1234567890",
				IFSCCode:        "IFSC0001234",
				BeneficiaryName: "John Doe",
				BeneficiaryBank: "Test Bank",
			}

			loan := &schema.Loan{
				Id:            loanId,
				Amount:        request.Amount,
				BeneficiaryId: nil,
				CreatedAt:     time.Now().Add(-24 * time.Hour),
				UpdatedAt:     time.Now().Add(-24 * time.Hour),
			}

			updatedLoan := &schema.Loan{
				Id:            loanId,
				Amount:        request.Amount,
				BeneficiaryId: &beneficiaryId,
				CreatedAt:     time.Now().Add(-24 * time.Hour),
				UpdatedAt:     time.Now(),
			}

			beneficiary := schema.Beneficiary{
				Id:        beneficiaryId,
				Name:      request.BeneficiaryName,
				Account:   request.AccountNumber,
				IFSC:      request.IFSCCode,
				Bank:      request.BeneficiaryBank,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			disbursement := schema.Disbursement{
				Id:         disbursementId,
				LoanId:     loanId,
				Amount:     request.Amount,
				Status:     models.DisbursementStatusInitiated,
				RetryCount: 0,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}

			mockDisbursement.On("Get", ctx, loanId).
				Return(nil, gorm.ErrRecordNotFound).Once()
			mockLoan.On("Get", ctx, loanId).
				Return(loan, nil).Once()
			mockIdGenerator.On("GenerateBeneficiaryId").Return(beneficiaryId).Once()
			mockBeneficiary.On("CreateOrGet", ctx, beneficiaryId, request.BeneficiaryName, request.AccountNumber, request.IFSCCode, request.BeneficiaryBank).
				Return(&beneficiary, nil).
				Once()
			mockLoan.On("Update", ctx, loanId, map[string]any{"beneficiary_id": beneficiaryId}).
				Return(updatedLoan, nil).Once()
			mockIdGenerator.On("GenerateDisbursementId").Return(disbursementId).Once()
			mockDisbursement.On("Create", ctx, disbursementId, loanId, models.PaymentChannelUPI, models.DisbursementStatusInitiated, request.Amount).
				Return(&disbursement, nil).
				Once()

			result, err := service.Disburse(ctx, request)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, disbursementId, result.DisbursementId)
			assert.Equal(t, models.DisbursementStatusInitiated, result.Status)

			// Drain the payment channel
			select {
			case <-paymentChan:
			default:
			}

			mockDisbursement.AssertExpectations(t)
			mockLoan.AssertExpectations(t)
			mockBeneficiary.AssertExpectations(t)
			mockIdGenerator.AssertExpectations(t)
		},
	)

	t.Run("returns existing disbursement when it already exists", func(t *testing.T) {
		mockIdGenerator := new(utils_test.MockIdGenerator)
		mockLoan := new(db_test.MockLoanRepository)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)

		paymentChan := make(chan string, 1)
		service := NewDisbursementService(
			mockIdGenerator,
			mockLoan,
			mockDisbursement,
			mockTransaction,
			mockBeneficiary,
			paymentChan,
		)

		loanId := "LOAN-123456789012"
		disbursementId := "DISB-EXISTING"

		request := &models.DisburseRequest{
			LoanId: loanId,
			Amount: 10000.0,
		}

		existingDisbursement := schema.Disbursement{
			Id:         disbursementId,
			LoanId:     loanId,
			Amount:     10000.0,
			Status:     models.DisbursementStatusProcessing,
			RetryCount: 0,
			CreatedAt:  time.Now().Add(-1 * time.Hour),
			UpdatedAt:  time.Now(),
		}

		mockDisbursement.On("Get", ctx, loanId).
			Return(&existingDisbursement, nil).Once()

		result, err := service.Disburse(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, disbursementId, result.DisbursementId)
		assert.Equal(t, models.DisbursementStatusProcessing, result.Status)
		assert.Equal(t, "Disbursement already exists", result.Message)

		mockDisbursement.AssertExpectations(t)
		mockLoan.AssertNotCalled(t, "Get")
	})

	t.Run("returns error when checking existing disbursement fails", func(t *testing.T) {
		mockIdGenerator := new(utils_test.MockIdGenerator)
		mockLoan := new(db_test.MockLoanRepository)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)

		paymentChan := make(chan string, 1)
		service := NewDisbursementService(
			mockIdGenerator,
			mockLoan,
			mockDisbursement,
			mockTransaction,
			mockBeneficiary,
			paymentChan,
		)

		loanId := "LOAN-123456789012"

		request := &models.DisburseRequest{
			LoanId: loanId,
			Amount: 10000.0,
		}

		repoError := errors.New("database connection error")

		mockDisbursement.On("Get", ctx, loanId).
			Return(nil, repoError).Once()

		result, err := service.Disburse(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to check existing disbursement")

		mockDisbursement.AssertExpectations(t)
		mockLoan.AssertNotCalled(t, "Get")
	})

	t.Run("returns error when loan not found", func(t *testing.T) {
		mockIdGenerator := new(utils_test.MockIdGenerator)
		mockLoan := new(db_test.MockLoanRepository)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)

		paymentChan := make(chan string, 1)
		service := NewDisbursementService(
			mockIdGenerator,
			mockLoan,
			mockDisbursement,
			mockTransaction,
			mockBeneficiary,
			paymentChan,
		)

		loanId := "LOAN-NONEXISTENT"

		request := &models.DisburseRequest{
			LoanId: loanId,
			Amount: 10000.0,
		}

		mockDisbursement.On("Get", ctx, loanId).
			Return(nil, gorm.ErrRecordNotFound).Once()
		mockLoan.On("Get", ctx, loanId).
			Return(nil, gorm.ErrRecordNotFound).Once()

		result, err := service.Disburse(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "invalid loan id", err.Error())

		mockDisbursement.AssertExpectations(t)
		mockLoan.AssertExpectations(t)
	})

	t.Run("returns error when loan amount does not match", func(t *testing.T) {
		mockIdGenerator := new(utils_test.MockIdGenerator)
		mockLoan := new(db_test.MockLoanRepository)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)

		paymentChan := make(chan string, 1)
		service := NewDisbursementService(
			mockIdGenerator,
			mockLoan,
			mockDisbursement,
			mockTransaction,
			mockBeneficiary,
			paymentChan,
		)

		loanId := "LOAN-123456789012"
		beneficiaryId := "BEN-987654321098"

		request := &models.DisburseRequest{
			LoanId: loanId,
			Amount: 10000.0,
		}

		loan := &schema.Loan{
			Id:            loanId,
			Amount:        15000.0, // Different amount
			BeneficiaryId: &beneficiaryId,
			CreatedAt:     time.Now().Add(-24 * time.Hour),
			UpdatedAt:     time.Now().Add(-24 * time.Hour),
		}

		mockDisbursement.On("Get", ctx, loanId).
			Return(nil, gorm.ErrRecordNotFound).Once()
		mockLoan.On("Get", ctx, loanId).
			Return(loan, nil).Once()

		result, err := service.Disburse(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "loan amount does not match disbursement amount")

		mockDisbursement.AssertExpectations(t)
		mockLoan.AssertExpectations(t)
	})

	t.Run("returns error when beneficiary creation fails", func(t *testing.T) {
		mockIdGenerator := new(utils_test.MockIdGenerator)
		mockLoan := new(db_test.MockLoanRepository)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)

		paymentChan := make(chan string, 1)
		service := NewDisbursementService(
			mockIdGenerator,
			mockLoan,
			mockDisbursement,
			mockTransaction,
			mockBeneficiary,
			paymentChan,
		)

		loanId := "LOAN-123456789012"
		beneficiaryId := "BEN-987654321098"

		request := &models.DisburseRequest{
			LoanId:          loanId,
			Amount:          10000.0,
			AccountNumber:   "1234567890",
			IFSCCode:        "IFSC0001234",
			BeneficiaryName: "John Doe",
			BeneficiaryBank: "Test Bank",
		}

		loan := &schema.Loan{
			Id:            loanId,
			Amount:        request.Amount,
			BeneficiaryId: nil,
			CreatedAt:     time.Now().Add(-24 * time.Hour),
			UpdatedAt:     time.Now().Add(-24 * time.Hour),
		}

		repoError := errors.New("database error")

		mockDisbursement.On("Get", ctx, loanId).
			Return(nil, gorm.ErrRecordNotFound).Once()
		mockLoan.On("Get", ctx, loanId).
			Return(loan, nil).Once()
		mockIdGenerator.On("GenerateBeneficiaryId").Return(beneficiaryId).Once()
		mockBeneficiary.On("CreateOrGet", ctx, beneficiaryId, request.BeneficiaryName, request.AccountNumber, request.IFSCCode, request.BeneficiaryBank).
			Return(nil, repoError).
			Once()

		result, err := service.Disburse(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to create or get beneficiary")

		mockDisbursement.AssertExpectations(t)
		mockLoan.AssertExpectations(t)
		mockBeneficiary.AssertExpectations(t)
	})

	t.Run("returns error when loan update fails after beneficiary creation", func(t *testing.T) {
		mockIdGenerator := new(utils_test.MockIdGenerator)
		mockLoan := new(db_test.MockLoanRepository)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)

		paymentChan := make(chan string, 1)
		service := NewDisbursementService(
			mockIdGenerator,
			mockLoan,
			mockDisbursement,
			mockTransaction,
			mockBeneficiary,
			paymentChan,
		)

		loanId := "LOAN-123456789012"
		beneficiaryId := "BEN-987654321098"

		request := &models.DisburseRequest{
			LoanId:          loanId,
			Amount:          10000.0,
			AccountNumber:   "1234567890",
			IFSCCode:        "IFSC0001234",
			BeneficiaryName: "John Doe",
			BeneficiaryBank: "Test Bank",
		}

		loan := &schema.Loan{
			Id:            loanId,
			Amount:        request.Amount,
			BeneficiaryId: nil,
			CreatedAt:     time.Now().Add(-24 * time.Hour),
			UpdatedAt:     time.Now().Add(-24 * time.Hour),
		}

		beneficiary := schema.Beneficiary{
			Id:        beneficiaryId,
			Name:      request.BeneficiaryName,
			Account:   request.AccountNumber,
			IFSC:      request.IFSCCode,
			Bank:      request.BeneficiaryBank,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		repoError := errors.New("database error")

		mockDisbursement.On("Get", ctx, loanId).
			Return(nil, gorm.ErrRecordNotFound).Once()
		mockLoan.On("Get", ctx, loanId).
			Return(loan, nil).Once()
		mockIdGenerator.On("GenerateBeneficiaryId").Return(beneficiaryId).Once()
		mockBeneficiary.On("CreateOrGet", ctx, beneficiaryId, request.BeneficiaryName, request.AccountNumber, request.IFSCCode, request.BeneficiaryBank).
			Return(&beneficiary, nil).
			Once()
		mockLoan.On("Update", ctx, loanId, map[string]any{"beneficiary_id": beneficiaryId}).
			Return(nil, repoError).Once()

		result, err := service.Disburse(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to update loan")

		mockDisbursement.AssertExpectations(t)
		mockLoan.AssertExpectations(t)
		mockBeneficiary.AssertExpectations(t)
	})

	t.Run("returns error when disbursement creation fails", func(t *testing.T) {
		mockIdGenerator := new(utils_test.MockIdGenerator)
		mockLoan := new(db_test.MockLoanRepository)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)

		paymentChan := make(chan string, 1)
		service := NewDisbursementService(
			mockIdGenerator,
			mockLoan,
			mockDisbursement,
			mockTransaction,
			mockBeneficiary,
			paymentChan,
		)

		loanId := "LOAN-123456789012"
		beneficiaryId := "BEN-987654321098"
		disbursementId := "DISB-123456789012"

		request := &models.DisburseRequest{
			LoanId: loanId,
			Amount: 10000.0,
		}

		loan := &schema.Loan{
			Id:            loanId,
			Amount:        request.Amount,
			BeneficiaryId: &beneficiaryId,
			CreatedAt:     time.Now().Add(-24 * time.Hour),
			UpdatedAt:     time.Now().Add(-24 * time.Hour),
		}

		repoError := errors.New("database error")

		mockDisbursement.On("Get", ctx, loanId).
			Return(nil, gorm.ErrRecordNotFound).Once()
		mockLoan.On("Get", ctx, loanId).
			Return(loan, nil).Once()
		mockIdGenerator.On("GenerateDisbursementId").Return(disbursementId).Once()
		mockDisbursement.On("Create", ctx, disbursementId, loanId, models.PaymentChannelUPI, models.DisbursementStatusInitiated, request.Amount).
			Return(nil, repoError).
			Once()

		result, err := service.Disburse(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to create disbursement")

		mockDisbursement.AssertExpectations(t)
		mockLoan.AssertExpectations(t)
		mockIdGenerator.AssertExpectations(t)
	})

	t.Run("selects UPI channel for amount <= 100000", func(t *testing.T) {
		mockIdGenerator := new(utils_test.MockIdGenerator)
		mockLoan := new(db_test.MockLoanRepository)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)
		paymentChan := make(chan string, 1)

		service := NewDisbursementService(
			mockIdGenerator,
			mockLoan,
			mockDisbursement,
			mockTransaction,
			mockBeneficiary,
			paymentChan,
		)

		loanId := "LOAN-123456789012"
		beneficiaryId := "BEN-987654321098"
		disbursementId := "DISB-123456789012"

		request := &models.DisburseRequest{
			LoanId: loanId,
			Amount: 100000.0, // Exactly 100000
		}

		loan := &schema.Loan{
			Id:            loanId,
			Amount:        request.Amount,
			BeneficiaryId: &beneficiaryId,
			CreatedAt:     time.Now().Add(-24 * time.Hour),
			UpdatedAt:     time.Now().Add(-24 * time.Hour),
		}

		disbursement := schema.Disbursement{
			Id:         disbursementId,
			LoanId:     loanId,
			Amount:     request.Amount,
			Status:     models.DisbursementStatusInitiated,
			RetryCount: 0,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		mockDisbursement.On("Get", ctx, loanId).
			Return(nil, gorm.ErrRecordNotFound).Once()
		mockLoan.On("Get", ctx, loanId).
			Return(loan, nil).Once()
		mockIdGenerator.On("GenerateDisbursementId").Return(disbursementId).Once()
		mockDisbursement.On("Create", ctx, disbursementId, loanId, models.PaymentChannelUPI, models.DisbursementStatusInitiated, request.Amount).
			Return(&disbursement, nil).
			Once()

		result, err := service.Disburse(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		select {
		case <-paymentChan:
		default:
		}

		mockDisbursement.AssertExpectations(t)
		mockLoan.AssertExpectations(t)
		mockIdGenerator.AssertExpectations(t)
	})

	t.Run("selects IMPS channel for amount > 100000 and <= 500000", func(t *testing.T) {
		mockIdGenerator := new(utils_test.MockIdGenerator)
		mockLoan := new(db_test.MockLoanRepository)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)
		paymentChan := make(chan string, 1)

		service := NewDisbursementService(
			mockIdGenerator,
			mockLoan,
			mockDisbursement,
			mockTransaction,
			mockBeneficiary,
			paymentChan,
		)

		loanId := "LOAN-123456789012"
		beneficiaryId := "BEN-987654321098"
		disbursementId := "DISB-123456789012"

		request := &models.DisburseRequest{
			LoanId: loanId,
			Amount: 300000.0, // Between 100000 and 500000
		}

		loan := &schema.Loan{
			Id:            loanId,
			Amount:        request.Amount,
			BeneficiaryId: &beneficiaryId,
			CreatedAt:     time.Now().Add(-24 * time.Hour),
			UpdatedAt:     time.Now().Add(-24 * time.Hour),
		}

		disbursement := schema.Disbursement{
			Id:         disbursementId,
			LoanId:     loanId,
			Amount:     request.Amount,
			Status:     models.DisbursementStatusInitiated,
			RetryCount: 0,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		mockDisbursement.On("Get", ctx, loanId).
			Return(nil, gorm.ErrRecordNotFound).Once()
		mockLoan.On("Get", ctx, loanId).
			Return(loan, nil).Once()
		mockIdGenerator.On("GenerateDisbursementId").Return(disbursementId).Once()
		mockDisbursement.On("Create", ctx, disbursementId, loanId, models.PaymentChannelIMPS, models.DisbursementStatusInitiated, request.Amount).
			Return(&disbursement, nil).
			Once()

		result, err := service.Disburse(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		select {
		case <-paymentChan:
		default:
		}

		mockDisbursement.AssertExpectations(t)
		mockLoan.AssertExpectations(t)
		mockIdGenerator.AssertExpectations(t)
	})

	t.Run("selects NEFT channel for amount > 500000", func(t *testing.T) {
		mockIdGenerator := new(utils_test.MockIdGenerator)
		mockLoan := new(db_test.MockLoanRepository)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)
		paymentChan := make(chan string, 1)

		service := NewDisbursementService(
			mockIdGenerator,
			mockLoan,
			mockDisbursement,
			mockTransaction,
			mockBeneficiary,
			paymentChan,
		)

		loanId := "LOAN-123456789012"
		beneficiaryId := "BEN-987654321098"
		disbursementId := "DISB-123456789012"

		request := &models.DisburseRequest{
			LoanId: loanId,
			Amount: 600000.0, // Greater than 500000
		}

		loan := &schema.Loan{
			Id:            loanId,
			Amount:        request.Amount,
			BeneficiaryId: &beneficiaryId,
			CreatedAt:     time.Now().Add(-24 * time.Hour),
			UpdatedAt:     time.Now().Add(-24 * time.Hour),
		}

		disbursement := schema.Disbursement{
			Id:         disbursementId,
			LoanId:     loanId,
			Amount:     request.Amount,
			Status:     models.DisbursementStatusInitiated,
			RetryCount: 0,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		mockDisbursement.On("Get", ctx, loanId).
			Return(nil, gorm.ErrRecordNotFound).Once()
		mockLoan.On("Get", ctx, loanId).
			Return(loan, nil).Once()
		mockIdGenerator.On("GenerateDisbursementId").Return(disbursementId).Once()
		mockDisbursement.On("Create", ctx, disbursementId, loanId, models.PaymentChannelNEFT, models.DisbursementStatusInitiated, request.Amount).
			Return(&disbursement, nil).
			Once()

		result, err := service.Disburse(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		select {
		case <-paymentChan:
		default:
		}

		mockDisbursement.AssertExpectations(t)
		mockLoan.AssertExpectations(t)
		mockIdGenerator.AssertExpectations(t)
	})
}

func TestDisbursementService_Fetch(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully fetches disbursement with transactions", func(t *testing.T) {
		mockIdGenerator := new(utils_test.MockIdGenerator)
		mockLoan := new(db_test.MockLoanRepository)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)

		paymentChan := make(chan string, 1)
		service := NewDisbursementService(
			mockIdGenerator,
			mockLoan,
			mockDisbursement,
			mockTransaction,
			mockBeneficiary,
			paymentChan,
		)

		disbursementId := "DISB-123456789012"
		loanId := "LOAN-123456789012"

		disbursement := schema.Disbursement{
			Id:         disbursementId,
			LoanId:     loanId,
			Amount:     10000.0,
			Status:     models.DisbursementStatusProcessing,
			RetryCount: 1,
			CreatedAt:  time.Now().Add(-2 * time.Hour),
			UpdatedAt:  time.Now().Add(-1 * time.Hour),
		}

		transactions := []schema.Transaction{
			{
				Id:             "TXN-001",
				DisbursementId: disbursementId,
				ReferenceId:    "REF-001",
				Amount:         5000.0,
				Channel:        models.PaymentChannelUPI,
				Status:         models.TransactionStatusSuccess,
				Message:        nil,
				CreatedAt:      time.Now().Add(-1 * time.Hour),
				UpdatedAt:      time.Now().Add(-30 * time.Minute),
			},
			{
				Id:             "TXN-002",
				DisbursementId: disbursementId,
				ReferenceId:    "REF-002",
				Amount:         5000.0,
				Channel:        models.PaymentChannelNEFT,
				Status:         models.TransactionStatusFailed,
				Message:        stringPtr("Insufficient balance"),
				CreatedAt:      time.Now().Add(-45 * time.Minute),
				UpdatedAt:      time.Now().Add(-20 * time.Minute),
			},
		}

		mockDisbursement.On("Get", ctx, disbursementId).
			Return(&disbursement, nil).Once()
		mockTransaction.On("ListByDisbursement", ctx, disbursementId).
			Return(transactions, nil).Once()

		result, err := service.Fetch(ctx, disbursementId)

		assert.NoError(t, err)
		assert.NotNil(t, result)

		disbursementResult, ok := result.(models.Disbursement)
		assert.True(t, ok)
		assert.Equal(t, disbursementId, disbursementResult.DisbursementId)
		assert.Equal(t, loanId, disbursementResult.LoanId)
		assert.Equal(t, 10000.0, disbursementResult.Amount)
		assert.Equal(t, models.DisbursementStatusProcessing, disbursementResult.Status)
		assert.Len(t, disbursementResult.Transaction, 2)
		assert.Equal(t, "TXN-001", disbursementResult.Transaction[0].TransactionId)
		assert.Equal(t, models.TransactionStatusSuccess, disbursementResult.Transaction[0].Status)
		assert.Equal(t, "TXN-002", disbursementResult.Transaction[1].TransactionId)
		assert.Equal(t, models.TransactionStatusFailed, disbursementResult.Transaction[1].Status)

		mockDisbursement.AssertExpectations(t)
		mockTransaction.AssertExpectations(t)
	})

	t.Run("successfully fetches disbursement with no transactions", func(t *testing.T) {
		mockIdGenerator := new(utils_test.MockIdGenerator)
		mockLoan := new(db_test.MockLoanRepository)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)

		paymentChan := make(chan string, 1)
		service := NewDisbursementService(
			mockIdGenerator,
			mockLoan,
			mockDisbursement,
			mockTransaction,
			mockBeneficiary,
			paymentChan,
		)

		disbursementId := "DISB-123456789012"
		loanId := "LOAN-123456789012"

		disbursement := schema.Disbursement{
			Id:         disbursementId,
			LoanId:     loanId,
			Amount:     10000.0,
			Status:     models.DisbursementStatusInitiated,
			RetryCount: 0,
			CreatedAt:  time.Now().Add(-1 * time.Hour),
			UpdatedAt:  time.Now().Add(-1 * time.Hour),
		}

		transactions := []schema.Transaction{}

		mockDisbursement.On("Get", ctx, disbursementId).
			Return(&disbursement, nil).Once()
		mockTransaction.On("ListByDisbursement", ctx, disbursementId).
			Return(transactions, nil).Once()

		result, err := service.Fetch(ctx, disbursementId)

		assert.NoError(t, err)
		assert.NotNil(t, result)

		disbursementResult, ok := result.(models.Disbursement)
		assert.True(t, ok)
		assert.Equal(t, disbursementId, disbursementResult.DisbursementId)
		assert.Len(t, disbursementResult.Transaction, 0)

		mockDisbursement.AssertExpectations(t)
		mockTransaction.AssertExpectations(t)
	})

	t.Run("returns error when disbursement not found", func(t *testing.T) {
		mockIdGenerator := new(utils_test.MockIdGenerator)
		mockLoan := new(db_test.MockLoanRepository)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)

		paymentChan := make(chan string, 1)
		service := NewDisbursementService(
			mockIdGenerator,
			mockLoan,
			mockDisbursement,
			mockTransaction,
			mockBeneficiary,
			paymentChan,
		)

		disbursementId := "DISB-NONEXISTENT"

		mockDisbursement.On("Get", ctx, disbursementId).
			Return(nil, gorm.ErrRecordNotFound).Once()

		result, err := service.Fetch(ctx, disbursementId)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to fetch disbursement")

		mockDisbursement.AssertExpectations(t)
		mockTransaction.AssertNotCalled(t, "ListByDisbursement")
	})

	t.Run("returns error when transaction list fails", func(t *testing.T) {
		mockIdGenerator := new(utils_test.MockIdGenerator)
		mockLoan := new(db_test.MockLoanRepository)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)

		paymentChan := make(chan string, 1)
		service := NewDisbursementService(
			mockIdGenerator,
			mockLoan,
			mockDisbursement,
			mockTransaction,
			mockBeneficiary,
			paymentChan,
		)

		disbursementId := "DISB-123456789012"

		disbursement := schema.Disbursement{
			Id:         disbursementId,
			LoanId:     "LOAN-123456789012",
			Amount:     10000.0,
			Status:     models.DisbursementStatusInitiated,
			RetryCount: 0,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		repoError := errors.New("database error")

		mockDisbursement.On("Get", ctx, disbursementId).
			Return(&disbursement, nil).Once()
		mockTransaction.On("ListByDisbursement", ctx, disbursementId).
			Return([]schema.Transaction{}, repoError).Once()

		result, err := service.Fetch(ctx, disbursementId)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to fetch transactions")

		mockDisbursement.AssertExpectations(t)
		mockTransaction.AssertExpectations(t)
	})
}

func TestDisbursementService_Retry(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully retries disbursement", func(t *testing.T) {
		mockIdGenerator := new(utils_test.MockIdGenerator)
		mockLoan := new(db_test.MockLoanRepository)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)

		paymentChan := make(chan string, 1)
		service := NewDisbursementService(
			mockIdGenerator,
			mockLoan,
			mockDisbursement,
			mockTransaction,
			mockBeneficiary,
			paymentChan,
		)

		disbursementId := "DISB-123456789012"

		disbursement := schema.Disbursement{
			Id:         disbursementId,
			LoanId:     "LOAN-123456789012",
			Amount:     10000.0,
			Status:     models.DisbursementStatusFailed,
			RetryCount: 2,
			LastError:  stringPtr("Network error"),
			CreatedAt:  time.Now().Add(-2 * time.Hour),
			UpdatedAt:  time.Now().Add(-1 * time.Hour),
		}

		mockDisbursement.On("Get", ctx, disbursementId).
			Return(&disbursement, nil).Once()
		mockDisbursement.On("Update", ctx, disbursementId, mock.MatchedBy(func(fields map[string]any) bool {
			return fields["status"] == string(models.DisbursementStatusInitiated) &&
				fields["last_error"] == nil &&
				fields["updated_at"] != nil
		})).
			Return(nil).
			Once()

		result, err := service.Retry(ctx, disbursementId)

		assert.NoError(t, err)
		assert.NotNil(t, result)

		response, ok := result.(*models.DisbursementResponse)
		assert.True(t, ok)
		assert.Equal(t, disbursementId, response.DisbursementId)
		assert.Equal(t, models.DisbursementStatusInitiated, response.Status)
		assert.Equal(t, "Disbursement retried", response.Message)

		mockDisbursement.AssertExpectations(t)
	})

	t.Run("returns error when disbursement not found", func(t *testing.T) {
		mockIdGenerator := new(utils_test.MockIdGenerator)
		mockLoan := new(db_test.MockLoanRepository)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)

		paymentChan := make(chan string, 1)
		service := NewDisbursementService(
			mockIdGenerator,
			mockLoan,
			mockDisbursement,
			mockTransaction,
			mockBeneficiary,
			paymentChan,
		)

		disbursementId := "DISB-NONEXISTENT"

		mockDisbursement.On("Get", ctx, disbursementId).
			Return(nil, gorm.ErrRecordNotFound).Once()

		result, err := service.Retry(ctx, disbursementId)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get disbursement")

		mockDisbursement.AssertExpectations(t)
	})

	t.Run("returns error when disbursement is in progress", func(t *testing.T) {
		mockIdGenerator := new(utils_test.MockIdGenerator)
		mockLoan := new(db_test.MockLoanRepository)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)

		paymentChan := make(chan string, 1)
		service := NewDisbursementService(
			mockIdGenerator,
			mockLoan,
			mockDisbursement,
			mockTransaction,
			mockBeneficiary,
			paymentChan,
		)

		disbursementId := "DISB-123456789012"

		disbursement := schema.Disbursement{
			Id:         disbursementId,
			LoanId:     "LOAN-123456789012",
			Amount:     10000.0,
			Status:     models.DisbursementStatusProcessing,
			RetryCount: 1,
			CreatedAt:  time.Now().Add(-1 * time.Hour),
			UpdatedAt:  time.Now(),
		}

		mockDisbursement.On("Get", ctx, disbursementId).
			Return(&disbursement, nil).Once()

		result, err := service.Retry(ctx, disbursementId)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "disbursement is in-progress")

		mockDisbursement.AssertExpectations(t)
		mockDisbursement.AssertNotCalled(t, "Update")
	})

	t.Run("returns error when disbursement is completed", func(t *testing.T) {
		mockIdGenerator := new(utils_test.MockIdGenerator)
		mockLoan := new(db_test.MockLoanRepository)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)

		paymentChan := make(chan string, 1)
		service := NewDisbursementService(
			mockIdGenerator,
			mockLoan,
			mockDisbursement,
			mockTransaction,
			mockBeneficiary,
			paymentChan,
		)

		disbursementId := "DISB-123456789012"

		disbursement := schema.Disbursement{
			Id:         disbursementId,
			LoanId:     "LOAN-123456789012",
			Amount:     10000.0,
			Status:     models.DisbursementStatusSuccess,
			RetryCount: 0,
			CreatedAt:  time.Now().Add(-2 * time.Hour),
			UpdatedAt:  time.Now().Add(-1 * time.Hour),
		}

		mockDisbursement.On("Get", ctx, disbursementId).
			Return(&disbursement, nil).Once()

		result, err := service.Retry(ctx, disbursementId)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "disbursement is completed")

		mockDisbursement.AssertExpectations(t)
		mockDisbursement.AssertNotCalled(t, "Update")
	})

	t.Run("returns error when disbursement update fails", func(t *testing.T) {
		mockIdGenerator := new(utils_test.MockIdGenerator)
		mockLoan := new(db_test.MockLoanRepository)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)

		paymentChan := make(chan string, 1)
		service := NewDisbursementService(
			mockIdGenerator,
			mockLoan,
			mockDisbursement,
			mockTransaction,
			mockBeneficiary,
			paymentChan,
		)

		disbursementId := "DISB-123456789012"

		disbursement := schema.Disbursement{
			Id:         disbursementId,
			LoanId:     "LOAN-123456789012",
			Amount:     10000.0,
			Status:     models.DisbursementStatusFailed,
			RetryCount: 1,
			LastError:  stringPtr("Network error"),
			CreatedAt:  time.Now().Add(-2 * time.Hour),
			UpdatedAt:  time.Now().Add(-1 * time.Hour),
		}

		repoError := errors.New("database error")

		mockDisbursement.On("Get", ctx, disbursementId).
			Return(&disbursement, nil).Once()
		mockDisbursement.On("Update", ctx, disbursementId, mock.Anything).
			Return(repoError).Once()

		result, err := service.Retry(ctx, disbursementId)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to update disbursement")

		mockDisbursement.AssertExpectations(t)
	})

	t.Run("allows retry for initiated status", func(t *testing.T) {
		mockIdGenerator := new(utils_test.MockIdGenerator)
		mockLoan := new(db_test.MockLoanRepository)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)

		paymentChan := make(chan string, 1)
		service := NewDisbursementService(
			mockIdGenerator,
			mockLoan,
			mockDisbursement,
			mockTransaction,
			mockBeneficiary,
			paymentChan,
		)

		disbursementId := "DISB-123456789012"

		disbursement := schema.Disbursement{
			Id:         disbursementId,
			LoanId:     "LOAN-123456789012",
			Amount:     10000.0,
			Status:     models.DisbursementStatusInitiated,
			RetryCount: 0,
			LastError:  nil,
			CreatedAt:  time.Now().Add(-1 * time.Hour),
			UpdatedAt:  time.Now().Add(-1 * time.Hour),
		}

		mockDisbursement.On("Get", ctx, disbursementId).
			Return(&disbursement, nil).Once()
		mockDisbursement.On("Update", ctx, disbursementId, mock.MatchedBy(func(fields map[string]any) bool {
			return fields["status"] == string(models.DisbursementStatusInitiated) &&
				fields["last_error"] == nil &&
				fields["updated_at"] != nil
		})).
			Return(nil).
			Once()

		result, err := service.Retry(ctx, disbursementId)

		assert.NoError(t, err)
		assert.NotNil(t, result)

		mockDisbursement.AssertExpectations(t)
	})

	t.Run("allows retry for failed status", func(t *testing.T) {
		mockIdGenerator := new(utils_test.MockIdGenerator)
		mockLoan := new(db_test.MockLoanRepository)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)

		paymentChan := make(chan string, 1)
		service := NewDisbursementService(
			mockIdGenerator,
			mockLoan,
			mockDisbursement,
			mockTransaction,
			mockBeneficiary,
			paymentChan,
		)

		disbursementId := "DISB-123456789012"

		disbursement := schema.Disbursement{
			Id:         disbursementId,
			LoanId:     "LOAN-123456789012",
			Amount:     10000.0,
			Status:     models.DisbursementStatusFailed,
			RetryCount: 3,
			LastError:  stringPtr("Payment gateway timeout"),
			CreatedAt:  time.Now().Add(-3 * time.Hour),
			UpdatedAt:  time.Now().Add(-30 * time.Minute),
		}

		mockDisbursement.On("Get", ctx, disbursementId).
			Return(&disbursement, nil).Once()
		mockDisbursement.On("Update", ctx, disbursementId, mock.MatchedBy(func(fields map[string]any) bool {
			return fields["status"] == string(models.DisbursementStatusInitiated) &&
				fields["last_error"] == nil &&
				fields["updated_at"] != nil
		})).
			Return(nil).
			Once()

		result, err := service.Retry(ctx, disbursementId)

		assert.NoError(t, err)
		assert.NotNil(t, result)

		mockDisbursement.AssertExpectations(t)
	})
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
