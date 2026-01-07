package handlers

import (
	"encoding/json"
	"net/http"

	"loan-disbursement-service/api/services"
	"loan-disbursement-service/models"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type LoanHandler struct {
	BaseHandler
	service *services.LoanService
}

func NewLoanHandler(service *services.LoanService) *LoanHandler {
	return &LoanHandler{
		service: service,
	}
}

func (l LoanHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.LoanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	loan, err := l.service.Create(r.Context(), req.Amount)
	if err != nil {
		l.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	l.JSONResponse(w, loan)
}

func (l LoanHandler) Update(w http.ResponseWriter, r *http.Request) {
	loanId := mux.Vars(r)["id"]
	var req models.LoanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	loan, err := l.service.Update(r.Context(), loanId, map[string]any{"amount": req.Amount})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			l.ErrorResponse(w, http.StatusNotFound, "loan not found")
			return
		}
		l.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	l.JSONResponse(w, loan)
}

func (l LoanHandler) Get(w http.ResponseWriter, r *http.Request) {
	loanId := mux.Vars(r)["id"]
	loan, err := l.service.Get(r.Context(), loanId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			l.ErrorResponse(w, http.StatusNotFound, "loan not found")
			return
		}
		l.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	l.JSONResponse(w, loan)
}

func (l LoanHandler) List(w http.ResponseWriter, r *http.Request) {
	loans, err := l.service.List(r.Context())
	if err != nil {
		l.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	l.JSONResponse(w, loans)
}
