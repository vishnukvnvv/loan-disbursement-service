package schema

import "time"

type Account struct {
	Id        string  `gorm:"primaryKey"`
	Name      string  `gorm:"uniqueIndex:idx_account_unique"`
	Balance   float64 `gorm:"default:0"`
	Threshold float64 `gorm:"default:0"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
