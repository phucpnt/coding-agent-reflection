#!/usr/bin/env bash
# Claude Code UserPromptSubmit hook — captures each user prompt as it's sent.

set -euo pipefail

COLLECTOR_URL="${COLLECTOR_URL:-http://localhost:9000}"

INPUT=$(cat)

SESSION_ID=$(echo "$INPUT" | jq -r '.session_id // empty')
CWD=$(echo "$INPUT" | jq -r '.cwd // empty')
PROMPT=$(echo "$INPUT" | jq -r '.prompt // empty')

# Skip empty prompts
[[ -z "$PROMPT" ]] && exit 0

jq -n \
  --arg sid "$SESSION_ID" \
  --arg cwd "$CWD" \
  --arg prompt "$PROMPT" \
  '{
    session_id: $sid,
    cwd: $cwd,
    user_prompt: $prompt
  }' | curl -s -o /dev/null -w '' \
    -X POST "${COLLECTOR_URL}/ingest/claude" \
    -H "Content-Type: application/json" \
    -d @- || true

exit 0
