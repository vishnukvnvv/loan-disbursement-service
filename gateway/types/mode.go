package types

type PaymentMode string

const (
	UPI  PaymentMode = "UPI"
	IMPS PaymentMode = "IMPS"
	NEFT PaymentMode = "NEFT"
)
