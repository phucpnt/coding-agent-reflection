package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port                string
	DBPath              string
	RetentionDays       int
	ReflectionCLI       string
	ReflectionSchedule  string
	ReflectionOutputDir string
	ReflectionPrompt    string
}

func Load() Config {
	return Config{
		Port:               envOr("COLLECTOR_PORT", "19321"),
		DBPath:             envOr("COLLECTOR_DB_PATH", "./data/ai_interactions.db"),
		RetentionDays:      envOrInt("COLLECTOR_RETENTION_DAYS", 0),
		ReflectionCLI:      envOr("COLLECTOR_REFLECTION_CLI", "claude --print"),
		ReflectionSchedule: envOr("COLLECTOR_REFLECTION_SCHEDULE", "daily"),
		ReflectionOutputDir: envOr("COLLECTOR_REFLECTION_DIR", "./data/reflections"),
		ReflectionPrompt:    envOr("COLLECTOR_REFLECTION_PROMPT", ""),
	}
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
