package ingestion

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	db "github.com/brunosimplicio/ask-my-cv/internal/db/sqlc"
	domainchunk "github.com/brunosimplicio/ask-my-cv/internal/domain/chunk"
	"github.com/brunosimplicio/ask-my-cv/internal/service/embedding"
)

type DocumentRepository interface {
	GetByPath(ctx context.Context, sourcePath string) (db.Document, bool, error)
	StoreDocumentChunks(ctx context.Context, sourcePath string, title string, contentHash string, embeddingModel string, chunks []domainchunk.CreateInput) (db.Document, error)
}

type Embedder interface {
	Model() string
	Embed(ctx context.Context, input string) (embedding.Result, error)
}

type Service struct {
	documents DocumentRepository
	embedder  Embedder
}

type Summary struct {
	FilesSeen     int
	FilesSkipped  int
	FilesIngested int
	ChunksCreated int
}

func New(documents DocumentRepository, embedder Embedder) *Service {
	return &Service{documents: documents, embedder: embedder}
}

func (s *Service) IngestDirectory(ctx context.Context, dir string) (Summary, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return Summary{}, err
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	var summary Summary
	for _, entry := range entries {
		if entry.IsDir() || strings.ToLower(filepath.Ext(entry.Name())) != ".md" {
			continue
		}
		summary.FilesSeen++
		result, err := s.ingestFile(ctx, filepath.Join(dir, entry.Name()))
		if err != nil {
			return summary, err
		}
		if result.skipped {
			summary.FilesSkipped++
			continue
		}
		summary.FilesIngested++
		summary.ChunksCreated += result.chunks
	}
	return summary, nil
}

type ingestFileResult struct {
	skipped bool
	chunks  int
}

func (s *Service) ingestFile(ctx context.Context, path string) (ingestFileResult, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return ingestFileResult{}, err
	}

	sourcePath := cleanSourcePath(path)
	parsed := ParseMarkdown(sourcePath, string(raw))
	contentHash := hashString(parsed.Body)

	existing, found, err := s.documents.GetByPath(ctx, sourcePath)
	if err != nil {
		return ingestFileResult{}, err
	}
	if found && existing.ContentHash == contentHash && existing.EmbeddingModel == s.embedder.Model() {
		return ingestFileResult{skipped: true}, nil
	}

	markdownChunks := ChunkMarkdown(parsed.Body)
	chunks := make([]domainchunk.CreateInput, 0, len(markdownChunks))
	for index, item := range markdownChunks {
		embedding, err := s.embedder.Embed(ctx, item.Content)
		if err != nil {
			return ingestFileResult{}, fmt.Errorf("embed %s chunk %d: %w", sourcePath, index, err)
		}

		metadata, err := json.Marshal(map[string]any{
			"heading":    item.Heading,
			"word_count": item.WordCount,
		})
		if err != nil {
			return ingestFileResult{}, err
		}

		chunks = append(chunks, domainchunk.CreateInput{
			Content:     item.Content,
			Embedding:   embedding.Literal,
			Metadata:    metadata,
			SourceTitle: parsed.Title,
			SourcePath:  sourcePath,
			ChunkIndex:  int32(index),
			ContentHash: hashString(item.Content),
		})
	}

	_, err = s.documents.StoreDocumentChunks(ctx, sourcePath, parsed.Title, contentHash, s.embedder.Model(), chunks)
	if err != nil {
		return ingestFileResult{}, err
	}

	return ingestFileResult{chunks: len(chunks)}, nil
}

func hashString(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func cleanSourcePath(path string) string {
	clean := filepath.Clean(path)
	if rel, err := filepath.Rel(".", clean); err == nil && !strings.HasPrefix(rel, "..") {
		clean = rel
	}
	return filepath.ToSlash(clean)
}
