#!/usr/bin/env bash
# Claude Code Stop hook adapter — reads hook payload from stdin,
# extracts fields, and POSTs to the collector service.

set -euo pipefail

COLLECTOR_URL="${COLLECTOR_URL:-http://localhost:9000}"

INPUT=$(cat)

SESSION_ID=$(echo "$INPUT" | jq -r '.session_id // empty')
CWD=$(echo "$INPUT" | jq -r '.cwd // empty')
LAST_OUTPUT=$(echo "$INPUT" | jq -r '.last_assistant_message // empty')
TRANSCRIPT_PATH=$(echo "$INPUT" | jq -r '.transcript_path // empty')

# Expand ~ in transcript path
TRANSCRIPT_PATH="${TRANSCRIPT_PATH/#\~/$HOME}"

# Extract last user prompt from transcript (JSONL)
# Transcript format: each line has .type, .message.role, .message.content
# User messages: .type=="user" and .message.role=="user"
# Content is either a string or array of {type:"text", text:"..."}
USER_PROMPT=""
if [[ -f "$TRANSCRIPT_PATH" ]]; then
  # Get last user message with actual text (skip tool_result-only messages)
  USER_PROMPT=$(jq -rs '
    [.[] | select(
      .type == "user" and .message.role == "user" and .message.content != null
      and (
        (.message.content | type == "string") or
        (.message.content | type == "array" and any(.[]; .type == "text"))
      )
    )]
    | last
    | .message.content
    | if type == "array" then
        [.[] | select(.type == "text") | .text] | join("\n")
      elif type == "string" then .
      else ""
      end
  ' "$TRANSCRIPT_PATH" 2>/dev/null || true)
fi

# Extract tool names used in the session
# Assistant messages: .type=="assistant" and .message.content is array with tool_use items
TOOLS_USED="[]"
if [[ -f "$TRANSCRIPT_PATH" ]]; then
  TOOLS_USED=$(jq -rs '
    [.[] | select(.type == "assistant" and .message.content != null)
     | .message.content[]? | select(.type == "tool_use") | .name]
    | unique
  ' "$TRANSCRIPT_PATH" 2>/dev/null || echo "[]")
fi

# Build payload and POST to collector
jq -n \
  --arg sid "$SESSION_ID" \
  --arg cwd "$CWD" \
  --arg prompt "$USER_PROMPT" \
  --arg output "$LAST_OUTPUT" \
  --argjson tools "$TOOLS_USED" \
  '{
    session_id: $sid,
    cwd: $cwd,
    user_prompt: $prompt,
    agent_output: $output,
    tools_used: $tools
  }' | curl -s -o /dev/null -w '' \
    -X POST "${COLLECTOR_URL}/ingest/claude" \
    -H "Content-Type: application/json" \
    -d @- || true

exit 0
