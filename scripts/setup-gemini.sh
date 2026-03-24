#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

if ! command -v jq &>/dev/null; then
  echo "Error: jq is required but not installed."
  echo "Install it: brew install jq / apt install jq / pacman -S jq"
  exit 1
fi

# Gemini CLI settings location
SETTINGS_DIR="${HOME}/.gemini"
SETTINGS_FILE="${SETTINGS_DIR}/settings.json"

echo "Configuring Gemini CLI hooks in: ${SETTINGS_FILE}"

mkdir -p "$SETTINGS_DIR"

if [[ -f "$SETTINGS_FILE" ]]; then
  cp "$SETTINGS_FILE" "${SETTINGS_FILE}.bak"
  echo "Backed up existing settings to ${SETTINGS_FILE}.bak"
else
  echo '{}' > "$SETTINGS_FILE"
fi

COLLECTOR_URL="http://localhost:9000/ingest/gemini"

jq --arg url "$COLLECTOR_URL" '
  .hooks //= {} |
  .hooks.AfterAgent //= [] |
  (if (.hooks.AfterAgent | map(select(.url == $url)) | length) == 0
   then .hooks.AfterAgent += [{"type": "http", "url": $url, "timeout": 5000}]
   else . end)
' "$SETTINGS_FILE" > "${SETTINGS_FILE}.tmp" && mv "${SETTINGS_FILE}.tmp" "$SETTINGS_FILE"

echo "Done! Gemini CLI hooks configured."
echo "Restart your Gemini CLI session for hooks to take effect."
