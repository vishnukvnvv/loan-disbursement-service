package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"payment-gateway/api"
	"payment-gateway/api/service"
	"payment-gateway/db"
	httpclient "payment-gateway/http"
	"payment-gateway/models"
	"payment-gateway/utils"
	"payment-gateway/worker"
	"syscall"

	"github.com/rs/zerolog/log"
)

func main() {
	db, err := db.New(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}

	idGenerator := utils.NewIdGenerator()

	processor := make(chan models.ProcessorMessage)

	serviceFactory := service.NewServiceFactory(db, processor, idGenerator)
	server := api.NewGatewayServer("8080", serviceFactory)

	worker := worker.NewWorker(db, processor, httpclient.NewHTTPClient())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go worker.StartNotifier(ctx)
	go worker.StartProcessor(ctx)

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
	server.Close()

	log.Info().Msg("Payment gateway stopped")
}
