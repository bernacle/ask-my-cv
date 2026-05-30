package conversation

import (
	"context"
	"errors"

	db "github.com/brunosimplicio/ask-my-cv/internal/db/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	q *db.Queries
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{q: db.New(pool)}
}

func (r *Repository) Create(ctx context.Context) (db.Conversation, error) {
	return r.q.CreateConversation(ctx)
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (db.Conversation, bool, error) {
	conversation, err := r.q.GetConversation(ctx, id)
	if err == nil {
		return conversation, true, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return db.Conversation{}, false, nil
	}
	return db.Conversation{}, false, err
}

func (r *Repository) Ensure(ctx context.Context, id *uuid.UUID) (db.Conversation, error) {
	if id == nil {
		return r.Create(ctx)
	}
	conversation, found, err := r.Get(ctx, *id)
	if err != nil {
		return db.Conversation{}, err
	}
	if found {
		return conversation, nil
	}
	return r.Create(ctx)
}

func (r *Repository) Touch(ctx context.Context, id uuid.UUID) error {
	return r.q.TouchConversation(ctx, id)
}
