package services

import (
	"loan-disbursement-service/db/daos"
	"loan-disbursement-service/db/schema"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t testing.TB) (*testDatabase, func()) {
	// Use in-memory SQLite for fast tests
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Silence logs during tests
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Auto-migrate all schemas
	err = gormDB.AutoMigrate(
		&schema.Beneficiary{},
		&schema.Loan{},
		&schema.Disbursement{},
		&schema.Transaction{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	// Create DAOs directly
	loanDAO := daos.NewLoanDAO(gormDB)
	beneficiaryDAO := daos.NewBeneficiaryDAO(gormDB)
	disbursementDAO := daos.NewDisbursementDAO(gormDB)
	transactionDAO := daos.NewTransactionDAO(gormDB)

	// Cleanup function
	cleanup := func() {
		sqlDB, err := gormDB.DB()
		if err == nil {
			sqlDB.Close()
		}
	}

	// Return a custom struct that provides DAOs
	return &testDatabase{
		loanDAO:         loanDAO,
		beneficiaryDAO:  beneficiaryDAO,
		disbursementDAO: disbursementDAO,
		transactionDAO:  transactionDAO,
		cleanup:         cleanup,
	}, cleanup
}

// testDatabase wraps DAOs for testing
type testDatabase struct {
	loanDAO         *daos.LoanDAO
	beneficiaryDAO  *daos.BeneficiaryDAO
	disbursementDAO *daos.DisbursementDAO
	transactionDAO  *daos.TransactionDAO
	cleanup         func()
}

