package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"
)

// Config holds all runtime configuration for the compliance engine.
type Config struct {
	DatabaseURL   string
	HTTPPort      string
	StaleWindow   time.Duration // metrics older than this are treated as unknown
	LogLevel      slog.Level
}

// Load reads configuration from environment variables.
// DATABASE_URL is required; all other fields have defaults.
func Load() (Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}

	staleMinutes := 10.0
	if s := os.Getenv("STALE_WINDOW_MINUTES"); s != "" {
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return Config{}, fmt.Errorf("invalid STALE_WINDOW_MINUTES %q: %w", s, err)
		}
		staleMinutes = v
	}

	level := slog.LevelInfo
	if s := os.Getenv("LOG_LEVEL"); s != "" {
		if err := level.UnmarshalText([]byte(s)); err != nil {
			return Config{}, fmt.Errorf("invalid LOG_LEVEL %q: %w", s, err)
		}
	}

	return Config{
		DatabaseURL: dbURL,
		HTTPPort:    port,
		StaleWindow: time.Duration(staleMinutes * float64(time.Minute)),
		LogLevel:    level,
	}, nil
}
