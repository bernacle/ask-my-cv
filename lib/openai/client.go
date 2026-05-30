package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

type Usage struct {
	InputTokens  int
	OutputTokens int
	TotalTokens  int
}

type EmbeddingResult struct {
	Embedding []float64
	Usage     Usage
}

type AnswerRequest struct {
	Model           string
	SystemPrompt    string
	UserPrompt      string
	ReasoningEffort string
	TextVerbosity   string
	MaxOutputTokens int
}

type AnswerResult struct {
	Text  string
	Usage Usage
}

func NewClient(apiKey, baseURL string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: strings.TrimRight(baseURL, "/"),
		http: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

func (c *Client) CreateEmbedding(ctx context.Context, model string, input string) (EmbeddingResult, error) {
	var response struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
		Usage struct {
			PromptTokens int `json:"prompt_tokens"`
			TotalTokens  int `json:"total_tokens"`
		} `json:"usage"`
	}

	err := c.postJSON(ctx, "/v1/embeddings", map[string]any{
		"model": model,
		"input": input,
	}, &response)
	if err != nil {
		return EmbeddingResult{}, err
	}
	if len(response.Data) == 0 || len(response.Data[0].Embedding) == 0 {
		return EmbeddingResult{}, fmt.Errorf("openai returned no embedding")
	}
	return EmbeddingResult{
		Embedding: response.Data[0].Embedding,
		Usage: Usage{
			InputTokens: response.Usage.PromptTokens,
			TotalTokens: response.Usage.TotalTokens,
		},
	}, nil
}

func (c *Client) CreateAnswer(ctx context.Context, req AnswerRequest) (AnswerResult, error) {
	payload := map[string]any{
		"model": req.Model,
		"input": []map[string]any{
			{
				"role": "system",
				"content": []map[string]string{
					{"type": "input_text", "text": req.SystemPrompt},
				},
			},
			{
				"role": "user",
				"content": []map[string]string{
					{"type": "input_text", "text": req.UserPrompt},
				},
			},
		},
		"max_output_tokens": req.MaxOutputTokens,
	}
	if req.ReasoningEffort != "" {
		payload["reasoning"] = map[string]string{"effort": req.ReasoningEffort}
	}
	if req.TextVerbosity != "" {
		payload["text"] = map[string]string{"verbosity": req.TextVerbosity}
	}

	var response struct {
		Output []struct {
			Type    string `json:"type"`
			Role    string `json:"role"`
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
		OutputText string `json:"output_text"`
		Usage      struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
			TotalTokens  int `json:"total_tokens"`
		} `json:"usage"`
	}

	err := c.postJSON(ctx, "/v1/responses", payload, &response)
	if err != nil {
		return AnswerResult{}, err
	}

	text := strings.TrimSpace(response.OutputText)
	if text == "" {
		var parts []string
		for _, item := range response.Output {
			if item.Type != "message" || item.Role != "assistant" {
				continue
			}
			for _, content := range item.Content {
				if content.Type == "output_text" && strings.TrimSpace(content.Text) != "" {
					parts = append(parts, strings.TrimSpace(content.Text))
				}
			}
		}
		text = strings.Join(parts, "\n\n")
	}
	if text == "" {
		return AnswerResult{}, fmt.Errorf("openai returned no answer text")
	}

	return AnswerResult{
		Text: text,
		Usage: Usage{
			InputTokens:  response.Usage.InputTokens,
			OutputTokens: response.Usage.OutputTokens,
			TotalTokens:  response.Usage.TotalTokens,
		},
	}, nil
}

func (c *Client) postJSON(ctx context.Context, path string, payload any, target any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+c.apiKey)
	request.Header.Set("Content-Type", "application/json")

	response, err := c.http.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 4<<20))
	if err != nil {
		return err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("openai status %d: %s", response.StatusCode, strings.TrimSpace(string(responseBody)))
	}
	if err := json.Unmarshal(responseBody, target); err != nil {
		return err
	}
	return nil
}
