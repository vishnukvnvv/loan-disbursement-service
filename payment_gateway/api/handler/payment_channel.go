package handler

import (
	"encoding/json"
	"net/http"
	"payment-gateway/api/service"
	"payment-gateway/models"
	"time"

	"github.com/gorilla/mux"
)

type PaymentChannelHandler struct {
	BaseHandler
	service  service.PaymentChannelService
	schedule service.AvailabilitySchedule
}

func NewPaymentChannelHandler(
	service service.PaymentChannelService,
	schedule service.AvailabilitySchedule,
) *PaymentChannelHandler {
	return &PaymentChannelHandler{
		service:  service,
		schedule: schedule,
	}
}

func (h PaymentChannelHandler) CreatePaymentChannel(w http.ResponseWriter, r *http.Request) {
	var req models.CreatePaymentChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	paymentChannel, err := h.service.CreatePaymentChannel(r.Context(), req)
	if err != nil {
		h.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.JSONResponse(w, paymentChannel)
}

func (h PaymentChannelHandler) ListPaymentChannels(w http.ResponseWriter, r *http.Request) {
	paymentChannels, err := h.service.ListPaymentChannels(r.Context())
	if err != nil {
		h.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.JSONResponse(w, paymentChannels)
}

func (h PaymentChannelHandler) UpdatePaymentChannel(w http.ResponseWriter, r *http.Request) {
	channel := models.PaymentChannel(mux.Vars(r)["channel"])
	var req models.UpdatePaymentChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	paymentChannel, err := h.service.UpdatePaymentChannel(r.Context(), channel, req)
	if err != nil {
		h.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.JSONResponse(w, paymentChannel)
}

func (h PaymentChannelHandler) CheckAvailability(w http.ResponseWriter, r *http.Request) {
	channel := models.PaymentChannel(mux.Vars(r)["channel"])
	time := time.Now()
	available := h.schedule.IsAvailable(channel, time)
	h.JSONResponse(w, map[string]bool{"available": available})
}
