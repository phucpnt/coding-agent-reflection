## Context

This is a greenfield Go service for personal use. There are three coding agent CLIs in play — Claude Code, Gemini CLI, and Codex — each with different hook/telemetry mechanisms. Currently, interaction data from these tools is ephemeral and siloed. The service runs locally on a single machine alongside the coding agents.

DuckDB is chosen over SQLite or Postgres because it handles analytical queries (aggregations over interaction history) well, requires zero infrastructure, and stores everything in a single file. The service is not multi-user or multi-tenant.

## Goals / Non-Goals

**Goals:**
- Unified ingest from Claude Code (HTTP hooks), Gemini CLI (HTTP hooks), and Codex (OTLP/HTTP)
- Normalize all provider payloads into a single `Interaction` struct before storage
- Persist interactions in DuckDB with enough context to understand what happened without replaying sessions
- Run a daily reflection job that calls an LLM to identify patterns and persist summaries
- Configurable retention: keep raw interactions for N days, keep reflections indefinitely

**Non-Goals:**
- Real-time dashboards or UI — query DuckDB directly or build later
- Multi-user or remote access — localhost only, no auth
- Streaming/live processing of agent output — ingest happens on completed events only
- Supporting agents beyond Claude, Gemini, and Codex in the initial version
- High availability or fault tolerance — acceptable to lose a few events on crash

## Decisions

### 1. Single Go binary with embedded HTTP server

**Choice**: One `main.go` binary using `net/http` stdlib, no framework.

**Rationale**: The service has ~4 routes. A framework (Gin, Echo) adds dependency weight for no benefit. Stdlib `net/http` handles routing, JSON decoding, and graceful shutdown. Keeps the binary small and dependency-free where possible.

**Alternatives considered**:
- Gin/Echo: Unnecessary for 4 routes. Would add middleware ecosystem we don't need.
- Separate binaries per provider: Defeats the purpose of unified collection.

### 2. Provider-specific HTTP handlers with shared normalizer

**Choice**: Three ingest handlers (`/ingest/claude`, `/ingest/gemini`, `/v1/traces`) each with a provider-specific normalizer that produces a common `Interaction` struct, then a shared `Store.Insert()` call.

**Rationale**: Each provider sends fundamentally different payloads (Claude hook JSON, Gemini hook JSON, OTLP protobuf spans). Trying to accept a generic payload would require complex conditional logic. Dedicated handlers keep parsing clean; the shared struct keeps storage unified.

**Alternatives considered**:
- Single `/ingest` endpoint with provider detection: Fragile — payload shapes overlap in unpredictable ways.
- Provider-specific tables: Defeats unified querying for reflection.

### 3. DuckDB with `go-duckdb` driver

**Choice**: Use `github.com/marcboeker/go-duckdb` via `database/sql` interface. Single `.duckdb` file in `./data/`.

**Rationale**: DuckDB is ideal for append-heavy writes with analytical reads (daily reflection queries). The Go driver is mature and implements `database/sql`. No server process needed — the file is the database.

**Alternatives considered**:
- SQLite: Weaker at analytical aggregations, but would also work. DuckDB's columnar storage is better for "scan all rows from yesterday" queries.
- PostgreSQL: Overkill infrastructure for a personal local service.

### 4. OTLP HTTP receiver for Codex (not gRPC)

**Choice**: Accept OTLP over HTTP (`/v1/traces`) with JSON or protobuf encoding. No gRPC.

**Rationale**: Codex's OTel export can target HTTP endpoints. HTTP is simpler to implement — no gRPC dependency, no protobuf code generation for the receiver. Parse spans, extract relevant attributes (`user_prompt`, `completion`, repo, etc.), and normalize.

**Alternatives considered**:
- Full OTel Collector sidecar: Too heavy for personal use. We only need a handful of span attributes.
- gRPC OTLP: Adds protobuf + gRPC dependencies for no benefit when HTTP works.

### 5. Daily reflection via HTTP trigger + LLM call

**Choice**: Expose `POST /jobs/daily-reflection` that queries yesterday's interactions, builds a prompt, calls an LLM API (Claude or similar), and writes the result to a `reflections` table. Triggered externally by cron/systemd timer.

**Rationale**: Keeping the scheduler external (cron) means the Go service stays stateless between requests — no internal timers, no goroutine lifecycle management. The reflection endpoint is idempotent (re-running for the same day overwrites).

**Alternatives considered**:
- Internal scheduler (e.g., `robfig/cron`): Adds statefulness and crash-recovery concerns.
- Batch export to markdown files: Loses queryability. DuckDB table is better for trend analysis.

### 6. Project layout

```
cmd/collector/main.go       — entry point, wire routes
internal/
  ingest/
    claude.go               — Claude hook handler + normalizer
    gemini.go               — Gemini hook handler + normalizer
    codex.go                — OTLP handler + normalizer
  model/
    interaction.go          — Interaction struct
    reflection.go           — Reflection struct
  store/
    duckdb.go               — DuckDB repository (insert, query, prune)
  reflection/
    job.go                  — Daily reflection job logic
config/
  config.go                 — Config loading (port, DB path, retention days, LLM API key)
data/                       — DuckDB file lives here (gitignored)
```

Flat internal packages, no interfaces until a second implementation exists.

## Risks / Trade-offs

- **[DuckDB concurrent writes]** DuckDB has a single-writer model. If two ingest requests hit simultaneously, one blocks. → Mitigation: Use a write mutex in the store layer. At personal-use volume (~100 events/day), contention is negligible.

- **[OTLP payload complexity]** Codex OTel spans may contain deeply nested attributes that change across versions. → Mitigation: Extract only the fields we need (prompt, output, repo, tokens). Log and skip unrecognized span types rather than failing.

- **[LLM API dependency for reflection]** Daily reflection fails if the LLM API is down or the key expires. → Mitigation: Reflection job logs the error and exits non-zero so cron can alert. Raw data is retained regardless — reflection can be re-run.

- **[Schema drift across CLI versions]** Claude/Gemini hook payload shapes may change in future CLI updates. → Mitigation: Normalize defensively — extract known fields, ignore unknown ones. Version the normalizers if breaking changes occur.

- **[Single-machine, single-file]** DuckDB file corruption or disk failure loses all data. → Mitigation: Acceptable for personal use. Can add periodic backup of the `.duckdb` file to cloud storage later.

## Open Questions

- Which LLM to use for the reflection job — Claude API, local model, or configurable? Start with Claude API and make it swappable.
- Should the reflection prompt be hardcoded or user-configurable? Start hardcoded, extract to a template file if needed.
- Exact Codex OTel span attribute names — need to verify against actual Codex telemetry output once enabled.
