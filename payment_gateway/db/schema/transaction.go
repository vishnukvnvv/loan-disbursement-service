package schema

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"payment-gateway/models"
	"time"
)

// JSONB is a custom type for handling JSONB columns in PostgreSQL
type JSONB map[string]any

// Value implements the driver.Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("failed to unmarshal JSONB value")
	}

	result := make(map[string]any)
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}

	*j = JSONB(result)
	return nil
}

type Transaction struct {
	ID              string `gorm:"primaryKey"`
	ReferenceID     string `gorm:"index"`
	Amount          float64
	Channel         models.PaymentChannel
	Fee             float64
	BeneficiaryName string
	AccountNumber   string
	IFSCCode        string
	BankName        string
	Metadata        JSONB `gorm:"type:jsonb"`
	Status          models.TransactionStatus
	Message         *string
	CreatedAt       time.Time
	ProcessedAt     *time.Time
	NotifiedAt      *time.Time
	UpdatedAt       time.Time
}
