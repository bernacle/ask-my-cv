package config

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv                string
	HTTPAddr              string
	DatabaseURL           string
	OpenAIAPIKey          string
	OpenAIBaseURL         string
	OpenAIEmbeddingModel  string
	OpenAIChatModel       string
	OpenAIReasoningEffort string
	OpenAITextVerbosity   string
	CORSAllowedOrigins    []string
	MaxQuestionChars      int
	MaxRetrievedChunks    int
	SimilarityThreshold   float64
	MaxOutputTokens       int
	RateLimitPerMinute    int
	RateLimitPerDay       int
	CacheTTL              time.Duration
}

func Load() (Config, error) {
	loadDotEnv(".env")

	cfg := Config{
		AppEnv:                getEnv("APP_ENV", "development"),
		HTTPAddr:              getHTTPAddr(),
		DatabaseURL:           getEnv("DATABASE_URL", "postgres://ask_my_cv:ask_my_cv@localhost:5432/ask_my_cv?sslmode=disable"),
		OpenAIAPIKey:          os.Getenv("OPENAI_API_KEY"),
		OpenAIBaseURL:         strings.TrimRight(getEnv("OPENAI_BASE_URL", "https://api.openai.com"), "/"),
		OpenAIEmbeddingModel:  getEnv("OPENAI_EMBEDDING_MODEL", "text-embedding-3-small"),
		OpenAIChatModel:       getEnv("OPENAI_CHAT_MODEL", "gpt-6-luna"),
		OpenAIReasoningEffort: getEnv("OPENAI_REASONING_EFFORT", "low"),
		OpenAITextVerbosity:   getEnv("OPENAI_TEXT_VERBOSITY", "low"),
		CORSAllowedOrigins:    splitCSV(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:3001,https://brunosimplicio.com")),
		MaxQuestionChars:      getEnvInt("MAX_QUESTION_CHARS", 1000),
		MaxRetrievedChunks:    getEnvInt("MAX_RETRIEVED_CHUNKS", 6),
		SimilarityThreshold:   getEnvFloat("SIMILARITY_THRESHOLD", 0.60),
		MaxOutputTokens:       getEnvInt("MAX_OUTPUT_TOKENS", 500),
		RateLimitPerMinute:    getEnvInt("RATE_LIMIT_PER_MINUTE", 10),
		RateLimitPerDay:       getEnvInt("RATE_LIMIT_PER_DAY", 100),
		CacheTTL:              getEnvDuration("CACHE_TTL", 24*time.Hour),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	return cfg, nil
}

func (c Config) ValidateOpenAI() error {
	if strings.TrimSpace(c.OpenAIAPIKey) == "" {
		return errors.New("OPENAI_API_KEY is required")
	}
	return nil
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getHTTPAddr() string {
	if value := strings.TrimSpace(os.Getenv("HTTP_ADDR")); value != "" {
		return value
	}
	if port := strings.TrimSpace(os.Getenv("PORT")); port != "" {
		return ":" + strings.TrimPrefix(port, ":")
	}
	return ":8080"
}

func getEnvInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func getEnvFloat(key string, fallback float64) float64 {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return fallback
	}
	return value
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}
	return value
}

func splitCSV(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value != "" {
			values = append(values, value)
		}
	}
	return values
}

func loadDotEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if key == "" || os.Getenv(key) != "" {
			continue
		}
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		_ = os.Setenv(key, value)
	}
}
