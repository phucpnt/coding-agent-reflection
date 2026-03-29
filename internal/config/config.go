package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Reflection struct {
	CLI       string `yaml:"cli"`
	Schedule  string `yaml:"schedule"`
	OutputDir string `yaml:"output_dir"`
	Prompt    string `yaml:"prompt"`
}

type Config struct {
	Port          int        `yaml:"port"`
	DBPath        string     `yaml:"db_path"`
	RetentionDays int        `yaml:"retention_days"`
	Reflection    Reflection `yaml:"reflection"`
}

func Defaults() Config {
	return Config{
		Port:          19321,
		DBPath:        filepath.Join(DataDir(), "ai_interactions.db"),
		RetentionDays: 0,
		Reflection: Reflection{
			CLI:       "claude --print",
			Schedule:  "daily",
			OutputDir: filepath.Join(DataDir(), "reflections"),
			Prompt:    "",
		},
	}
}

func ConfigDir() string {
	if d := os.Getenv("XDG_CONFIG_HOME"); d != "" {
		return filepath.Join(d, "ai-collector")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "ai-collector")
}

func DataDir() string {
	if d := os.Getenv("XDG_DATA_HOME"); d != "" {
		return filepath.Join(d, "ai-collector")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "ai-collector")
}

func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.yaml")
}

func PidPath() string {
	return filepath.Join(DataDir(), "collector.pid")
}

func LogPath() string {
	return filepath.Join(DataDir(), "collector.log")
}

func Load() (Config, error) {
	cfg := Defaults()

	data, err := os.ReadFile(ConfigPath())
	if err == nil {
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return cfg, fmt.Errorf("parse config: %w", err)
		}
	}
	// env var overrides
	applyEnvOverrides(&cfg)

	return cfg, nil
}

func Save(cfg Config) error {
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	return os.WriteFile(ConfigPath(), data, 0o644)
}

func SetValue(key, value string) error {
	cfg, _ := Load()

	switch key {
	case "port":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("port must be a number")
		}
		cfg.Port = n
	case "db_path":
		cfg.DBPath = value
	case "retention_days":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("retention_days must be a number")
		}
		cfg.RetentionDays = n
	case "reflection.cli":
		cfg.Reflection.CLI = value
	case "reflection.schedule":
		cfg.Reflection.Schedule = value
	case "reflection.output_dir":
		cfg.Reflection.OutputDir = value
	case "reflection.prompt":
		cfg.Reflection.Prompt = value
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}

	return Save(cfg)
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("COLLECTOR_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Port = n
		}
	}
	if v := os.Getenv("COLLECTOR_DB_PATH"); v != "" {
		cfg.DBPath = v
	}
	if v := os.Getenv("COLLECTOR_RETENTION_DAYS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.RetentionDays = n
		}
	}
	if v := os.Getenv("COLLECTOR_REFLECTION_CLI"); v != "" {
		cfg.Reflection.CLI = v
	}
	if v := os.Getenv("COLLECTOR_REFLECTION_SCHEDULE"); v != "" {
		cfg.Reflection.Schedule = v
	}
	if v := os.Getenv("COLLECTOR_REFLECTION_DIR"); v != "" {
		cfg.Reflection.OutputDir = v
	}
	if v := os.Getenv("COLLECTOR_REFLECTION_PROMPT"); v != "" {
		cfg.Reflection.Prompt = v
	}
}
