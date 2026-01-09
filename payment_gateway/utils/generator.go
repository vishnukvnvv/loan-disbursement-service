package utils

import (
	"fmt"

	"github.com/google/uuid"
)

type IdGenerator interface {
	GenerateAccountId() string
	GeneratePaymentChannelId() string
	GenerateUPITransactionId() string
	GenerateNEFTTransactionId() string
	GenerateIMPSTransactionId() string
}

type IdGeneratorImpl struct {
}

func NewIdGenerator() IdGenerator {
	return &IdGeneratorImpl{}
}

func (g *IdGeneratorImpl) GenerateAccountId() string {
	return fmt.Sprintf("ACC-%s", uuid.New().String()[:12])
}

func (g *IdGeneratorImpl) GeneratePaymentChannelId() string {
	return fmt.Sprintf("CH-%s", uuid.New().String()[:12])
}

func (g *IdGeneratorImpl) GenerateUPITransactionId() string {
	return fmt.Sprintf("UPI-TXN-%s", uuid.New().String()[:12])
}

func (g *IdGeneratorImpl) GenerateNEFTTransactionId() string {
	return fmt.Sprintf("NEFT-TXN-%s", uuid.New().String()[:12])
}

func (g *IdGeneratorImpl) GenerateIMPSTransactionId() string {
	return fmt.Sprintf("IMPS-TXN-%s", uuid.New().String()[:12])
}
