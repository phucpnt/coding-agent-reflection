#!/usr/bin/env bash
set -euo pipefail

# Codex config location
CONFIG_DIR="${HOME}/.codex"
CONFIG_FILE="${CONFIG_DIR}/config.toml"

echo "Configuring Codex OTel export in: ${CONFIG_FILE}"

mkdir -p "$CONFIG_DIR"

if [[ -f "$CONFIG_FILE" ]]; then
  cp "$CONFIG_FILE" "${CONFIG_FILE}.bak"
  echo "Backed up existing config to ${CONFIG_FILE}.bak"
fi

# Check if telemetry section already exists
if [[ -f "$CONFIG_FILE" ]] && grep -q 'endpoint.*localhost:9000' "$CONFIG_FILE" 2>/dev/null; then
  echo "Codex OTel export already configured."
  exit 0
fi

# Append telemetry config
cat >> "$CONFIG_FILE" <<'EOF'

# AI Interaction Collector - OTel export
[telemetry]
enabled = true
exporter = "otlp-http"
endpoint = "http://localhost:9000"
EOF

echo "Done! Codex OTel export configured."
echo "Restart Codex for changes to take effect."
