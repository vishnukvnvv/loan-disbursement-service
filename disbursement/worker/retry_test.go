package worker

import (
	"context"
	"errors"
	"loan-disbursement-service/db/schema"
	"loan-disbursement-service/models"
	db_test "loan-disbursement-service/test/db"
	"testing"

	"github.com/stretchr/testify/mock"
)

// MockPaymentService for testing
type MockPaymentService struct {
	mock.Mock
}

func (m *MockPaymentService) Process(ctx context.Context, disbursement *schema.Disbursement) error {
	args := m.Called(ctx, disbursement)
	return args.Error(0)
}

func (m *MockPaymentService) HandleNotification(
	ctx context.Context,
	notification models.PaymentNotificationRequest,
) error {
	args := m.Called(ctx, notification)
	return args.Error(0)
}

func (m *MockPaymentService) HandleFailure(
	ctx context.Context,
	disbursement *schema.Disbursement,
	transaction *schema.Transaction,
	channel models.PaymentChannel,
	err error,
) error {
	args := m.Called(ctx, disbursement, transaction, channel, err)
	return args.Error(0)
}

func (m *MockPaymentService) HanleSuccess(
	ctx context.Context,
	disbursementId, transactionId string,
	channel models.PaymentChannel,
) error {
	args := m.Called(ctx, disbursementId, transactionId, channel)
	return args.Error(0)
}

