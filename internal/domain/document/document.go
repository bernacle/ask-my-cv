package document

import (
	"time"

	"github.com/google/uuid"
)

type Source struct {
	ID             uuid.UUID
	SourcePath     string
	Title          string
	ContentHash    string
	EmbeddingModel string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	ChunkCount     int32
}
