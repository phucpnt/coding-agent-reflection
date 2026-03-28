#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
COLLECTOR_PORT="${COLLECTOR_PORT:-19321}"

if ! command -v jq &>/dev/null; then
  echo "Error: jq is required but not installed."
  echo "Install it: brew install jq / apt install jq / pacman -S jq"
  exit 1
fi

# Determine scope: global or project
SCOPE="${1:-}"
if [[ -z "$SCOPE" ]]; then
  echo "Where should the hooks be installed?"
  echo ""
  echo "  1) global   — all Claude Code sessions (~/.claude/settings.json)"
  echo "  2) project  — this project only (.claude/settings.json)"
  echo ""
  read -rp "Choose [1/2]: " choice
  case "$choice" in
    1|global)  SCOPE="global" ;;
    2|project) SCOPE="project" ;;
    *)
      echo "Invalid choice. Use: $0 [global|project]"
      exit 1
      ;;
  esac
fi

case "$SCOPE" in
  global)
    SETTINGS_FILE="${HOME}/.claude/settings.json"
    ;;
  project|--project)
    SETTINGS_FILE="${PROJECT_DIR}/.claude/settings.json"
    ;;
  *)
    echo "Usage: $0 [global|project]"
    exit 1
    ;;
esac

echo "Configuring Claude Code hooks in: ${SETTINGS_FILE}"
echo "Collector port: ${COLLECTOR_PORT}"

mkdir -p "$(dirname "$SETTINGS_FILE")"

# Back up existing settings
if [[ -f "$SETTINGS_FILE" ]]; then
  cp "$SETTINGS_FILE" "${SETTINGS_FILE}.bak"
  echo "Backed up existing settings to ${SETTINGS_FILE}.bak"
else
  echo '{}' > "$SETTINGS_FILE"
fi

STOP_HOOK="${PROJECT_DIR}/scripts/claude-hook.sh"

# Merge Stop hook into existing settings without overwriting
jq --arg stop_hook "$STOP_HOOK" '
  .hooks //= {} |
  .hooks.Stop //= [] |
  (if (.hooks.Stop | map(select(.hooks[]?.command == $stop_hook)) | length) == 0
   then .hooks.Stop += [{"hooks": [{"type": "command", "command": $stop_hook, "async": true}]}]
   else . end)
' "$SETTINGS_FILE" > "${SETTINGS_FILE}.tmp" && mv "${SETTINGS_FILE}.tmp" "$SETTINGS_FILE"

echo ""
echo "Done! Claude Code Stop hook configured (${SCOPE})."
echo "Restart your Claude Code session for hooks to take effect."
