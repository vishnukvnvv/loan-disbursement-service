package worker

import (
	"context"
	"loan-disbursement-service/models"

	"github.com/rs/zerolog/log"
)

func (w Worker) ProcessPaymentBatch(ctx context.Context, disbursmentId string) {
	disbursement, err := w.disbursement.Get(
		ctx,
		disbursmentId,
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to get disbursement")
		return
	}
	if disbursement.Channel == models.PaymentChannelNEFT {
		log.Info().Msg("Not processing NEFT in payment worker")
		return
	}

	err = w.paymentService.Process(ctx, disbursement)
	if err != nil {
		log.Error().Err(err).Msg("failed to process disbursement")
	}
}
