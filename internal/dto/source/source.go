package source

import "time"

type Source struct {
	ID             string    `json:"id"`
	SourcePath     string    `json:"source_path"`
	Title          string    `json:"title"`
	EmbeddingModel string    `json:"embedding_model"`
	ChunkCount     int32     `json:"chunk_count"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type ListResponse struct {
	Sources []Source `json:"sources"`
}
