package daos

import (
	"context"
	"loan-disbursement-service/db/schema"

	"gorm.io/gorm"
)

type LoanRepository interface {
	Create(ctx context.Context, loanId string, amount float64) (*schema.Loan, error)
	Update(ctx context.Context, loanId string, data map[string]any) (*schema.Loan, error)
	List(ctx context.Context) ([]schema.Loan, error)
	Get(ctx context.Context, loanId string) (*schema.Loan, error)
}

func NewLoanRepository(db *gorm.DB) LoanRepository {
	return &LoanDAO{
		db: db,
	}
}

type LoanDAO struct {
	db *gorm.DB
}

func (l LoanDAO) Create(ctx context.Context, loanId string, amount float64) (*schema.Loan, error) {
	loan := &schema.Loan{
		Id:     loanId,
		Amount: amount,
	}

	if err := l.db.WithContext(ctx).Model(&schema.Loan{}).Create(loan).Error; err != nil {
		return nil, err
	}
	return loan, nil
}

func (l LoanDAO) Update(
	ctx context.Context,
	loanId string,
	data map[string]any,
) (*schema.Loan, error) {
	if err := l.db.WithContext(ctx).Model(&schema.Loan{}).
		Where("id = ?", loanId).
		Updates(data).Error; err != nil {
		return nil, err
	}

	var loan schema.Loan
	if err := l.db.WithContext(ctx).Model(&schema.Loan{}).
		Where("id = ?", loanId).
		First(&loan).Error; err != nil {
		return nil, err
	}
	return &loan, nil
}

func (l LoanDAO) List(ctx context.Context) ([]schema.Loan, error) {
	var loans []schema.Loan
	if err := l.db.WithContext(ctx).Model(&schema.Loan{}).Find(&loans).Error; err != nil {
		return nil, err
	}
	return loans, nil
}

func (l LoanDAO) Get(ctx context.Context, loanId string) (*schema.Loan, error) {
	var loan schema.Loan
	if err := l.db.WithContext(ctx).Model(&schema.Loan{}).
		Where("id = ?", loanId).
		First(&loan).Error; err != nil {
		return nil, err
	}
	return &loan, nil
}
