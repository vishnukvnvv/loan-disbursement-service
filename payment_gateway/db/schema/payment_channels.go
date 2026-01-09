package schema

import (
	"payment-gateway/models"
	"time"
)

type PaymentChannel struct {
	Id          string                `gorm:"primaryKey"`
	Name        models.PaymentChannel `gorm:"uniqueIndex:idx_payment_channel_unique"`
	Limit       float64               `gorm:"default:0"`
	SuccessRate float64               `gorm:"default:1.0"`
	Fee         float64               `gorm:"default:0"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
