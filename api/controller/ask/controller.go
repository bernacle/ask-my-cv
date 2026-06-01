package ask

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/brunosimplicio/ask-my-cv/api/middleware"
	"github.com/brunosimplicio/ask-my-cv/api/response"
	askdto "github.com/brunosimplicio/ask-my-cv/internal/dto/ask"
	"github.com/brunosimplicio/ask-my-cv/internal/service/rag"
	apperrors "github.com/brunosimplicio/ask-my-cv/pkg/errors"
	"github.com/brunosimplicio/ask-my-cv/pkg/telemetry"
	"github.com/google/uuid"
)

type Controller struct {
	service          *rag.Service
	maxQuestionChars int
}

func New(service *rag.Service, maxQuestionChars int) *Controller {
	return &Controller{service: service, maxQuestionChars: maxQuestionChars}
}

func (c *Controller) Ask(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var request askdto.Request
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.Error(w, r, apperrors.Invalid("invalid JSON request body"))
		return
	}

	question := strings.TrimSpace(request.Question)
	if question == "" {
		response.Error(w, r, apperrors.Invalid("question is required"))
		return
	}
	if c.maxQuestionChars > 0 && utf8.RuneCountInString(question) > c.maxQuestionChars {
		response.Error(w, r, apperrors.Invalid("question is too long"))
		return
	}

	var conversationID *uuid.UUID
	if strings.TrimSpace(request.ConversationID) != "" {
		parsed, err := uuid.Parse(request.ConversationID)
		if err != nil {
			response.Error(w, r, apperrors.Invalid("conversation_id must be a UUID"))
			return
		}
		conversationID = &parsed
	}

	result, err := c.service.Ask(r.Context(), rag.AskInput{
		Question:       question,
		ConversationID: conversationID,
		RequestID:      telemetry.RequestID(r.Context()),
		IP:             middleware.ClientIP(r),
		UserAgent:      r.UserAgent(),
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "ask failed",
			"request_id", telemetry.RequestID(r.Context()),
			"error", err,
		)
		response.Error(w, r, err)
		return
	}

	response.JSON(w, http.StatusOK, askdto.Response{
		Answer:         result.Answer,
		Citations:      result.Citations,
		ConversationID: result.ConversationID.String(),
		SourcesUsed:    result.SourcesUsed,
		TokenUsage:     result.TokenUsage,
	})
}
