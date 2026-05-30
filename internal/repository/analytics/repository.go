package analytics

import (
	"context"

	db "github.com/brunosimplicio/ask-my-cv/internal/db/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	q *db.Queries
}

type QuestionEvent struct {
	ConversationID     uuid.UUID
	RequestID          string
	Question           string
	NormalizedQuestion string
	RetrievedChunks    int32
	UsedChunks         int32
	MaxSimilarity      float64
	Answered           bool
	IPHash             string
	UserAgent          string
}

type TokenUsage struct {
	ConversationID uuid.UUID
	RequestID      string
	Provider       string
	Model          string
	Operation      string
	InputTokens    int32
	OutputTokens   int32
	TotalTokens    int32
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{q: db.New(pool)}
}

func (r *Repository) CreateQuestionEvent(ctx context.Context, event QuestionEvent) error {
	_, err := r.q.CreateQuestionEvent(ctx, db.CreateQuestionEventParams{
		ConversationID:     uuidToPg(event.ConversationID),
		RequestID:          event.RequestID,
		Question:           event.Question,
		NormalizedQuestion: event.NormalizedQuestion,
		RetrievedChunks:    event.RetrievedChunks,
		UsedChunks:         event.UsedChunks,
		MaxSimilarity:      event.MaxSimilarity,
		Answered:           event.Answered,
		IpHash:             textToPg(event.IPHash),
		UserAgent:          textToPg(event.UserAgent),
	})
	return err
}

func (r *Repository) CreateTokenUsage(ctx context.Context, usage TokenUsage) error {
	if usage.TotalTokens == 0 {
		usage.TotalTokens = usage.InputTokens + usage.OutputTokens
	}
	_, err := r.q.CreateTokenUsage(ctx, db.CreateTokenUsageParams{
		ConversationID: uuidToPg(usage.ConversationID),
		RequestID:      usage.RequestID,
		Provider:       usage.Provider,
		Model:          usage.Model,
		Operation:      usage.Operation,
		InputTokens:    usage.InputTokens,
		OutputTokens:   usage.OutputTokens,
		TotalTokens:    usage.TotalTokens,
	})
	return err
}

func uuidToPg(id uuid.UUID) pgtype.UUID {
	if id == uuid.Nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: id, Valid: true}
}

func textToPg(value string) pgtype.Text {
	if value == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: value, Valid: true}
}
