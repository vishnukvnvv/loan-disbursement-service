package worker

import (
	"context"
	"loan-disbursement-service/models"

	"github.com/rs/zerolog/log"
)

func (w *Worker) ProcessNEFTBatch(ctx context.Context) {
	status := []models.DisbursementStatus{
		models.DisbursementStatusInitiated,
		models.DisbursementStatusSuspended,
	}
	offset := 0
	channels := []models.PaymentChannel{
		models.PaymentChannelNEFT,
	}
	for {
		disbursements, err := w.disbursement.List(
			ctx,
			offset,
			w.neftBatchSize,
			status,
			channels,
		)
		log.Info().Msgf("NEFT disbursements worker: %v", len(disbursements))
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

		if len(disbursements) < w.neftBatchSize {
			log.Info().Msg("no more disbursements to process")
			break
		}

		offset += w.neftBatchSize
	}
}
