package daos

import (
	"context"
	"payment-gateway/db/schema"

	"gorm.io/gorm"
)

type AccountRepository interface {
	Create(
		ctx context.Context,
		id, name string,
		balance, threshold float64,
	) (*schema.Account, error)
	List(ctx context.Context) ([]schema.Account, error)
	Update(
		ctx context.Context,
		id string,
		fields map[string]any,
	) (*schema.Account, error)
	Get(ctx context.Context, id string) (*schema.Account, error)
}

type AccountDAO struct {
	db *gorm.DB
}

func NewAccountRepository(db *gorm.DB) AccountRepository {
	return &AccountDAO{db: db}
}

func (a AccountDAO) Create(
	ctx context.Context,
	id, name string,
	balance, threshold float64,
) (*schema.Account, error) {
	account := &schema.Account{
		Id:        id,
		Name:      name,
		Balance:   balance,
		Threshold: threshold,
	}
	if err := a.db.WithContext(ctx).
		Model(&schema.Account{}).
		Create(account).Error; err != nil {
		return nil, err
	}
	return account, nil
}

func (a AccountDAO) List(ctx context.Context) ([]schema.Account, error) {
	var accounts []schema.Account
	if err := a.db.WithContext(ctx).
		Model(&schema.Account{}).
		Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

func (a AccountDAO) Update(
	ctx context.Context,
	id string,
	fields map[string]any,
) (*schema.Account, error) {
	if err := a.db.WithContext(ctx).
		Model(&schema.Account{}).
		Where("id = ?", id).
		Updates(fields).Error; err != nil {
		return nil, err
	}
	return a.Get(ctx, id)
}

func (a AccountDAO) Get(ctx context.Context, id string) (*schema.Account, error) {
	var account schema.Account
	if err := a.db.WithContext(ctx).
		Model(&schema.Account{}).
		Where("id = ?", id).
		First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}
