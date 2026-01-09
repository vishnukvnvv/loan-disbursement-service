package services

import (
	"context"
	"errors"
	"loan-disbursement-service/db"
	"loan-disbursement-service/db/schema"
	"loan-disbursement-service/models"
	db_test "loan-disbursement-service/test/db"
	provider_test "loan-disbursement-service/test/providers"
	utils_test "loan-disbursement-service/test/utils"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) *db.Database {
	sqliteDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	testDB := &db.Database{}
	v := reflect.ValueOf(testDB).Elem()
	dbField := v.FieldByName("db")
	if dbField.IsValid() {
		ptr := unsafe.Pointer(testDB)
		dbPtr := (*unsafe.Pointer)(unsafe.Pointer(uintptr(ptr) + 0))
		*dbPtr = unsafe.Pointer(sqliteDB)
	} else {
		t.Fatalf("Failed to set unexported db field using reflection")
	}
	return testDB
}

func TestPaymentService_Process(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully processes initiated disbursement with UPI channel", func(t *testing.T) {
		mockDB := setupMockDB(t)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockLoan := new(db_test.MockLoanRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)
		mockRetryPolicy := new(MockRetryPolicy)
		mockGatewayProvider := new(provider_test.MockGatewayProvider)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewPaymentService(
			mockDB,
			mockDisbursement,
			mockTransaction,
			mockLoan,
			mockBeneficiary,
			mockRetryPolicy,
			mockGatewayProvider,
			mockIdGenerator,
			"https://example.com/webhook",
		)

		disbursement := &schema.Disbursement{
			Id:         "DISB-123",
			LoanId:     "LOAN-123",
			Status:     models.DisbursementStatusInitiated,
			RetryCount: 0,
			UpdatedAt:  time.Now(),
		}

		loan := &schema.Loan{
			Id:            "LOAN-123",
			Amount:        50000.0,
			BeneficiaryId: stringPtr("BEN-123"),
		}

		beneficiary := &schema.Beneficiary{
			Id:      "BEN-123",
			Name:    "John Doe",
			Account: "1234567890",
			IFSC:    "IFSC0001234",
			Bank:    "Test Bank",
		}

		transactionId := "TXN-123"
		referenceId := "REF-123"

		mockLoan.On("Get", ctx, "LOAN-123").Return(loan, nil).Once()
		mockBeneficiary.On("GetById", ctx, "BEN-123").Return(beneficiary, nil).Once()
		mockGatewayProvider.On("IsActive", ctx, models.PaymentChannelUPI).Return(true, nil).Once()
		mockDisbursement.On("Update", ctx, disbursement.Id, mock.MatchedBy(func(fields map[string]any) bool {
			return fields["status"] == models.DisbursementStatusProcessing &&
				fields["channel"] == models.PaymentChannelUPI
		})).
			Return(nil).
			Once()
		mockIdGenerator.On("GenerateTransactionId").Return(transactionId).Once()
		mockIdGenerator.On("GenerateReferenceId").Return(referenceId).Once()
		mockTransaction.On("Create", ctx, mock.MatchedBy(func(txn schema.Transaction) bool {
			return txn.Id == transactionId &&
				txn.ReferenceId == referenceId &&
				txn.Channel == models.PaymentChannelUPI &&
				txn.Amount == loan.Amount &&
				txn.Status == models.TransactionStatusInitiated
		})).Return(&schema.Transaction{Id: transactionId}, nil).Once()

		paymentResponse := models.PaymentResponse{
			Status: models.TransactionStatusSuccess,
		}

		mockGatewayProvider.On("Transfer", ctx, mock.Anything).Return(paymentResponse, nil).Once()
		mockTransaction.On("Update", ctx, transactionId, mock.MatchedBy(func(fields map[string]any) bool {
			return fields["status"] == models.TransactionStatusSuccess
		})).
			Return(nil).
			Once()

		err := service.Process(ctx, disbursement)

		assert.NoError(t, err)
		mockLoan.AssertExpectations(t)
		mockBeneficiary.AssertExpectations(t)
		mockGatewayProvider.AssertExpectations(t)
		mockDisbursement.AssertExpectations(t)
		mockTransaction.AssertExpectations(t)
		mockIdGenerator.AssertExpectations(t)
	})

	t.Run("returns nil when disbursement status is processing", func(t *testing.T) {
		mockDB := setupMockDB(t)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockLoan := new(db_test.MockLoanRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)
		mockRetryPolicy := new(MockRetryPolicy)
		mockGatewayProvider := new(provider_test.MockGatewayProvider)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewPaymentService(
			mockDB,
			mockDisbursement,
			mockTransaction,
			mockLoan,
			mockBeneficiary,
			mockRetryPolicy,
			mockGatewayProvider,
			mockIdGenerator,
			"https://example.com/webhook",
		)

		disbursement := &schema.Disbursement{
			Id:     "DISB-123",
			Status: models.DisbursementStatusProcessing,
		}

		err := service.Process(ctx, disbursement)

		assert.NoError(t, err)
		mockLoan.AssertNotCalled(t, "Get")
		mockDisbursement.AssertNotCalled(t, "Update")
	})

	t.Run("returns nil when disbursement status is success", func(t *testing.T) {
		mockDB := setupMockDB(t)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockLoan := new(db_test.MockLoanRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)
		mockRetryPolicy := new(MockRetryPolicy)
		mockGatewayProvider := new(provider_test.MockGatewayProvider)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewPaymentService(
			mockDB,
			mockDisbursement,
			mockTransaction,
			mockLoan,
			mockBeneficiary,
			mockRetryPolicy,
			mockGatewayProvider,
			mockIdGenerator,
			"https://example.com/webhook",
		)

		disbursement := &schema.Disbursement{
			Id:     "DISB-123",
			Status: models.DisbursementStatusSuccess,
		}

		err := service.Process(ctx, disbursement)

		assert.NoError(t, err)
		mockLoan.AssertNotCalled(t, "Get")
	})

	t.Run("processes suspended disbursement when retry is eligible", func(t *testing.T) {
		mockDB := setupMockDB(t)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockLoan := new(db_test.MockLoanRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)
		mockRetryPolicy := new(MockRetryPolicy)
		mockGatewayProvider := new(provider_test.MockGatewayProvider)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewPaymentService(
			mockDB,
			mockDisbursement,
			mockTransaction,
			mockLoan,
			mockBeneficiary,
			mockRetryPolicy,
			mockGatewayProvider,
			mockIdGenerator,
			"https://example.com/webhook",
		)

		disbursement := &schema.Disbursement{
			Id:         "DISB-123",
			LoanId:     "LOAN-123",
			Status:     models.DisbursementStatusSuspended,
			RetryCount: 1,
			UpdatedAt:  time.Now().Add(-2 * time.Hour),
		}

		loan := &schema.Loan{
			Id:            "LOAN-123",
			Amount:        50000.0,
			BeneficiaryId: stringPtr("BEN-123"),
		}

		beneficiary := &schema.Beneficiary{
			Id:      "BEN-123",
			Name:    "John Doe",
			Account: "1234567890",
			IFSC:    "IFSC0001234",
			Bank:    "Test Bank",
		}

		transactionId := "TXN-123"
		referenceId := "REF-123"

		mockRetryPolicy.On("IsRetryEligible", disbursement.UpdatedAt, disbursement.RetryCount).
			Return(true).
			Once()
		mockLoan.On("Get", ctx, "LOAN-123").Return(loan, nil).Once()
		mockBeneficiary.On("GetById", ctx, "BEN-123").Return(beneficiary, nil).Once()
		mockGatewayProvider.On("IsActive", ctx, models.PaymentChannelNEFT).Return(true, nil).Once()
		mockDisbursement.On("Update", ctx, disbursement.Id, mock.MatchedBy(func(fields map[string]any) bool {
			return fields["status"] == models.DisbursementStatusProcessing &&
				fields["channel"] == models.PaymentChannelNEFT
		})).
			Return(nil).
			Once()
		mockIdGenerator.On("GenerateTransactionId").Return(transactionId).Once()
		mockIdGenerator.On("GenerateReferenceId").Return(referenceId).Once()
		mockTransaction.On("Create", ctx, mock.Anything).
			Return(&schema.Transaction{Id: transactionId}, nil).
			Once()

		paymentResponse := models.PaymentResponse{
			Status: models.TransactionStatusSuccess,
		}

		mockGatewayProvider.On("Transfer", ctx, mock.Anything).Return(paymentResponse, nil).Once()
		mockTransaction.On("Update", ctx, transactionId, mock.Anything).Return(nil).Once()

		err := service.Process(ctx, disbursement)

		assert.NoError(t, err)
		mockRetryPolicy.AssertExpectations(t)
	})

	t.Run("returns nil when suspended disbursement retry is not eligible", func(t *testing.T) {
		mockDB := setupMockDB(t)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockLoan := new(db_test.MockLoanRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)
		mockRetryPolicy := new(MockRetryPolicy)
		mockGatewayProvider := new(provider_test.MockGatewayProvider)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewPaymentService(
			mockDB,
			mockDisbursement,
			mockTransaction,
			mockLoan,
			mockBeneficiary,
			mockRetryPolicy,
			mockGatewayProvider,
			mockIdGenerator,
			"https://example.com/webhook",
		)

		disbursement := &schema.Disbursement{
			Id:         "DISB-123",
			Status:     models.DisbursementStatusSuspended,
			RetryCount: 1,
			UpdatedAt:  time.Now(),
		}

		mockRetryPolicy.On("IsRetryEligible", disbursement.UpdatedAt, disbursement.RetryCount).
			Return(false).
			Once()

		err := service.Process(ctx, disbursement)

		assert.NoError(t, err)
		mockLoan.AssertNotCalled(t, "Get")
	})

	t.Run("returns error when loan not found", func(t *testing.T) {
		mockDB := setupMockDB(t)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockLoan := new(db_test.MockLoanRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)
		mockRetryPolicy := new(MockRetryPolicy)
		mockGatewayProvider := new(provider_test.MockGatewayProvider)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewPaymentService(
			mockDB,
			mockDisbursement,
			mockTransaction,
			mockLoan,
			mockBeneficiary,
			mockRetryPolicy,
			mockGatewayProvider,
			mockIdGenerator,
			"https://example.com/webhook",
		)

		disbursement := &schema.Disbursement{
			Id:         "DISB-123",
			LoanId:     "LOAN-123",
			Status:     models.DisbursementStatusInitiated,
			RetryCount: 0,
		}

		mockLoan.On("Get", ctx, "LOAN-123").Return(nil, errors.New("loan not found")).Once()

		err := service.Process(ctx, disbursement)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get loan")
	})

	t.Run("returns error when beneficiary not found", func(t *testing.T) {
		mockDB := setupMockDB(t)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockLoan := new(db_test.MockLoanRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)
		mockRetryPolicy := new(MockRetryPolicy)
		mockGatewayProvider := new(provider_test.MockGatewayProvider)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewPaymentService(
			mockDB,
			mockDisbursement,
			mockTransaction,
			mockLoan,
			mockBeneficiary,
			mockRetryPolicy,
			mockGatewayProvider,
			mockIdGenerator,
			"https://example.com/webhook",
		)

		disbursement := &schema.Disbursement{
			Id:         "DISB-123",
			LoanId:     "LOAN-123",
			Status:     models.DisbursementStatusInitiated,
			RetryCount: 0,
		}

		loan := &schema.Loan{
			Id:            "LOAN-123",
			Amount:        50000.0,
			BeneficiaryId: stringPtr("BEN-123"),
		}

		mockLoan.On("Get", ctx, "LOAN-123").Return(loan, nil).Once()
		mockBeneficiary.On("GetById", ctx, "BEN-123").
			Return(nil, errors.New("beneficiary not found")).
			Once()

		err := service.Process(ctx, disbursement)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get beneficiary")
	})

	t.Run("handles channel fallback when UPI is inactive", func(t *testing.T) {
		mockDB := setupMockDB(t)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockLoan := new(db_test.MockLoanRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)
		mockRetryPolicy := new(MockRetryPolicy)
		mockGatewayProvider := new(provider_test.MockGatewayProvider)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewPaymentService(
			mockDB,
			mockDisbursement,
			mockTransaction,
			mockLoan,
			mockBeneficiary,
			mockRetryPolicy,
			mockGatewayProvider,
			mockIdGenerator,
			"https://example.com/webhook",
		)

		disbursement := &schema.Disbursement{
			Id:         "DISB-123",
			LoanId:     "LOAN-123",
			Status:     models.DisbursementStatusInitiated,
			RetryCount: 0,
			UpdatedAt:  time.Now(),
		}

		loan := &schema.Loan{
			Id:            "LOAN-123",
			Amount:        50000.0,
			BeneficiaryId: stringPtr("BEN-123"),
		}

		beneficiary := &schema.Beneficiary{
			Id:      "BEN-123",
			Name:    "John Doe",
			Account: "1234567890",
			IFSC:    "IFSC0001234",
			Bank:    "Test Bank",
		}

		transactionId := "TXN-123"
		referenceId := "REF-123"

		mockLoan.On("Get", ctx, "LOAN-123").Return(loan, nil).Once()
		mockBeneficiary.On("GetById", ctx, "BEN-123").Return(beneficiary, nil).Once()
		mockDisbursement.On("Update", ctx, disbursement.Id, mock.MatchedBy(func(fields map[string]any) bool {
			return fields["status"] == models.DisbursementStatusProcessing &&
				fields["channel"] == models.PaymentChannelUPI
		})).
			Return(nil).
			Once()
		mockGatewayProvider.On("IsActive", ctx, models.PaymentChannelUPI).Return(false, nil).Once()
		mockIdGenerator.On("GenerateTransactionId").Return(transactionId).Once()
		mockIdGenerator.On("GenerateReferenceId").Return(referenceId).Once()
		mockTransaction.On("Create", ctx, mock.MatchedBy(func(txn schema.Transaction) bool {
			return txn.Channel == models.PaymentChannelIMPS
		})).Return(&schema.Transaction{Id: transactionId}, nil).Once()

		paymentResponse := models.PaymentResponse{
			Status: models.TransactionStatusSuccess,
		}

		mockGatewayProvider.On("Transfer", ctx, mock.Anything).Return(paymentResponse, nil).Once()
		mockTransaction.On("Update", ctx, transactionId, mock.Anything).Return(nil).Once()

		err := service.Process(ctx, disbursement)

		assert.NoError(t, err)
		mockGatewayProvider.AssertExpectations(t)
	})

	t.Run("handles transfer failure and calls HandleFailure", func(t *testing.T) {
		mockDB := setupMockDB(t)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockLoan := new(db_test.MockLoanRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)
		mockRetryPolicy := new(MockRetryPolicy)
		mockGatewayProvider := new(provider_test.MockGatewayProvider)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewPaymentService(
			mockDB,
			mockDisbursement,
			mockTransaction,
			mockLoan,
			mockBeneficiary,
			mockRetryPolicy,
			mockGatewayProvider,
			mockIdGenerator,
			"https://example.com/webhook",
		)

		disbursement := &schema.Disbursement{
			Id:         "DISB-123",
			LoanId:     "LOAN-123",
			Status:     models.DisbursementStatusInitiated,
			RetryCount: 0,
			UpdatedAt:  time.Now(),
		}

		loan := &schema.Loan{
			Id:            "LOAN-123",
			Amount:        50000.0,
			BeneficiaryId: stringPtr("BEN-123"),
		}

		beneficiary := &schema.Beneficiary{
			Id:      "BEN-123",
			Name:    "John Doe",
			Account: "1234567890",
			IFSC:    "IFSC0001234",
			Bank:    "Test Bank",
		}

		transactionId := "TXN-123"
		referenceId := "REF-123"
		transaction := &schema.Transaction{
			Id:          transactionId,
			ReferenceId: referenceId,
			Channel:     models.PaymentChannelUPI,
		}

		mockLoan.On("Get", ctx, "LOAN-123").Return(loan, nil).Once()
		mockBeneficiary.On("GetById", ctx, "BEN-123").Return(beneficiary, nil).Once()
		mockGatewayProvider.On("IsActive", ctx, models.PaymentChannelUPI).Return(true, nil).Once()
		mockDisbursement.On("Update", ctx, disbursement.Id, mock.Anything).Return(nil).Once()
		mockIdGenerator.On("GenerateTransactionId").Return(transactionId).Once()
		mockIdGenerator.On("GenerateReferenceId").Return(referenceId).Once()
		mockTransaction.On("Create", ctx, mock.Anything).Return(transaction, nil).Once()

		transferError := models.NETWORK_ERROR
		mockGatewayProvider.On("Transfer", ctx, mock.Anything).
			Return(models.PaymentResponse{}, transferError).
			Once()

		mockTransaction.On("Update", ctx, transactionId, mock.Anything).Return(nil).Once()
		mockDisbursement.On("Update", ctx, disbursement.Id, mock.Anything).Return(nil).Once()

		err := service.Process(ctx, disbursement)

		assert.NoError(t, err)
		mockTransaction.AssertExpectations(t)
		mockDisbursement.AssertExpectations(t)
	})

	t.Run("selects IMPS channel for amount between 100000 and 500000", func(t *testing.T) {
		mockDB := setupMockDB(t)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockLoan := new(db_test.MockLoanRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)
		mockRetryPolicy := new(MockRetryPolicy)
		mockGatewayProvider := new(provider_test.MockGatewayProvider)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewPaymentService(
			mockDB,
			mockDisbursement,
			mockTransaction,
			mockLoan,
			mockBeneficiary,
			mockRetryPolicy,
			mockGatewayProvider,
			mockIdGenerator,
			"https://example.com/webhook",
		)

		disbursement := &schema.Disbursement{
			Id:         "DISB-123",
			LoanId:     "LOAN-123",
			Status:     models.DisbursementStatusInitiated,
			RetryCount: 0,
			UpdatedAt:  time.Now(),
		}

		loan := &schema.Loan{
			Id:            "LOAN-123",
			Amount:        200000.0,
			BeneficiaryId: stringPtr("BEN-123"),
		}

		beneficiary := &schema.Beneficiary{
			Id:      "BEN-123",
			Name:    "John Doe",
			Account: "1234567890",
			IFSC:    "IFSC0001234",
			Bank:    "Test Bank",
		}

		transactionId := "TXN-123"
		referenceId := "REF-123"

		mockLoan.On("Get", ctx, "LOAN-123").Return(loan, nil).Once()
		mockBeneficiary.On("GetById", ctx, "BEN-123").Return(beneficiary, nil).Once()
		mockGatewayProvider.On("IsActive", ctx, models.PaymentChannelIMPS).Return(true, nil).Once()
		mockDisbursement.On("Update", ctx, disbursement.Id, mock.MatchedBy(func(fields map[string]any) bool {
			return fields["channel"] == models.PaymentChannelIMPS
		})).
			Return(nil).
			Once()
		mockIdGenerator.On("GenerateTransactionId").Return(transactionId).Once()
		mockIdGenerator.On("GenerateReferenceId").Return(referenceId).Once()
		mockTransaction.On("Create", ctx, mock.MatchedBy(func(txn schema.Transaction) bool {
			return txn.Channel == models.PaymentChannelIMPS
		})).Return(&schema.Transaction{Id: transactionId}, nil).Once()

		paymentResponse := models.PaymentResponse{
			Status: models.TransactionStatusSuccess,
		}

		mockGatewayProvider.On("Transfer", ctx, mock.Anything).Return(paymentResponse, nil).Once()
		mockTransaction.On("Update", ctx, transactionId, mock.Anything).Return(nil).Once()

		err := service.Process(ctx, disbursement)

		assert.NoError(t, err)
	})

	t.Run("selects NEFT channel for amount above 500000", func(t *testing.T) {
		mockDB := setupMockDB(t)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockLoan := new(db_test.MockLoanRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)
		mockRetryPolicy := new(MockRetryPolicy)
		mockGatewayProvider := new(provider_test.MockGatewayProvider)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewPaymentService(
			mockDB,
			mockDisbursement,
			mockTransaction,
			mockLoan,
			mockBeneficiary,
			mockRetryPolicy,
			mockGatewayProvider,
			mockIdGenerator,
			"https://example.com/webhook",
		)

		disbursement := &schema.Disbursement{
			Id:         "DISB-123",
			LoanId:     "LOAN-123",
			Status:     models.DisbursementStatusInitiated,
			RetryCount: 0,
			UpdatedAt:  time.Now(),
		}

		loan := &schema.Loan{
			Id:            "LOAN-123",
			Amount:        600000.0,
			BeneficiaryId: stringPtr("BEN-123"),
		}

		beneficiary := &schema.Beneficiary{
			Id:      "BEN-123",
			Name:    "John Doe",
			Account: "1234567890",
			IFSC:    "IFSC0001234",
			Bank:    "Test Bank",
		}

		transactionId := "TXN-123"
		referenceId := "REF-123"

		mockLoan.On("Get", ctx, "LOAN-123").Return(loan, nil).Once()
		mockBeneficiary.On("GetById", ctx, "BEN-123").Return(beneficiary, nil).Once()
		mockGatewayProvider.On("IsActive", ctx, models.PaymentChannelNEFT).Return(true, nil).Once()
		mockDisbursement.On("Update", ctx, disbursement.Id, mock.MatchedBy(func(fields map[string]any) bool {
			return fields["channel"] == models.PaymentChannelNEFT
		})).
			Return(nil).
			Once()
		mockIdGenerator.On("GenerateTransactionId").Return(transactionId).Once()
		mockIdGenerator.On("GenerateReferenceId").Return(referenceId).Once()
		mockTransaction.On("Create", ctx, mock.MatchedBy(func(txn schema.Transaction) bool {
			return txn.Channel == models.PaymentChannelNEFT
		})).Return(&schema.Transaction{Id: transactionId}, nil).Once()

		paymentResponse := models.PaymentResponse{
			Status: models.TransactionStatusSuccess,
		}

		mockGatewayProvider.On("Transfer", ctx, mock.Anything).Return(paymentResponse, nil).Once()
		mockTransaction.On("Update", ctx, transactionId, mock.Anything).Return(nil).Once()

		err := service.Process(ctx, disbursement)

		assert.NoError(t, err)
	})
}

