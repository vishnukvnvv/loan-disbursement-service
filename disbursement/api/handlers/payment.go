package handlers

import (
	"encoding/json"
	"loan-disbursement-service/api/services"
	"loan-disbursement-service/models"
	"net/http"
)

type PaymentHandler struct {
	BaseHandler
	service services.PaymentService
}

func NewPaymentHandler(service services.PaymentService) *PaymentHandler {
	return &PaymentHandler{service: service}
}

func (h PaymentHandler) ProcessPayment(w http.ResponseWriter, r *http.Request) {
	var req models.PaymentNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	err := h.service.HandleNotification(r.Context(), req)
	if err != nil {
		h.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.JSONResponse(w, nil)
}
