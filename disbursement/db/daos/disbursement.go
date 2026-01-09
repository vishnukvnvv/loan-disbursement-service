package daos

import (
	"context"
	"loan-disbursement-service/db/schema"
	"loan-disbursement-service/models"

	"gorm.io/gorm"
)

type DisbursementRepository interface {
	Create(
		ctx context.Context,
		id, loanId string,
		channel models.PaymentChannel,
		status models.DisbursementStatus,
		amount float64,
	) (*schema.Disbursement, error)
	Update(ctx context.Context, id string, fields map[string]any) error
	Get(ctx context.Context, id string) (*schema.Disbursement, error)
	List(
		ctx context.Context,
		offset, limit int,
		status []models.DisbursementStatus,
		channels []models.PaymentChannel,
	) ([]schema.Disbursement, error)
	ListByLoan(
		ctx context.Context,
		loanId string,
	) ([]schema.Disbursement, error)
}
type DisbursementDAO struct {
	db *gorm.DB
}

func NewDisbursementRepository(db *gorm.DB) DisbursementRepository {
	return &DisbursementDAO{db: db}
}

func (d DisbursementDAO) Create(
	ctx context.Context,
	id, loanId string,
	channel models.PaymentChannel,
	status models.DisbursementStatus,
	amount float64,
) (*schema.Disbursement, error) {
	disbursement := &schema.Disbursement{
		Id:         id,
		LoanId:     loanId,
		Channel:    channel,
		Amount:     amount,
		Status:     status,
		RetryCount: 0,
		LastError:  nil,
	}

	if err := d.db.WithContext(ctx).Create(disbursement).Error; err != nil {
		return nil, err
	}
	return disbursement, nil
}

func (d DisbursementDAO) Update(ctx context.Context, id string, fields map[string]any) error {
	return d.db.WithContext(ctx).Model(&schema.Disbursement{}).
		Where("id = ?", id).
		Updates(fields).Error
}

func (d DisbursementDAO) Get(ctx context.Context, id string) (*schema.Disbursement, error) {
	var disbursement schema.Disbursement
	if err := d.db.WithContext(ctx).Where("id = ?", id).First(&disbursement).Error; err != nil {
		return nil, err
	}
	return &disbursement, nil
}

func (d DisbursementDAO) List(
	ctx context.Context,
	offset, limit int,
	status []models.DisbursementStatus,
	channels []models.PaymentChannel,
) ([]schema.Disbursement, error) {
	var disbursements []schema.Disbursement
	query := d.db.WithContext(ctx).Model(&schema.Disbursement{})

	if len(status) > 0 {
		query = query.Where("status IN ?", status)
	}

	if len(channels) > 0 {
		query = query.Where("channel IN ?", channels)
	}

	if err := query.
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&disbursements).Error; err != nil {
		return nil, err
	}

	return disbursements, nil
}

func (d DisbursementDAO) ListByLoan(
	ctx context.Context,
	loanId string,
) ([]schema.Disbursement, error) {
	var disbursements []schema.Disbursement
	if err := d.db.WithContext(ctx).Where("loan_id = ?", loanId).Find(&disbursements).Error; err != nil {
		return nil, err
	}
	return disbursements, nil
}
