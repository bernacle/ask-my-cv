package document

import (
	"context"
	"errors"

	db "github.com/brunosimplicio/ask-my-cv/internal/db/sqlc"
	domainchunk "github.com/brunosimplicio/ask-my-cv/internal/domain/chunk"
	domaindocument "github.com/brunosimplicio/ask-my-cv/internal/domain/document"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool, q: db.New(pool)}
}

func (r *Repository) GetByPath(ctx context.Context, sourcePath string) (db.Document, bool, error) {
	doc, err := r.q.GetDocumentByPath(ctx, sourcePath)
	if err == nil {
		return doc, true, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return db.Document{}, false, nil
	}
	return db.Document{}, false, err
}

func (r *Repository) StoreDocumentChunks(ctx context.Context, sourcePath string, title string, contentHash string, embeddingModel string, chunks []domainchunk.CreateInput) (db.Document, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return db.Document{}, err
	}
	defer tx.Rollback(ctx)

	qtx := r.q.WithTx(tx)
	doc, err := qtx.UpsertDocument(ctx, db.UpsertDocumentParams{
		SourcePath:     sourcePath,
		Title:          title,
		ContentHash:    contentHash,
		EmbeddingModel: embeddingModel,
	})
	if err != nil {
		return db.Document{}, err
	}

	if err := qtx.DeleteChunksByDocument(ctx, doc.ID); err != nil {
		return db.Document{}, err
	}

	for _, item := range chunks {
		_, err := qtx.InsertChunk(ctx, db.InsertChunkParams{
			DocumentID:  doc.ID,
			Content:     item.Content,
			Embedding:   item.Embedding,
			Metadata:    item.Metadata,
			SourceTitle: item.SourceTitle,
			SourcePath:  item.SourcePath,
			ChunkIndex:  item.ChunkIndex,
			ContentHash: item.ContentHash,
		})
		if err != nil {
			return db.Document{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return db.Document{}, err
	}
	return doc, nil
}

func (r *Repository) ListSources(ctx context.Context) ([]domaindocument.Source, error) {
	rows, err := r.q.ListSources(ctx)
	if err != nil {
		return nil, err
	}
	sources := make([]domaindocument.Source, 0, len(rows))
	for _, row := range rows {
		sources = append(sources, domaindocument.Source{
			ID:             row.ID,
			SourcePath:     row.SourcePath,
			Title:          row.Title,
			ContentHash:    row.ContentHash,
			EmbeddingModel: row.EmbeddingModel,
			CreatedAt:      row.CreatedAt.Time,
			UpdatedAt:      row.UpdatedAt.Time,
			ChunkCount:     row.ChunkCount,
		})
	}
	return sources, nil
}

func (r *Repository) DeleteChunksByDocument(ctx context.Context, documentID uuid.UUID) error {
	return r.q.DeleteChunksByDocument(ctx, documentID)
}
