package setup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func SetupGemini(port int) error {
	home, _ := os.UserHomeDir()
	settingsPath := filepath.Join(home, ".gemini", "settings.json")

	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		return fmt.Errorf("create settings dir: %w", err)
	}

	var settings map[string]any
	data, err := os.ReadFile(settingsPath)
	if err == nil {
		os.WriteFile(settingsPath+".bak", data, 0o644)
		json.Unmarshal(data, &settings)
	}
	if settings == nil {
		settings = map[string]any{}
	}

	hooks, _ := settings["hooks"].(map[string]any)
	if hooks == nil {
		hooks = map[string]any{}
		settings["hooks"] = hooks
	}

	url := fmt.Sprintf("http://localhost:%d/ingest/gemini", port)

	afterAgent, _ := hooks["AfterAgent"].([]any)
	for _, entry := range afterAgent {
		entryMap, _ := entry.(map[string]any)
		if entryMap["url"] == url {
			return nil
		}
	}

	afterAgent = append(afterAgent, map[string]any{
		"type":    "http",
		"url":     url,
		"timeout": 5000,
	})
	hooks["AfterAgent"] = afterAgent

	out, _ := json.MarshalIndent(settings, "", "  ")
	return os.WriteFile(settingsPath, append(out, '\n'), 0o644)
}
