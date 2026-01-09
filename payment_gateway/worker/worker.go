package worker

import (
	"context"
	"payment-gateway/db"
	"payment-gateway/db/daos"
	httpclient "payment-gateway/http"
	"payment-gateway/models"

	"github.com/rs/zerolog/log"
)

type Worker struct {
	db          *db.Database
	account     daos.AccountRepository
	transaction daos.TransactionRepository
	stopChan    chan struct{}
	processor   chan models.ProcessorMessage
	notifier    chan string
	httpClient  httpclient.HTTPClient
}

func NewWorker(
	database *db.Database,
	processor chan models.ProcessorMessage,
	httpClient httpclient.HTTPClient,
) *Worker {
	return &Worker{
		db:          database,
		account:     database.GetAccountRepository(),
		transaction: database.GetTransactionRepository(),
		stopChan:    make(chan struct{}),
		notifier:    make(chan string),
		processor:   processor,
		httpClient:  httpClient,
	}
}

func (w *Worker) StartProcessor(ctx context.Context) {
	log.Info().Msg("Starting processor worker")
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Worker stopped by context")
			close(w.stopChan)
			return
		case <-w.stopChan:
			log.Info().Msg("Worker stopped by signal")
			return
		case message := <-w.processor:
			log.Info().Msgf("Received message to process transaction: %s", message.TransactionID)
			w.Process(ctx, message)
		}
	}
}

func (w *Worker) StartNotifier(ctx context.Context) {
	log.Info().Msg("Starting notifier worker")
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Worker stopped by context")
			close(w.stopChan)
			return
		case <-w.stopChan:
			log.Info().Msg("Worker stopped by signal")
			return
		case transactionID := <-w.notifier:
			log.Info().Msgf("Received message to notify transaction: %s", transactionID)
			w.Notify(ctx, transactionID)
		}
	}

}

func (w *Worker) Stop(ctx context.Context) {
	close(w.stopChan)
	log.Info().Msg("Stopping worker")
}
