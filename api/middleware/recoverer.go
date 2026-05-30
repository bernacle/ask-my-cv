package middleware

import (
	"log/slog"
	"net/http"

	"github.com/brunosimplicio/ask-my-cv/api/response"
	apperrors "github.com/brunosimplicio/ask-my-cv/pkg/errors"
	"github.com/brunosimplicio/ask-my-cv/pkg/telemetry"
)

func Recoverer(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					log.ErrorContext(r.Context(), "panic recovered",
						"request_id", telemetry.RequestID(r.Context()),
						"panic", recovered,
					)
					response.Error(w, r, apperrors.Internal("internal server error", nil))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
