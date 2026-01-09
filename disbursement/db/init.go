package db

import (
	"loan-disbursement-service/db/daos"
	"loan-disbursement-service/db/schema"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	db           *gorm.DB
	loan         daos.LoanRepository
	beneficiary  daos.BeneficiaryRepository
	disbursement daos.DisbursementRepository
	transaction  daos.TransactionRepository
}

func New(dsn string) (*Database, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		&schema.Beneficiary{},
		&schema.Loan{},
		&schema.Disbursement{},
		&schema.Transaction{},
	); err != nil {
		return nil, err
	}

	return &Database{
		db:           db,
		loan:         daos.NewLoanRepository(db),
		beneficiary:  daos.NewBeneficiaryRepository(db),
		disbursement: daos.NewDisbursementRepository(db),
		transaction:  daos.NewTransactionRepository(db),
	}, nil
}
func (d *Database) GetDB() *gorm.DB {
	return d.db
}
func (d *Database) GetLoanRepository() daos.LoanRepository {
	return d.loan
}

func (d *Database) GetBeneficiaryRepository() daos.BeneficiaryRepository {
	return d.beneficiary
}

func (d *Database) GetDisbursementRepository() daos.DisbursementRepository {
	return d.disbursement
}

func (d *Database) GetTransactionRepository() daos.TransactionRepository {
	return d.transaction
}
