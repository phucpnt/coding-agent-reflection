package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port          string
	DBPath        string
	RetentionDays int
	LLMAPIKey     string
	LLMModel      string
}

func Load() Config {
	c := Config{
		Port:          envOr("COLLECTOR_PORT", "9000"),
		DBPath:        envOr("COLLECTOR_DB_PATH", "./data/ai_interactions.db"),
		RetentionDays: envOrInt("COLLECTOR_RETENTION_DAYS", 0),
		LLMAPIKey:     os.Getenv("ANTHROPIC_API_KEY"),
		LLMModel:      envOr("COLLECTOR_LLM_MODEL", "claude-sonnet-4-20250514"),
	}
	return c
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envOrInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
