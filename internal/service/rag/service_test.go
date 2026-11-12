package rag

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	db "github.com/brunosimplicio/ask-my-cv/internal/db/sqlc"
	domainchunk "github.com/brunosimplicio/ask-my-cv/internal/domain/chunk"
	analyticsrepo "github.com/brunosimplicio/ask-my-cv/internal/repository/analytics"
	"github.com/brunosimplicio/ask-my-cv/internal/service/chat"
	"github.com/brunosimplicio/ask-my-cv/internal/service/embedding"
	"github.com/brunosimplicio/ask-my-cv/lib/openai"
	"github.com/google/uuid"
)

func TestAskNoSourceDoesNotCallChat(t *testing.T) {
	conversationID := uuid.New()
	analytics := &fakeAnalytics{}
	messages := &fakeMessages{}
	chatService := &fakeChat{}

	service := New(
		fakeEmbedder{},
		fakeChunkSearcher{results: []domainchunk.SearchResult{
			{Similarity: 0.5, SourceTitle: "Resume", SourcePath: "data/resume.md"},
		}},
		fakeConversations{id: conversationID},
		messages,
		analytics,
		chatService,
		6,
		0.72,
		0,
	)

	result, err := service.Ask(context.Background(), AskInput{
		Question:  "What fintech experience does Bruno have?",
		RequestID: "req-1",
		IP:        "127.0.0.1",
		UserAgent: "test",
	})
	if err != nil {
		t.Fatalf("Ask() error = %v", err)
	}
	if result.Answer != noSourceAnswer {
		t.Fatalf("Answer = %q, want no-source answer", result.Answer)
	}
	if chatService.called {
		t.Fatal("chat service should not be called when no chunks pass threshold")
	}
	if len(messages.roles) != 2 || messages.roles[0] != "user" || messages.roles[1] != "assistant" {
		t.Fatalf("messages roles = %#v, want user and assistant", messages.roles)
	}
	if len(analytics.questionEvents) != 1 {
		t.Fatalf("question events = %d, want 1", len(analytics.questionEvents))
	}
	if analytics.questionEvents[0].Answered {
		t.Fatal("question event should be marked unanswered")
	}
	if len(analytics.tokenUsages) != 1 || analytics.tokenUsages[0].Operation != "embedding" {
		t.Fatalf("token usages = %#v, want one embedding usage", analytics.tokenUsages)
	}
	if result.TokenUsage == nil || result.TokenUsage.TotalTokens != 7 {
		t.Fatalf("token usage = %#v, want total 7", result.TokenUsage)
	}
}

func TestAskNormalizesDashesInAnswer(t *testing.T) {
	messages := &fakeMessages{}
	chatService := &fakeChat{text: "Bruno led the WorkOS migration — enabling SSO and SCIM."}

	service := New(
		fakeEmbedder{},
		fakeChunkSearcher{results: []domainchunk.SearchResult{
			{Similarity: 0.9, SourceTitle: "Resume", SourcePath: "data/resume.md", Content: "WorkOS migration"},
		}},
		fakeConversations{id: uuid.New()},
		messages,
		&fakeAnalytics{},
		chatService,
		6,
		0.72,
		time.Hour,
	)

	input := AskInput{Question: "Did Bruno work with WorkOS?", RequestID: "req-1"}
	want := "Bruno led the WorkOS migration, enabling SSO and SCIM."

	for attempt := 1; attempt <= 2; attempt++ {
		result, err := service.Ask(context.Background(), input)
		if err != nil {
			t.Fatalf("Ask() attempt %d error = %v", attempt, err)
		}
		if result.Answer != want {
			t.Fatalf("Answer attempt %d = %q, want %q", attempt, result.Answer, want)
		}
	}
	if got := messages.contents[1]; got != want {
		t.Fatalf("persisted answer = %q, want %q", got, want)
	}
}

func TestSelectChunksUsesBestChunkForBroadQuestions(t *testing.T) {
	used, maxSimilarity := selectChunks([]domainchunk.SearchResult{
		{Similarity: 0.55, SourceTitle: "Bio", SourcePath: "data/bio.md"},
		{Similarity: 0.50, SourceTitle: "Resume", SourcePath: "data/resume.md"},
	}, 0.60)

	if maxSimilarity != 0.55 {
		t.Fatalf("maxSimilarity = %v, want 0.55", maxSimilarity)
	}
	if len(used) != 1 || used[0].SourcePath != "data/bio.md" {
		t.Fatalf("used = %#v, want best fallback chunk", used)
	}
}

