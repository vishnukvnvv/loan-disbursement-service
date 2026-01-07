package worker

import (
	"context"
	"loan-disbursement-service/api/services"
	"loan-disbursement-service/db/daos"
	"time"

	"github.com/rs/zerolog/log"
)

type Worker struct {
	disbursementDAO *daos.DisbursementDAO
	paymentService  *services.PaymentService
	pollInterval    time.Duration
	batchSize       int
	stopChan        chan struct{}
}

func NewWorker(
	disbursementDAO *daos.DisbursementDAO,
	paymentService *services.PaymentService,
) *Worker {
	return &Worker{
		disbursementDAO: disbursementDAO,
		paymentService:  paymentService,
		pollInterval:    5 * time.Second,
		batchSize:       10,
		stopChan:        make(chan struct{}),
	}
}

func (w Worker) Start(ctx context.Context) {
	log.Info().Msg("Starting disbursement worker")

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Worker stopped by context")
			close(w.stopChan)
			return
		case <-w.stopChan:
			log.Info().Msg("Worker stopped by signal")
			return
		case <-ticker.C:
			log.Info().Msg("Processing batch")
			w.processBatch(ctx)
		}
	}
}

func (w Worker) Stop(ctx context.Context) {
	close(w.stopChan)
	log.Info().Msg("Stopping disbursement worker")
}
