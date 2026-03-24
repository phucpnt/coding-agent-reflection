## 1. Project Setup

- [x] 1.1 Initialize Go module (`go mod init`) and create directory structure: `cmd/collector/`, `internal/ingest/`, `internal/model/`, `internal/store/`, `internal/reflection/`, `config/`, `data/`
- [x] 1.2 Add `go-duckdb` driver dependency and verify it compiles
- [x] 1.3 Create `config/config.go` — load config from env vars or TOML: port, DB path, retention days, LLM API key, LLM model name
- [x] 1.4 Add `data/` to `.gitignore`

## 2. Data Model

- [x] 2.1 Create `internal/model/interaction.go` — `Interaction` struct with fields: `ID` (UUID), `Ts`, `Provider`, `SessionID`, `Project`, `UserPrompt`, `AgentOutput`, `Context`, `TokensPrompt`, `TokensOutput`, `ToolsUsed`
- [x] 2.2 Create `internal/model/reflection.go` — `Reflection` struct with fields: `ID` (UUID), `Date`, `Summary`, `ShouldDo`, `ShouldNotDo`, `ConfigChanges`, `CreatedAt`

## 3. DuckDB Storage Layer

- [x] 3.1 Create `internal/store/duckdb.go` — `Store` struct holding `*sql.DB` and `sync.Mutex`; constructor opens DuckDB file and runs `CREATE TABLE IF NOT EXISTS` for both `interactions` and `reflections` tables
- [x] 3.2 Implement `Store.Insert(interaction)` — mutex-protected single-row insert into `interactions`
- [x] 3.3 Implement `Store.QueryByDateRange(from, to)` — returns `[]Interaction` ordered by `ts` ascending
- [x] 3.4 Implement `Store.UpsertReflection(reflection)` — insert or replace keyed by `date`
- [x] 3.5 Implement `Store.PruneInteractions(retentionDays)` — delete rows older than retention window
- [x] 3.6 Implement `Store.HealthCheck()` — run a simple query to verify DuckDB is accessible
- [x] 3.7 Write tests for Store using a temporary DuckDB file

## 4. Ingest Handlers

- [x] 4.1 Create `internal/ingest/claude.go` — HTTP handler for `POST /ingest/claude`: parse Claude hook JSON, extract fields, produce `Interaction`, call `Store.Insert()`
- [x] 4.2 Create `internal/ingest/gemini.go` — HTTP handler for `POST /ingest/gemini`: parse Gemini hook JSON, extract fields, produce `Interaction`, call `Store.Insert()`
- [x] 4.3 Create `internal/ingest/codex.go` — HTTP handler for `POST /v1/traces`: parse OTLP JSON, iterate spans, extract LLM attributes, produce `Interaction` per relevant span, call `Store.Insert()`. Skip unrecognized spans with debug log.
- [x] 4.4 Write tests for each handler using `httptest` with sample payloads

## 5. Daily Reflection Job

- [x] 5.1 Create `internal/reflection/job.go` — `RunReflection(store, llmClient, targetDate)` function: query interactions for the date, build LLM prompt, call API, parse response into `Reflection` struct, upsert
- [x] 5.2 Implement LLM client wrapper — call Claude API with the reflection prompt, return structured response. Make the model configurable.
- [x] 5.3 Implement reflection HTTP handler for `POST /jobs/daily-reflection`: accept optional `{"date": "..."}` body, default to yesterday, call `RunReflection`, optionally prune if retention configured
- [x] 5.4 Write tests for reflection job with a mocked LLM client

## 6. HTTP Server and Wiring

- [x] 6.1 Create `cmd/collector/main.go` — load config, open DuckDB store, register routes (`/health`, `/ingest/claude`, `/ingest/gemini`, `/v1/traces`, `/jobs/daily-reflection`), start HTTP server with graceful shutdown
- [x] 6.2 Implement `GET /health` handler — call `Store.HealthCheck()`, return `{"status": "ok"}` or 503
- [x] 6.3 Add structured logging (stdlib `slog`) for ingest events, errors, and reflection job results

## 7. CLI Hook Configuration

- [x] 7.1 Create example Claude Code `hooks.json` config pointing `AfterAgent` to `http://localhost:9000/ingest/claude`
- [x] 7.2 Create example Gemini CLI hooks config pointing `AfterAgent` to `http://localhost:9000/ingest/gemini`
- [x] 7.3 Create example Codex `config.toml` snippet enabling OTel HTTP export to `http://localhost:9000`
- [x] 7.4 Document example cron entry for daily reflection: `0 2 * * * curl -X POST http://localhost:9000/jobs/daily-reflection`

## 8. Integration Testing

- [x] 8.1 End-to-end test: start server, POST sample Claude payload, verify row in DuckDB
- [x] 8.2 End-to-end test: POST sample Gemini payload, verify row
- [x] 8.3 End-to-end test: POST sample OTLP trace, verify row
- [x] 8.4 End-to-end test: insert interactions, trigger reflection, verify reflection row exists
