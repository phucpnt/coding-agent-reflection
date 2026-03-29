package setup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func SetupClaude(scope string, hookScriptPath string) error {
	var settingsPath string
	switch scope {
	case "global":
		home, _ := os.UserHomeDir()
		settingsPath = filepath.Join(home, ".claude", "settings.json")
	case "project":
		settingsPath = filepath.Join(".claude", "settings.json")
	default:
		return fmt.Errorf("invalid scope: %s (use global or project)", scope)
	}

	return mergeClaudeHook(settingsPath, hookScriptPath)
}

func mergeClaudeHook(settingsPath, hookScriptPath string) error {
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		return fmt.Errorf("create settings dir: %w", err)
	}

	// Read existing settings
	var settings map[string]any
	data, err := os.ReadFile(settingsPath)
	if err == nil {
		// Backup
		if err := os.WriteFile(settingsPath+".bak", data, 0o644); err != nil {
			return fmt.Errorf("backup settings: %w", err)
		}
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("parse settings: %w", err)
		}
	} else {
		settings = map[string]any{}
	}

	// Ensure hooks.Stop exists
	hooks, _ := settings["hooks"].(map[string]any)
	if hooks == nil {
		hooks = map[string]any{}
		settings["hooks"] = hooks
	}

	stopHooks, _ := hooks["Stop"].([]any)

	// Check if our hook is already present
	for _, entry := range stopHooks {
		entryMap, _ := entry.(map[string]any)
		hooksList, _ := entryMap["hooks"].([]any)
		for _, h := range hooksList {
			hMap, _ := h.(map[string]any)
			if hMap["command"] == hookScriptPath {
				return nil // already configured
			}
		}
	}

	// Add our hook
	newHook := map[string]any{
		"hooks": []any{
			map[string]any{
				"type":    "command",
				"command": hookScriptPath,
				"async":   true,
			},
		},
	}
	stopHooks = append(stopHooks, newHook)
	hooks["Stop"] = stopHooks

	// Write back
	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}
	return os.WriteFile(settingsPath, append(out, '\n'), 0o644)
}

func FindHookScript() string {
	// Try to find the hook script relative to the executable
	exe, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(exe)
		candidates := []string{
			filepath.Join(dir, "..", "scripts", "claude-hook.sh"),
			filepath.Join(dir, "scripts", "claude-hook.sh"),
		}
		for _, c := range candidates {
			if abs, err := filepath.Abs(c); err == nil {
				if _, err := os.Stat(abs); err == nil {
					return abs
				}
			}
		}
	}
	// Fallback: look in current directory
	if abs, err := filepath.Abs("scripts/claude-hook.sh"); err == nil {
		if _, err := os.Stat(abs); err == nil {
			return abs
		}
	}
	return ""
}
