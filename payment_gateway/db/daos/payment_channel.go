package daos

import (
	"context"
	"payment-gateway/db/schema"
	"payment-gateway/models"

	"gorm.io/gorm"
)

type PaymentChannelRepository interface {
	Create(
		ctx context.Context,
		id string, name models.PaymentChannel,
		limit, successRate, fee float64,
	) (*schema.PaymentChannel, error)
	List(ctx context.Context) ([]schema.PaymentChannel, error)
	Update(
		ctx context.Context,
		channel models.PaymentChannel,
		fields map[string]any,
	) (*schema.PaymentChannel, error)
	Get(ctx context.Context, channel models.PaymentChannel) (*schema.PaymentChannel, error)
}

type PaymentChannelDAO struct {
	db *gorm.DB
}

func NewPaymentChannelRepository(db *gorm.DB) PaymentChannelRepository {
	return &PaymentChannelDAO{db: db}
}

func (p PaymentChannelDAO) Create(
	ctx context.Context,
	id string, name models.PaymentChannel,
	limit, successRate, fee float64,
) (*schema.PaymentChannel, error) {
	paymentChannel := &schema.PaymentChannel{
		Id:          id,
		Name:        name,
		Limit:       limit,
		SuccessRate: successRate,
		Fee:         fee,
	}
	if err := p.db.WithContext(ctx).
		Model(&schema.PaymentChannel{}).
		Create(paymentChannel).Error; err != nil {
		return nil, err
	}
	return paymentChannel, nil
}

func (p PaymentChannelDAO) List(ctx context.Context) ([]schema.PaymentChannel, error) {
	var paymentChannels []schema.PaymentChannel
	if err := p.db.WithContext(ctx).
		Model(&schema.PaymentChannel{}).
		Find(&paymentChannels).Error; err != nil {
		return nil, err
	}
	return paymentChannels, nil
}

func (p PaymentChannelDAO) Update(
	ctx context.Context,
	channel models.PaymentChannel,
	fields map[string]any,
) (*schema.PaymentChannel, error) {
	if err := p.db.WithContext(ctx).
		Model(&schema.PaymentChannel{}).
		Where("name = ?", channel).Updates(fields).Error; err != nil {
		return nil, err
	}
	return p.Get(ctx, channel)
}

func (p PaymentChannelDAO) Get(
	ctx context.Context,
	channel models.PaymentChannel,
) (*schema.PaymentChannel, error) {
	var paymentChannel schema.PaymentChannel
	if err := p.db.WithContext(ctx).
		Model(&schema.PaymentChannel{}).
		Where("name = ?", channel).
		First(&paymentChannel).Error; err != nil {
		return nil, err
	}
	return &paymentChannel, nil
}
