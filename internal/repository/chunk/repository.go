package chunk

import (
	"context"

	db "github.com/brunosimplicio/ask-my-cv/internal/db/sqlc"
	domainchunk "github.com/brunosimplicio/ask-my-cv/internal/domain/chunk"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	q *db.Queries
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{q: db.New(pool)}
}

func (r *Repository) Search(ctx context.Context, embeddingLiteral string, embeddingModel string, limit int32) ([]domainchunk.SearchResult, error) {
	rows, err := r.q.SearchChunks(ctx, db.SearchChunksParams{
		Embedding:      embeddingLiteral,
		EmbeddingModel: embeddingModel,
		MatchCount:     limit,
	})
	if err != nil {
		return nil, err
	}
	results := make([]domainchunk.SearchResult, 0, len(rows))
	for _, row := range rows {
		results = append(results, domainchunk.SearchResult{
			ID:          row.ID,
			DocumentID:  row.DocumentID,
			Content:     row.Content,
			Metadata:    row.Metadata,
			SourceTitle: row.SourceTitle,
			SourcePath:  row.SourcePath,
			ChunkIndex:  row.ChunkIndex,
			Distance:    row.Distance,
			Similarity:  row.Similarity,
		})
	}
	return results, nil
}
