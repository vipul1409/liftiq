package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"
)

// Config holds all runtime configuration loaded from environment variables.
type Config struct {
	// SimulatorURL is the base URL of the elevator simulator HTTP API.
	// Env: SIMULATOR_URL (default: http://localhost:8000)
	SimulatorURL string

	// DatabaseURL is the PostgreSQL DSN for TimescaleDB.
	// Env: DATABASE_URL (required)
	DatabaseURL string

	// PollInterval is how often to fetch snapshots from the simulator.
	// Env: POLL_INTERVAL_SECONDS (default: 5)
	PollInterval time.Duration

	// LogLevel controls structured log verbosity.
	// Env: LOG_LEVEL — "debug" | "info" | "warn" | "error" (default: info)
	LogLevel slog.Level
}

// Load reads configuration from environment variables.
// Returns an error if any required variable is missing or any value is invalid.
func Load() (Config, error) {
	cfg := Config{
		SimulatorURL: getEnv("SIMULATOR_URL", "http://localhost:8000"),
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		PollInterval: 5 * time.Second,
		LogLevel:     slog.LevelInfo,
	}

	if cfg.DatabaseURL == "" {
		return cfg, fmt.Errorf("DATABASE_URL is required")
	}

	if v := os.Getenv("POLL_INTERVAL_SECONDS"); v != "" {
		secs, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return cfg, fmt.Errorf("invalid POLL_INTERVAL_SECONDS %q: %w", v, err)
		}
		if secs <= 0 {
			return cfg, fmt.Errorf("POLL_INTERVAL_SECONDS must be > 0, got %g", secs)
		}
		cfg.PollInterval = time.Duration(secs * float64(time.Second))
	}

	if v := os.Getenv("LOG_LEVEL"); v != "" {
		lvl, err := parseLogLevel(v)
		if err != nil {
			return cfg, err
		}
		cfg.LogLevel = lvl
	}

	return cfg, nil
}

func parseLogLevel(s string) (slog.Level, error) {
	switch s {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("invalid LOG_LEVEL %q: must be debug/info/warn/error", s)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
