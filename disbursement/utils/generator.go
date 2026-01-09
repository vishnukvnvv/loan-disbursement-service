package utils

import (
	"fmt"

	"github.com/google/uuid"
)

type IdGenerator interface {
	GenerateLoanId() string
	GenerateTransactionId() string
	GenerateReferenceId() string
	GenerateBeneficiaryId() string
	GenerateDisbursementId() string
	GenerateReconciliationId() string
}

type IdGeneratorImpl struct{}

func NewIdGenerator() IdGenerator {
	return &IdGeneratorImpl{}
}

func (g *IdGeneratorImpl) GenerateLoanId() string {
	return fmt.Sprintf("LOAN-%s", uuid.New().String()[:12])
}

func (g *IdGeneratorImpl) GenerateTransactionId() string {
	return fmt.Sprintf("TXN-%s", uuid.New().String()[:12])
}

func (g *IdGeneratorImpl) GenerateBeneficiaryId() string {
	return fmt.Sprintf("BEN-%s", uuid.New().String()[:12])
}

func (g *IdGeneratorImpl) GenerateDisbursementId() string {
	return fmt.Sprintf("DIS-%s", uuid.New().String()[:12])
}

func (g *IdGeneratorImpl) GenerateReferenceId() string {
	return fmt.Sprintf("REF-%s", uuid.New().String()[:12])
}

func (g *IdGeneratorImpl) GenerateReconciliationId() string {
	return fmt.Sprintf("RECON-%s", uuid.New().String()[:12])
}
