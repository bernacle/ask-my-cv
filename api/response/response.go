package response

import (
	"encoding/json"
	"errors"
	"net/http"

	apperrors "github.com/brunosimplicio/ask-my-cv/pkg/errors"
	"github.com/brunosimplicio/ask-my-cv/pkg/telemetry"
)

type errorEnvelope struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func Error(w http.ResponseWriter, r *http.Request, err error) {
	var appErr *apperrors.AppError
	status := http.StatusInternalServerError
	code := string(apperrors.CodeInternal)
	message := "internal server error"

	if errors.As(err, &appErr) {
		status = appErr.HTTPStatus
		code = string(appErr.Code)
		message = appErr.Message
	}

	JSON(w, status, errorEnvelope{
		Error: errorBody{
			Code:      code,
			Message:   message,
			RequestID: telemetry.RequestID(r.Context()),
		},
	})
}
