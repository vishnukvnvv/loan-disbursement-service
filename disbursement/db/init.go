package db

import (
	"loan-disbursement-service/db/daos"
	"loan-disbursement-service/db/schema"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	db           *gorm.DB
	loan         *daos.LoanDAO
	beneficiary  *daos.BeneficiaryDAO
	disbursement *daos.DisbursementDAO
	transaction  *daos.TransactionDAO
}

func New(dsn string) (*Database, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
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
		db: db,
	}, nil
}

func (d *Database) GetLoanDAO() *daos.LoanDAO {
	if d.loan == nil {
		d.loan = daos.NewLoanDAO(d.db)
	}
	return d.loan
}

func (d *Database) GetBeneficiaryDAO() *daos.BeneficiaryDAO {
	if d.beneficiary == nil {
		d.beneficiary = daos.NewBeneficiaryDAO(d.db)
	}
	return d.beneficiary
}

func (d *Database) GetDisbursementDAO() *daos.DisbursementDAO {
	if d.disbursement == nil {
		d.disbursement = daos.NewDisbursementDAO(d.db)
	}
	return d.disbursement
}

func (d *Database) GetTransactionDAO() *daos.TransactionDAO {
	if d.transaction == nil {
		d.transaction = daos.NewTransactionDAO(d.db)
	}
	return d.transaction
}
