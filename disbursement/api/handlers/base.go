package handlers

import (
	"encoding/json"
	"net/http"
)

type BaseHandler struct{}

func (b *BaseHandler) JSONResponse(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(v)
}

func (b *BaseHandler) ErrorResponse(w http.ResponseWriter, status int, error string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": error})
}
