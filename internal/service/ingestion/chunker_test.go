package ingestion

import (
	"strconv"
	"strings"
	"testing"
)

func TestParseMarkdownUsesFrontMatterTitle(t *testing.T) {
	parsed := ParseMarkdown("data/resume.md", "---\ntitle: Resume\n---\n\n# Ignored\nBody")
	if parsed.Title != "Resume" {
		t.Fatalf("Title = %q, want Resume", parsed.Title)
	}
	if strings.Contains(parsed.Body, "title:") {
		t.Fatalf("front matter should be removed from body: %q", parsed.Body)
	}
}

func TestChunkMarkdownSplitsLongSectionsWithHeading(t *testing.T) {
	words := make([]string, 0, 25)
	for i := 0; i < 25; i++ {
		words = append(words, "word"+strconv.Itoa(i))
	}
	body := "# Backend\n\n" + strings.Join(words, " ")

	chunks := chunkMarkdown(body, 10, 2)
	if len(chunks) != 4 {
		t.Fatalf("len(chunks) = %d, want 4", len(chunks))
	}
	for _, chunk := range chunks {
		if chunk.Heading != "Backend" {
			t.Fatalf("chunk heading = %q, want Backend", chunk.Heading)
		}
		if !strings.Contains(chunk.Content, "Backend") {
			t.Fatalf("chunk content should preserve heading: %q", chunk.Content)
		}
	}
}
