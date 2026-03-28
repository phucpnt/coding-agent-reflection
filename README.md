# AI Interaction Collector

A Go service that captures interactions from coding agents (Claude Code, Gemini CLI, Codex) into SQLite and generates daily self-reflections using a configurable CLI.

## Prerequisites

- Go 1.24+
- `jq` (for setup scripts)
- At least one coding CLI: `claude`, `gemini`, or `codex`

## Quick Start

```bash
git clone <repo-url> && cd coding-agent-reflection

# Build binaries
make install

# Configure hooks for your CLI(s)
make setup-claude    # Claude Code
make setup-gemini    # Gemini CLI
make setup-codex     # Codex

# Start the collector
make run

# Verify everything is wired up
make verify
```

Restart your coding CLI session after setup for hooks to take effect.

## Configuration

All configuration is via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `COLLECTOR_PORT` | `9000` | HTTP server port |
| `COLLECTOR_DB_PATH` | `./data/ai_interactions.db` | SQLite database path |
| `COLLECTOR_RETENTION_DAYS` | `0` (keep all) | Auto-prune interactions older than N days |
| `COLLECTOR_REFLECTION_CLI` | `claude --print` | CLI command for generating reflections |
| `COLLECTOR_REFLECTION_SCHEDULE` | `daily` | `daily`, `hourly`, or `off` |
| `COLLECTOR_REFLECTION_DIR` | `./data/reflections` | Directory for reflection markdown files |
| `COLLECTOR_REFLECTION_PROMPT` | _(built-in)_ | Path to custom prompt template file |

Example with custom settings:

```bash
COLLECTOR_REFLECTION_CLI="gemini" COLLECTOR_REFLECTION_SCHEDULE=hourly make run
```

### Custom Reflection Prompt

Override the built-in reflection prompt by pointing to your own template file:

```bash
COLLECTOR_REFLECTION_PROMPT=./prompts/reflection.md make run
```

The template file uses `{{INTERACTIONS}}` as a placeholder that gets replaced with the formatted interaction data. See `prompts/reflection.md` for the default template.

Your template should instruct the LLM to respond with `## Summary`, `## Should Do`, `## Should Not Do`, and `## Config Changes` section headers — the parser extracts these sections from the response.

## Usage

```bash
# Query recent interactions
make status

# Trigger a reflection manually
make reflect

# Use the query tool directly
./query today        # today's interactions
./query week         # last 7 days
./query stats        # summary with counts
./query reflections  # recent reflections
```

## How It Works

1. **Hooks** fire on each coding CLI interaction and POST data to the collector
2. **Collector** normalizes payloads and stores them in SQLite
3. **Scheduler** triggers reflections on your configured schedule
4. **Reflection** pipes the day's interactions to your chosen CLI and saves the result as a `YYYYMMDD-NNN.md` file

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Health check |
| `POST` | `/ingest/claude` | Claude Code hook ingest |
| `POST` | `/ingest/gemini` | Gemini CLI hook ingest |
| `POST` | `/v1/traces` | Codex OTLP trace ingest |
| `GET` | `/interactions` | Query recent interactions |
| `POST` | `/jobs/daily-reflection` | Trigger reflection manually |

## Project Structure

```
cmd/collector/    — HTTP server entry point
cmd/query/        — CLI query tool
config/           — Environment-based configuration
internal/
  ingest/         — Provider-specific HTTP handlers
  model/          — Data structs (Interaction, Reflection)
  reflection/     — CLI completer, scheduler, file writer
  store/          — SQLite repository
scripts/          — Setup and hook scripts
examples/         — Example config snippets
data/             — SQLite DB and reflection files (gitignored)
```
