package handler

import (
	"encoding/json"
	"net/http"

	"payment-gateway/api/service"
	"payment-gateway/models"

	"github.com/gorilla/mux"
)

type AccountHandler struct {
	BaseHandler
	service service.AccountService
}

func NewAccountHandler(service service.AccountService) *AccountHandler {
	return &AccountHandler{
		service: service,
	}
}

func (h AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req models.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	account, err := h.service.CreateAccount(r.Context(), req)
	if err != nil {
		h.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.JSONResponse(w, account)
}

func (h AccountHandler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.service.ListAccounts(r.Context())
	if err != nil {
		h.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.JSONResponse(w, accounts)
}

func (h AccountHandler) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var req models.UpdateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	account, err := h.service.UpdateAccount(r.Context(), id, req)
	if err != nil {
		h.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.JSONResponse(w, account)
}
