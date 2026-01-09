package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"loan-disbursement-service/api"
	"loan-disbursement-service/api/services"
	"loan-disbursement-service/db"
	httpclient "loan-disbursement-service/http"
	"loan-disbursement-service/providers"
	"loan-disbursement-service/utils"
	"loan-disbursement-service/worker"

	"github.com/rs/zerolog/log"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal().Msg("DATABASE_URL env var is required")
	}

	database, err := db.New(dsn)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect database")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	idGenerator := utils.NewIdGenerator()

	paymentProvider, err := providers.NewPaymentProvider(
		os.Getenv("PAYMENT_PROVIDER_URL"),
		httpclient.NewHTTPClient(),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create payment provider")
	}
	notificationURL := os.Getenv("NOTIFICATION_URL")

	serviceFactory := services.New(database, idGenerator, paymentProvider, notificationURL)

	worker := worker.NewWorker(
		database.GetDisbursementRepository(),
		serviceFactory.GetPaymentService(),
	)
	go worker.StartPaymentDisbursement(ctx)
	go worker.StartRetryDisbursement(ctx)
	go worker.StartNEFTDisbursement(ctx)

	server := api.New("7070", serviceFactory)

	go func() {
		if err := server.Serve(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server exited with error")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Received interrupt signal, shutting down...")
	worker.Stop(ctx)
	server.Close(ctx)

	log.Info().Msg("Disbursement service stopped")
}
