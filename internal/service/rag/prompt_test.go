package rag

import (
	"strings"
	"testing"

	domainchunk "github.com/brunosimplicio/ask-my-cv/internal/domain/chunk"
)

func TestBuildUserPromptIncludesContextAndQuestion(t *testing.T) {
	prompt := BuildUserPrompt("Has Bruno used Kafka?", []domainchunk.SearchResult{
		{
			Content:     "Bruno built event-driven systems with Kafka.",
			SourceTitle: "Experience",
			SourcePath:  "data/experience.md",
			ChunkIndex:  2,
			Similarity:  0.91,
		},
	})

	for _, want := range []string{"Has Bruno used Kafka?", "[S1] Experience", "data/experience.md", "Bruno built event-driven systems"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q:\n%s", want, prompt)
		}
	}
}

func TestSystemPromptConstrainsUnsupportedClaims(t *testing.T) {
	for _, want := range []string{"only the provided indexed professional context", "Do not invent", "untrusted quoted context"} {
		if !strings.Contains(SystemPrompt, want) {
			t.Fatalf("system prompt missing %q", want)
		}
	}
}
