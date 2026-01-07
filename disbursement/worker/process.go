package worker

import (
	"context"
	"loan-disbursement-service/models"

	"github.com/rs/zerolog/log"
)

func (w Worker) processBatch(ctx context.Context) {
	status := []string{
		string(models.DisbursementStatusInitiated),
		string(models.DisbursementStatusSuspended),
	}
	offset := 0
	for {
		disbursements, err := w.disbursementDAO.List(ctx, offset, w.batchSize, status)
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

		if len(disbursements) < w.batchSize {
			log.Info().Msg("no more disbursements to process")
			break
		}

		offset += w.batchSize
	}
}
