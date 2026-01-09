package api

import (
	"net/http"
	"payment-gateway/api/handler"

	"github.com/gorilla/mux"
)

func (g *GatewayServer) routes() http.Handler {
	router := mux.NewRouter()

	subRoute := router.PathPrefix("/api/v1").Subrouter()

	accountHandler := handler.NewAccountHandler(g.serviceFactory.GetAccountService())
	accountSubRoute := subRoute.PathPrefix("/account").Subrouter()

	accountSubRoute.HandleFunc("", accountHandler.CreateAccount).Methods(http.MethodPost)
	accountSubRoute.HandleFunc("", accountHandler.ListAccounts).Methods(http.MethodGet)
	accountSubRoute.HandleFunc("/{id}", accountHandler.UpdateAccount).Methods(http.MethodPut)

	paymentChannelHandler := handler.NewPaymentChannelHandler(
		g.serviceFactory.GetPaymentChannelService(),
		g.serviceFactory.GetAvailabilitySchedule(),
	)
	paymentChannelSubRoute := subRoute.PathPrefix("/channel").Subrouter()

	paymentChannelSubRoute.HandleFunc("", paymentChannelHandler.CreatePaymentChannel).
		Methods(http.MethodPost)
	paymentChannelSubRoute.HandleFunc("", paymentChannelHandler.ListPaymentChannels).
		Methods(http.MethodGet)
	paymentChannelSubRoute.HandleFunc("/{channel}", paymentChannelHandler.UpdatePaymentChannel).
		Methods(http.MethodPut)
	paymentChannelSubRoute.HandleFunc("/{channel}/status", paymentChannelHandler.CheckAvailability).
		Methods(http.MethodGet)

	paymentHandler := handler.NewPaymentHandler(g.serviceFactory.GetPaymentService())
	paymentSubRoute := subRoute.PathPrefix("/payment").Subrouter()

	paymentSubRoute.HandleFunc("", paymentHandler.ProcessPayment).Methods(http.MethodPost)
	paymentSubRoute.HandleFunc("/{channel}/txn/{id}", paymentHandler.GetTransaction).
		Methods(http.MethodGet)

	return router
}
