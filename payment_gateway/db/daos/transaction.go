package daos

import (
	"context"
	"payment-gateway/db/schema"
	"payment-gateway/models"

	"gorm.io/gorm"
)

type TransactionRepository interface {
	Create(ctx context.Context, transaction schema.Transaction) (schema.Transaction, error)
	Update(ctx context.Context, id string, fields map[string]any) (schema.Transaction, error)
	Get(ctx context.Context, id string) (schema.Transaction, error)
	GetByReferenceID(ctx context.Context, referenceID string) (*schema.Transaction, error)
	List(
		ctx context.Context,
		offset, limit int,
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
) (schema.Transaction, error) {
	if err := t.db.WithContext(ctx).Model(&schema.Transaction{}).Create(&transaction).Error; err != nil {
		return schema.Transaction{}, err
	}
	return transaction, nil
}

func (t TransactionDAO) Update(
	ctx context.Context,
	id string,
	fields map[string]any,
) (schema.Transaction, error) {
	if err := t.db.WithContext(ctx).Model(&schema.Transaction{}).
		Where("id = ?", id).
		Updates(fields).Error; err != nil {
		return schema.Transaction{}, err
	}
	return t.Get(ctx, id)
}

func (t TransactionDAO) Get(ctx context.Context, id string) (schema.Transaction, error) {
	var transaction schema.Transaction
	if err := t.db.WithContext(ctx).Model(&schema.Transaction{}).
		Where("id = ?", id).
		First(&transaction).Error; err != nil {
		return schema.Transaction{}, err
	}
	return transaction, nil
}

func (t TransactionDAO) GetByReferenceID(
	ctx context.Context,
	referenceID string,
) (*schema.Transaction, error) {
	var transaction schema.Transaction
	if err := t.db.WithContext(ctx).Model(&schema.Transaction{}).
		Where("reference_id = ?", referenceID).
		First(&transaction).Error; err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (t TransactionDAO) List(
	ctx context.Context,
	offset, limit int,
	status []models.TransactionStatus,
) ([]schema.Transaction, error) {
	var transactions []schema.Transaction
	query := t.db.WithContext(ctx).Model(&schema.Transaction{})

	if len(status) > 0 {
		query = query.Where("status IN ?", status)
	}

	if err := query.
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}
