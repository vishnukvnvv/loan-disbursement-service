package worker

import (
	"context"
	"loan-disbursement-service/models"

	"github.com/rs/zerolog/log"
)

func (w Worker) ProcessRetryBatch(ctx context.Context) {
	status := []models.DisbursementStatus{
		models.DisbursementStatusSuspended,
	}
	offset := 0
	channels := []models.PaymentChannel{
		models.PaymentChannelUPI,
		models.PaymentChannelIMPS,
	}
	for {
		disbursements, err := w.disbursement.List(
			ctx,
			offset,
			w.retryBatchSize,
			status,
			channels,
		)
		log.Info().Msgf("disbursements worker: %v", len(disbursements))
		if err != nil {
			log.Error().Err(err).Msg("failed to list disbursements")
			return
		}

		for _, disbursement := range disbursements {
			err := w.paymentService.Process(ctx, &disbursement)
			if err != nil {
				log.Error().Err(err).Msg("failed to process disbursement")
			}
		}

		if len(disbursements) < w.retryBatchSize {
			log.Info().Msg("no more disbursements to process")
			break
		}

		offset += w.retryBatchSize
	}
}
