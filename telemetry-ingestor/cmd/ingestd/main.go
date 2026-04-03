package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/liftiq/telemetry-ingestor/internal/config"
	"github.com/liftiq/telemetry-ingestor/internal/ingest"
	"github.com/liftiq/telemetry-ingestor/internal/simulator"
	"github.com/liftiq/telemetry-ingestor/internal/store"
)

func main() {
	// Bootstrap logger at INFO until we know the configured level.
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		log.Error("config error", "err", err)
		os.Exit(1)
	}

	// Reconfigure logger at the requested level.
	log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))
	slog.SetDefault(log)

	log.Info("starting telemetry ingestor",
		"simulator_url", cfg.SimulatorURL,
		"poll_interval", cfg.PollInterval,
	)

	// Graceful shutdown on SIGINT / SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Connect to TimescaleDB.
	pool, err := store.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("database connect failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Apply schema migrations (idempotent — safe on every start).
	if err := store.Migrate(ctx, pool); err != nil {
		log.Error("migration failed", "err", err)
		os.Exit(1)
	}
	log.Info("schema migrations applied")

	// Wire dependencies.
	simClient := simulator.NewHTTPClient(cfg.SimulatorURL)
	tsStore := store.NewPGXStore(pool)
	poller := ingest.New(simClient, tsStore, cfg.PollInterval, log)

	// Run until signal.
	if err := poller.Run(ctx); err != nil {
		log.Error("poller exited with error", "err", err)
		os.Exit(1)
	}

	log.Info("shutdown complete")
}
