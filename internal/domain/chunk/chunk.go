package chunk

import (
	"encoding/json"

	"github.com/google/uuid"
)

type SearchResult struct {
	ID          uuid.UUID
	DocumentID  uuid.UUID
	Content     string
	Metadata    json.RawMessage
	SourceTitle string
	SourcePath  string
	ChunkIndex  int32
	Distance    float64
	Similarity  float64
}

type CreateInput struct {
	Content     string
	Embedding   string
	Metadata    json.RawMessage
	SourceTitle string
	SourcePath  string
	ChunkIndex  int32
	ContentHash string
}
