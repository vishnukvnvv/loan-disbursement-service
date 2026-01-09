package models

type PaymentChannel string

const (
	PaymentChannelUPI  PaymentChannel = "UPI"
	PaymentChannelNEFT PaymentChannel = "NEFT"
	PaymentChannelIMPS PaymentChannel = "IMPS"
)

type CreatePaymentChannelRequest struct {
	Channel     PaymentChannel `json:"channel"      binding:"required"`
	Limit       float64        `json:"limit"        binding:"required"`
	SuccessRate float64        `json:"success_rate" binding:"required"`
	Fee         float64        `json:"fee"          binding:"required"`
}

type UpdatePaymentChannelRequest struct {
	Limit       *float64 `json:"limit"        binding:"required"`
	SuccessRate *float64 `json:"success_rate" binding:"required"`
	Fee         *float64 `json:"fee"          binding:"required"`
}

type PaymentChannelResponse struct {
	Id          string         `json:"id"`
	Channel     PaymentChannel `json:"channel"`
	Limit       float64        `json:"limit"`
	SuccessRate float64        `json:"success_rate"`
	Fee         float64        `json:"fee"`
}
