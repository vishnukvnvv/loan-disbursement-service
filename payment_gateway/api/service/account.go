package service

import (
	"context"
	"errors"
	"payment-gateway/db/daos"
	"payment-gateway/db/schema"
	"payment-gateway/models"
	"payment-gateway/utils"
)

type AccountService interface {
	CreateAccount(ctx context.Context, account models.CreateAccountRequest) (*models.Account, error)
	ListAccounts(ctx context.Context) ([]models.Account, error)
	UpdateAccount(
		ctx context.Context,
		accountId string,
		account models.UpdateAccountRequest,
	) (*models.Account, error)
}

type AccountServiceImpl struct {
	accountRepo daos.AccountRepository
	idGenerator utils.IdGenerator
}

func NewAccountService(
	accountRepo daos.AccountRepository,
	idGenerator utils.IdGenerator,
) *AccountServiceImpl {
	return &AccountServiceImpl{accountRepo: accountRepo, idGenerator: idGenerator}
}

func (s *AccountServiceImpl) CreateAccount(
	ctx context.Context,
	account models.CreateAccountRequest,
) (*models.Account, error) {
	accounts, err := s.accountRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	if len(accounts) > 0 {
		return nil, errors.New("account already exists")
	}
	newAccount, err := s.accountRepo.Create(
		ctx,
		s.idGenerator.GenerateAccountId(),
		account.Name,
		account.Balance,
		account.Threshold,
	)
	if err != nil {
		return nil, err
	}
	return s.toModel(newAccount), nil
}

func (s *AccountServiceImpl) ListAccounts(ctx context.Context) ([]models.Account, error) {
	accounts, err := s.accountRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]models.Account, 0, len(accounts))
	for i := range accounts {
		if mapped := s.toModel(&accounts[i]); mapped != nil {
			result = append(result, *mapped)
		}
	}
	return result, nil
}

func (s *AccountServiceImpl) UpdateAccount(
	ctx context.Context,
	accountId string,
	account models.UpdateAccountRequest,
) (*models.Account, error) {
	existingAccount, err := s.accountRepo.Get(ctx, accountId)
	if err != nil {
		return nil, err
	}
	if existingAccount == nil {
		return nil, errors.New("account not found")
	}

	threshold := existingAccount.Threshold
	if account.Threshold != nil {
		threshold = *account.Threshold
	}

	updatedAccount, err := s.accountRepo.Update(ctx, accountId, map[string]any{
		"balance":   existingAccount.Balance + account.Balance,
		"threshold": threshold,
	})
	if err != nil {
		return nil, err
	}
	return s.toModel(updatedAccount), nil
}

func (s *AccountServiceImpl) toModel(account *schema.Account) *models.Account {
	if account == nil {
		return nil
	}
	return &models.Account{
		Id:        account.Id,
		Name:      account.Name,
		Balance:   account.Balance,
		Threshold: account.Threshold,
	}
}
