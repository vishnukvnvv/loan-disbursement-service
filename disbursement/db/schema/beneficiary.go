package schema

import "time"

type Beneficiary struct {
	Id        string `gorm:"primaryKey"`
	Name      string `gorm:"uniqueIndex:idx_beneficiary_unique"`
	Account   string `gorm:"uniqueIndex:idx_beneficiary_unique"`
	IFSC      string `gorm:"uniqueIndex:idx_beneficiary_unique"`
	Bank      string `gorm:"uniqueIndex:idx_beneficiary_unique"`
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}
