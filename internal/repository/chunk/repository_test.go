package chunk

import (
	"context"
	"os"
	"testing"

	db "github.com/brunosimplicio/ask-my-cv/internal/db/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestSearchByKeywordsRanksRelevantCareerFactsBeforeGenericMatches(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is required for the PostgreSQL integration test")
	}
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(ctx)
	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	// Temporary tables isolate the query test without requiring pgvector or migrations.
	_, err = tx.Exec(ctx, `
		CREATE TEMP TABLE documents (
			id uuid PRIMARY KEY,
			embedding_model text NOT NULL
		) ON COMMIT DROP;
		CREATE TEMP TABLE chunks (
			id uuid PRIMARY KEY,
			document_id uuid NOT NULL,
			content text NOT NULL,
			metadata jsonb NOT NULL DEFAULT '{}',
			source_title text NOT NULL,
			source_path text NOT NULL,
			chunk_index integer NOT NULL
		) ON COMMIT DROP;
	`)
	if err != nil {
		t.Fatal(err)
	}

	model := "text-embedding-3-small"
	documentID := uuid.New()
	otherDocumentID := uuid.New()
	_, err = tx.Exec(ctx, `INSERT INTO documents (id, embedding_model) VALUES ($1, $2), ($3, $4)`, documentID, model, otherDocumentID, "other-model")
	if err != nil {
		t.Fatal(err)
	}
	insert := func(id, document uuid.UUID, path string, index int, content string) {
		t.Helper()
		_, err := tx.Exec(ctx, `INSERT INTO chunks (id, document_id, content, source_title, source_path, chunk_index) VALUES ($1, $2, $3, $4, $5, $6)`, id, document, content, path, path, index)
		if err != nil {
			t.Fatal(err)
		}
	}
	for index := 0; index < 6; index++ {
		insert(uuid.New(), documentID, "/app/data/ai-learning.md", index, "Bruno is interested in AI engineering and work on product workflows.")
	}
	wantID := uuid.New()
	insert(wantID, documentID, "/app/data/resume.md", 0, "Bruno joined Notabene as its second engineering hire and worked there from 2021 to 2025.")
	insert(uuid.New(), otherDocumentID, "/app/data/other.md", 0, "engineering hire was notabene when work there")

	repository := &Repository{q: db.New(tx)}
	results, err := repository.SearchByKeywords(ctx, []string{"engineering", "hire", "was", "notabene", "when", "work", "there"}, model, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].ID != wantID {
		t.Fatalf("keyword search = %#v, want the indexed Notabene career fact", results)
	}
}
