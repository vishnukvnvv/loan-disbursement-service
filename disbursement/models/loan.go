package models

import "time"

type Loan struct {
	Id        string    `json:"id"`
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LoanRequest struct {
	Amount float64 `json:"amount"`
}
