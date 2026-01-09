package api

import (
	"loan-disbursement-service/api/handlers"
	"net/http"

	"github.com/gorilla/mux"
)

func (d *DisbursementServer) routes() http.Handler {
	router := mux.NewRouter()

	loanService := d.serviceFactory.GetLoanService()
	loanHandler := handlers.NewLoanHandler(loanService)

	subRoute := router.PathPrefix("/api/v1").Subrouter()

	loanSubRoute := subRoute.PathPrefix("/loan").Subrouter()

	loanSubRoute.HandleFunc("", loanHandler.Create).Methods(http.MethodPost)
	loanSubRoute.HandleFunc("", loanHandler.List).Methods(http.MethodGet)
	loanSubRoute.HandleFunc("/{id}", loanHandler.Update).Methods(http.MethodPut)
	loanSubRoute.HandleFunc("/{id}", loanHandler.Get).Methods(http.MethodGet)

	disbursementService := d.serviceFactory.GetDisbursementService()
	disbursementHandler := handlers.NewDisbursementHandler(disbursementService)

	disbursementSubRoute := subRoute.PathPrefix("/disburse").Subrouter()
	disbursementSubRoute.HandleFunc("", disbursementHandler.Disburse).Methods(http.MethodPost)
	disbursementSubRoute.HandleFunc("/{id}", disbursementHandler.Fetch).Methods(http.MethodGet)
	disbursementSubRoute.HandleFunc("/{id}", disbursementHandler.Retry).Methods(http.MethodPost)

	paymentService := d.serviceFactory.GetPaymentService()
	paymentHandler := handlers.NewPaymentHandler(paymentService)

	paymentSubRoute := subRoute.PathPrefix("/payment").Subrouter()
	paymentSubRoute.HandleFunc("/notify", paymentHandler.ProcessPayment).Methods(http.MethodPost)

	return router
}
