package services

import (
	"context"
	"fmt"
	"loan-disbursement-service/db/daos"
	"loan-disbursement-service/db/schema"
	"loan-disbursement-service/models"

	"github.com/google/uuid"
)

type LoanService struct {
	loanDAO *daos.LoanDAO
}

func NewLoanService(loanDAO *daos.LoanDAO) *LoanService {
	return &LoanService{loanDAO: loanDAO}
}

func (s *LoanService) Create(ctx context.Context, amount float64) (*models.Loan, error) {
	loanId := fmt.Sprintf("LOAN%s", uuid.New().String()[:12])
	loan, err := s.loanDAO.Create(ctx, loanId, amount)
	if err != nil {
		return nil, err
	}
	return s.toModel(loan), nil
}

func (s *LoanService) Update(
	ctx context.Context,
	loanId string,
	fields map[string]any,
) (*models.Loan, error) {
	loan, err := s.loanDAO.Update(ctx, loanId, fields)
	if err != nil {
		return nil, err
	}
	return s.toModel(loan), nil
}

func (s *LoanService) List(ctx context.Context) ([]models.Loan, error) {
	loans, err := s.loanDAO.List(ctx)
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

func (s *LoanService) Get(ctx context.Context, loanId string) (*models.Loan, error) {
	loan, err := s.loanDAO.Get(ctx, loanId)
	if err != nil {
		return nil, err
	}
	return s.toModel(loan), nil
}

func (s *LoanService) toModel(loan *schema.Loan) *models.Loan {
	if loan == nil {
		return nil
	}
	return &models.Loan{
		Id:        loan.Id,
		Amount:    loan.Amount,
		Disbursed: loan.Disbursement,
		CreatedAt: loan.CreatedAt,
		UpdatedAt: loan.UpdatedAt,
	}
}
