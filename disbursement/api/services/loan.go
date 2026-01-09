package services

import (
	"context"
	"loan-disbursement-service/db/daos"
	"loan-disbursement-service/db/schema"
	"loan-disbursement-service/models"
	"loan-disbursement-service/utils"
)

type LoanService interface {
	Create(ctx context.Context, amount float64) (*models.Loan, error)
	Update(ctx context.Context, loanId string, fields map[string]any) (*models.Loan, error)
	List(ctx context.Context) ([]models.Loan, error)
	Get(ctx context.Context, loanId string) (*models.Loan, error)
}

type LoanServiceImpl struct {
	loan        daos.LoanRepository
	idGenerator utils.IdGenerator
}

func NewLoanService(loan daos.LoanRepository, idGenerator utils.IdGenerator) LoanService {
	return &LoanServiceImpl{loan: loan, idGenerator: idGenerator}
}

func (s LoanServiceImpl) Create(ctx context.Context, amount float64) (*models.Loan, error) {
	loanId := s.idGenerator.GenerateLoanId()
	loan, err := s.loan.Create(ctx, loanId, amount)
	if err != nil {
		return nil, err
	}
	return s.toModel(loan), nil
}

func (s *LoanServiceImpl) Update(
	ctx context.Context,
	loanId string,
	fields map[string]any,
) (*models.Loan, error) {
	loan, err := s.loan.Update(ctx, loanId, fields)
	if err != nil {
		return nil, err
	}
	return s.toModel(loan), nil
}

func (s *LoanServiceImpl) List(ctx context.Context) ([]models.Loan, error) {
	loans, err := s.loan.List(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]models.Loan, 0, len(loans))
	for i := range loans {
		if mapped := s.toModel(&loans[i]); mapped != nil {
			result = append(result, *mapped)
		}
	}
	return result, nil
}

func (s *LoanServiceImpl) Get(ctx context.Context, loanId string) (*models.Loan, error) {
	loan, err := s.loan.Get(ctx, loanId)
	if err != nil {
		return nil, err
	}
	return s.toModel(loan), nil
}

func (s *LoanServiceImpl) toModel(loan *schema.Loan) *models.Loan {
	if loan == nil {
		return nil
	}
	return &models.Loan{
		Id:        loan.Id,
		Amount:    loan.Amount,
		CreatedAt: loan.CreatedAt,
		UpdatedAt: loan.UpdatedAt,
	}
}
