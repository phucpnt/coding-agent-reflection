#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

if ! command -v jq &>/dev/null; then
  echo "Error: jq is required but not installed."
  echo "Install it: brew install jq / apt install jq / pacman -S jq"
  exit 1
fi

SETTINGS_FILE="${HOME}/.claude/settings.json"
if [[ "${1:-}" == "--project" ]]; then
  SETTINGS_FILE="${PROJECT_DIR}/.claude/settings.json"
fi

echo "Configuring Claude Code hooks in: ${SETTINGS_FILE}"

mkdir -p "$(dirname "$SETTINGS_FILE")"

# Back up existing settings
if [[ -f "$SETTINGS_FILE" ]]; then
  cp "$SETTINGS_FILE" "${SETTINGS_FILE}.bak"
  echo "Backed up existing settings to ${SETTINGS_FILE}.bak"
else
  echo '{}' > "$SETTINGS_FILE"
fi

PROMPT_HOOK="${PROJECT_DIR}/scripts/claude-prompt-hook.sh"
STOP_HOOK="${PROJECT_DIR}/scripts/claude-hook.sh"

# Merge hooks into existing settings without overwriting
jq --arg prompt_hook "$PROMPT_HOOK" --arg stop_hook "$STOP_HOOK" '
  .hooks //= {} |
  .hooks.UserPromptSubmit //= [] |
  .hooks.Stop //= [] |
  # Add prompt hook if not already present
  (if (.hooks.UserPromptSubmit | map(select(.hooks[]?.command == $prompt_hook)) | length) == 0
   then .hooks.UserPromptSubmit += [{"hooks": [{"type": "command", "command": $prompt_hook, "async": true}]}]
   else . end) |
  # Add stop hook if not already present
  (if (.hooks.Stop | map(select(.hooks[]?.command == $stop_hook)) | length) == 0
   then .hooks.Stop += [{"hooks": [{"type": "command", "command": $stop_hook, "async": true}]}]
   else . end)
' "$SETTINGS_FILE" > "${SETTINGS_FILE}.tmp" && mv "${SETTINGS_FILE}.tmp" "$SETTINGS_FILE"

echo "Done! Claude Code hooks configured."
echo "Restart your Claude Code session for hooks to take effect."
