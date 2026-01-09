package failures

import "errors"

var (
	INVALID_IFSC                   = errors.New("Invalid IFSC code")
	BENEFICIARY_BANK_DOWN          = errors.New("Beneficiary Bank is Down")
	INACTIVE_ACCOUNT               = errors.New("Inactive Beneficiary Account")
	SERVICE_UNAVAILABLE            = errors.New("Service Unavailable")
	LIMIT_EXCEEDED                 = errors.New("Limit Exceeded")
	REFERENCE_ID_ALREADY_PROCESSED = errors.New("Reference ID already processed")
	TRANSACTION_NOT_FOUND          = errors.New("Transaction not found")
	INVALID_PAYMENT_CHANNEL        = errors.New("Invalid Payment Channel")
	UNKNOWN_ERROR                  = errors.New("Unknown Error")
	INSUFFICIENT_BALANCE           = errors.New("Insufficient Balance")
)

var TRANSACTION_FAILURES = []error{
	BENEFICIARY_BANK_DOWN,
	SERVICE_UNAVAILABLE,
	LIMIT_EXCEEDED,
}

var INVALID_IFSC_CODES = []string{
	"INVALID_IFSC_CODE_1",
	"INVALID_IFSC_CODE_2",
	"INVALID_IFSC_CODE_3",
}

var INVALID_ACCOUNT_NUMBERS = []string{
	"0000000000000000",
	"1111111111111111",
	"2222222222222222",
}
