package daos

import (
	"context"
	"errors"
	"fmt"
	"loan-disbursement-service/db/schema"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BeneficiaryDAO struct {
	db *gorm.DB
}

func NewBeneficiaryDAO(db *gorm.DB) *BeneficiaryDAO {
	return &BeneficiaryDAO{
		db: db,
	}
}

func (b BeneficiaryDAO) Create(
	ctx context.Context,
	id, account, ifsc, bank string,
) (*schema.Beneficiary, error) {
	beneficiary := &schema.Beneficiary{
		Id:      id,
		Account: account,
		IFSC:    ifsc,
		Bank:    bank,
		Status:  "Active",
	}

	if err := b.db.WithContext(ctx).Model(&schema.Beneficiary{}).Create(beneficiary).Error; err != nil {
		return nil, err
	}
	return beneficiary, nil
}

func (b BeneficiaryDAO) CreateOrGet(
	ctx context.Context,
	name, account, ifsc, bank string,
) (*schema.Beneficiary, error) {
	beneficiary, err := b.Get(ctx, account, ifsc, bank)
	if err == nil && beneficiary != nil {
		return beneficiary, nil
	}
	// If Get failed with ErrRecordNotFound, create a new beneficiary
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	id := fmt.Sprintf("BEN%s", uuid.New().String()[:12])
	beneficiary, err = b.Create(ctx, id, account, ifsc, bank)
	if err != nil {
		return nil, err
	}
	return beneficiary, nil
}

func (b BeneficiaryDAO) Get(
	ctx context.Context,
	account, ifsc, bank string,
) (*schema.Beneficiary, error) {
	var beneficiary schema.Beneficiary
	if err := b.db.WithContext(ctx).Model(&schema.Beneficiary{}).
		Where("account = ? AND ifsc = ? AND bank = ?", account, ifsc, bank).
		First(&beneficiary).Error; err != nil {
		return nil, err
	}
	return &beneficiary, nil
}

func (b BeneficiaryDAO) GetById(
	ctx context.Context,
	id string,
) (*schema.Beneficiary, error) {
	var beneficiary schema.Beneficiary
	if err := b.db.WithContext(ctx).Model(&schema.Beneficiary{}).
		Where("id = ?", id).
		First(&beneficiary).Error; err != nil {
		return nil, err
	}
	return &beneficiary, nil
}

func (b BeneficiaryDAO) Update(ctx context.Context, id string, fields map[string]any) error {
	return b.db.WithContext(ctx).Model(&schema.Beneficiary{}).
		Where("id = ?", id).
		Updates(fields).Error
}
