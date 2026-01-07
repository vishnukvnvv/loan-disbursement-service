package handlers

import (
	"encoding/json"
	"net/http"

	"loan-disbursement-service/api/services"
	"loan-disbursement-service/models"

	"github.com/gorilla/mux"
)

type DisbursementHandler struct {
	BaseHandler
	service *services.DisbursementService
}

func NewDisbursementHandler(service *services.DisbursementService) *DisbursementHandler {
	return &DisbursementHandler{
		service: service,
	}
}

func (d DisbursementHandler) Disburse(w http.ResponseWriter, r *http.Request) {
	var req models.DisburseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		d.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := d.service.Disburse(r.Context(), &req)
	if err != nil {
		d.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	d.JSONResponse(w, result)
}

func (d DisbursementHandler) Fetch(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	result, err := d.service.Fetch(r.Context(), id)
	if err != nil {
		d.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	d.JSONResponse(w, result)
}

func (d DisbursementHandler) Retry(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	result, err := d.service.Retry(r.Context(), id)
	if err != nil {
		d.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	d.JSONResponse(w, result)
}
