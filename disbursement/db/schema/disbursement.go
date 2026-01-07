package schema

import "time"

type Disbursement struct {
	Id         string `gorm:"primaryKey"`
	LoanId     string `gorm:"index"`
	Amount     float64
	Loan       Loan `gorm:"foreignKey:LoanId;references:Id"`
	RetryCount int  `gorm:"default:0"`
	Status     string
	LastError  *string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
