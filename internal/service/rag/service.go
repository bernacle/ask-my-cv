package rag

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"time"

	db "github.com/brunosimplicio/ask-my-cv/internal/db/sqlc"
	domainchunk "github.com/brunosimplicio/ask-my-cv/internal/domain/chunk"
	"github.com/brunosimplicio/ask-my-cv/internal/domain/message"
	analyticsrepo "github.com/brunosimplicio/ask-my-cv/internal/repository/analytics"
	"github.com/brunosimplicio/ask-my-cv/internal/service/chat"
	"github.com/brunosimplicio/ask-my-cv/internal/service/embedding"
	apperrors "github.com/brunosimplicio/ask-my-cv/pkg/errors"
	"github.com/google/uuid"
)

const noSourceAnswer = "I don't have enough information in Bruno's CV to answer that."

const fallbackSimilarityThreshold = 0.52

type Embedder interface {
	Model() string
	Embed(ctx context.Context, input string) (embedding.Result, error)
}

type ChunkSearcher interface {
	Search(ctx context.Context, embeddingLiteral string, embeddingModel string, limit int32) ([]domainchunk.SearchResult, error)
	SearchByKeywords(ctx context.Context, keywords []string, embeddingModel string, limit int32) ([]domainchunk.SearchResult, error)
}

type ConversationRepository interface {
	Ensure(ctx context.Context, id *uuid.UUID) (db.Conversation, error)
	Touch(ctx context.Context, id uuid.UUID) error
}

type MessageRepository interface {
	Create(ctx context.Context, conversationID uuid.UUID, role string, content string, citations json.RawMessage) (db.Message, error)
}

type AnalyticsRepository interface {
	CreateQuestionEvent(ctx context.Context, event analyticsrepo.QuestionEvent) error
	CreateTokenUsage(ctx context.Context, usage analyticsrepo.TokenUsage) error
}

type ChatService interface {
	Model() string
	Answer(ctx context.Context, systemPrompt, userPrompt string) (chat.Result, error)
}

type Service struct {
	embedder            Embedder
	chunks              ChunkSearcher
	conversations       ConversationRepository
	messages            MessageRepository
	analytics           AnalyticsRepository
	chat                ChatService
	maxRetrievedChunks  int
	similarityThreshold float64
	cacheTTL            time.Duration
	cache               map[string]cacheEntry
	cacheMu             sync.Mutex
}

type AskInput struct {
	Question       string
	ConversationID *uuid.UUID
	RequestID      string
	IP             string
	UserAgent      string
}

type AskResult struct {
	Answer         string
	ConversationID uuid.UUID
	Citations      []message.Citation
	SourcesUsed    []message.Citation
	TokenUsage     *message.TokenUsage
}

type cacheEntry struct {
	answer     string
	citations  []message.Citation
	expiresAt  time.Time
	answered   bool
	maxSim     float64
	usedChunks int32
}

func New(
	embedder Embedder,
	chunks ChunkSearcher,
	conversations ConversationRepository,
	messages MessageRepository,
	analytics AnalyticsRepository,
	chat ChatService,
	maxRetrievedChunks int,
	similarityThreshold float64,
	cacheTTL time.Duration,
) *Service {
	if maxRetrievedChunks <= 0 {
		maxRetrievedChunks = 6
	}
	return &Service{
		embedder:            embedder,
		chunks:              chunks,
		conversations:       conversations,
		messages:            messages,
		analytics:           analytics,
		chat:                chat,
		maxRetrievedChunks:  maxRetrievedChunks,
		similarityThreshold: similarityThreshold,
		cacheTTL:            cacheTTL,
		cache:               make(map[string]cacheEntry),
	}
}

