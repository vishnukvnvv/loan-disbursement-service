package models

const (
	BeneficiaryStatusActive   string = "active"
	BeneficiaryStatusInactive string = "inactive"
)

type Beneficiary struct {
	Name    string `json:"name"`
	Account string `json:"account"`
	IFSC    string `json:"ifsc"`
	Bank    string `json:"bank"`
}
