package ingestion

import (
	"path/filepath"
	"strings"
)

const (
	defaultMaxWords     = 520
	defaultOverlapWords = 80
)

type ParsedMarkdown struct {
	Title string
	Body  string
}

type MarkdownChunk struct {
	Content   string
	Heading   string
	WordCount int
}

func ParseMarkdown(path string, raw string) ParsedMarkdown {
	body := strings.TrimSpace(raw)
	title := ""

	if strings.HasPrefix(body, "---") {
		if end := strings.Index(body[3:], "\n---"); end >= 0 {
			frontMatter := body[3 : end+3]
			body = strings.TrimSpace(body[end+7:])
			for _, line := range strings.Split(frontMatter, "\n") {
				key, value, ok := strings.Cut(line, ":")
				if !ok || strings.TrimSpace(strings.ToLower(key)) != "title" {
					continue
				}
				title = strings.Trim(strings.TrimSpace(value), `"'`)
				break
			}
		}
	}

	if title == "" {
		for _, line := range strings.Split(body, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "# ") {
				title = strings.TrimSpace(strings.TrimPrefix(line, "# "))
				break
			}
		}
	}
	if title == "" {
		base := filepath.Base(path)
		title = strings.TrimSuffix(base, filepath.Ext(base))
		title = strings.ReplaceAll(title, "-", " ")
	}

	return ParsedMarkdown{Title: title, Body: body}
}

func ChunkMarkdown(body string) []MarkdownChunk {
	return chunkMarkdown(body, defaultMaxWords, defaultOverlapWords)
}

func chunkMarkdown(body string, maxWords int, overlapWords int) []MarkdownChunk {
	body = strings.TrimSpace(body)
	if body == "" {
		return nil
	}
	if maxWords <= 0 {
		maxWords = defaultMaxWords
	}
	if overlapWords < 0 {
		overlapWords = 0
	}
	if overlapWords >= maxWords {
		overlapWords = maxWords / 5
	}

	sections := splitSections(body)
	var chunks []MarkdownChunk
	for _, section := range sections {
		words := strings.Fields(section.content)
		if len(words) == 0 {
			continue
		}
		if len(words) <= maxWords {
			chunks = append(chunks, MarkdownChunk{
				Content:   strings.TrimSpace(section.content),
				Heading:   section.heading,
				WordCount: len(words),
			})
			continue
		}

		step := maxWords - overlapWords
		for start := 0; start < len(words); start += step {
			end := start + maxWords
			if end > len(words) {
				end = len(words)
			}
			content := strings.Join(words[start:end], " ")
			if section.heading != "" && !strings.HasPrefix(content, section.heading) {
				content = section.heading + "\n\n" + content
			}
			chunks = append(chunks, MarkdownChunk{
				Content:   strings.TrimSpace(content),
				Heading:   section.heading,
				WordCount: end - start,
			})
			if end == len(words) {
				break
			}
		}
	}
	return chunks
}

type markdownSection struct {
	heading string
	content string
}

func splitSections(body string) []markdownSection {
	var sections []markdownSection
	var currentHeading string
	var current []string

	flush := func() {
		content := strings.TrimSpace(strings.Join(current, "\n"))
		if content != "" {
			sections = append(sections, markdownSection{
				heading: currentHeading,
				content: content,
			})
		}
		current = nil
	}

	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") && strings.Contains(trimmed, " ") {
			flush()
			currentHeading = strings.TrimSpace(strings.TrimLeft(trimmed, "#"))
		}
		current = append(current, line)
	}
	flush()

	if len(sections) == 0 {
		return []markdownSection{{content: body}}
	}
	return sections
}
