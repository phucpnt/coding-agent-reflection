## Context

The collector service exists and works — it ingests interactions from Claude/Gemini/Codex into SQLite and has a reflection endpoint. But setup is entirely manual (copy JSON, edit configs), the reflection job requires a Claude API key, and there's no documentation. The user has Claude Code, Gemini CLI, and/or Codex already installed locally.

## Goals / Non-Goals

**Goals:**
- Zero-to-working in `git clone && make install && make setup-claude && make run`
- Reflection via CLI subprocess (no API key) — user picks which CLI to use
- Reflection files saved as `YYYYMMDD-NNN.md` in a configurable directory
- Collector handles its own reflection schedule internally
- Verification that setup is correct before the user has to debug

**Non-Goals:**
- GUI or web UI for configuration
- Supporting CLIs beyond claude/codex/gemini
- Daemonizing the collector (systemd/launchd) — user runs it however they want
- Migrating existing DuckDB data to SQLite

## Decisions

### 1. CLI subprocess via `exec.CommandContext` with piped stdin

**Choice**: Replace `ClaudeClient` (HTTP API) with a `CLICompleter` that runs a configurable command, pipes the prompt to stdin, and captures stdout. The `LLMClient` interface stays the same (`Complete(ctx, prompt) (string, error)`).

**Rationale**: All three CLIs support non-interactive piped input:
- `echo "prompt" | claude --print` — outputs response to stdout, no UI
- `echo "prompt" | codex --quiet` — quiet mode, stdout only
- `echo "prompt" | gemini` — reads stdin when piped

Using `exec.CommandContext` gives us timeout control via context. The command string is configurable so users can add their own flags.

**Alternatives considered**:
- Keep API client as option alongside CLI: Over-engineering for now. CLI works without extra credentials.
- Write prompt to temp file, pass as arg: Unnecessary — stdin piping works for all three CLIs.

### 2. Reflection file naming: `YYYYMMDD-NNN.md`

**Choice**: Save each reflection as a markdown file in a configurable output directory. Filename format: `20260324-001.md`, `20260324-002.md`, etc. The sequence number `NNN` increments per-day by scanning existing files with the same date prefix.

**Rationale**: User asked for this format. It allows multiple reflections per day (e.g., manual re-runs) and sorts chronologically. Markdown is human-readable and can be committed to a vault/repo.

The file contains the same structured sections as the SQLite row: summary, should do, should not do, config changes — plus a YAML frontmatter with metadata (date, provider count, interaction count).

### 3. Configurable output directory via `COLLECTOR_REFLECTION_DIR`

**Choice**: Default to `./data/reflections/`. User overrides via env var `COLLECTOR_REFLECTION_DIR`. The collector creates the directory on startup if it doesn't exist.

**Rationale**: User explicitly asked to choose the location. Env var is the simplest config mechanism (consistent with existing config pattern). Default keeps reflections alongside the DB for portability.

### 4. Internal scheduler via `time.Ticker` goroutine

**Choice**: The collector starts a background goroutine that triggers reflection on a configurable schedule. Config: `COLLECTOR_REFLECTION_SCHEDULE` — accepts `daily` (default, runs at 02:00 local time), `hourly`, or `off` (disabled).

On each tick, it checks if a reflection already exists for the target period. If yes, skip. If no, run the reflection job.

**Rationale**: Simpler than cron for personal use. The collector is already long-running. A goroutine with `time.Ticker` is the lightest approach — no third-party scheduler dependency.

**Alternatives considered**:
- `robfig/cron`: Adds a dependency for a single scheduled job. Overkill.
- External cron only: Already rejected — user wants it built-in.

### 5. Setup scripts that merge into existing settings

**Choice**: Each setup script (`setup-claude.sh`, etc.) reads the user's existing settings file, uses `jq` to merge the hook entry without overwriting existing hooks, and writes it back. If `jq` is not installed, the script exits with a clear message.

For Claude Code: merges into `~/.claude/settings.json` (global) or `.claude/settings.json` (project). The script asks which.

**Rationale**: Users may have existing hooks (like the `agent-deck` hooks we saw). Blindly overwriting would break their setup. `jq` is the standard tool for JSON manipulation in shell.

**Alternatives considered**:
- Go tool for setup: More code to maintain for a one-time operation.
- Manual instructions only: Defeats the purpose of onboarding.

### 6. Makefile as the single entry point

**Choice**:
```
make install        # go build both binaries
make setup-claude   # inject hooks into Claude settings
make setup-gemini   # inject hooks into Gemini settings
make setup-codex    # inject hooks into Codex settings
make run            # start collector (foreground)
make reflect        # trigger reflection manually now
make verify         # check collector is running, hooks are configured
make status         # curl /interactions summary
```

**Rationale**: Makefile is universally available, self-documenting (targets show available actions), and doesn't require learning a new tool. Each target is a thin wrapper around a script or go command.

### 7. Migration path for existing config

**Choice**: Remove `LLMAPIKey` and `LLMModel` from config. Add `ReflectionCLI` (default: `claude --print`), `ReflectionSchedule` (default: `daily`), `ReflectionOutputDir` (default: `./data/reflections/`). The HTTP reflection endpoint still works for manual triggers.

`ANTHROPIC_API_KEY` is no longer read. If set, it's ignored.

## Risks / Trade-offs

- **[CLI availability]** If the configured CLI isn't installed, reflection fails silently on the scheduled tick. → Mitigation: `make verify` checks that the CLI binary exists in PATH. Scheduled reflection logs errors clearly.

- **[CLI output format]** Different CLI versions may produce different output formats (markdown vs plain text, extra preamble). → Mitigation: The reflection parser already handles freeform text with a fallback (stuff everything into `summary` if sections aren't found).

- **[Long-running subprocess]** A CLI reflection call could hang. → Mitigation: `exec.CommandContext` with a 5-minute timeout. Kill the process on timeout.

- **[jq dependency for setup]** Setup scripts require `jq`. → Mitigation: Check for `jq` at script start, print install instructions if missing. `jq` is available on all major package managers.

- **[Merging hooks into existing settings]** Complex jq merge could corrupt settings. → Mitigation: Scripts back up the original file before modifying (e.g., `settings.json.bak`).

## Open Questions

- Should `make setup-claude` configure hooks globally (`~/.claude/settings.json`) or per-project (`.claude/settings.json`)? Lean towards global with a `--project` flag.
- Exact `codex --quiet` and `gemini` stdin/stdout behavior needs verification against current CLI versions.
