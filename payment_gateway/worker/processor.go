package worker

import (
	"context"
	"errors"
	"math/rand"
	"payment-gateway/db/schema"
	"payment-gateway/failures"
	"payment-gateway/models"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func (w *Worker) Process(ctx context.Context, message models.ProcessorMessage) {
	transaction, err := w.transaction.Get(ctx, message.TransactionID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get transaction")
		return
	}

	if transaction.Status != models.TransactionStatusInitiated {
		log.Info().Msgf("Transaction already processed: %s", message.TransactionID)
		return
	}

	log.Info().Msgf("Processing transaction: %s", message.TransactionID)
	_, err = w.transaction.Update(ctx, message.TransactionID, map[string]any{
		"status":     models.TransactionStatusProcessing,
		"updated_at": time.Now(),
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to update transaction")
		return
	}

	time.Sleep(time.Duration(message.Delay) * time.Millisecond)

	if rand.Float64() > message.SuccessRate {
		log.Info().Msgf("Transaction failed: %s", message.TransactionID)
		reason := w.getFailureReason()
		w.markTransactionAsFailed(ctx, message.TransactionID, reason)
		return
	}

	accounts, err := w.account.List(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get account")
		w.markTransactionAsFailed(ctx, message.TransactionID, failures.UNKNOWN_ERROR.Error())
		return
	}

	if len(accounts) == 0 {
		log.Error().Msg("no account found")
		w.markTransactionAsFailed(ctx, message.TransactionID, failures.UNKNOWN_ERROR.Error())
		return
	}

	account := accounts[0]
	totalAmount := transaction.Amount + transaction.Fee

	if account.Balance < totalAmount {
		log.Warn().Msgf("Transaction failed due to insufficient balance: %s", message.TransactionID)
		w.markTransactionAsFailed(ctx, message.TransactionID, failures.INSUFFICIENT_BALANCE.Error())
		return
	}

	err = w.db.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&schema.Account{}).
			Where("id = ? AND balance >= ?", account.Id, totalAmount).
			Update("balance", gorm.Expr("balance - ?", totalAmount))

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			log.Warn().
				Msgf("Account balance update affected 0 rows - balance likely changed by concurrent transaction: %s", message.TransactionID)
			return failures.INSUFFICIENT_BALANCE
		}

		processedAt := time.Now()
		if err := tx.Model(&schema.Transaction{}).
			Where("id = ?", message.TransactionID).
			Updates(map[string]any{
				"status":       models.TransactionStatusSuccess,
				"processed_at": processedAt,
				"updated_at":   processedAt,
			}).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		if errors.Is(err, failures.INSUFFICIENT_BALANCE) {
			log.Warn().
				Err(err).
				Msgf("Transaction failed due to insufficient balance (concurrent update): %s", message.TransactionID)
			w.markTransactionAsFailed(
				ctx,
				message.TransactionID,
				failures.INSUFFICIENT_BALANCE.Error(),
			)
		} else {
			log.Error().Err(err).Msg("failed to process transaction in database transaction")
			w.markTransactionAsFailed(ctx, message.TransactionID, failures.UNKNOWN_ERROR.Error())
		}
		return
	}

	log.Info().Msgf("Transaction processed successfully: %s", message.TransactionID)
	w.notifier <- message.TransactionID
}

func (w *Worker) markTransactionAsFailed(ctx context.Context, transactionID, message string) {
	_, err := w.transaction.Update(ctx, transactionID, map[string]any{
		"status":     models.TransactionStatusFailed,
		"message":    message,
		"updated_at": time.Now(),
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to update transaction status")
	}
}

func (w *Worker) getFailureReason() string {
	failure := failures.TRANSACTION_FAILURES[rand.Intn(len(failures.TRANSACTION_FAILURES))]
	return failure.Error()
}
