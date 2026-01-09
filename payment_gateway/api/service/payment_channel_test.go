package service

import (
	"context"
	"errors"
	"payment-gateway/db/schema"
	"payment-gateway/models"
	db_test "payment-gateway/test/db"
	utils_test "payment-gateway/test/utils"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestPaymentChannelService_CreatePaymentChannel(t *testing.T) {
	ctx := context.Background()

	t.Run("successful payment channel creation", func(t *testing.T) {
		mockRepo := new(db_test.MockPaymentChannelRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewPaymentChannelService(mockRepo, mockIdGenerator)

		channelId := "CH-123456789012"
		req := models.CreatePaymentChannelRequest{
			Channel:     models.PaymentChannelUPI,
			Limit:       100000.0,
			SuccessRate: 0.95,
			Fee:         5.0,
		}

		expectedChannel := &schema.PaymentChannel{
			Id:          channelId,
			Name:        req.Channel,
			Limit:       req.Limit,
			SuccessRate: req.SuccessRate,
			Fee:         req.Fee,
		}

		mockRepo.On("Get", ctx, req.Channel).Return(nil, gorm.ErrRecordNotFound).Once()
		mockIdGenerator.On("GeneratePaymentChannelId").Return(channelId).Once()
		mockRepo.On("Create", ctx, channelId, req.Channel, req.Limit, req.SuccessRate, req.Fee).
			Return(expectedChannel, nil).Once()

		result, err := service.CreatePaymentChannel(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, channelId, result.Id)
		assert.Equal(t, req.Channel, result.Channel)
		assert.Equal(t, req.Limit, result.Limit)
		assert.Equal(t, req.SuccessRate, result.SuccessRate)
		assert.Equal(t, req.Fee, result.Fee)

		mockRepo.AssertExpectations(t)
		mockIdGenerator.AssertExpectations(t)
	})

	t.Run("returns error when payment channel already exists", func(t *testing.T) {
		mockRepo := new(db_test.MockPaymentChannelRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewPaymentChannelService(mockRepo, mockIdGenerator)

		req := models.CreatePaymentChannelRequest{
			Channel:     models.PaymentChannelUPI,
			Limit:       100000.0,
			SuccessRate: 0.95,
			Fee:         5.0,
		}

		existingChannel := &schema.PaymentChannel{
			Id:          "CH-EXISTING",
			Name:        models.PaymentChannelUPI,
			Limit:       50000.0,
			SuccessRate: 0.90,
			Fee:         3.0,
		}

		mockRepo.On("Get", ctx, req.Channel).Return(existingChannel, nil).Once()

		result, err := service.CreatePaymentChannel(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "payment channel already exists")

		mockRepo.AssertExpectations(t)
		mockIdGenerator.AssertNotCalled(t, "GeneratePaymentChannelId")
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run(
		"returns error when repository get fails with non-record-not-found error",
		func(t *testing.T) {
			mockRepo := new(db_test.MockPaymentChannelRepository)
			mockIdGenerator := new(utils_test.MockIdGenerator)

			service := NewPaymentChannelService(mockRepo, mockIdGenerator)

			req := models.CreatePaymentChannelRequest{
				Channel:     models.PaymentChannelUPI,
				Limit:       100000.0,
				SuccessRate: 0.95,
				Fee:         5.0,
			}

			repoError := errors.New("database connection error")
			mockRepo.On("Get", ctx, req.Channel).Return(nil, repoError).Once()

			result, err := service.CreatePaymentChannel(ctx, req)

			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), "failed to get payment channel")

			mockRepo.AssertExpectations(t)
			mockIdGenerator.AssertNotCalled(t, "GeneratePaymentChannelId")
			mockRepo.AssertNotCalled(t, "Create")
		},
	)

	t.Run("returns error when repository create fails", func(t *testing.T) {
		mockRepo := new(db_test.MockPaymentChannelRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewPaymentChannelService(mockRepo, mockIdGenerator)

		channelId := "CH-123456789012"
		req := models.CreatePaymentChannelRequest{
			Channel:     models.PaymentChannelUPI,
			Limit:       100000.0,
			SuccessRate: 0.95,
			Fee:         5.0,
		}

		repoError := errors.New("database error")

		mockRepo.On("Get", ctx, req.Channel).Return(nil, gorm.ErrRecordNotFound).Once()
		mockIdGenerator.On("GeneratePaymentChannelId").Return(channelId).Once()
		mockRepo.On("Create", ctx, channelId, req.Channel, req.Limit, req.SuccessRate, req.Fee).
			Return(nil, repoError).Once()

		result, err := service.CreatePaymentChannel(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, repoError, err)

		mockRepo.AssertExpectations(t)
		mockIdGenerator.AssertExpectations(t)
	})

	t.Run("creates payment channel for different channel types", func(t *testing.T) {
		mockRepo := new(db_test.MockPaymentChannelRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewPaymentChannelService(mockRepo, mockIdGenerator)

		testCases := []struct {
			name    string
			channel models.PaymentChannel
		}{
			{"UPI", models.PaymentChannelUPI},
			{"NEFT", models.PaymentChannelNEFT},
			{"IMPS", models.PaymentChannelIMPS},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				channelId := "CH-123456789012"
				req := models.CreatePaymentChannelRequest{
					Channel:     tc.channel,
					Limit:       100000.0,
					SuccessRate: 0.95,
					Fee:         5.0,
				}

				expectedChannel := &schema.PaymentChannel{
					Id:          channelId,
					Name:        req.Channel,
					Limit:       req.Limit,
					SuccessRate: req.SuccessRate,
					Fee:         req.Fee,
				}

				mockRepo.On("Get", ctx, req.Channel).Return(nil, gorm.ErrRecordNotFound).Once()
				mockIdGenerator.On("GeneratePaymentChannelId").Return(channelId).Once()
				mockRepo.On("Create", ctx, channelId, req.Channel, req.Limit, req.SuccessRate, req.Fee).
					Return(expectedChannel, nil).
					Once()

				result, err := service.CreatePaymentChannel(ctx, req)

				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.channel, result.Channel)

				mockRepo.AssertExpectations(t)
				mockIdGenerator.AssertExpectations(t)
			})
		}
	})
}

func TestPaymentChannelService_ListPaymentChannels(t *testing.T) {
	ctx := context.Background()

	t.Run("successful list payment channels", func(t *testing.T) {
		mockRepo := new(db_test.MockPaymentChannelRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewPaymentChannelService(mockRepo, mockIdGenerator)

		channels := []schema.PaymentChannel{
			{
				Id:          "CH-001",
				Name:        models.PaymentChannelUPI,
				Limit:       100000.0,
				SuccessRate: 0.95,
				Fee:         5.0,
			},
			{
				Id:          "CH-002",
				Name:        models.PaymentChannelIMPS,
				Limit:       500000.0,
				SuccessRate: 0.98,
				Fee:         10.0,
			},
		}

		mockRepo.On("List", ctx).Return(channels, nil).Once()

		result, err := service.ListPaymentChannels(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 2)
		assert.Equal(t, "CH-001", result[0].Id)
		assert.Equal(t, models.PaymentChannelUPI, result[0].Channel)
		assert.Equal(t, 100000.0, result[0].Limit)
		assert.Equal(t, 0.95, result[0].SuccessRate)
		assert.Equal(t, 5.0, result[0].Fee)
		assert.Equal(t, "CH-002", result[1].Id)
		assert.Equal(t, models.PaymentChannelIMPS, result[1].Channel)
		assert.Equal(t, 500000.0, result[1].Limit)
		assert.Equal(t, 0.98, result[1].SuccessRate)
		assert.Equal(t, 10.0, result[1].Fee)

		mockRepo.AssertExpectations(t)
	})

	t.Run("returns empty list when no payment channels exist", func(t *testing.T) {
		mockRepo := new(db_test.MockPaymentChannelRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewPaymentChannelService(mockRepo, mockIdGenerator)

		mockRepo.On("List", ctx).Return([]schema.PaymentChannel{}, nil).Once()

		result, err := service.ListPaymentChannels(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 0)

		mockRepo.AssertExpectations(t)
	})

	t.Run("returns error when repository list fails", func(t *testing.T) {
		mockRepo := new(db_test.MockPaymentChannelRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewPaymentChannelService(mockRepo, mockIdGenerator)

		repoError := errors.New("database error")
		mockRepo.On("List", ctx).Return(nil, repoError).Once()

		result, err := service.ListPaymentChannels(ctx)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, repoError, err)

		mockRepo.AssertExpectations(t)
	})
}

func TestPaymentChannelService_UpdatePaymentChannel(t *testing.T) {
	ctx := context.Background()

	t.Run("successful payment channel update with all fields", func(t *testing.T) {
		mockRepo := new(db_test.MockPaymentChannelRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewPaymentChannelService(mockRepo, mockIdGenerator)

		channel := models.PaymentChannelUPI
		existingChannel := &schema.PaymentChannel{
			Id:          "CH-001",
			Name:        channel,
			Limit:       100000.0,
			SuccessRate: 0.95,
			Fee:         5.0,
		}

		newLimit := 200000.0
		newSuccessRate := 0.98
		newFee := 7.0
		req := models.UpdatePaymentChannelRequest{
			Limit:       &newLimit,
			SuccessRate: &newSuccessRate,
			Fee:         &newFee,
		}

		updatedChannel := &schema.PaymentChannel{
			Id:          "CH-001",
			Name:        channel,
			Limit:       200000.0,
			SuccessRate: 0.98,
			Fee:         7.0,
		}

		mockRepo.On("Get", ctx, channel).Return(existingChannel, nil).Once()
		mockRepo.On("Update", ctx, channel, map[string]any{
			"limit":        200000.0,
			"success_rate": 0.98,
			"fee":          7.0,
		}).Return(updatedChannel, nil).Once()

		result, err := service.UpdatePaymentChannel(ctx, channel, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "CH-001", result.Id)
		assert.Equal(t, 200000.0, result.Limit)
		assert.Equal(t, 0.98, result.SuccessRate)
		assert.Equal(t, 7.0, result.Fee)

		mockRepo.AssertExpectations(t)
	})

	t.Run("successful payment channel update with partial fields (nil values)", func(t *testing.T) {
		mockRepo := new(db_test.MockPaymentChannelRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewPaymentChannelService(mockRepo, mockIdGenerator)

		channel := models.PaymentChannelUPI
		existingChannel := &schema.PaymentChannel{
			Id:          "CH-001",
			Name:        channel,
			Limit:       100000.0,
			SuccessRate: 0.95,
			Fee:         5.0,
		}

		newLimit := 200000.0
		req := models.UpdatePaymentChannelRequest{
			Limit:       &newLimit,
			SuccessRate: nil, // preserve existing
			Fee:         nil, // preserve existing
		}

		updatedChannel := &schema.PaymentChannel{
			Id:          "CH-001",
			Name:        channel,
			Limit:       200000.0,
			SuccessRate: 0.95, // unchanged
			Fee:         5.0,  // unchanged
		}

		mockRepo.On("Get", ctx, channel).Return(existingChannel, nil).Once()
		mockRepo.On("Update", ctx, channel, map[string]any{
			"limit":        200000.0,
			"success_rate": 0.95,
			"fee":          5.0,
		}).Return(updatedChannel, nil).Once()

		result, err := service.UpdatePaymentChannel(ctx, channel, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 200000.0, result.Limit)
		assert.Equal(t, 0.95, result.SuccessRate) // unchanged
		assert.Equal(t, 5.0, result.Fee)          // unchanged

		mockRepo.AssertExpectations(t)
	})

	t.Run("successful payment channel update with only success rate changed", func(t *testing.T) {
		mockRepo := new(db_test.MockPaymentChannelRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewPaymentChannelService(mockRepo, mockIdGenerator)

		channel := models.PaymentChannelNEFT
		existingChannel := &schema.PaymentChannel{
			Id:          "CH-002",
			Name:        channel,
			Limit:       500000.0,
			SuccessRate: 0.90,
			Fee:         10.0,
		}

		newSuccessRate := 0.99
		req := models.UpdatePaymentChannelRequest{
			Limit:       nil,
			SuccessRate: &newSuccessRate,
			Fee:         nil,
		}

		updatedChannel := &schema.PaymentChannel{
			Id:          "CH-002",
			Name:        channel,
			Limit:       500000.0, // unchanged
			SuccessRate: 0.99,
			Fee:         10.0, // unchanged
		}

		mockRepo.On("Get", ctx, channel).Return(existingChannel, nil).Once()
		mockRepo.On("Update", ctx, channel, map[string]any{
			"limit":        500000.0,
			"success_rate": 0.99,
			"fee":          10.0,
		}).Return(updatedChannel, nil).Once()

		result, err := service.UpdatePaymentChannel(ctx, channel, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 500000.0, result.Limit) // unchanged
		assert.Equal(t, 0.99, result.SuccessRate)
		assert.Equal(t, 10.0, result.Fee) // unchanged

		mockRepo.AssertExpectations(t)
	})

	t.Run("returns error when payment channel not found", func(t *testing.T) {
		mockRepo := new(db_test.MockPaymentChannelRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewPaymentChannelService(mockRepo, mockIdGenerator)

		channel := models.PaymentChannelUPI
		req := models.UpdatePaymentChannelRequest{
			Limit:       nil,
			SuccessRate: nil,
			Fee:         nil,
		}

		mockRepo.On("Get", ctx, channel).Return(nil, nil).Once()

		result, err := service.UpdatePaymentChannel(ctx, channel, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "payment channel not found")

		mockRepo.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "Update")
	})

	t.Run("returns error when repository get fails", func(t *testing.T) {
		mockRepo := new(db_test.MockPaymentChannelRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewPaymentChannelService(mockRepo, mockIdGenerator)

		channel := models.PaymentChannelUPI
		req := models.UpdatePaymentChannelRequest{
			Limit:       nil,
			SuccessRate: nil,
			Fee:         nil,
		}

		repoError := errors.New("database error")
		mockRepo.On("Get", ctx, channel).Return(nil, repoError).Once()

		result, err := service.UpdatePaymentChannel(ctx, channel, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, repoError, err)

		mockRepo.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "Update")
	})

	t.Run("returns error when repository update fails", func(t *testing.T) {
		mockRepo := new(db_test.MockPaymentChannelRepository)
		mockIdGenerator := new(utils_test.MockIdGenerator)

		service := NewPaymentChannelService(mockRepo, mockIdGenerator)

		channel := models.PaymentChannelUPI
		existingChannel := &schema.PaymentChannel{
			Id:          "CH-001",
			Name:        channel,
			Limit:       100000.0,
			SuccessRate: 0.95,
			Fee:         5.0,
		}

		newLimit := 200000.0
		req := models.UpdatePaymentChannelRequest{
			Limit:       &newLimit,
			SuccessRate: nil,
			Fee:         nil,
		}

		repoError := errors.New("database error")

		mockRepo.On("Get", ctx, channel).Return(existingChannel, nil).Once()
		mockRepo.On("Update", ctx, channel, map[string]any{
			"limit":        200000.0,
			"success_rate": 0.95,
			"fee":          5.0,
		}).Return(nil, repoError).Once()

		result, err := service.UpdatePaymentChannel(ctx, channel, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, repoError, err)

		mockRepo.AssertExpectations(t)
	})
}
