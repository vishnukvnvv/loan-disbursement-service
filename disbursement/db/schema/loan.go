package schema

import "time"

type Loan struct {
	Id            string `gorm:"primaryKey"`
	Amount        float64
	Disbursement  float64
	BeneficiaryId *string      `gorm:"index"`
	Beneficiary   *Beneficiary `gorm:"foreignKey:BeneficiaryId;references:Id"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