func (s *Service) Ask(ctx context.Context, input AskInput) (AskResult, error) {
	question := strings.TrimSpace(input.Question)
	if question == "" {
		return AskResult{}, apperrors.Invalid("question is required")
	}
	normalized := normalizeQuestion(question)
	retrievalQuery := expandRetrievalQuery(question)

	conversation, err := s.conversations.Ensure(ctx, input.ConversationID)
	if err != nil {
		return AskResult{}, err
	}

	if cached, ok := s.getCached(normalized); ok {
		if err := s.persistExchange(ctx, conversation.ID, question, cached.answer, cached.citations); err != nil {
			return AskResult{}, err
		}
		_ = s.conversations.Touch(ctx, conversation.ID)
		if err := s.analytics.CreateQuestionEvent(ctx, analyticsrepo.QuestionEvent{
			ConversationID:     conversation.ID,
			RequestID:          input.RequestID,
			Question:           question,
			NormalizedQuestion: normalized,
			RetrievedChunks:    cached.usedChunks,
			UsedChunks:         cached.usedChunks,
			MaxSimilarity:      cached.maxSim,
			Answered:           cached.answered,
			IPHash:             hashIP(input.IP),
			UserAgent:          input.UserAgent,
		}); err != nil {
			return AskResult{}, err
		}
		return AskResult{
			Answer:         cached.answer,
			ConversationID: conversation.ID,
			Citations:      cached.citations,
			SourcesUsed:    cached.citations,
		}, nil
	}

	questionEmbedding, err := s.embedder.Embed(ctx, retrievalQuery)
	if err != nil {
		return AskResult{}, apperrors.OpenAI("failed to embed question", err)
	}
	tokenUsage := message.TokenUsage{
		InputTokens:  questionEmbedding.Usage.InputTokens,
		OutputTokens: questionEmbedding.Usage.OutputTokens,
		TotalTokens:  questionEmbedding.Usage.TotalTokens,
	}
	if err := s.analytics.CreateTokenUsage(ctx, analyticsrepo.TokenUsage{
		ConversationID: conversation.ID,
		RequestID:      input.RequestID,
		Provider:       "openai",
		Model:          questionEmbedding.Model,
		Operation:      "embedding",
		InputTokens:    int32(questionEmbedding.Usage.InputTokens),
		OutputTokens:   int32(questionEmbedding.Usage.OutputTokens),
		TotalTokens:    int32(questionEmbedding.Usage.TotalTokens),
	}); err != nil {
		return AskResult{}, err
	}

	retrieved, err := s.chunks.Search(ctx, questionEmbedding.Literal, questionEmbedding.Model, int32(s.maxRetrievedChunks))
	if err != nil {
		return AskResult{}, err
	}
	keywordResults, err := s.chunks.SearchByKeywords(ctx, extractKeywords(retrievalQuery), questionEmbedding.Model, int32(s.maxRetrievedChunks))
	if err != nil {
		return AskResult{}, err
	}
	retrieved = mergeChunkResults(keywordResults, retrieved, s.maxRetrievedChunks)
	used, maxSimilarity := selectChunks(retrieved, s.similarityThreshold)

	if len(used) == 0 {
		if err := s.persistExchange(ctx, conversation.ID, question, noSourceAnswer, nil); err != nil {
			return AskResult{}, err
		}
		_ = s.conversations.Touch(ctx, conversation.ID)
		if err := s.analytics.CreateQuestionEvent(ctx, analyticsrepo.QuestionEvent{
			ConversationID:     conversation.ID,
			RequestID:          input.RequestID,
			Question:           question,
			NormalizedQuestion: normalized,
			RetrievedChunks:    int32(len(retrieved)),
			UsedChunks:         0,
			MaxSimilarity:      maxSimilarity,
			Answered:           false,
			IPHash:             hashIP(input.IP),
			UserAgent:          input.UserAgent,
		}); err != nil {
			return AskResult{}, err
		}
		return AskResult{
			Answer:         noSourceAnswer,
			ConversationID: conversation.ID,
			Citations:      nil,
			SourcesUsed:    nil,
			TokenUsage:     &tokenUsage,
		}, nil
	}

	answer, err := s.chat.Answer(ctx, SystemPrompt, BuildUserPrompt(question, used))
	if err != nil {
		return AskResult{}, apperrors.OpenAI("failed to generate answer", err)
	}
	tokenUsage.InputTokens += answer.Usage.InputTokens
	tokenUsage.OutputTokens += answer.Usage.OutputTokens
	tokenUsage.TotalTokens += answer.Usage.TotalTokens
	if err := s.analytics.CreateTokenUsage(ctx, analyticsrepo.TokenUsage{
		ConversationID: conversation.ID,
		RequestID:      input.RequestID,
		Provider:       "openai",
		Model:          answer.Model,
		Operation:      "chat",
		InputTokens:    int32(answer.Usage.InputTokens),
		OutputTokens:   int32(answer.Usage.OutputTokens),
		TotalTokens:    int32(answer.Usage.TotalTokens),
	}); err != nil {
		return AskResult{}, err
	}

	citations := citationsFromChunks(used)
	if err := s.persistExchange(ctx, conversation.ID, question, answer.Text, citations); err != nil {
		return AskResult{}, err
	}
	_ = s.conversations.Touch(ctx, conversation.ID)
	if err := s.analytics.CreateQuestionEvent(ctx, analyticsrepo.QuestionEvent{
		ConversationID:     conversation.ID,
		RequestID:          input.RequestID,
		Question:           question,
		NormalizedQuestion: normalized,
		RetrievedChunks:    int32(len(retrieved)),
		UsedChunks:         int32(len(used)),
		MaxSimilarity:      maxSimilarity,
		Answered:           true,
		IPHash:             hashIP(input.IP),
		UserAgent:          input.UserAgent,
	}); err != nil {
		return AskResult{}, err
	}

	s.setCached(normalized, cacheEntry{
		answer:     answer.Text,
		citations:  citations,
		expiresAt:  time.Now().Add(s.cacheTTL),
		answered:   true,
		maxSim:     maxSimilarity,
		usedChunks: int32(len(used)),
	})

	return AskResult{
		Answer:         answer.Text,
		ConversationID: conversation.ID,
		Citations:      citations,
		SourcesUsed:    citations,
		TokenUsage:     &tokenUsage,
	}, nil
}

