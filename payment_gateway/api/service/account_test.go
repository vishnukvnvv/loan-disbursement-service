package service

import (
	"context"
	"errors"
	"payment-gateway/db/schema"
	"payment-gateway/models"
	db_test "payment-gateway/test/db"
	utils_test "payment-gateway/test/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccountService_CreateAccount(t *testing.T) {
	ctx := context.Background()

	t.Run("successful account creation", func(t *testing.T) {
		mockRepo := new(db_test.MockAccountRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewAccountService(mockRepo, mockIdGenerator)

		accountId := "ACC-123456789012"
		req := models.CreateAccountRequest{
			Name:      "Test Account",
			Balance:   1000.0,
			Threshold: 100.0,
		}

		expectedAccount := &schema.Account{
			Id:        accountId,
			Name:      req.Name,
			Balance:   req.Balance,
			Threshold: req.Threshold,
		}

		mockRepo.On("List", ctx).Return([]schema.Account{}, nil).Once()
		mockIdGenerator.On("GenerateAccountId").Return(accountId).Once()
		mockRepo.On("Create", ctx, accountId, req.Name, req.Balance, req.Threshold).
			Return(expectedAccount, nil).Once()

		result, err := service.CreateAccount(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, accountId, result.Id)
		assert.Equal(t, req.Name, result.Name)
		assert.Equal(t, req.Balance, result.Balance)
		assert.Equal(t, req.Threshold, result.Threshold)

		mockRepo.AssertExpectations(t)
		mockIdGenerator.AssertExpectations(t)
	})

	t.Run("returns error when account already exists", func(t *testing.T) {
		mockRepo := new(db_test.MockAccountRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewAccountService(mockRepo, mockIdGenerator)

		req := models.CreateAccountRequest{
			Name:      "Test Account",
			Balance:   1000.0,
			Threshold: 100.0,
		}

		existingAccount := schema.Account{
			Id:        "ACC-EXISTING",
			Name:      "Existing Account",
			Balance:   500.0,
			Threshold: 50.0,
		}

		mockRepo.On("List", ctx).Return([]schema.Account{existingAccount}, nil).Once()

		result, err := service.CreateAccount(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "account already exists")

		mockRepo.AssertExpectations(t)
		mockIdGenerator.AssertNotCalled(t, "GenerateAccountId")
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("returns error when repository list fails", func(t *testing.T) {
		mockRepo := new(db_test.MockAccountRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewAccountService(mockRepo, mockIdGenerator)

		req := models.CreateAccountRequest{
			Name:      "Test Account",
			Balance:   1000.0,
			Threshold: 100.0,
		}

		repoError := errors.New("database error")
		mockRepo.On("List", ctx).Return(nil, repoError).Once()

		result, err := service.CreateAccount(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, repoError, err)

		mockRepo.AssertExpectations(t)
		mockIdGenerator.AssertNotCalled(t, "GenerateAccountId")
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("returns error when repository create fails", func(t *testing.T) {
		mockRepo := new(db_test.MockAccountRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewAccountService(mockRepo, mockIdGenerator)

		accountId := "ACC-123456789012"
		req := models.CreateAccountRequest{
			Name:      "Test Account",
			Balance:   1000.0,
			Threshold: 100.0,
		}

		repoError := errors.New("database error")

		mockRepo.On("List", ctx).Return([]schema.Account{}, nil).Once()
		mockIdGenerator.On("GenerateAccountId").Return(accountId).Once()
		mockRepo.On("Create", ctx, accountId, req.Name, req.Balance, req.Threshold).
			Return(nil, repoError).Once()

		result, err := service.CreateAccount(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, repoError, err)

		mockRepo.AssertExpectations(t)
		mockIdGenerator.AssertExpectations(t)
	})
}

func TestAccountService_ListAccounts(t *testing.T) {
	ctx := context.Background()

	t.Run("successful list accounts", func(t *testing.T) {
		mockRepo := new(db_test.MockAccountRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewAccountService(mockRepo, mockIdGenerator)

		accounts := []schema.Account{
			{
				Id:        "ACC-001",
				Name:      "Account 1",
				Balance:   1000.0,
				Threshold: 100.0,
			},
			{
				Id:        "ACC-002",
				Name:      "Account 2",
				Balance:   2000.0,
				Threshold: 200.0,
			},
		}

		mockRepo.On("List", ctx).Return(accounts, nil).Once()

		result, err := service.ListAccounts(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 2)
		assert.Equal(t, "ACC-001", result[0].Id)
		assert.Equal(t, "Account 1", result[0].Name)
		assert.Equal(t, 1000.0, result[0].Balance)
		assert.Equal(t, 100.0, result[0].Threshold)
		assert.Equal(t, "ACC-002", result[1].Id)
		assert.Equal(t, "Account 2", result[1].Name)
		assert.Equal(t, 2000.0, result[1].Balance)
		assert.Equal(t, 200.0, result[1].Threshold)

		mockRepo.AssertExpectations(t)
	})

	t.Run("returns empty list when no accounts exist", func(t *testing.T) {
		mockRepo := new(db_test.MockAccountRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewAccountService(mockRepo, mockIdGenerator)

		mockRepo.On("List", ctx).Return([]schema.Account{}, nil).Once()

		result, err := service.ListAccounts(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 0)

		mockRepo.AssertExpectations(t)
	})

	t.Run("returns error when repository list fails", func(t *testing.T) {
		mockRepo := new(db_test.MockAccountRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewAccountService(mockRepo, mockIdGenerator)

		repoError := errors.New("database error")
		mockRepo.On("List", ctx).Return(nil, repoError).Once()

		result, err := service.ListAccounts(ctx)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, repoError, err)

		mockRepo.AssertExpectations(t)
	})
}

func TestAccountService_UpdateAccount(t *testing.T) {
	ctx := context.Background()

	t.Run("successful account update with threshold", func(t *testing.T) {
		mockRepo := new(db_test.MockAccountRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewAccountService(mockRepo, mockIdGenerator)

		accountId := "ACC-001"
		existingAccount := &schema.Account{
			Id:        accountId,
			Name:      "Test Account",
			Balance:   1000.0,
			Threshold: 100.0,
		}

		newThreshold := 200.0
		req := models.UpdateAccountRequest{
			Balance:   500.0,
			Threshold: &newThreshold,
		}

		updatedAccount := &schema.Account{
			Id:        accountId,
			Name:      "Test Account",
			Balance:   1500.0, // 1000.0 + 500.0
			Threshold: 200.0,
		}

		mockRepo.On("Get", ctx, accountId).Return(existingAccount, nil).Once()
		mockRepo.On("Update", ctx, accountId, map[string]any{
			"balance":   1500.0,
			"threshold": 200.0,
		}).Return(updatedAccount, nil).Once()

		result, err := service.UpdateAccount(ctx, accountId, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, accountId, result.Id)
		assert.Equal(t, 1500.0, result.Balance)
		assert.Equal(t, 200.0, result.Threshold)

		mockRepo.AssertExpectations(t)
	})

	t.Run("successful account update without threshold", func(t *testing.T) {
		mockRepo := new(db_test.MockAccountRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewAccountService(mockRepo, mockIdGenerator)

		accountId := "ACC-001"
		existingAccount := &schema.Account{
			Id:        accountId,
			Name:      "Test Account",
			Balance:   1000.0,
			Threshold: 100.0,
		}

		req := models.UpdateAccountRequest{
			Balance:   500.0,
			Threshold: nil,
		}

		updatedAccount := &schema.Account{
			Id:        accountId,
			Name:      "Test Account",
			Balance:   1500.0, // 1000.0 + 500.0
			Threshold: 100.0,  // unchanged
		}

		mockRepo.On("Get", ctx, accountId).Return(existingAccount, nil).Once()
		mockRepo.On("Update", ctx, accountId, map[string]any{
			"balance":   1500.0,
			"threshold": 100.0,
		}).Return(updatedAccount, nil).Once()

		result, err := service.UpdateAccount(ctx, accountId, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, accountId, result.Id)
		assert.Equal(t, 1500.0, result.Balance)
		assert.Equal(t, 100.0, result.Threshold) // unchanged

		mockRepo.AssertExpectations(t)
	})

	t.Run("successful account update with negative balance (withdrawal)", func(t *testing.T) {
		mockRepo := new(db_test.MockAccountRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewAccountService(mockRepo, mockIdGenerator)

		accountId := "ACC-001"
		existingAccount := &schema.Account{
			Id:        accountId,
			Name:      "Test Account",
			Balance:   1000.0,
			Threshold: 100.0,
		}

		req := models.UpdateAccountRequest{
			Balance:   -300.0, // withdrawal
			Threshold: nil,
		}

		updatedAccount := &schema.Account{
			Id:        accountId,
			Name:      "Test Account",
			Balance:   700.0, // 1000.0 - 300.0
			Threshold: 100.0,
		}

		mockRepo.On("Get", ctx, accountId).Return(existingAccount, nil).Once()
		mockRepo.On("Update", ctx, accountId, map[string]any{
			"balance":   700.0,
			"threshold": 100.0,
		}).Return(updatedAccount, nil).Once()

		result, err := service.UpdateAccount(ctx, accountId, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 700.0, result.Balance)

		mockRepo.AssertExpectations(t)
	})

	t.Run("returns error when account not found", func(t *testing.T) {
		mockRepo := new(db_test.MockAccountRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewAccountService(mockRepo, mockIdGenerator)

		accountId := "ACC-NONEXISTENT"
		req := models.UpdateAccountRequest{
			Balance:   500.0,
			Threshold: nil,
		}

		mockRepo.On("Get", ctx, accountId).Return(nil, nil).Once()

		result, err := service.UpdateAccount(ctx, accountId, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "account not found")

		mockRepo.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "Update")
	})

	t.Run("returns error when repository get fails", func(t *testing.T) {
		mockRepo := new(db_test.MockAccountRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewAccountService(mockRepo, mockIdGenerator)

		accountId := "ACC-001"
		req := models.UpdateAccountRequest{
			Balance:   500.0,
			Threshold: nil,
		}

		repoError := errors.New("database error")
		mockRepo.On("Get", ctx, accountId).Return(nil, repoError).Once()

		result, err := service.UpdateAccount(ctx, accountId, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, repoError, err)

		mockRepo.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "Update")
	})

	t.Run("returns error when repository update fails", func(t *testing.T) {
		mockRepo := new(db_test.MockAccountRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewAccountService(mockRepo, mockIdGenerator)

		accountId := "ACC-001"
		existingAccount := &schema.Account{
			Id:        accountId,
			Name:      "Test Account",
			Balance:   1000.0,
			Threshold: 100.0,
		}

		req := models.UpdateAccountRequest{
			Balance:   500.0,
			Threshold: nil,
		}

		repoError := errors.New("database error")

		mockRepo.On("Get", ctx, accountId).Return(existingAccount, nil).Once()
		mockRepo.On("Update", ctx, accountId, map[string]any{
			"balance":   1500.0,
			"threshold": 100.0,
		}).Return(nil, repoError).Once()

		result, err := service.UpdateAccount(ctx, accountId, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, repoError, err)

		mockRepo.AssertExpectations(t)
	})
}
