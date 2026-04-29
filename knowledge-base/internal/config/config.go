package config

import (
	"fmt"
	"log/slog"
	"os"
)

// Config holds all configuration for the knowledge-base service.
type Config struct {
	DatabaseURL     string
	HTTPPort        string
	OllamaURL       string
	OllamaEmbedModel string
	OllamaLLMModel  string
	LogLevel        slog.Level
}

// Load reads configuration from environment variables.
func Load() (Config, error) {
	cfg := Config{
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		HTTPPort:        "8085",
		OllamaURL:       "http://localhost:11434",
		OllamaEmbedModel: "nomic-embed-text",
		OllamaLLMModel:  "llama3.2",
		LogLevel:        slog.LevelInfo,
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	if s := os.Getenv("HTTP_PORT"); s != "" {
		cfg.HTTPPort = s
	}
	if s := os.Getenv("OLLAMA_URL"); s != "" {
		cfg.OllamaURL = s
	}
	if s := os.Getenv("OLLAMA_EMBED_MODEL"); s != "" {
		cfg.OllamaEmbedModel = s
	}
	if s := os.Getenv("OLLAMA_LLM_MODEL"); s != "" {
		cfg.OllamaLLMModel = s
	}
	if s := os.Getenv("LOG_LEVEL"); s != "" {
		if err := cfg.LogLevel.UnmarshalText([]byte(s)); err != nil {
			return Config{}, fmt.Errorf("invalid LOG_LEVEL %q: %w", s, err)
		}
	}

	return cfg, nil
}
