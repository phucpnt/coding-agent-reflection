package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func SetupCodex(port int) error {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".codex", "config.toml")

	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	endpoint := fmt.Sprintf("http://localhost:%d", port)

	// Check if already configured
	if data, err := os.ReadFile(configPath); err == nil {
		os.WriteFile(configPath+".bak", data, 0o644)
		if strings.Contains(string(data), endpoint) {
			return nil
		}
	}

	f, err := os.OpenFile(configPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	fmt.Fprintf(f, `
# AI Interaction Collector - OTel export
[telemetry]
enabled = true
exporter = "otlp-http"
endpoint = "%s"
`, endpoint)

	return nil
}
