#!/usr/bin/env bash
set -euo pipefail

COLLECTOR_URL="${COLLECTOR_URL:-http://localhost:9000}"
REFLECTION_CLI="${COLLECTOR_REFLECTION_CLI:-claude --print}"
PASS=0
FAIL=0

check() {
  local name="$1"
  local result="$2"
  if [[ "$result" == "ok" ]]; then
    echo "  ✓ $name"
    ((PASS++))
  else
    echo "  ✗ $name — $result"
    ((FAIL++))
  fi
}

echo "Verifying AI Interaction Collector setup..."
echo

# Check collector health
echo "Collector:"
if curl -sf "${COLLECTOR_URL}/health" >/dev/null 2>&1; then
  check "Collector running" "ok"
else
  check "Collector running" "not reachable at ${COLLECTOR_URL}"
fi

# Check reflection CLI
echo
echo "Reflection CLI:"
CLI_BIN=$(echo "$REFLECTION_CLI" | awk '{print $1}')
if command -v "$CLI_BIN" &>/dev/null; then
  check "$CLI_BIN in PATH" "ok"
else
  check "$CLI_BIN in PATH" "not found"
fi

# Check Claude hooks
echo
echo "Claude Code hooks:"
CLAUDE_GLOBAL="${HOME}/.claude/settings.json"
if [[ -f "$CLAUDE_GLOBAL" ]] && grep -q "claude-hook\|claude-prompt-hook\|ingest/claude" "$CLAUDE_GLOBAL" 2>/dev/null; then
  check "Global hooks configured" "ok"
else
  check "Global hooks configured" "not found in ${CLAUDE_GLOBAL}"
fi

# Check jq
echo
echo "Dependencies:"
if command -v jq &>/dev/null; then
  check "jq installed" "ok"
else
  check "jq installed" "not found"
fi

if command -v go &>/dev/null; then
  check "go installed" "ok"
else
  check "go installed" "not found"
fi

echo
echo "Results: ${PASS} passed, ${FAIL} failed"
[[ "$FAIL" -eq 0 ]] && exit 0 || exit 1
