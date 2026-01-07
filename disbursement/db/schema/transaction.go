package schema

import "time"

type Transaction struct {
	Id             string       `gorm:"primaryKey"`
	DisbursementId string       `gorm:"index"`
	Disbursement   Disbursement `gorm:"foreignKey:DisbursementId;references:Id"`
	ReferenceId    string       `gorm:"uniqueIndex"`
	Amount         float64
	Mode           string
	Status         string
	Message        *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
