package daos

import (
	"context"
	"loan-disbursement-service/db/schema"
	"loan-disbursement-service/models"
	"time"

	"gorm.io/gorm"
)

type TransactionRepository interface {
	Create(
		ctx context.Context,
		transaction schema.Transaction,
	) (*schema.Transaction, error)
	Update(ctx context.Context, id string, fields map[string]any) error
	Get(ctx context.Context, id string) (*schema.Transaction, error)
	GetByReferenceID(ctx context.Context, referenceID string) (*schema.Transaction, error)
	ListByDisbursement(ctx context.Context, disbursementId string) ([]schema.Transaction, error)
	ListByDate(
		ctx context.Context,
		date time.Time,
		status []models.TransactionStatus,
	) ([]schema.Transaction, error)
}

type TransactionDAO struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &TransactionDAO{db: db}
}

func (t TransactionDAO) Create(
	ctx context.Context,
	transaction schema.Transaction,
) (*schema.Transaction, error) {
	if err := t.db.WithContext(ctx).Model(&schema.Transaction{}).Create(&transaction).Error; err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (t TransactionDAO) Update(ctx context.Context, id string, fields map[string]any) error {
	return t.db.WithContext(ctx).Model(&schema.Transaction{}).
		Where("id = ?", id).
		Updates(fields).Error
}

func (t TransactionDAO) Get(ctx context.Context, id string) (*schema.Transaction, error) {
	var tx schema.Transaction
	if err := t.db.WithContext(ctx).Model(&schema.Transaction{}).
		Where("id = ?", id).
		First(&tx).Error; err != nil {
		return nil, err
	}
	return &tx, nil
}

func (t TransactionDAO) ListByDisbursement(
	ctx context.Context,
	disbursementId string,
) ([]schema.Transaction, error) {
	var txs []schema.Transaction
	if err := t.db.WithContext(ctx).Model(&schema.Transaction{}).
		Where("disbursement_id = ?", disbursementId).
		Find(&txs).Error; err != nil {
		return nil, err
	}
	return txs, nil
}

func (t TransactionDAO) GetByReferenceID(
	ctx context.Context,
	referenceID string,
) (*schema.Transaction, error) {
	var tx schema.Transaction
	if err := t.db.WithContext(ctx).Model(&schema.Transaction{}).
		Where("reference_id = ?", referenceID).
		First(&tx).Error; err != nil {
		return nil, err
	}
	return &tx, nil
}

func (t TransactionDAO) ListByDate(
	ctx context.Context,
	date time.Time,
	status []models.TransactionStatus,
) ([]schema.Transaction, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var txs []schema.Transaction

	query := t.db.WithContext(ctx).Model(&schema.Transaction{}).
		Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay)

	if len(status) > 0 {
		query = query.Where("status IN ?", status)
	}

	if err := query.Find(&txs).Error; err != nil {
		return nil, err
	}

	return txs, nil
}
