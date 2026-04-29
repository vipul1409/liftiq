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

	"github.com/liftiq/knowledge-base/internal/api"
	"github.com/liftiq/knowledge-base/internal/config"
	"github.com/liftiq/knowledge-base/internal/embed"
	"github.com/liftiq/knowledge-base/internal/llm"
	"github.com/liftiq/knowledge-base/internal/store"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "knowledged: %v\n", err)
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

	if err := store.Migrate(ctx, pool); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	log.Info("database connection established, migrations applied")

	st := store.NewPGXStore(pool)
	embedder := embed.NewOllamaEmbedder(cfg.OllamaURL, cfg.OllamaEmbedModel)
	generator := llm.NewOllamaGenerator(cfg.OllamaURL, cfg.OllamaLLMModel)

	handler := api.New(st, embedder, generator, log)

	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 120 * time.Second, // LLM generation can be slow
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("knowledge base listening",
			"port", cfg.HTTPPort,
			"ollama_url", cfg.OllamaURL,
			"embed_model", cfg.OllamaEmbedModel,
			"llm_model", cfg.OllamaLLMModel)
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
	log.Info("knowledged stopped")
	return nil
}
