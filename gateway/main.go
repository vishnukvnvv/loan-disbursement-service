package main

import (
	"log"
	"math/rand"
	"mock-payment-gateway/api"
	"mock-payment-gateway/config"
	"mock-payment-gateway/payment"
	"time"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration : ", err)
	}

	rand.Seed(time.Now().UnixNano())
	payment.Initialize(cfg.PaymentModes)

	app := api.New(cfg)
	app.Serve()
}
