package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/webb1es/foosballcup/internal/config"
	"github.com/webb1es/foosballcup/internal/seed"
	"github.com/webb1es/foosballcup/internal/server"
	"github.com/webb1es/foosballcup/internal/store"
)

//go:embed seed/*.csv
var seedFiles embed.FS

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := store.Open(ctx, cfg.DatabasePath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	if err := importSeed(ctx, logger, db); err != nil {
		return err
	}

	return server.New(cfg.Addr(), logger, db.PingContext).Run(ctx)
}

func importSeed(ctx context.Context, logger *slog.Logger, db *sql.DB) error {
	files, err := fs.Sub(seedFiles, "seed")
	if err != nil {
		return fmt.Errorf("open embedded seed: %w", err)
	}
	data, err := seed.Load(files)
	if err != nil {
		return fmt.Errorf("load seed: %w", err)
	}
	imported, err := seed.ImportIfEmpty(ctx, db, data, time.Now())
	if err != nil {
		return fmt.Errorf("import seed: %w", err)
	}
	if imported {
		logger.Info("seed imported", "players", len(data.Players), "matches", len(data.Matches))
	}
	return nil
}
