package handler

import (
	"encoding/json"
	"net/http"
	"payment-gateway/api/service"
	"payment-gateway/models"

	"github.com/gorilla/mux"
)

type PaymentHandler struct {
	BaseHandler
	service service.PaymentService
}

func NewPaymentHandler(service service.PaymentService) *PaymentHandler {
	return &PaymentHandler{service: service}
}

func (h PaymentHandler) ProcessPayment(w http.ResponseWriter, r *http.Request) {
	request := models.PaymentRequest{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	transaction, err := h.service.Process(r.Context(), request)
	if err != nil {
		h.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	h.JSONResponse(w, transaction)
}

func (h PaymentHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	transactionID := mux.Vars(r)["id"]
	channel := models.PaymentChannel(mux.Vars(r)["channel"])
	transaction, err := h.service.GetTransaction(r.Context(), channel, transactionID)
	if err != nil {
		h.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	h.JSONResponse(w, transaction)
}
