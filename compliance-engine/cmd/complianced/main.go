package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/liftiq/compliance-engine/internal/api"
	"github.com/liftiq/compliance-engine/internal/config"
	"github.com/liftiq/compliance-engine/internal/store"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "complianced: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Info("connecting to TimescaleDB", "url", cfg.DatabaseURL)
	pool, err := store.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer pool.Close()
	log.Info("database connection established")

	st := store.NewPGXStore(pool)
	handler := api.New(st, cfg.StaleWindow, log)

	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("compliance engine listening", "port", cfg.HTTPPort,
			"stale_window_minutes", cfg.StaleWindow.Minutes())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")
	case err := <-errCh:
		return fmt.Errorf("http server: %w", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	log.Info("complianced stopped")
	return nil
}
