package cmd

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	askcontroller "github.com/brunosimplicio/ask-my-cv/api/controller/ask"
	"github.com/brunosimplicio/ask-my-cv/api/controller/health"
	sourcescontroller "github.com/brunosimplicio/ask-my-cv/api/controller/sources"
	"github.com/brunosimplicio/ask-my-cv/api/route"
	"github.com/brunosimplicio/ask-my-cv/internal/config"
	"github.com/brunosimplicio/ask-my-cv/internal/db/postgres"
	analyticsrepo "github.com/brunosimplicio/ask-my-cv/internal/repository/analytics"
	chunkrepo "github.com/brunosimplicio/ask-my-cv/internal/repository/chunk"
	conversationrepo "github.com/brunosimplicio/ask-my-cv/internal/repository/conversation"
	documentrepo "github.com/brunosimplicio/ask-my-cv/internal/repository/document"
	messagerepo "github.com/brunosimplicio/ask-my-cv/internal/repository/message"
	"github.com/brunosimplicio/ask-my-cv/internal/service/chat"
	"github.com/brunosimplicio/ask-my-cv/internal/service/embedding"
	"github.com/brunosimplicio/ask-my-cv/internal/service/rag"
	"github.com/brunosimplicio/ask-my-cv/lib/openai"
	"github.com/brunosimplicio/ask-my-cv/pkg/logger"
)

func runServeAPI(_ []string) int {
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
	embeddingService := embedding.New(openaiClient, cfg.OpenAIEmbeddingModel)
	chatService := chat.New(openaiClient, cfg.OpenAIChatModel, cfg.OpenAIReasoningEffort, cfg.OpenAITextVerbosity, cfg.MaxOutputTokens)

	documentRepository := documentrepo.New(pool)
	ragService := rag.New(
		embeddingService,
		chunkrepo.New(pool),
		conversationrepo.New(pool),
		messagerepo.New(pool),
		analyticsrepo.New(pool),
		chatService,
		cfg.MaxRetrievedChunks,
		cfg.SimilarityThreshold,
		cfg.CacheTTL,
	)

	handler := route.New(cfg, log, route.Controllers{
		Ask:     askcontroller.New(ragService, cfg.MaxQuestionChars),
		Health:  health.New(),
		Sources: sourcescontroller.New(documentRepository),
	})

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("starting api", "addr", cfg.HTTPAddr)
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Error("shutdown api", "error", err)
			return 1
		}
		log.Info("api stopped")
		return 0
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return 0
		}
		log.Error("api failed", "error", err)
		return 1
	}
}
