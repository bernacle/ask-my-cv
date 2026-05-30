package cmd

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/brunosimplicio/ask-my-cv/internal/config"
	"github.com/brunosimplicio/ask-my-cv/internal/db/postgres"
	documentrepo "github.com/brunosimplicio/ask-my-cv/internal/repository/document"
	"github.com/brunosimplicio/ask-my-cv/internal/service/embedding"
	"github.com/brunosimplicio/ask-my-cv/internal/service/ingestion"
	"github.com/brunosimplicio/ask-my-cv/lib/openai"
	"github.com/brunosimplicio/ask-my-cv/pkg/logger"
)

func runIngest(args []string) int {
	dataDir := "./data"
	if len(args) > 0 {
		dataDir = args[0]
	}

	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		return 1
	}
	log := logger.New(cfg.AppEnv)
	if err := cfg.ValidateOpenAI(); err != nil {
		log.Error("invalid OpenAI config", "error", err)
		return 1
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("connect postgres", "error", err)
		return 1
	}
	defer pool.Close()

	openaiClient := openai.NewClient(cfg.OpenAIAPIKey, cfg.OpenAIBaseURL)
	service := ingestion.New(
		documentrepo.New(pool),
		embedding.New(openaiClient, cfg.OpenAIEmbeddingModel),
	)

	summary, err := service.IngestDirectory(ctx, dataDir)
	if err != nil {
		log.Error("ingestion failed", "error", err)
		return 1
	}

	log.Info("ingestion complete",
		"files_seen", summary.FilesSeen,
		"files_skipped", summary.FilesSkipped,
		"files_ingested", summary.FilesIngested,
		"chunks_created", summary.ChunksCreated,
	)
	return 0
}
