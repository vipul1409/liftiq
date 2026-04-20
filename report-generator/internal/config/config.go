package config

import (
	"fmt"
	"log/slog"
	"os"
)

// Config holds runtime configuration for the report generator.
type Config struct {
	HTTPPort string
	LogLevel slog.Level
}

// Load reads configuration from environment variables.
func Load() (Config, error) {
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8082"
	}

	level := slog.LevelInfo
	if s := os.Getenv("LOG_LEVEL"); s != "" {
		if err := level.UnmarshalText([]byte(s)); err != nil {
			return Config{}, fmt.Errorf("invalid LOG_LEVEL %q: %w", s, err)
		}
	}

	return Config{HTTPPort: port, LogLevel: level}, nil
}
