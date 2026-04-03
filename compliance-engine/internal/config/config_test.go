package config_test

import (
	"log/slog"
	"testing"
	"time"

	"github.com/liftiq/compliance-engine/internal/config"
)

func setenv(t *testing.T, key, val string) {
	t.Helper()
	t.Setenv(key, val)
}

func TestLoad_RequiresDatabaseURL(t *testing.T) {
	// Ensure DATABASE_URL is absent for this test.
	t.Setenv("DATABASE_URL", "")
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when DATABASE_URL is empty, got nil")
	}
}

func TestLoad_Defaults(t *testing.T) {
	setenv(t, "DATABASE_URL", "postgres://x:y@localhost/db")
	t.Setenv("HTTP_PORT", "")
	t.Setenv("STALE_WINDOW_MINUTES", "")
	t.Setenv("LOG_LEVEL", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.HTTPPort != "8080" {
		t.Errorf("HTTPPort: got %q, want %q", cfg.HTTPPort, "8080")
	}
	if cfg.StaleWindow != 10*time.Minute {
		t.Errorf("StaleWindow: got %v, want %v", cfg.StaleWindow, 10*time.Minute)
	}
	if cfg.LogLevel != slog.LevelInfo {
		t.Errorf("LogLevel: got %v, want %v", cfg.LogLevel, slog.LevelInfo)
	}
}

func TestLoad_CustomHTTPPort(t *testing.T) {
	setenv(t, "DATABASE_URL", "postgres://x:y@localhost/db")
	setenv(t, "HTTP_PORT", "9090")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.HTTPPort != "9090" {
		t.Errorf("HTTPPort: got %q, want %q", cfg.HTTPPort, "9090")
	}
}

func TestLoad_CustomStaleWindow(t *testing.T) {
	setenv(t, "DATABASE_URL", "postgres://x:y@localhost/db")
	setenv(t, "STALE_WINDOW_MINUTES", "5")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.StaleWindow != 5*time.Minute {
		t.Errorf("StaleWindow: got %v, want %v", cfg.StaleWindow, 5*time.Minute)
	}
}

func TestLoad_InvalidStaleWindow(t *testing.T) {
	setenv(t, "DATABASE_URL", "postgres://x:y@localhost/db")
	setenv(t, "STALE_WINDOW_MINUTES", "not-a-number")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for invalid STALE_WINDOW_MINUTES, got nil")
	}
}

func TestLoad_CustomLogLevel(t *testing.T) {
	setenv(t, "DATABASE_URL", "postgres://x:y@localhost/db")
	setenv(t, "LOG_LEVEL", "debug")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.LogLevel != slog.LevelDebug {
		t.Errorf("LogLevel: got %v, want %v", cfg.LogLevel, slog.LevelDebug)
	}
}

func TestLoad_InvalidLogLevel(t *testing.T) {
	setenv(t, "DATABASE_URL", "postgres://x:y@localhost/db")
	setenv(t, "LOG_LEVEL", "VERBOSE")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for invalid LOG_LEVEL, got nil")
	}
}

func TestLoad_DatabaseURLPreserved(t *testing.T) {
	dsn := "postgres://liftiq:liftiq@localhost:5432/liftiq?sslmode=disable"
	setenv(t, "DATABASE_URL", dsn)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DatabaseURL != dsn {
		t.Errorf("DatabaseURL: got %q, want %q", cfg.DatabaseURL, dsn)
	}
}