func TestPaymentService_HandleNotification(t *testing.T) {
	ctx := context.Background()

	t.Run("handles success notification", func(t *testing.T) {
		mockDB := setupMockDB(t)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockLoan := new(db_test.MockLoanRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)
		mockRetryPolicy := new(MockRetryPolicy)
		mockGatewayProvider := new(provider_test.MockGatewayProvider)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewPaymentService(
			mockDB,
			mockDisbursement,
			mockTransaction,
			mockLoan,
			mockBeneficiary,
			mockRetryPolicy,
			mockGatewayProvider,
			mockIdGenerator,
			"https://example.com/webhook",
		)

		notification := models.PaymentNotificationRequest{
			ReferenceID: "REF-123",
			Status:      models.TransactionStatusSuccess,
			Channel:     models.PaymentChannelUPI,
		}

		transaction := &schema.Transaction{
			Id:             "TXN-123",
			DisbursementId: "DISB-123",
			ReferenceId:    "REF-123",
		}

		disbursement := &schema.Disbursement{
			Id: "DISB-123",
		}

		mockTransaction.On("GetByReferenceID", ctx, "REF-123").Return(transaction, nil).Once()
		mockDisbursement.On("Get", ctx, "DISB-123").Return(disbursement, nil).Once()
		mockTransaction.On("Update", ctx, transaction.Id, mock.MatchedBy(func(fields map[string]any) bool {
			return fields["status"] == models.TransactionStatusSuccess
		})).
			Return(nil).
			Once()
		mockDisbursement.On("Update", ctx, disbursement.Id, mock.MatchedBy(func(fields map[string]any) bool {
			return fields["status"] == models.DisbursementStatusSuccess &&
				fields["channel"] == notification.Channel
		})).
			Return(nil).
			Once()

		err := service.HandleNotification(ctx, notification)

		assert.NoError(t, err)
		mockTransaction.AssertExpectations(t)
		mockDisbursement.AssertExpectations(t)
	})

	t.Run("handles failure notification", func(t *testing.T) {
		mockDB := setupMockDB(t)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockLoan := new(db_test.MockLoanRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)
		mockRetryPolicy := new(MockRetryPolicy)
		mockGatewayProvider := new(provider_test.MockGatewayProvider)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewPaymentService(
			mockDB,
			mockDisbursement,
			mockTransaction,
			mockLoan,
			mockBeneficiary,
			mockRetryPolicy,
			mockGatewayProvider,
			mockIdGenerator,
			"https://example.com/webhook",
		)

		notification := models.PaymentNotificationRequest{
			ReferenceID: "REF-123",
			Status:      models.TransactionStatusFailed,
			Message:     "Payment failed",
			Channel:     models.PaymentChannelUPI,
		}

		transaction := &schema.Transaction{
			Id:             "TXN-123",
			DisbursementId: "DISB-123",
			ReferenceId:    "REF-123",
		}

		disbursement := &schema.Disbursement{
			Id:         "DISB-123",
			RetryCount: 0,
		}

		mockTransaction.On("GetByReferenceID", ctx, "REF-123").Return(transaction, nil).Once()
		mockDisbursement.On("Get", ctx, "DISB-123").Return(disbursement, nil).Once()
		mockTransaction.On("Update", ctx, transaction.Id, mock.Anything).Return(nil).Once()
		mockDisbursement.On("Update", ctx, disbursement.Id, mock.Anything).Return(nil).Once()

		err := service.HandleNotification(ctx, notification)

		assert.NoError(t, err)
		mockTransaction.AssertExpectations(t)
		mockDisbursement.AssertExpectations(t)
	})

	t.Run("returns error when transaction not found", func(t *testing.T) {
		mockDB := setupMockDB(t)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockLoan := new(db_test.MockLoanRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)
		mockRetryPolicy := new(MockRetryPolicy)
		mockGatewayProvider := new(provider_test.MockGatewayProvider)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewPaymentService(
			mockDB,
			mockDisbursement,
			mockTransaction,
			mockLoan,
			mockBeneficiary,
			mockRetryPolicy,
			mockGatewayProvider,
			mockIdGenerator,
			"https://example.com/webhook",
		)

		notification := models.PaymentNotificationRequest{
			ReferenceID: "REF-123",
			Status:      models.TransactionStatusSuccess,
		}

		mockTransaction.On("GetByReferenceID", ctx, "REF-123").
			Return(nil, errors.New("not found")).
			Once()

		err := service.HandleNotification(ctx, notification)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get transaction")
	})

	t.Run("returns error when disbursement not found", func(t *testing.T) {
		mockDB := setupMockDB(t)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockLoan := new(db_test.MockLoanRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)
		mockRetryPolicy := new(MockRetryPolicy)
		mockGatewayProvider := new(provider_test.MockGatewayProvider)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewPaymentService(
			mockDB,
			mockDisbursement,
			mockTransaction,
			mockLoan,
			mockBeneficiary,
			mockRetryPolicy,
			mockGatewayProvider,
			mockIdGenerator,
			"https://example.com/webhook",
		)

		notification := models.PaymentNotificationRequest{
			ReferenceID: "REF-123",
			Status:      models.TransactionStatusSuccess,
		}

		transaction := &schema.Transaction{
			Id:             "TXN-123",
			DisbursementId: "DISB-123",
			ReferenceId:    "REF-123",
		}

		mockTransaction.On("GetByReferenceID", ctx, "REF-123").Return(transaction, nil).Once()
		mockDisbursement.On("Get", ctx, "DISB-123").Return(nil, errors.New("not found")).Once()

		err := service.HandleNotification(ctx, notification)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get disbursement")
	})
}

