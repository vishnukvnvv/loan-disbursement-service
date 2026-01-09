package schema

import (
	"loan-disbursement-service/models"
	"time"
)

type Transaction struct {
	Id             string       `gorm:"primaryKey"`
	DisbursementId string       `gorm:"index"`
	Disbursement   Disbursement `gorm:"foreignKey:DisbursementId;references:Id"`
	ReferenceId    string       `gorm:"uniqueIndex"`
	Amount         float64
	Channel        models.PaymentChannel
	Status         models.TransactionStatus
	Message        *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
