package setup

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSetupClaude_NewSettings(t *testing.T) {
	dir := t.TempDir()
	settingsPath := filepath.Join(dir, ".claude", "settings.json")
	hookScript := "/path/to/claude-hook.sh"

	if err := mergeClaudeHook(settingsPath, hookScript); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatal(err)
	}

	var settings map[string]any
	json.Unmarshal(data, &settings)

	hooks := settings["hooks"].(map[string]any)
	stopHooks := hooks["Stop"].([]any)
	if len(stopHooks) != 1 {
		t.Fatalf("expected 1 stop hook, got %d", len(stopHooks))
	}
}

func TestSetupClaude_ExistingHooks(t *testing.T) {
	dir := t.TempDir()
	settingsPath := filepath.Join(dir, "settings.json")
	os.MkdirAll(filepath.Dir(settingsPath), 0o755)

	existing := `{"hooks":{"Stop":[{"hooks":[{"type":"command","command":"existing-hook"}]}]}}`
	os.WriteFile(settingsPath, []byte(existing), 0o644)

	if err := mergeClaudeHook(settingsPath, "/path/to/claude-hook.sh"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(settingsPath)
	var settings map[string]any
	json.Unmarshal(data, &settings)

	hooks := settings["hooks"].(map[string]any)
	stopHooks := hooks["Stop"].([]any)
	if len(stopHooks) != 2 {
		t.Fatalf("expected 2 stop hooks (existing + new), got %d", len(stopHooks))
	}

	// Verify backup was created
	if _, err := os.Stat(settingsPath + ".bak"); os.IsNotExist(err) {
		t.Fatal("backup file not created")
	}
}

func TestSetupClaude_Idempotent(t *testing.T) {
	dir := t.TempDir()
	settingsPath := filepath.Join(dir, "settings.json")

	mergeClaudeHook(settingsPath, "/path/to/hook.sh")
	mergeClaudeHook(settingsPath, "/path/to/hook.sh")

	data, _ := os.ReadFile(settingsPath)
	var settings map[string]any
	json.Unmarshal(data, &settings)

	hooks := settings["hooks"].(map[string]any)
	stopHooks := hooks["Stop"].([]any)
	if len(stopHooks) != 1 {
		t.Fatalf("expected 1 stop hook (idempotent), got %d", len(stopHooks))
	}
}
