package middlewares

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

func CorrelationIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = fmt.Sprintf("req-%s", uuid.New().String())
		}
		ctx := context.WithValue(r.Context(), "request_id", id)
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
