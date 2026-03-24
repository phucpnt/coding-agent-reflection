## Why

There is no unified way to track, analyze, and learn from interactions across multiple coding agents (Claude Code, Gemini CLI, Codex). Each tool operates in isolation with no shared observability. A single collector service would enable self-reflection on prompt patterns, repeated mistakes, and effective techniques — turning raw usage data into a personal coaching loop.

## What Changes

- **New Go service** ("collector") that accepts interaction data from three coding agent CLIs via HTTP hooks and OpenTelemetry
- **Unified DuckDB storage** normalizing interactions from all providers into a single `interactions` table
- **Provider-specific ingest endpoints** (`/ingest/claude`, `/ingest/gemini`, OTLP `/v1/traces` for Codex) with per-provider normalization
- **Daily self-reflection job** that summarizes interactions, identifies patterns, and persists reflections
- **Data lifecycle management** with configurable retention and pruning of raw logs

## Capabilities

### New Capabilities

- `ingest-api`: HTTP ingest endpoints for Claude and Gemini hooks, plus OTLP receiver for Codex. Normalizes provider-specific payloads into a unified interaction struct (provider, session, project, prompt, output, tokens, tools).
- `duckdb-storage`: DuckDB-backed persistence layer. Schema for `interactions` and `reflections` tables. Insert, query, and prune operations.
- `daily-reflection`: Scheduled job that queries recent interactions, calls an LLM to summarize patterns (what worked, what didn't, reusable techniques), and persists reflections. Includes configurable retention/pruning.

### Modified Capabilities

_None — this is a greenfield service._

## Impact

- **New Go module** with dependencies on DuckDB driver, net/http, and an OTLP receiver library
- **CLI hook configuration** needed for Claude Code (`hooks.json`) and Gemini CLI to point at the collector
- **Codex config** needs OTel export enabled pointing at the collector's OTLP endpoint
- **Local port 9000** reserved for the collector service
- **Cron or systemd timer** needed to trigger daily reflection endpoint
