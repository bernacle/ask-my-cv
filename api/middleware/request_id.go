package middleware

import (
	"net/http"

	"github.com/brunosimplicio/ask-my-cv/pkg/telemetry"
	"github.com/google/uuid"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r.WithContext(telemetry.WithRequestID(r.Context(), requestID)))
	})
}
