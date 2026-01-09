package models

type Account struct {
	Id        string  `json:"id"`
	Name      string  `json:"name"`
	Balance   float64 `json:"balance"`
	Threshold float64 `json:"threshold"`
}

type CreateAccountRequest struct {
	Name      string  `json:"name"      binding:"required"`
	Balance   float64 `json:"balance"   binding:"required"`
	Threshold float64 `json:"threshold" binding:"required"`
}

type UpdateAccountRequest struct {
	Balance   float64  `json:"balance"   binding:"required"`
	Threshold *float64 `json:"threshold"`
}
