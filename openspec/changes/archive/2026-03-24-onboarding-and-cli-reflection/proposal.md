## Why

A new user cloning this repo has no guided path from download to working setup. There's no Makefile, no README, no setup script, and hook configuration requires manual JSON editing. Additionally, the reflection job currently depends on the Claude API (requires an API key), when it could just shell out to whichever coding CLI the user already has installed.

## What Changes

- **Makefile** with `install`, `setup-claude`, `setup-gemini`, `setup-codex`, `run`, `reflect`, and `status` targets
- **Setup scripts** that auto-inject hooks into Claude Code, Gemini CLI, and Codex configurations
- **README** documenting the full onboarding flow from clone to working setup
- **CLI-based reflection** — **BREAKING**: replace the Claude API HTTP client with a subprocess call to a configurable CLI (`claude --print`, `codex --quiet`, `gemini`). No API key needed.
- **Scheduled reflection** — the collector runs reflection on a configurable interval (default: daily) via an internal goroutine, no external cron needed
- **Reflection output to files** — reflections saved as markdown files named `YYYYMMDD-NNN.md` in a user-configurable directory (in addition to SQLite)
- **Auto-create `data/` directory** on collector startup if missing
- **Verification command** — `make verify` or a health-check script to confirm the collector is running and hooks are wired

## Capabilities

### New Capabilities

- `onboarding`: Makefile, setup scripts for auto-injecting hooks into Claude/Gemini/Codex configs, README, and verification command
- `cli-reflection`: Replace API-based LLM client with CLI subprocess execution. Support `claude --print`, `codex`, `gemini` as configurable backends. Save reflections as `YYYYMMDD-NNN.md` files to a user-chosen directory. Schedule reflection via internal timer in the collector.

### Modified Capabilities

- `daily-reflection`: Reflection trigger changes from external cron to internal scheduler. LLM call changes from HTTP API to CLI subprocess. Reflection output adds file persistence alongside SQLite.

## Impact

- **`internal/reflection/llm.go`** — replaced entirely: API client becomes CLI subprocess executor
- **`internal/reflection/handler.go`** — updated to support scheduled trigger
- **`cmd/collector/main.go`** — adds goroutine for scheduled reflection, auto-creates data dir
- **`config/config.go`** — new fields: `ReflectionCLI`, `ReflectionSchedule`, `ReflectionOutputDir`
- **New files**: `Makefile`, `scripts/setup-claude.sh`, `scripts/setup-gemini.sh`, `scripts/setup-codex.sh`, `scripts/verify.sh`, `README.md`
- **Removed dependency**: `ANTHROPIC_API_KEY` env var no longer required
