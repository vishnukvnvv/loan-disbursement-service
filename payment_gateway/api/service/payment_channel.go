package service

import (
	"context"
	"errors"
	"payment-gateway/db/daos"
	"payment-gateway/db/schema"
	"payment-gateway/models"
	"payment-gateway/utils"

	"gorm.io/gorm"
)

type PaymentChannelService interface {
	CreatePaymentChannel(
		ctx context.Context,
		paymentChannel models.CreatePaymentChannelRequest,
	) (*models.PaymentChannelResponse, error)
	ListPaymentChannels(ctx context.Context) ([]models.PaymentChannelResponse, error)
	UpdatePaymentChannel(
		ctx context.Context,
		channel models.PaymentChannel,
		paymentChannel models.UpdatePaymentChannelRequest,
	) (*models.PaymentChannelResponse, error)
}

type PaymentChannelServiceImpl struct {
	paymentChannelRepo daos.PaymentChannelRepository
	idGenerator        utils.IdGenerator
}

func NewPaymentChannelService(
	paymentChannelRepo daos.PaymentChannelRepository,
	idGenerator utils.IdGenerator,
) PaymentChannelService {
	return &PaymentChannelServiceImpl{
		paymentChannelRepo: paymentChannelRepo,
		idGenerator:        idGenerator,
	}
}

func (s *PaymentChannelServiceImpl) CreatePaymentChannel(
	ctx context.Context,
	paymentChannel models.CreatePaymentChannelRequest,
) (*models.PaymentChannelResponse, error) {
	existingPaymentChannel, err := s.paymentChannelRepo.Get(ctx, paymentChannel.Channel)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("failed to get payment channel")
	}
	if existingPaymentChannel != nil {
		return nil, errors.New("payment channel already exists")
	}
	id := s.idGenerator.GeneratePaymentChannelId()
	newPaymentChannel, err := s.paymentChannelRepo.Create(
		ctx,
		id,
		paymentChannel.Channel,
		paymentChannel.Limit,
		paymentChannel.SuccessRate,
		paymentChannel.Fee,
	)
	if err != nil {
		return nil, err
	}
	return s.toModel(newPaymentChannel), nil
}

func (s *PaymentChannelServiceImpl) UpdatePaymentChannel(
	ctx context.Context,
	channel models.PaymentChannel,
	paymentChannel models.UpdatePaymentChannelRequest,
) (*models.PaymentChannelResponse, error) {
	existingPaymentChannel, err := s.paymentChannelRepo.Get(ctx, channel)
	if err != nil {
		return nil, err
	}
	if existingPaymentChannel == nil {
		return nil, errors.New("payment channel not found")
	}
	updatedPaymentChannel, err := s.paymentChannelRepo.Update(
		ctx,
		channel,
		map[string]any{
			"limit": s.getFloat64(existingPaymentChannel.Limit, paymentChannel.Limit),
			"success_rate": s.getFloat64(
				existingPaymentChannel.SuccessRate,
				paymentChannel.SuccessRate,
			),
			"fee": s.getFloat64(
				existingPaymentChannel.Fee,
				paymentChannel.Fee,
			),
		},
	)
	if err != nil {
		return nil, err
	}
	return s.toModel(updatedPaymentChannel), nil
}

func (s PaymentChannelServiceImpl) ListPaymentChannels(
	ctx context.Context,
) ([]models.PaymentChannelResponse, error) {
	paymentChannels, err := s.paymentChannelRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]models.PaymentChannelResponse, 0, len(paymentChannels))
	for i := range paymentChannels {
		result = append(result, *s.toModel(&paymentChannels[i]))
	}
	return result, nil
}

func (s *PaymentChannelServiceImpl) toModel(
	paymentChannel *schema.PaymentChannel,
) *models.PaymentChannelResponse {
	if paymentChannel == nil {
		return nil
	}
	return &models.PaymentChannelResponse{
		Id:          paymentChannel.Id,
		Channel:     models.PaymentChannel(paymentChannel.Name),
		Limit:       paymentChannel.Limit,
		SuccessRate: paymentChannel.SuccessRate,
		Fee:         paymentChannel.Fee,
	}
}

func (s *PaymentChannelServiceImpl) getFloat64(existingValue float64, newValue *float64) float64 {
	if newValue == nil {
		return existingValue
	}
	return *newValue
}