func TestPaymentService_HandleFailure(t *testing.T) {
	ctx := context.Background()

	t.Run(
		"handles reference ID already processed and fetches payment successfully",
		func(t *testing.T) {
			mockDB := setupMockDB(t)
			mockDisbursement := new(db_test.MockDisbursementRepository)
			mockTransaction := new(db_test.MockTransactionRepository)
			mockLoan := new(db_test.MockLoanRepository)
			mockBeneficiary := new(db_test.MockBeneficiaryRepository)
			mockRetryPolicy := new(MockRetryPolicy)
			mockGatewayProvider := new(provider_test.MockGatewayProvider)
			mockIdGenerator := new(utils_test.MockIdGenerator)

			service := NewPaymentService(
				mockDB,
				mockDisbursement,
				mockTransaction,
				mockLoan,
				mockBeneficiary,
				mockRetryPolicy,
				mockGatewayProvider,
				mockIdGenerator,
				"https://example.com/webhook",
			)

			disbursement := &schema.Disbursement{
				Id:         "DISB-123",
				RetryCount: 0,
			}

			transaction := &schema.Transaction{
				Id:          "TXN-123",
				ReferenceId: "REF-123",
				Channel:     models.PaymentChannelUPI,
			}

			paymentResponse := models.PaymentResponse{
				Status: models.TransactionStatusSuccess,
			}

			mockGatewayProvider.On("Fetch", ctx, transaction.Channel, transaction.ReferenceId).
				Return(paymentResponse, nil).
				Once()
			mockTransaction.On("Update", ctx, transaction.Id, mock.MatchedBy(func(fields map[string]any) bool {
				return fields["status"] == models.TransactionStatusSuccess
			})).
				Return(nil).
				Once()

			err := service.HandleFailure(
				ctx,
				disbursement,
				transaction,
				models.PaymentChannelUPI,
				models.REFERENCE_ID_ALREADY_PROCESSED,
			)

			assert.NoError(t, err)
			mockGatewayProvider.AssertExpectations(t)
			mockTransaction.AssertExpectations(t)
		},
	)

	t.Run("handles reference ID already processed but payment fetch fails", func(t *testing.T) {
		mockDB := setupMockDB(t)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockLoan := new(db_test.MockLoanRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)
		mockRetryPolicy := new(MockRetryPolicy)
		mockGatewayProvider := new(provider_test.MockGatewayProvider)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewPaymentService(
			mockDB,
			mockDisbursement,
			mockTransaction,
			mockLoan,
			mockBeneficiary,
			mockRetryPolicy,
			mockGatewayProvider,
			mockIdGenerator,
			"https://example.com/webhook",
		)

		disbursement := &schema.Disbursement{
			Id:         "DISB-123",
			RetryCount: 0,
		}

		transaction := &schema.Transaction{
			Id:          "TXN-123",
			ReferenceId: "REF-123",
			Channel:     models.PaymentChannelUPI,
		}

		mockGatewayProvider.On("Fetch", ctx, transaction.Channel, transaction.ReferenceId).
			Return(models.PaymentResponse{}, errors.New("fetch error")).
			Once()

		err := service.HandleFailure(
			ctx,
			disbursement,
			transaction,
			models.PaymentChannelUPI,
			models.REFERENCE_ID_ALREADY_PROCESSED,
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get payment")
	})

	t.Run(
		"handles reference ID already processed but payment status is not success",
		func(t *testing.T) {
			mockDB := setupMockDB(t)
			mockDisbursement := new(db_test.MockDisbursementRepository)
			mockTransaction := new(db_test.MockTransactionRepository)
			mockLoan := new(db_test.MockLoanRepository)
			mockBeneficiary := new(db_test.MockBeneficiaryRepository)
			mockRetryPolicy := new(MockRetryPolicy)
			mockGatewayProvider := new(provider_test.MockGatewayProvider)
			mockIdGenerator := new(utils_test.MockIdGenerator)

			service := NewPaymentService(
				mockDB,
				mockDisbursement,
				mockTransaction,
				mockLoan,
				mockBeneficiary,
				mockRetryPolicy,
				mockGatewayProvider,
				mockIdGenerator,
				"https://example.com/webhook",
			)

			disbursement := &schema.Disbursement{
				Id:         "DISB-123",
				RetryCount: 0,
			}

			transaction := &schema.Transaction{
				Id:          "TXN-123",
				ReferenceId: "REF-123",
				Channel:     models.PaymentChannelUPI,
			}

			paymentResponse := models.PaymentResponse{
				Status: models.TransactionStatusFailed,
				Error:  models.NETWORK_ERROR,
			}

			mockGatewayProvider.On("Fetch", ctx, transaction.Channel, transaction.ReferenceId).
				Return(paymentResponse, nil).
				Once()
			mockTransaction.On("Update", ctx, transaction.Id, mock.Anything).Return(nil).Once()
			mockDisbursement.On("Update", ctx, disbursement.Id, mock.MatchedBy(func(fields map[string]any) bool {
				return fields["status"] == models.DisbursementStatusSuspended &&
					fields["retry_count"] == 1
			})).
				Return(nil).
				Once()

			err := service.HandleFailure(
				ctx,
				disbursement,
				transaction,
				models.PaymentChannelUPI,
				models.REFERENCE_ID_ALREADY_PROCESSED,
			)

			assert.NoError(t, err)
			mockGatewayProvider.AssertExpectations(t)
			mockTransaction.AssertExpectations(t)
			mockDisbursement.AssertExpectations(t)
		},
	)

	t.Run("handles network error and suspends disbursement", func(t *testing.T) {
		mockDB := setupMockDB(t)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockLoan := new(db_test.MockLoanRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)
		mockRetryPolicy := new(MockRetryPolicy)
		mockGatewayProvider := new(provider_test.MockGatewayProvider)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewPaymentService(
			mockDB,
			mockDisbursement,
			mockTransaction,
			mockLoan,
			mockBeneficiary,
			mockRetryPolicy,
			mockGatewayProvider,
			mockIdGenerator,
			"https://example.com/webhook",
		)

		disbursement := &schema.Disbursement{
			Id:         "DISB-123",
			RetryCount: 0,
		}

		transaction := &schema.Transaction{
			Id:          "TXN-123",
			ReferenceId: "REF-123",
			Channel:     models.PaymentChannelUPI,
		}

		mockTransaction.On("Update", ctx, transaction.Id, mock.Anything).Return(nil).Once()
		mockDisbursement.On("Update", ctx, disbursement.Id, mock.MatchedBy(func(fields map[string]any) bool {
			return fields["status"] == models.DisbursementStatusSuspended &&
				fields["retry_count"] == 1
		})).
			Return(nil).
			Once()

		err := service.HandleFailure(
			ctx,
			disbursement,
			transaction,
			models.PaymentChannelUPI,
			models.NETWORK_ERROR,
		)

		assert.NoError(t, err)
		mockTransaction.AssertExpectations(t)
		mockDisbursement.AssertExpectations(t)
	})

	t.Run("handles permanent failure and marks as failed", func(t *testing.T) {
		mockDB := setupMockDB(t)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockLoan := new(db_test.MockLoanRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)
		mockRetryPolicy := new(MockRetryPolicy)
		mockGatewayProvider := new(provider_test.MockGatewayProvider)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewPaymentService(
			mockDB,
			mockDisbursement,
			mockTransaction,
			mockLoan,
			mockBeneficiary,
			mockRetryPolicy,
			mockGatewayProvider,
			mockIdGenerator,
			"https://example.com/webhook",
		)

		disbursement := &schema.Disbursement{
			Id:         "DISB-123",
			RetryCount: 0,
		}

		transaction := &schema.Transaction{
			Id:          "TXN-123",
			ReferenceId: "REF-123",
			Channel:     models.PaymentChannelUPI,
		}

		permanentError := errors.New("invalid IFSC code")

		mockTransaction.On("Update", ctx, transaction.Id, mock.Anything).Return(nil).Once()
		mockDisbursement.On("Update", ctx, disbursement.Id, mock.MatchedBy(func(fields map[string]any) bool {
			return fields["status"] == models.DisbursementStatusFailed &&
				fields["retry_count"] == 1
		})).
			Return(nil).
			Once()

		err := service.HandleFailure(
			ctx,
			disbursement,
			transaction,
			models.PaymentChannelUPI,
			permanentError,
		)

		assert.NoError(t, err)
		mockTransaction.AssertExpectations(t)
		mockDisbursement.AssertExpectations(t)
	})

	t.Run("marks as failed when retry count exceeds max retries", func(t *testing.T) {
		mockDB := setupMockDB(t)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockLoan := new(db_test.MockLoanRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)
		mockRetryPolicy := new(MockRetryPolicy)
		mockGatewayProvider := new(provider_test.MockGatewayProvider)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewPaymentService(
			mockDB,
			mockDisbursement,
			mockTransaction,
			mockLoan,
			mockBeneficiary,
			mockRetryPolicy,
			mockGatewayProvider,
			mockIdGenerator,
			"https://example.com/webhook",
		)

		disbursement := &schema.Disbursement{
			Id:         "DISB-123",
			RetryCount: MaxRetries,
		}

		transaction := &schema.Transaction{
			Id:          "TXN-123",
			ReferenceId: "REF-123",
			Channel:     models.PaymentChannelUPI,
		}

		mockTransaction.On("Update", ctx, transaction.Id, mock.Anything).Return(nil).Once()
		mockDisbursement.On("Update", ctx, disbursement.Id, mock.MatchedBy(func(fields map[string]any) bool {
			return fields["status"] == models.DisbursementStatusFailed &&
				fields["retry_count"] == MaxRetries
		})).
			Return(nil).
			Once()

		err := service.HandleFailure(
			ctx,
			disbursement,
			transaction,
			models.PaymentChannelUPI,
			models.NETWORK_ERROR,
		)

		assert.NoError(t, err)
		mockTransaction.AssertExpectations(t)
		mockDisbursement.AssertExpectations(t)
	})
}

func TestPaymentService_HanleSuccess(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully handles success", func(t *testing.T) {
		mockDB := setupMockDB(t)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockLoan := new(db_test.MockLoanRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)
		mockRetryPolicy := new(MockRetryPolicy)
		mockGatewayProvider := new(provider_test.MockGatewayProvider)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewPaymentService(
			mockDB,
			mockDisbursement,
			mockTransaction,
			mockLoan,
			mockBeneficiary,
			mockRetryPolicy,
			mockGatewayProvider,
			mockIdGenerator,
			"https://example.com/webhook",
		)

		disbursementId := "DISB-123"
		transactionId := "TXN-123"
		channel := models.PaymentChannelUPI

		mockTransaction.On("Update", ctx, transactionId, mock.MatchedBy(func(fields map[string]any) bool {
			return fields["status"] == models.TransactionStatusSuccess
		})).
			Return(nil).
			Once()
		mockDisbursement.On("Update", ctx, disbursementId, mock.MatchedBy(func(fields map[string]any) bool {
			return fields["status"] == models.DisbursementStatusSuccess &&
				fields["channel"] == channel
		})).
			Return(nil).
			Once()

		err := service.HanleSuccess(ctx, disbursementId, transactionId, channel)

		assert.NoError(t, err)
		mockTransaction.AssertExpectations(t)
		mockDisbursement.AssertExpectations(t)
	})

	t.Run("returns error when transaction update fails", func(t *testing.T) {
		mockDB := setupMockDB(t)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockLoan := new(db_test.MockLoanRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)
		mockRetryPolicy := new(MockRetryPolicy)
		mockGatewayProvider := new(provider_test.MockGatewayProvider)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewPaymentService(
			mockDB,
			mockDisbursement,
			mockTransaction,
			mockLoan,
			mockBeneficiary,
			mockRetryPolicy,
			mockGatewayProvider,
			mockIdGenerator,
			"https://example.com/webhook",
		)

		disbursementId := "DISB-123"
		transactionId := "TXN-123"
		channel := models.PaymentChannelUPI

		mockTransaction.On("Update", ctx, transactionId, mock.Anything).
			Return(errors.New("update failed")).
			Once()

		err := service.HanleSuccess(ctx, disbursementId, transactionId, channel)

		assert.Error(t, err)
		mockDisbursement.AssertNotCalled(t, "Update")
	})

	t.Run("returns error when disbursement update fails", func(t *testing.T) {
		mockDB := setupMockDB(t)
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockTransaction := new(db_test.MockTransactionRepository)
		mockLoan := new(db_test.MockLoanRepository)
		mockBeneficiary := new(db_test.MockBeneficiaryRepository)
		mockRetryPolicy := new(MockRetryPolicy)
		mockGatewayProvider := new(provider_test.MockGatewayProvider)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewPaymentService(
			mockDB,
			mockDisbursement,
			mockTransaction,
			mockLoan,
			mockBeneficiary,
			mockRetryPolicy,
			mockGatewayProvider,
			mockIdGenerator,
			"https://example.com/webhook",
		)

		disbursementId := "DISB-123"
		transactionId := "TXN-123"
		channel := models.PaymentChannelUPI

		mockTransaction.On("Update", ctx, transactionId, mock.Anything).Return(nil).Once()
		mockDisbursement.On("Update", ctx, disbursementId, mock.Anything).
			Return(errors.New("update failed")).
			Once()

		err := service.HanleSuccess(ctx, disbursementId, transactionId, channel)

		assert.Error(t, err)
		mockTransaction.AssertExpectations(t)
		mockDisbursement.AssertExpectations(t)
	})
}
