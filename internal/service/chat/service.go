package chat

import (
	"context"

	"github.com/brunosimplicio/ask-my-cv/lib/openai"
)

type Client interface {
	CreateAnswer(ctx context.Context, req openai.AnswerRequest) (openai.AnswerResult, error)
}

type Service struct {
	client          Client
	model           string
	reasoningEffort string
	textVerbosity   string
	maxOutputTokens int
}

type Result struct {
	Text  string
	Usage openai.Usage
	Model string
}

func New(client Client, model, reasoningEffort, textVerbosity string, maxOutputTokens int) *Service {
	return &Service{
		client:          client,
		model:           model,
		reasoningEffort: reasoningEffort,
		textVerbosity:   textVerbosity,
		maxOutputTokens: maxOutputTokens,
	}
}

func (s *Service) Model() string {
	return s.model
}

func (s *Service) Answer(ctx context.Context, systemPrompt, userPrompt string) (Result, error) {
	result, err := s.client.CreateAnswer(ctx, openai.AnswerRequest{
		Model:           s.model,
		SystemPrompt:    systemPrompt,
		UserPrompt:      userPrompt,
		ReasoningEffort: s.reasoningEffort,
		TextVerbosity:   s.textVerbosity,
		MaxOutputTokens: s.maxOutputTokens,
	})
	if err != nil {
		return Result{}, err
	}
	return Result{Text: result.Text, Usage: result.Usage, Model: s.model}, nil
}
