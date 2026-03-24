## ADDED Requirements

### Requirement: CLI subprocess completer
The system SHALL implement an `LLMClient` that executes a configurable CLI command as a subprocess, pipes the prompt to stdin, and captures stdout as the response. The command MUST be configurable via the `COLLECTOR_REFLECTION_CLI` environment variable. Default: `claude --print`.

#### Scenario: Reflection via claude CLI
- **WHEN** `COLLECTOR_REFLECTION_CLI` is set to `claude --print` and a reflection is triggered
- **THEN** the system spawns `claude --print`, pipes the reflection prompt to stdin, and captures the stdout response

#### Scenario: Reflection via codex CLI
- **WHEN** `COLLECTOR_REFLECTION_CLI` is set to `codex --quiet`
- **THEN** the system spawns `codex --quiet` with the prompt piped to stdin

#### Scenario: Reflection via gemini CLI
- **WHEN** `COLLECTOR_REFLECTION_CLI` is set to `gemini`
- **THEN** the system spawns `gemini` with the prompt piped to stdin

#### Scenario: CLI subprocess timeout
- **WHEN** the CLI subprocess does not complete within 5 minutes
- **THEN** the process is killed and the reflection fails with a timeout error logged

#### Scenario: CLI not found
- **WHEN** the configured CLI binary is not found in PATH
- **THEN** the reflection fails with an error logged indicating the binary was not found

### Requirement: Reflection file output
The system SHALL save each successful reflection as a markdown file in the directory configured by `COLLECTOR_REFLECTION_DIR` (default: `./data/reflections/`). The filename format SHALL be `YYYYMMDD-NNN.md` where `YYYYMMDD` is the target date and `NNN` is a zero-padded sequence number starting at `001`, incrementing for each reflection on the same date.

#### Scenario: First reflection of the day
- **WHEN** a reflection completes for 2026-03-24 and no files with prefix `20260324-` exist
- **THEN** the file `20260324-001.md` is created in the reflection output directory

#### Scenario: Second reflection of the same day
- **WHEN** a reflection completes for 2026-03-24 and `20260324-001.md` already exists
- **THEN** the file `20260324-002.md` is created

#### Scenario: Reflection file content
- **WHEN** a reflection file is created
- **THEN** it contains YAML frontmatter (date, interaction count, provider) followed by the structured reflection sections (summary, should do, should not do, config changes)

### Requirement: Internal reflection scheduler
The system SHALL run a background scheduler that triggers reflection automatically. The schedule MUST be configurable via `COLLECTOR_REFLECTION_SCHEDULE` with values: `daily` (default — runs at 02:00 local time), `hourly`, or `off` (disabled).

#### Scenario: Daily scheduled reflection
- **WHEN** the collector is running with `COLLECTOR_REFLECTION_SCHEDULE=daily` and the clock reaches 02:00
- **THEN** a reflection is triggered for the previous day's interactions

#### Scenario: Skip if already reflected
- **WHEN** the scheduled reflection triggers and a reflection for the target date already exists in SQLite
- **THEN** the reflection is skipped

#### Scenario: Scheduler disabled
- **WHEN** `COLLECTOR_REFLECTION_SCHEDULE` is set to `off`
- **THEN** no automatic reflections are triggered (manual via HTTP endpoint still works)

#### Scenario: Hourly reflection
- **WHEN** `COLLECTOR_REFLECTION_SCHEDULE` is set to `hourly`
- **THEN** a reflection is triggered every hour for the current day's interactions so far
