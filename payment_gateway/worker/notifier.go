package worker

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

func (w *Worker) Notify(ctx context.Context, transactionID string) {
	transaction, err := w.transaction.Get(ctx, transactionID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get transaction")
		return
	}

	notificationURL, ok := transaction.Metadata["notification_url"].(string)
	if !ok {
		log.Error().
			Msgf("notification URL not found in metadata for transaction: %s", transactionID)
		return
	}

	payload := map[string]any{
		"transaction_id": transaction.ID,
		"reference_id":   transaction.ReferenceID,
		"status":         transaction.Status,
		"message":        transaction.Message,
		"amount":         transaction.Amount,
		"fee":            transaction.Fee,
		"channel":        transaction.Channel,
		"created_at":     transaction.CreatedAt,
		"updated_at":     transaction.UpdatedAt,
		"processed_at":   transaction.ProcessedAt,
	}

	resp, err := w.httpClient.POST(
		ctx,
		notificationURL,
		payload,
		map[string]string{
			"Content-Type": "application/json",
		},
	)
	if err != nil {
		log.Error().
			Err(err).
			Msgf("failed to send notification for transaction: %s to URL: %s", transactionID, notificationURL)
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Error().
			Int("status_code", resp.StatusCode).
			Str("url", notificationURL).
			Msgf("notification request failed for transaction: %s", transactionID)
		return
	}

	_, err = w.transaction.Update(ctx, transactionID, map[string]any{
		"notified_at": time.Now(),
		"updated_at":  time.Now(),
	})
	if err != nil {
		log.Error().Err(err).Msgf("failed to update notified_at for transaction: %s", transactionID)
		return
	}

	log.Info().
		Str("url", notificationURL).
		Int("status_code", resp.StatusCode).
		Msgf("Notification sent successfully for transaction: %s", transactionID)
}
