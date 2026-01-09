package db

import (
	"payment-gateway/db/daos"
	"payment-gateway/db/schema"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	db             *gorm.DB
	account        daos.AccountRepository
	paymentChannel daos.PaymentChannelRepository
	transaction    daos.TransactionRepository
}

func New(dsn string) (*Database, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		&schema.Account{},
		&schema.PaymentChannel{},
		&schema.Transaction{},
	); err != nil {
		return nil, err
	}

	return &Database{
		db:             db,
		account:        daos.NewAccountRepository(db),
		paymentChannel: daos.NewPaymentChannelRepository(db),
		transaction:    daos.NewTransactionRepository(db),
	}, nil
}

func (d Database) GetDB() *gorm.DB {
	return d.db
}

func (d Database) GetAccountRepository() daos.AccountRepository {
	return d.account
}

func (d Database) GetPaymentChannelRepository() daos.PaymentChannelRepository {
	return d.paymentChannel
}

func (d Database) GetTransactionRepository() daos.TransactionRepository {
	return d.transaction
}
