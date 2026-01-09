package models

type Beneficiary struct {
	Name    string `json:"name"`
	Account string `json:"account"`
	IFSC    string `json:"ifsc"`
	Bank    string `json:"bank"`
}
