package schema

import (
	"loan-disbursement-service/models"
	"time"
)

type Disbursement struct {
	Id         string `gorm:"primaryKey"`
	LoanId     string `gorm:"index"`
	Loan       Loan   `gorm:"foreignKey:LoanId;references:Id"`
	RetryCount int    `gorm:"default:0"`
	Channel    models.PaymentChannel
	Amount     float64
	Status     models.DisbursementStatus
	LastError  *string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
