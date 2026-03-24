## MODIFIED Requirements

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

## REMOVED Requirements

### Requirement: Claude API HTTP client
**Reason**: Replaced by CLI subprocess execution. Users no longer need an `ANTHROPIC_API_KEY`.
**Migration**: Set `COLLECTOR_REFLECTION_CLI` to the desired CLI command (default: `claude --print`). Remove `ANTHROPIC_API_KEY` from environment.
