package types

type PaymentStatus string

const (
	SUCCESS PaymentStatus = "success"
	FAILED  PaymentStatus = "failed"
)
