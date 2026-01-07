package config

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v2"

	"mock-payment-gateway/payment"
)

type Server struct {
	Port string `yaml:"port" validate:"required,gt=0"`
}

type Configuration struct {
	Server               Server                 `yaml:"server"`
	InvalidIFSC          []string               `yaml:"invalid_ifsc"`
	InvalidAccountNumber []string               `yaml:"invalid_account_number"`
	BeneficiaryBankDown  []string               `yaml:"beneficiary_bank_down"`
	PaymentModes         payment.PaymentMethods `yaml:"payment_modes"`
}

func LoadConfig() (*Configuration, error) {
	data, err := os.ReadFile("config/config.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &Configuration{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}
