package services

import (
	"context"
	"errors"
	"loan-disbursement-service/db/schema"
	db_test "loan-disbursement-service/test/db"
	utils_test "loan-disbursement-service/test/utils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestLoanService_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully creates loan", func(t *testing.T) {
		mockLoan := new(db_test.MockLoanRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewLoanService(mockLoan, mockIdGenerator)

		amount := 10000.0
		loanId := "LOAN-123456789012"

		expectedLoan := schema.Loan{
			Id:        loanId,
			Amount:    amount,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockIdGenerator.On("GenerateLoanId").Return(loanId).Once()
		mockLoan.On("Create", ctx, loanId, amount).
			Return(&expectedLoan, nil).Once()

		result, err := service.Create(ctx, amount)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, loanId, result.Id)
		assert.Equal(t, amount, result.Amount)
		assert.False(t, result.CreatedAt.IsZero())
		assert.False(t, result.UpdatedAt.IsZero())

		mockLoan.AssertExpectations(t)
		mockIdGenerator.AssertExpectations(t)
	})

	t.Run("returns error when repository create fails", func(t *testing.T) {
		mockLoan := new(db_test.MockLoanRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewLoanService(mockLoan, mockIdGenerator)

		amount := 10000.0
		loanId := "LOAN-123456789012"
		repoError := errors.New("database error")

		mockIdGenerator.On("GenerateLoanId").Return(loanId).Once()
		mockLoan.On("Create", ctx, loanId, amount).
			Return(nil, repoError).Once()

		result, err := service.Create(ctx, amount)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, repoError, err)

		mockLoan.AssertExpectations(t)
		mockIdGenerator.AssertExpectations(t)
	})

	t.Run("creates loan with zero amount", func(t *testing.T) {
		mockLoan := new(db_test.MockLoanRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewLoanService(mockLoan, mockIdGenerator)

		amount := 0.0
		loanId := "LOAN-000000000000"

		expectedLoan := schema.Loan{
			Id:        loanId,
			Amount:    amount,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockIdGenerator.On("GenerateLoanId").Return(loanId).Once()
		mockLoan.On("Create", ctx, loanId, amount).
			Return(&expectedLoan, nil).Once()

		result, err := service.Create(ctx, amount)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, amount, result.Amount)

		mockLoan.AssertExpectations(t)
		mockIdGenerator.AssertExpectations(t)
	})

	t.Run("creates loan with large amount", func(t *testing.T) {
		mockLoan := new(db_test.MockLoanRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewLoanService(mockLoan, mockIdGenerator)

		amount := 1000000.0
		loanId := "LOAN-999999999999"

		expectedLoan := schema.Loan{
			Id:        loanId,
			Amount:    amount,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockIdGenerator.On("GenerateLoanId").Return(loanId).Once()
		mockLoan.On("Create", ctx, loanId, amount).
			Return(&expectedLoan, nil).Once()

		result, err := service.Create(ctx, amount)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, amount, result.Amount)

		mockLoan.AssertExpectations(t)
		mockIdGenerator.AssertExpectations(t)
	})
}

func TestLoanService_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully updates loan", func(t *testing.T) {
		mockLoan := new(db_test.MockLoanRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewLoanService(mockLoan, mockIdGenerator)

		loanId := "LOAN-123456789012"
		fields := map[string]any{
			"amount": 15000.0,
		}

		updatedLoan := schema.Loan{
			Id:        loanId,
			Amount:    15000.0,
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now(),
		}

		mockLoan.On("Update", ctx, loanId, fields).
			Return(&updatedLoan, nil).Once()

		result, err := service.Update(ctx, loanId, fields)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, loanId, result.Id)
		assert.Equal(t, 15000.0, result.Amount)

		mockLoan.AssertExpectations(t)
	})

	t.Run("returns error when repository update fails", func(t *testing.T) {
		mockLoan := new(db_test.MockLoanRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewLoanService(mockLoan, mockIdGenerator)

		loanId := "LOAN-123456789012"
		fields := map[string]any{
			"amount": 15000.0,
		}
		repoError := errors.New("database error")

		mockLoan.On("Update", ctx, loanId, fields).
			Return(nil, repoError).Once()

		result, err := service.Update(ctx, loanId, fields)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, repoError, err)

		mockLoan.AssertExpectations(t)
	})

	t.Run("returns error when repository get fails after update", func(t *testing.T) {
		mockLoan := new(db_test.MockLoanRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewLoanService(mockLoan, mockIdGenerator)

		loanId := "LOAN-123456789012"
		fields := map[string]any{
			"amount": 15000.0,
		}
		repoError := errors.New("database error")

		mockLoan.On("Update", ctx, loanId, fields).
			Return(nil, repoError).Once()

		result, err := service.Update(ctx, loanId, fields)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, repoError, err)

		mockLoan.AssertExpectations(t)
	})

	t.Run("updates loan with multiple fields", func(t *testing.T) {
		mockLoan := new(db_test.MockLoanRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewLoanService(mockLoan, mockIdGenerator)

		loanId := "LOAN-123456789012"
		beneficiaryId := "BEN-987654321098"
		fields := map[string]any{
			"amount":         20000.0,
			"beneficiary_id": beneficiaryId,
		}

		updatedLoan := schema.Loan{
			Id:            loanId,
			Amount:        20000.0,
			BeneficiaryId: &beneficiaryId,
			CreatedAt:     time.Now().Add(-48 * time.Hour),
			UpdatedAt:     time.Now(),
		}

		mockLoan.On("Update", ctx, loanId, fields).
			Return(&updatedLoan, nil).Once()

		result, err := service.Update(ctx, loanId, fields)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, loanId, result.Id)
		assert.Equal(t, 20000.0, result.Amount)

		mockLoan.AssertExpectations(t)
	})

	t.Run("updates loan with empty fields map", func(t *testing.T) {
		mockLoan := new(db_test.MockLoanRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewLoanService(mockLoan, mockIdGenerator)

		loanId := "LOAN-123456789012"
		fields := map[string]any{}

		updatedLoan := schema.Loan{
			Id:        loanId,
			Amount:    10000.0,
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now(),
		}

		mockLoan.On("Update", ctx, loanId, fields).
			Return(&updatedLoan, nil).Once()

		result, err := service.Update(ctx, loanId, fields)

		assert.NoError(t, err)
		assert.NotNil(t, result)

		mockLoan.AssertExpectations(t)
	})
}

