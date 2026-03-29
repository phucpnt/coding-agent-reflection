package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaults(t *testing.T) {
	cfg := Defaults()
	if cfg.Port != 19321 {
		t.Errorf("expected port 19321, got %d", cfg.Port)
	}
	if cfg.Reflection.CLI != "claude --print" {
		t.Errorf("expected claude --print, got %s", cfg.Reflection.CLI)
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("XDG_DATA_HOME", dir)

	cfg := Defaults()
	cfg.Port = 8080
	cfg.Reflection.CLI = "gemini"

	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}

	// Verify file exists
	path := filepath.Join(dir, "ai-collector", "config.yaml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("config file not created")
	}

	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Port != 8080 {
		t.Errorf("expected port 8080, got %d", loaded.Port)
	}
	if loaded.Reflection.CLI != "gemini" {
		t.Errorf("expected gemini, got %s", loaded.Reflection.CLI)
	}
}

func TestEnvOverride(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("XDG_DATA_HOME", dir)
	t.Setenv("COLLECTOR_PORT", "9999")
	t.Setenv("COLLECTOR_REFLECTION_CLI", "codex --quiet")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 9999 {
		t.Errorf("expected port 9999, got %d", cfg.Port)
	}
	if cfg.Reflection.CLI != "codex --quiet" {
		t.Errorf("expected codex --quiet, got %s", cfg.Reflection.CLI)
	}
}

func TestSetValue(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("XDG_DATA_HOME", dir)

	if err := SetValue("port", "7777"); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 7777 {
		t.Errorf("expected port 7777, got %d", cfg.Port)
	}
}

func TestSetValueUnknownKey(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("XDG_DATA_HOME", dir)

	if err := SetValue("unknown.key", "value"); err == nil {
		t.Fatal("expected error for unknown key")
	}
}
