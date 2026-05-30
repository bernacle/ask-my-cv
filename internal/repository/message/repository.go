package message

import (
	"context"
	"encoding/json"

	db "github.com/brunosimplicio/ask-my-cv/internal/db/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	q *db.Queries
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{q: db.New(pool)}
}

func (r *Repository) Create(ctx context.Context, conversationID uuid.UUID, role string, content string, citations json.RawMessage) (db.Message, error) {
	return r.q.CreateMessage(ctx, db.CreateMessageParams{
		ConversationID: conversationID,
		Role:           role,
		Content:        content,
		Citations:      citations,
	})
}
