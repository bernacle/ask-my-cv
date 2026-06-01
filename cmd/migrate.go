package cmd

import (
	"database/sql"
	"log/slog"

	"github.com/brunosimplicio/ask-my-cv/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

const migrationsDir = "internal/db/migrations"

func runMigrateUp(_ []string) int {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		return 1
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		slog.Error("open postgres", "error", err)
		return 1
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		slog.Error("set goose dialect", "error", err)
		return 1
	}
	if err := goose.Up(db, migrationsDir); err != nil {
		slog.Error("migrate up", "error", err)
		return 1
	}

	slog.Info("migrations complete")
	return 0
}