func TestSelectChunksRejectsVeryWeakMatches(t *testing.T) {
	used, maxSimilarity := selectChunks([]domainchunk.SearchResult{
		{Similarity: 0.40, SourceTitle: "Bio", SourcePath: "data/bio.md"},
	}, 0.60)

	if maxSimilarity != 0.40 {
		t.Fatalf("maxSimilarity = %v, want 0.40", maxSimilarity)
	}
	if len(used) != 0 {
		t.Fatalf("used = %#v, want none", used)
	}
}

func TestMergeChunkResultsPrioritizesKeywordResults(t *testing.T) {
	keywordID := uuid.New()
	semanticID := uuid.New()
	merged := mergeChunkResults([]domainchunk.SearchResult{
		{ID: keywordID, SourcePath: "data/skills.md"},
	}, []domainchunk.SearchResult{
		{ID: semanticID, SourcePath: "data/resume.md"},
	}, 1)

	if len(merged) != 1 || merged[0].ID != keywordID {
		t.Fatalf("merged = %#v, want keyword result first", merged)
	}
}

func TestExpandRetrievalQueryForBroadProfileQuestion(t *testing.T) {
	got := expandRetrievalQuery("What does Bruno do?")
	if !strings.Contains(got, "professional profile senior software engineer") {
		t.Fatalf("expanded query = %q, want professional profile terms", got)
	}
}

func TestExtractKeywordsKeepsAI(t *testing.T) {
	got := extractKeywords("Does Bruno have AI or RAG experience?")
	want := map[string]bool{"ai": true, "rag": true, "experience": true}
	for _, keyword := range got {
		delete(want, keyword)
	}
	if len(want) != 0 {
		t.Fatalf("extractKeywords missing %#v from %#v", want, got)
	}
}

type fakeEmbedder struct{}

func (fakeEmbedder) Model() string { return "text-embedding-3-small" }

func (fakeEmbedder) Embed(context.Context, string) (embedding.Result, error) {
	return embedding.Result{
		Model:   "text-embedding-3-small",
		Literal: "[0.1,0.2,0.3]",
		Usage:   openai.Usage{InputTokens: 7, TotalTokens: 7},
	}, nil
}

type fakeChunkSearcher struct {
	results []domainchunk.SearchResult
}

func (f fakeChunkSearcher) Search(context.Context, string, string, int32) ([]domainchunk.SearchResult, error) {
	return f.results, nil
}

func (f fakeChunkSearcher) SearchByKeywords(context.Context, []string, string, int32) ([]domainchunk.SearchResult, error) {
	return nil, nil
}

type fakeConversations struct {
	id uuid.UUID
}

func (f fakeConversations) Ensure(context.Context, *uuid.UUID) (db.Conversation, error) {
	return db.Conversation{ID: f.id}, nil
}

func (fakeConversations) Touch(context.Context, uuid.UUID) error { return nil }

type fakeMessages struct {
	roles    []string
	contents []string
}

func (f *fakeMessages) Create(_ context.Context, _ uuid.UUID, role string, content string, _ json.RawMessage) (db.Message, error) {
	f.roles = append(f.roles, role)
	f.contents = append(f.contents, content)
	return db.Message{Role: role}, nil
}

type fakeAnalytics struct {
	questionEvents []analyticsrepo.QuestionEvent
	tokenUsages    []analyticsrepo.TokenUsage
}

func (f *fakeAnalytics) CreateQuestionEvent(_ context.Context, event analyticsrepo.QuestionEvent) error {
	f.questionEvents = append(f.questionEvents, event)
	return nil
}

func (f *fakeAnalytics) CreateTokenUsage(_ context.Context, usage analyticsrepo.TokenUsage) error {
	f.tokenUsages = append(f.tokenUsages, usage)
	return nil
}

type fakeChat struct {
	called bool
	text   string
}

func (f *fakeChat) Model() string { return "gpt-6-luna" }

func (f *fakeChat) Answer(context.Context, string, string) (chat.Result, error) {
	f.called = true
	text := f.text
	if text == "" {
		text = "answer"
	}
	return chat.Result{Text: text}, nil
}