func (s *Service) persistExchange(ctx context.Context, conversationID uuid.UUID, question string, answer string, citations []message.Citation) error {
	emptyCitations := json.RawMessage("[]")
	if _, err := s.messages.Create(ctx, conversationID, "user", question, emptyCitations); err != nil {
		return err
	}

	citationJSON := emptyCitations
	if len(citations) > 0 {
		raw, err := json.Marshal(citations)
		if err != nil {
			return err
		}
		citationJSON = raw
	}
	_, err := s.messages.Create(ctx, conversationID, "assistant", answer, citationJSON)
	return err
}

func selectChunks(chunks []domainchunk.SearchResult, threshold float64) ([]domainchunk.SearchResult, float64) {
	used := make([]domainchunk.SearchResult, 0, len(chunks))
	maxSimilarity := 0.0
	for _, chunk := range chunks {
		if chunk.Similarity > maxSimilarity {
			maxSimilarity = chunk.Similarity
		}
		if chunk.Similarity >= threshold {
			used = append(used, chunk)
		}
	}
	if len(used) == 0 && len(chunks) > 0 && maxSimilarity >= fallbackSimilarityThreshold {
		used = append(used, chunks[0])
	}
	return used, maxSimilarity
}

func mergeChunkResults(primary []domainchunk.SearchResult, secondary []domainchunk.SearchResult, limit int) []domainchunk.SearchResult {
	if len(secondary) == 0 {
		return primary
	}
	seen := make(map[uuid.UUID]struct{}, len(primary)+len(secondary))
	merged := make([]domainchunk.SearchResult, 0, len(primary)+len(secondary))
	for _, item := range primary {
		if _, ok := seen[item.ID]; ok {
			continue
		}
		seen[item.ID] = struct{}{}
		merged = append(merged, item)
	}
	for _, item := range secondary {
		if _, ok := seen[item.ID]; ok {
			continue
		}
		seen[item.ID] = struct{}{}
		merged = append(merged, item)
	}
	if limit > 0 && len(merged) > limit {
		return merged[:limit]
	}
	return merged
}

func citationsFromChunks(chunks []domainchunk.SearchResult) []message.Citation {
	seen := make(map[string]struct{}, len(chunks))
	citations := make([]message.Citation, 0, len(chunks))
	for _, chunk := range chunks {
		key := chunk.SourcePath + "#" + strconv.Itoa(int(chunk.ChunkIndex))
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		citations = append(citations, message.Citation{
			SourceTitle: chunk.SourceTitle,
			SourcePath:  chunk.SourcePath,
			ChunkIndex:  chunk.ChunkIndex,
		})
	}
	return citations
}

func normalizeQuestion(question string) string {
	return strings.ToLower(strings.Join(strings.Fields(question), " "))
}

func expandRetrievalQuery(question string) string {
	normalized := normalizeQuestion(question)
	switch {
	case (strings.Contains(normalized, "bruno") && strings.Contains(normalized, "what") && (strings.Contains(normalized, " do") || strings.Contains(normalized, "does"))) ||
		strings.Contains(normalized, "who is bruno"):
		return question + " Bruno Simplicio professional profile senior software engineer backend frontend cloud infrastructure"
	case strings.Contains(normalized, "study") || strings.Contains(normalized, "education") || strings.Contains(normalized, "degree"):
		return question + " education degree university"
	default:
		return question
	}
}

func extractKeywords(question string) []string {
	stopwords := map[string]struct{}{
		"a": {}, "about": {}, "an": {}, "and": {}, "are": {}, "as": {}, "at": {}, "be": {}, "bruno": {}, "did": {}, "do": {}, "does": {}, "for": {}, "from": {}, "has": {}, "have": {}, "he": {}, "his": {}, "in": {}, "is": {}, "know": {}, "of": {}, "on": {}, "or": {}, "the": {}, "to": {}, "what": {}, "where": {}, "which": {}, "with": {}, "worked": {},
	}
	fields := strings.FieldsFunc(strings.ToLower(question), func(r rune) bool {
		return !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9')
	})
	keywords := make([]string, 0, len(fields))
	seen := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if len(field) < 3 && field != "ai" {
			continue
		}
		if _, ok := stopwords[field]; ok {
			continue
		}
		if _, ok := seen[field]; ok {
			continue
		}
		seen[field] = struct{}{}
		keywords = append(keywords, field)
	}
	return keywords
}

func hashIP(ip string) string {
	if strings.TrimSpace(ip) == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(ip))
	return hex.EncodeToString(sum[:])
}

func (s *Service) getCached(key string) (cacheEntry, bool) {
	if s.cacheTTL <= 0 {
		return cacheEntry{}, false
	}
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	entry, ok := s.cache[key]
	if !ok {
		return cacheEntry{}, false
	}
	if time.Now().After(entry.expiresAt) {
		delete(s.cache, key)
		return cacheEntry{}, false
	}
	return entry, true
}

func (s *Service) setCached(key string, entry cacheEntry) {
	if s.cacheTTL <= 0 {
		return
	}
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	s.cache[key] = entry
}