func TestWorker_ProcessRetryBatch(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully processes single batch of disbursements", func(t *testing.T) {
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockPaymentService := new(MockPaymentService)

		worker := Worker{
			disbursement:   mockDisbursement,
			paymentService: mockPaymentService,
			retryBatchSize: 10,
		}

		disbursements := []schema.Disbursement{
			{
				Id:     "DISB-1",
				Status: models.DisbursementStatusSuspended,
			},
			{
				Id:     "DISB-2",
				Status: models.DisbursementStatusSuspended,
			},
		}

		mockDisbursement.On("List", ctx, 0, 10, []models.DisbursementStatus{
			models.DisbursementStatusSuspended,
		}, []models.PaymentChannel{
			models.PaymentChannelUPI,
			models.PaymentChannelIMPS,
		}).Return(disbursements, nil).Once()

		mockPaymentService.On("Process", ctx, &disbursements[0]).Return(nil).Once()
		mockPaymentService.On("Process", ctx, &disbursements[1]).Return(nil).Once()

		worker.ProcessRetryBatch(ctx)

		mockDisbursement.AssertExpectations(t)
		mockPaymentService.AssertExpectations(t)
	})

	t.Run("stops when batch size is less than retryBatchSize", func(t *testing.T) {
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockPaymentService := new(MockPaymentService)

		worker := Worker{
			disbursement:   mockDisbursement,
			paymentService: mockPaymentService,
			retryBatchSize: 10,
		}

		disbursements := []schema.Disbursement{
			{
				Id:     "DISB-1",
				Status: models.DisbursementStatusSuspended,
			},
		}

		mockDisbursement.On("List", ctx, 0, 10, []models.DisbursementStatus{
			models.DisbursementStatusSuspended,
		}, []models.PaymentChannel{
			models.PaymentChannelUPI,
			models.PaymentChannelIMPS,
		}).Return(disbursements, nil).Once()

		mockPaymentService.On("Process", ctx, &disbursements[0]).Return(nil).Once()

		worker.ProcessRetryBatch(ctx)

		mockDisbursement.AssertExpectations(t)
		mockPaymentService.AssertExpectations(t)
		// Should not call List again since batch size < retryBatchSize
		mockDisbursement.AssertNumberOfCalls(t, "List", 1)
	})

	t.Run("processes multiple batches with pagination", func(t *testing.T) {
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockPaymentService := new(MockPaymentService)

		worker := Worker{
			disbursement:   mockDisbursement,
			paymentService: mockPaymentService,
			retryBatchSize: 2,
		}

		firstBatch := []schema.Disbursement{
			{
				Id:     "DISB-1",
				Status: models.DisbursementStatusSuspended,
			},
			{
				Id:     "DISB-2",
				Status: models.DisbursementStatusSuspended,
			},
		}

		secondBatch := []schema.Disbursement{
			{
				Id:     "DISB-3",
				Status: models.DisbursementStatusSuspended,
			},
		}

		mockDisbursement.On("List", ctx, 0, 2, []models.DisbursementStatus{
			models.DisbursementStatusSuspended,
		}, []models.PaymentChannel{
			models.PaymentChannelUPI,
			models.PaymentChannelIMPS,
		}).Return(firstBatch, nil).Once()

		mockDisbursement.On("List", ctx, 2, 2, []models.DisbursementStatus{
			models.DisbursementStatusSuspended,
		}, []models.PaymentChannel{
			models.PaymentChannelUPI,
			models.PaymentChannelIMPS,
		}).Return(secondBatch, nil).Once()

		mockPaymentService.On("Process", ctx, &firstBatch[0]).Return(nil).Once()
		mockPaymentService.On("Process", ctx, &firstBatch[1]).Return(nil).Once()
		mockPaymentService.On("Process", ctx, &secondBatch[0]).Return(nil).Once()

		worker.ProcessRetryBatch(ctx)

		mockDisbursement.AssertExpectations(t)
		mockPaymentService.AssertExpectations(t)
	})

	t.Run("returns early when List returns error", func(t *testing.T) {
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockPaymentService := new(MockPaymentService)

		worker := Worker{
			disbursement:   mockDisbursement,
			paymentService: mockPaymentService,
			retryBatchSize: 10,
		}

		mockDisbursement.On("List", ctx, 0, 10, []models.DisbursementStatus{
			models.DisbursementStatusSuspended,
		}, []models.PaymentChannel{
			models.PaymentChannelUPI,
			models.PaymentChannelIMPS,
		}).Return([]schema.Disbursement{}, errors.New("database error")).Once()

		worker.ProcessRetryBatch(ctx)

		mockDisbursement.AssertExpectations(t)
		mockPaymentService.AssertNotCalled(t, "Process")
	})

	t.Run("continues processing when one disbursement fails", func(t *testing.T) {
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockPaymentService := new(MockPaymentService)

		worker := Worker{
			disbursement:   mockDisbursement,
			paymentService: mockPaymentService,
			retryBatchSize: 10,
		}

		disbursements := []schema.Disbursement{
			{
				Id:     "DISB-1",
				Status: models.DisbursementStatusSuspended,
			},
			{
				Id:     "DISB-2",
				Status: models.DisbursementStatusSuspended,
			},
			{
				Id:     "DISB-3",
				Status: models.DisbursementStatusSuspended,
			},
		}

		mockDisbursement.On("List", ctx, 0, 10, []models.DisbursementStatus{
			models.DisbursementStatusSuspended,
		}, []models.PaymentChannel{
			models.PaymentChannelUPI,
			models.PaymentChannelIMPS,
		}).Return(disbursements, nil).Once()

		mockPaymentService.On("Process", ctx, &disbursements[0]).Return(nil).Once()
		mockPaymentService.On("Process", ctx, &disbursements[1]).
			Return(errors.New("processing error")).
			Once()
		mockPaymentService.On("Process", ctx, &disbursements[2]).Return(nil).Once()

		worker.ProcessRetryBatch(ctx)

		mockDisbursement.AssertExpectations(t)
		mockPaymentService.AssertExpectations(t)
		// All three should be processed despite one failing
		mockPaymentService.AssertNumberOfCalls(t, "Process", 3)
	})

	t.Run("handles empty disbursement list", func(t *testing.T) {
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockPaymentService := new(MockPaymentService)

		worker := Worker{
			disbursement:   mockDisbursement,
			paymentService: mockPaymentService,
			retryBatchSize: 10,
		}

		emptyList := []schema.Disbursement{}

		mockDisbursement.On("List", ctx, 0, 10, []models.DisbursementStatus{
			models.DisbursementStatusSuspended,
		}, []models.PaymentChannel{
			models.PaymentChannelUPI,
			models.PaymentChannelIMPS,
		}).Return(emptyList, nil).Once()

		worker.ProcessRetryBatch(ctx)

		mockDisbursement.AssertExpectations(t)
		mockPaymentService.AssertNotCalled(t, "Process")
	})

	t.Run("uses correct status and channel filters", func(t *testing.T) {
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockPaymentService := new(MockPaymentService)

		worker := Worker{
			disbursement:   mockDisbursement,
			paymentService: mockPaymentService,
			retryBatchSize: 10,
		}

		disbursements := []schema.Disbursement{
			{
				Id:     "DISB-1",
				Status: models.DisbursementStatusSuspended,
			},
		}

		mockDisbursement.On("List", ctx, 0, 10,
			[]models.DisbursementStatus{models.DisbursementStatusSuspended},
			[]models.PaymentChannel{models.PaymentChannelUPI, models.PaymentChannelIMPS},
		).Return(disbursements, nil).Once()

		mockPaymentService.On("Process", ctx, &disbursements[0]).Return(nil).Once()

		worker.ProcessRetryBatch(ctx)

		mockDisbursement.AssertExpectations(t)
		mockPaymentService.AssertExpectations(t)
	})

	t.Run("handles exact batch size boundary", func(t *testing.T) {
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockPaymentService := new(MockPaymentService)

		worker := Worker{
			disbursement:   mockDisbursement,
			paymentService: mockPaymentService,
			retryBatchSize: 2,
		}

		exactBatch := []schema.Disbursement{
			{
				Id:     "DISB-1",
				Status: models.DisbursementStatusSuspended,
			},
			{
				Id:     "DISB-2",
				Status: models.DisbursementStatusSuspended,
			},
		}

		nextBatch := []schema.Disbursement{}

		mockDisbursement.On("List", ctx, 0, 2, []models.DisbursementStatus{
			models.DisbursementStatusSuspended,
		}, []models.PaymentChannel{
			models.PaymentChannelUPI,
			models.PaymentChannelIMPS,
		}).Return(exactBatch, nil).Once()

		mockDisbursement.On("List", ctx, 2, 2, []models.DisbursementStatus{
			models.DisbursementStatusSuspended,
		}, []models.PaymentChannel{
			models.PaymentChannelUPI,
			models.PaymentChannelIMPS,
		}).Return(nextBatch, nil).Once()

		mockPaymentService.On("Process", ctx, &exactBatch[0]).Return(nil).Once()
		mockPaymentService.On("Process", ctx, &exactBatch[1]).Return(nil).Once()

		worker.ProcessRetryBatch(ctx)

		mockDisbursement.AssertExpectations(t)
		mockPaymentService.AssertExpectations(t)
	})

	t.Run("processes large batch correctly", func(t *testing.T) {
		mockDisbursement := new(db_test.MockDisbursementRepository)
		mockPaymentService := new(MockPaymentService)

		worker := Worker{
			disbursement:   mockDisbursement,
			paymentService: mockPaymentService,
			retryBatchSize: 5,
		}

		largeBatch := make([]schema.Disbursement, 5)
		for i := 0; i < 5; i++ {
			largeBatch[i] = schema.Disbursement{
				Id:     "DISB-" + string(rune(i+1)),
				Status: models.DisbursementStatusSuspended,
			}
		}

		emptyBatch := []schema.Disbursement{}

		mockDisbursement.On("List", ctx, 0, 5, []models.DisbursementStatus{
			models.DisbursementStatusSuspended,
		}, []models.PaymentChannel{
			models.PaymentChannelUPI,
			models.PaymentChannelIMPS,
		}).Return(largeBatch, nil).Once()

		mockDisbursement.On("List", ctx, 5, 5, []models.DisbursementStatus{
			models.DisbursementStatusSuspended,
		}, []models.PaymentChannel{
			models.PaymentChannelUPI,
			models.PaymentChannelIMPS,
		}).Return(emptyBatch, nil).Once()

		for i := range largeBatch {
			mockPaymentService.On("Process", ctx, &largeBatch[i]).Return(nil).Once()
		}

		worker.ProcessRetryBatch(ctx)

		mockDisbursement.AssertExpectations(t)
		mockPaymentService.AssertExpectations(t)
		mockPaymentService.AssertNumberOfCalls(t, "Process", 5)
	})
}