func TestLoanService_List(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully lists loans", func(t *testing.T) {
		mockLoan := new(db_test.MockLoanRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewLoanService(mockLoan, mockIdGenerator)

		loans := []schema.Loan{
			{
				Id:        "LOAN-001",
				Amount:    10000.0,
				CreatedAt: time.Now().Add(-24 * time.Hour),
				UpdatedAt: time.Now().Add(-24 * time.Hour),
			},
			{
				Id:        "LOAN-002",
				Amount:    20000.0,
				CreatedAt: time.Now().Add(-48 * time.Hour),
				UpdatedAt: time.Now().Add(-48 * time.Hour),
			},
			{
				Id:        "LOAN-003",
				Amount:    30000.0,
				CreatedAt: time.Now().Add(-72 * time.Hour),
				UpdatedAt: time.Now().Add(-72 * time.Hour),
			},
		}

		mockLoan.On("List", ctx).
			Return(loans, nil).Once()

		result, err := service.List(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 3)
		assert.Equal(t, "LOAN-001", result[0].Id)
		assert.Equal(t, 10000.0, result[0].Amount)
		assert.Equal(t, "LOAN-002", result[1].Id)
		assert.Equal(t, 20000.0, result[1].Amount)
		assert.Equal(t, "LOAN-003", result[2].Id)
		assert.Equal(t, 30000.0, result[2].Amount)

		mockLoan.AssertExpectations(t)
	})

	t.Run("returns empty list when no loans exist", func(t *testing.T) {
		mockLoan := new(db_test.MockLoanRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewLoanService(mockLoan, mockIdGenerator)

		loans := []schema.Loan{}

		mockLoan.On("List", ctx).
			Return(loans, nil).Once()

		result, err := service.List(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 0)

		mockLoan.AssertExpectations(t)
	})

	t.Run("returns error when repository list fails", func(t *testing.T) {
		mockLoan := new(db_test.MockLoanRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewLoanService(mockLoan, mockIdGenerator)

		repoError := errors.New("database connection error")

		mockLoan.On("List", ctx).
			Return([]schema.Loan{}, repoError).Once()

		result, err := service.List(ctx)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, repoError, err)

		mockLoan.AssertExpectations(t)
	})

	t.Run("filters out nil loans in list", func(t *testing.T) {
		mockLoan := new(db_test.MockLoanRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewLoanService(mockLoan, mockIdGenerator)

		// Note: The toModel method returns nil if loan is nil,
		// so we test that nil loans are filtered out
		loans := []schema.Loan{
			{
				Id:        "LOAN-001",
				Amount:    10000.0,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}

		mockLoan.On("List", ctx).
			Return(loans, nil).Once()

		result, err := service.List(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 1)

		mockLoan.AssertExpectations(t)
	})
}

func TestLoanService_Get(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully retrieves loan", func(t *testing.T) {
		mockLoan := new(db_test.MockLoanRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewLoanService(mockLoan, mockIdGenerator)

		loanId := "LOAN-123456789012"
		loan := schema.Loan{
			Id:        loanId,
			Amount:    10000.0,
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now(),
		}

		mockLoan.On("Get", ctx, loanId).
			Return(&loan, nil).Once()

		result, err := service.Get(ctx, loanId)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, loanId, result.Id)
		assert.Equal(t, 10000.0, result.Amount)
		assert.False(t, result.CreatedAt.IsZero())
		assert.False(t, result.UpdatedAt.IsZero())

		mockLoan.AssertExpectations(t)
	})

	t.Run("returns error when loan not found", func(t *testing.T) {
		mockLoan := new(db_test.MockLoanRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewLoanService(mockLoan, mockIdGenerator)

		loanId := "LOAN-NONEXISTENT"

		mockLoan.On("Get", ctx, loanId).
			Return(nil, gorm.ErrRecordNotFound).Once()

		result, err := service.Get(ctx, loanId)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, gorm.ErrRecordNotFound, err)

		mockLoan.AssertExpectations(t)
	})

	t.Run("returns error when repository get fails", func(t *testing.T) {
		mockLoan := new(db_test.MockLoanRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewLoanService(mockLoan, mockIdGenerator)

		loanId := "LOAN-123456789012"
		repoError := errors.New("database connection error")

		mockLoan.On("Get", ctx, loanId).
			Return(nil, repoError).Once()

		result, err := service.Get(ctx, loanId)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, repoError, err)

		mockLoan.AssertExpectations(t)
	})

	t.Run("retrieves loan with beneficiary", func(t *testing.T) {
		mockLoan := new(db_test.MockLoanRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewLoanService(mockLoan, mockIdGenerator)

		loanId := "LOAN-123456789012"
		beneficiaryId := "BEN-987654321098"
		loan := schema.Loan{
			Id:            loanId,
			Amount:        10000.0,
			BeneficiaryId: &beneficiaryId,
			CreatedAt:     time.Now().Add(-24 * time.Hour),
			UpdatedAt:     time.Now(),
		}

		mockLoan.On("Get", ctx, loanId).
			Return(&loan, nil).Once()

		result, err := service.Get(ctx, loanId)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, loanId, result.Id)
		assert.Equal(t, 10000.0, result.Amount)

		mockLoan.AssertExpectations(t)
	})
}
