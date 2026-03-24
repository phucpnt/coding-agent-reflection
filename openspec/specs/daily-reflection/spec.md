## Requirements

### Requirement: Daily reflection HTTP trigger
The system SHALL expose `POST /jobs/daily-reflection` that triggers the reflection job for the previous day. The endpoint SHALL be idempotent — calling it multiple times for the same day overwrites the previous reflection.

#### Scenario: Trigger daily reflection
- **WHEN** a POST request arrives at `/jobs/daily-reflection` with no body
- **THEN** the system queries yesterday's interactions, generates a reflection via LLM, upserts it into the `reflections` table, and responds with HTTP 200 and the reflection summary

#### Scenario: Trigger reflection with custom date
- **WHEN** a POST request arrives at `/jobs/daily-reflection` with body `{"date": "2026-03-23"}`
- **THEN** the system generates a reflection for the specified date instead of yesterday

#### Scenario: No interactions for the target date
- **WHEN** the reflection job runs and there are zero interactions for the target date
- **THEN** the system responds with HTTP 200 and a message indicating no interactions to reflect on, without calling the LLM or creating a reflection row

### Requirement: LLM-based reflection generation
The system SHALL build a prompt from the day's interactions and call a coding CLI subprocess (configurable, default `claude --print`) to generate a structured reflection. The prompt SHALL ask the LLM to identify: patterns in prompts that worked vs failed, repeated mistakes, reusable techniques, and suggestions for config/workflow changes.

#### Scenario: Successful CLI reflection
- **WHEN** the reflection job pipes the prompt to the configured CLI subprocess
- **THEN** the CLI response is parsed into structured fields: `summary`, `should_do`, `should_not_do`, `config_changes`

#### Scenario: CLI subprocess failure
- **WHEN** the CLI subprocess exits with a non-zero code or times out
- **THEN** the system responds with HTTP 502, logs the error with full context, and does not create or overwrite any reflection row

### Requirement: Reflection persistence
The system SHALL persist the generated reflection into the `reflections` table using `Store.UpsertReflection()` AND save it as a markdown file named `YYYYMMDD-NNN.md` in the configured reflection output directory. Reflections in SQLite are never automatically pruned — they are retained indefinitely.

#### Scenario: Reflection stored successfully
- **WHEN** the CLI returns a valid reflection
- **THEN** the reflection is upserted into the `reflections` table, saved as a markdown file, and the HTTP response includes the stored reflection

### Requirement: Interaction pruning on reflection
The system SHALL optionally prune interactions older than the configured `retention_days` after a successful reflection. Pruning MUST only run after the reflection is persisted, never before.

#### Scenario: Pruning enabled after reflection
- **WHEN** the reflection job completes successfully and `retention_days` is configured (e.g., 30)
- **THEN** the system calls `Store.PruneInteractions(retentionDays)` to delete interactions older than the retention window

#### Scenario: Pruning disabled
- **WHEN** the reflection job completes and `retention_days` is set to 0 or not configured
- **THEN** no interactions are pruned
