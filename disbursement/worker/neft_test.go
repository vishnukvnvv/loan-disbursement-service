package worker

import (
	"context"
	"errors"
	"loan-disbursement-service/db/schema"
	"loan-disbursement-service/models"
	db_test "loan-disbursement-service/test/db"
	"testing"
)

func TestWorker_ProcessNEFTBatch(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully processes single batch of NEFT disbursements", func(t *testing.T) {
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockPaymentService := new(MockPaymentService)

		worker := Worker{
			disbursement:   mockDisbursement,
			paymentService: mockPaymentService,
			neftBatchSize:  10,
		}

		disbursements := []schema.Disbursement{
			{
				Id:      "DISB-1",
				Status:  models.DisbursementStatusInitiated,
				Channel: models.PaymentChannelNEFT,
			},
			{
				Id:      "DISB-2",
				Status:  models.DisbursementStatusSuspended,
				Channel: models.PaymentChannelNEFT,
			},
		}

		mockDisbursement.On("List", ctx, 0, 10, []models.DisbursementStatus{
			models.DisbursementStatusInitiated,
			models.DisbursementStatusSuspended,
		}, []models.PaymentChannel{
			models.PaymentChannelNEFT,
		}).Return(disbursements, nil).Once()

		mockPaymentService.On("Process", ctx, &disbursements[0]).Return(nil).Once()
		mockPaymentService.On("Process", ctx, &disbursements[1]).Return(nil).Once()

		worker.ProcessNEFTBatch(ctx)

		mockDisbursement.AssertExpectations(t)
		mockPaymentService.AssertExpectations(t)
	})

	t.Run("stops when batch size is less than neftBatchSize", func(t *testing.T) {
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockPaymentService := new(MockPaymentService)

		worker := Worker{
			disbursement:   mockDisbursement,
			paymentService: mockPaymentService,
			neftBatchSize:  10,
		}

		disbursements := []schema.Disbursement{
			{
				Id:      "DISB-1",
				Status:  models.DisbursementStatusInitiated,
				Channel: models.PaymentChannelNEFT,
			},
		}

		mockDisbursement.On("List", ctx, 0, 10, []models.DisbursementStatus{
			models.DisbursementStatusInitiated,
			models.DisbursementStatusSuspended,
		}, []models.PaymentChannel{
			models.PaymentChannelNEFT,
		}).Return(disbursements, nil).Once()

		mockPaymentService.On("Process", ctx, &disbursements[0]).Return(nil).Once()

		worker.ProcessNEFTBatch(ctx)

		mockDisbursement.AssertExpectations(t)
		mockPaymentService.AssertExpectations(t)
		mockDisbursement.AssertNumberOfCalls(t, "List", 1)
	})

	t.Run("processes multiple batches with pagination", func(t *testing.T) {
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockPaymentService := new(MockPaymentService)

		worker := Worker{
			disbursement:   mockDisbursement,
			paymentService: mockPaymentService,
			neftBatchSize:  2,
		}

		firstBatch := []schema.Disbursement{
			{
				Id:      "DISB-1",
				Status:  models.DisbursementStatusInitiated,
				Channel: models.PaymentChannelNEFT,
			},
			{
				Id:      "DISB-2",
				Status:  models.DisbursementStatusSuspended,
				Channel: models.PaymentChannelNEFT,
			},
		}

		secondBatch := []schema.Disbursement{
			{
				Id:      "DISB-3",
				Status:  models.DisbursementStatusInitiated,
				Channel: models.PaymentChannelNEFT,
			},
		}

		mockDisbursement.On("List", ctx, 0, 2, []models.DisbursementStatus{
			models.DisbursementStatusInitiated,
			models.DisbursementStatusSuspended,
		}, []models.PaymentChannel{
			models.PaymentChannelNEFT,
		}).Return(firstBatch, nil).Once()

		mockDisbursement.On("List", ctx, 2, 2, []models.DisbursementStatus{
			models.DisbursementStatusInitiated,
			models.DisbursementStatusSuspended,
		}, []models.PaymentChannel{
			models.PaymentChannelNEFT,
		}).Return(secondBatch, nil).Once()

		mockPaymentService.On("Process", ctx, &firstBatch[0]).Return(nil).Once()
		mockPaymentService.On("Process", ctx, &firstBatch[1]).Return(nil).Once()
		mockPaymentService.On("Process", ctx, &secondBatch[0]).Return(nil).Once()

		worker.ProcessNEFTBatch(ctx)

		mockDisbursement.AssertExpectations(t)
		mockPaymentService.AssertExpectations(t)
	})

	t.Run("returns early when List returns error", func(t *testing.T) {
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockPaymentService := new(MockPaymentService)

		worker := Worker{
			disbursement:   mockDisbursement,
			paymentService: mockPaymentService,
			neftBatchSize:  10,
		}

		mockDisbursement.On("List", ctx, 0, 10, []models.DisbursementStatus{
			models.DisbursementStatusInitiated,
			models.DisbursementStatusSuspended,
		}, []models.PaymentChannel{
			models.PaymentChannelNEFT,
		}).Return([]schema.Disbursement{}, errors.New("database error")).Once()

		worker.ProcessNEFTBatch(ctx)

		mockDisbursement.AssertExpectations(t)
		mockPaymentService.AssertNotCalled(t, "Process")
	})

	t.Run("continues processing when one disbursement fails", func(t *testing.T) {
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockPaymentService := new(MockPaymentService)

		worker := Worker{
			disbursement:   mockDisbursement,
			paymentService: mockPaymentService,
			neftBatchSize:  10,
		}

		disbursements := []schema.Disbursement{
			{
				Id:      "DISB-1",
				Status:  models.DisbursementStatusInitiated,
				Channel: models.PaymentChannelNEFT,
			},
			{
				Id:      "DISB-2",
				Status:  models.DisbursementStatusSuspended,
				Channel: models.PaymentChannelNEFT,
			},
			{
				Id:      "DISB-3",
				Status:  models.DisbursementStatusInitiated,
				Channel: models.PaymentChannelNEFT,
			},
		}

		mockDisbursement.On("List", ctx, 0, 10, []models.DisbursementStatus{
			models.DisbursementStatusInitiated,
			models.DisbursementStatusSuspended,
		}, []models.PaymentChannel{
			models.PaymentChannelNEFT,
		}).Return(disbursements, nil).Once()

		mockPaymentService.On("Process", ctx, &disbursements[0]).Return(nil).Once()
		mockPaymentService.On("Process", ctx, &disbursements[1]).
			Return(errors.New("processing error")).
			Once()
		mockPaymentService.On("Process", ctx, &disbursements[2]).Return(nil).Once()

		worker.ProcessNEFTBatch(ctx)

		mockDisbursement.AssertExpectations(t)
		mockPaymentService.AssertExpectations(t)
		mockPaymentService.AssertNumberOfCalls(t, "Process", 3)
	})

	t.Run("handles empty disbursement list", func(t *testing.T) {
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockPaymentService := new(MockPaymentService)

		worker := Worker{
			disbursement:   mockDisbursement,
			paymentService: mockPaymentService,
			neftBatchSize:  10,
		}

		emptyList := []schema.Disbursement{}

		mockDisbursement.On("List", ctx, 0, 10, []models.DisbursementStatus{
			models.DisbursementStatusInitiated,
			models.DisbursementStatusSuspended,
		}, []models.PaymentChannel{
			models.PaymentChannelNEFT,
		}).Return(emptyList, nil).Once()

		worker.ProcessNEFTBatch(ctx)

		mockDisbursement.AssertExpectations(t)
		mockPaymentService.AssertNotCalled(t, "Process")
	})

	t.Run("uses correct status and channel filters", func(t *testing.T) {
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockPaymentService := new(MockPaymentService)

		worker := Worker{
			disbursement:   mockDisbursement,
			paymentService: mockPaymentService,
			neftBatchSize:  10,
		}

		disbursements := []schema.Disbursement{
			{
				Id:      "DISB-1",
				Status:  models.DisbursementStatusInitiated,
				Channel: models.PaymentChannelNEFT,
			},
		}

		mockDisbursement.On("List", ctx, 0, 10,
			[]models.DisbursementStatus{
				models.DisbursementStatusInitiated,
				models.DisbursementStatusSuspended,
			},
			[]models.PaymentChannel{models.PaymentChannelNEFT},
		).Return(disbursements, nil).Once()

		mockPaymentService.On("Process", ctx, &disbursements[0]).Return(nil).Once()

		worker.ProcessNEFTBatch(ctx)

		mockDisbursement.AssertExpectations(t)
		mockPaymentService.AssertExpectations(t)
	})

	t.Run("handles exact batch size boundary", func(t *testing.T) {
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockPaymentService := new(MockPaymentService)

		worker := Worker{
			disbursement:   mockDisbursement,
			paymentService: mockPaymentService,
			neftBatchSize:  2,
		}

		exactBatch := []schema.Disbursement{
			{
				Id:      "DISB-1",
				Status:  models.DisbursementStatusInitiated,
				Channel: models.PaymentChannelNEFT,
			},
			{
				Id:      "DISB-2",
				Status:  models.DisbursementStatusSuspended,
				Channel: models.PaymentChannelNEFT,
			},
		}

		nextBatch := []schema.Disbursement{}

		mockDisbursement.On("List", ctx, 0, 2, []models.DisbursementStatus{
			models.DisbursementStatusInitiated,
			models.DisbursementStatusSuspended,
		}, []models.PaymentChannel{
			models.PaymentChannelNEFT,
		}).Return(exactBatch, nil).Once()

		mockDisbursement.On("List", ctx, 2, 2, []models.DisbursementStatus{
			models.DisbursementStatusInitiated,
			models.DisbursementStatusSuspended,
		}, []models.PaymentChannel{
			models.PaymentChannelNEFT,
		}).Return(nextBatch, nil).Once()

		mockPaymentService.On("Process", ctx, &exactBatch[0]).Return(nil).Once()
		mockPaymentService.On("Process", ctx, &exactBatch[1]).Return(nil).Once()

		worker.ProcessNEFTBatch(ctx)

		mockDisbursement.AssertExpectations(t)
		mockPaymentService.AssertExpectations(t)
	})

	t.Run("processes large batch correctly", func(t *testing.T) {
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockPaymentService := new(MockPaymentService)

		worker := Worker{
			disbursement:   mockDisbursement,
			paymentService: mockPaymentService,
			neftBatchSize:  5,
		}

		largeBatch := make([]schema.Disbursement, 5)
		for i := 0; i < 5; i++ {
			largeBatch[i] = schema.Disbursement{
				Id:      "DISB-" + string(rune(i+1)),
				Status:  models.DisbursementStatusInitiated,
				Channel: models.PaymentChannelNEFT,
			}
		}

		emptyBatch := []schema.Disbursement{}

		mockDisbursement.On("List", ctx, 0, 5, []models.DisbursementStatus{
			models.DisbursementStatusInitiated,
			models.DisbursementStatusSuspended,
		}, []models.PaymentChannel{
			models.PaymentChannelNEFT,
		}).Return(largeBatch, nil).Once()

		mockDisbursement.On("List", ctx, 5, 5, []models.DisbursementStatus{
			models.DisbursementStatusInitiated,
			models.DisbursementStatusSuspended,
		}, []models.PaymentChannel{
			models.PaymentChannelNEFT,
		}).Return(emptyBatch, nil).Once()

		for i := range largeBatch {
			mockPaymentService.On("Process", ctx, &largeBatch[i]).Return(nil).Once()
		}

		worker.ProcessNEFTBatch(ctx)

		mockDisbursement.AssertExpectations(t)
		mockPaymentService.AssertExpectations(t)
		mockPaymentService.AssertNumberOfCalls(t, "Process", 5)
	})

	t.Run("processes both initiated and suspended statuses", func(t *testing.T) {
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockPaymentService := new(MockPaymentService)

		worker := Worker{
			disbursement:   mockDisbursement,
			paymentService: mockPaymentService,
			neftBatchSize:  10,
		}

		disbursements := []schema.Disbursement{
			{
				Id:      "DISB-1",
				Status:  models.DisbursementStatusInitiated,
				Channel: models.PaymentChannelNEFT,
			},
			{
				Id:      "DISB-2",
				Status:  models.DisbursementStatusSuspended,
				Channel: models.PaymentChannelNEFT,
			},
		}

		mockDisbursement.On("List", ctx, 0, 10, []models.DisbursementStatus{
			models.DisbursementStatusInitiated,
			models.DisbursementStatusSuspended,
		}, []models.PaymentChannel{
			models.PaymentChannelNEFT,
		}).Return(disbursements, nil).Once()

		mockPaymentService.On("Process", ctx, &disbursements[0]).Return(nil).Once()
		mockPaymentService.On("Process", ctx, &disbursements[1]).Return(nil).Once()

		worker.ProcessNEFTBatch(ctx)

		mockDisbursement.AssertExpectations(t)
		mockPaymentService.AssertExpectations(t)
	})
}
