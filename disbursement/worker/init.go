package worker

import (
	"context"
	"loan-disbursement-service/api/services"
	"loan-disbursement-service/db/daos"
	"time"

	"github.com/rs/zerolog/log"
)

type Worker struct {
	disbursement      daos.DisbursementRepository
	paymentService    services.PaymentService
	neftBatchSize     int
	retryBatchSize    int
	neftPollInterval  time.Duration
	retryPollInterval time.Duration
	paymentChan       chan string
	stopChan          chan struct{}
}

func NewWorker(
	disbursement daos.DisbursementRepository,
	paymentService services.PaymentService,
) *Worker {
	return &Worker{
		neftPollInterval:  30 * time.Second,
		retryPollInterval: 5 * time.Second,
		paymentChan:       make(chan string),
		stopChan:          make(chan struct{}),
		disbursement:      disbursement,
		paymentService:    paymentService,
	}
}

func (w Worker) StartPaymentDisbursement(ctx context.Context) {
	log.Ctx(ctx).Info().Msg("Starting payment disbursement worker")
	for {
		select {
		case <-ctx.Done():
			log.Ctx(ctx).Info().Msg("Worker stopped by context")
			close(w.stopChan)
			return
		case <-w.stopChan:
			log.Ctx(ctx).Info().Msg("Worker stopped by signal")
			return
		case disbursementId := <-w.paymentChan:
			log.Ctx(ctx).Info().Msgf("Processing payment for disbursement: %s", disbursementId)
			w.ProcessPaymentBatch(ctx, disbursementId)
		}
	}
}

func (w Worker) StartRetryDisbursement(ctx context.Context) {
	log.Ctx(ctx).Info().Msg("Starting retry disbursement worker")
	ticker := time.NewTicker(w.retryPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Ctx(ctx).Info().Msg("Worker stopped by context")
			close(w.stopChan)
			return
		case <-w.stopChan:
			log.Ctx(ctx).Info().Msg("Worker stopped by signal")
			return
		case <-ticker.C:
			log.Ctx(ctx).Info().Msg("Processing retry disbursement")
			w.ProcessRetryBatch(ctx)
		}
	}
}

func (w Worker) StartNEFTDisbursement(ctx context.Context) {
	log.Ctx(ctx).Info().Msg("Starting neft disbursement worker")

	ticker := time.NewTicker(w.neftPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Ctx(ctx).Info().Msg("Worker stopped by context")
			close(w.stopChan)
			return
		case <-w.stopChan:
			log.Ctx(ctx).Info().Msg("Worker stopped by signal")
			return
		case <-ticker.C:
			log.Ctx(ctx).Info().Msg("Processing neft disbursement")
			w.ProcessNEFTBatch(ctx)
		}
	}
}

func (w Worker) Stop(ctx context.Context) {
	close(w.stopChan)
	log.Ctx(ctx).Info().Msg("Stopping disbursement worker")
}
