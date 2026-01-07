package failures

import "errors"

var (
	LIMIT_EXCEEDED        = errors.New("Limit Exceeded")
	INVALID_IFSC          = errors.New("Invalid IFSC code")
	INACTIVE_ACCOUNT      = errors.New("Inactive Beneficiary Account")
	BENEFICIARY_BANK_DOWN = errors.New("Beneficiary Bank is Down")
)
