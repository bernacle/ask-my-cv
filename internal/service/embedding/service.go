package embedding

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/brunosimplicio/ask-my-cv/lib/openai"
)

type Client interface {
	CreateEmbedding(ctx context.Context, model string, input string) (openai.EmbeddingResult, error)
}

type Service struct {
	client Client
	model  string
}

type Result struct {
	Model   string
	Values  []float64
	Literal string
	Usage   openai.Usage
}

func (r Result) GetLiteral() string {
	return r.Literal
}

func New(client Client, model string) *Service {
	return &Service{client: client, model: model}
}

func (s *Service) Model() string {
	return s.model
}

func (s *Service) Embed(ctx context.Context, input string) (Result, error) {
	result, err := s.client.CreateEmbedding(ctx, s.model, input)
	if err != nil {
		return Result{}, err
	}
	return Result{
		Model:   s.model,
		Values:  result.Embedding,
		Literal: VectorLiteral(result.Embedding),
		Usage:   result.Usage,
	}, nil
}

func VectorLiteral(values []float64) string {
	parts := make([]string, len(values))
	for i, value := range values {
		parts[i] = strconv.FormatFloat(value, 'g', -1, 64)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func ValidateDimension(values []float64, expected int) error {
	if len(values) != expected {
		return fmt.Errorf("expected embedding dimension %d, got %d", expected, len(values))
	}
	return nil
}
