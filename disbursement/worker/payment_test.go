package worker

import (
	"context"
	"errors"
	"loan-disbursement-service/db/schema"
	"loan-disbursement-service/models"
	db_test "loan-disbursement-service/test/db"
	"testing"

	"gorm.io/gorm"
)

func TestWorker_ProcessPaymentBatch(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully processes non-NEFT disbursement", func(t *testing.T) {
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockPaymentService := new(MockPaymentService)

		worker := Worker{
			disbursement:   mockDisbursement,
			paymentService: mockPaymentService,
		}

		disbursementId := "DISB-123"
		disbursement := &schema.Disbursement{
			Id:      disbursementId,
			Channel: models.PaymentChannelUPI,
		}

		mockDisbursement.On("Get", ctx, disbursementId).Return(disbursement, nil).Once()
		mockPaymentService.On("Process", ctx, disbursement).Return(nil).Once()

		worker.ProcessPaymentBatch(ctx, disbursementId)

		mockDisbursement.AssertExpectations(t)
		mockPaymentService.AssertExpectations(t)
	})

	t.Run("returns early when disbursement not found", func(t *testing.T) {
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockPaymentService := new(MockPaymentService)

		worker := Worker{
			disbursement:   mockDisbursement,
			paymentService: mockPaymentService,
		}

		disbursementId := "DISB-123"

		mockDisbursement.On("Get", ctx, disbursementId).
			Return(nil, gorm.ErrRecordNotFound).
			Once()

		worker.ProcessPaymentBatch(ctx, disbursementId)

		mockDisbursement.AssertExpectations(t)
		mockPaymentService.AssertNotCalled(t, "Process")
	})

	t.Run("returns early when disbursement channel is NEFT", func(t *testing.T) {
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockPaymentService := new(MockPaymentService)

		worker := Worker{
			disbursement:   mockDisbursement,
			paymentService: mockPaymentService,
		}

		disbursementId := "DISB-123"
		disbursement := &schema.Disbursement{
			Id:      disbursementId,
			Channel: models.PaymentChannelNEFT,
		}

		mockDisbursement.On("Get", ctx, disbursementId).Return(disbursement, nil).Once()

		worker.ProcessPaymentBatch(ctx, disbursementId)

		mockDisbursement.AssertExpectations(t)
		mockPaymentService.AssertNotCalled(t, "Process")
	})

	t.Run("handles processing error", func(t *testing.T) {
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockPaymentService := new(MockPaymentService)

		worker := Worker{
			disbursement:   mockDisbursement,
			paymentService: mockPaymentService,
		}

		disbursementId := "DISB-123"
		disbursement := &schema.Disbursement{
			Id:      disbursementId,
			Channel: models.PaymentChannelUPI,
		}

		mockDisbursement.On("Get", ctx, disbursementId).Return(disbursement, nil).Once()
		mockPaymentService.On("Process", ctx, disbursement).
			Return(errors.New("processing error")).
			Once()

		worker.ProcessPaymentBatch(ctx, disbursementId)

		mockDisbursement.AssertExpectations(t)
		mockPaymentService.AssertExpectations(t)
	})

	t.Run("handles database error when getting disbursement", func(t *testing.T) {
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockPaymentService := new(MockPaymentService)

		worker := Worker{
			disbursement:   mockDisbursement,
			paymentService: mockPaymentService,
		}

		disbursementId := "DISB-123"

		mockDisbursement.On("Get", ctx, disbursementId).
			Return(nil, errors.New("database connection error")).
			Once()

		worker.ProcessPaymentBatch(ctx, disbursementId)

		mockDisbursement.AssertExpectations(t)
		mockPaymentService.AssertNotCalled(t, "Process")
	})

	t.Run("processes IMPS channel disbursement", func(t *testing.T) {
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockPaymentService := new(MockPaymentService)

		worker := Worker{
			disbursement:   mockDisbursement,
			paymentService: mockPaymentService,
		}

		disbursementId := "DISB-123"
		disbursement := &schema.Disbursement{
			Id:      disbursementId,
			Channel: models.PaymentChannelIMPS,
		}

		mockDisbursement.On("Get", ctx, disbursementId).Return(disbursement, nil).Once()
		mockPaymentService.On("Process", ctx, disbursement).Return(nil).Once()

		worker.ProcessPaymentBatch(ctx, disbursementId)

		mockDisbursement.AssertExpectations(t)
		mockPaymentService.AssertExpectations(t)
	})

	t.Run("processes UPI channel disbursement", func(t *testing.T) {
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockPaymentService := new(MockPaymentService)

		worker := Worker{
			disbursement:   mockDisbursement,
			paymentService: mockPaymentService,
		}

		disbursementId := "DISB-123"
		disbursement := &schema.Disbursement{
			Id:      disbursementId,
			Channel: models.PaymentChannelUPI,
		}

		mockDisbursement.On("Get", ctx, disbursementId).Return(disbursement, nil).Once()
		mockPaymentService.On("Process", ctx, disbursement).Return(nil).Once()

		worker.ProcessPaymentBatch(ctx, disbursementId)

		mockDisbursement.AssertExpectations(t)
		mockPaymentService.AssertExpectations(t)
	})
}
