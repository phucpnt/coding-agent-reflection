## ADDED Requirements

### Requirement: Interactions table schema
The system SHALL create an `interactions` table on startup if it does not exist, with columns: `id` (UUID), `ts` (TIMESTAMP), `provider` (TEXT), `session_id` (TEXT), `project` (TEXT), `user_prompt` (TEXT), `agent_output` (TEXT), `context` (TEXT), `tokens_prompt` (INT, nullable), `tokens_output` (INT, nullable), `tools_used` (TEXT, nullable).

#### Scenario: First startup with no existing database
- **WHEN** the service starts and the DuckDB file does not exist
- **THEN** the system creates the file and the `interactions` table with the specified schema

#### Scenario: Startup with existing database
- **WHEN** the service starts and the DuckDB file already contains the `interactions` table
- **THEN** the system proceeds without modifying the existing table

### Requirement: Reflections table schema
The system SHALL create a `reflections` table on startup if it does not exist, with columns: `id` (UUID), `date` (DATE, unique), `summary` (TEXT), `should_do` (TEXT), `should_not_do` (TEXT), `config_changes` (TEXT), `created_at` (TIMESTAMP).

#### Scenario: Reflections table created on first startup
- **WHEN** the service starts and the `reflections` table does not exist
- **THEN** the system creates the table with the specified schema

### Requirement: Insert interaction
The system SHALL insert an `Interaction` struct into the `interactions` table as a single row. Inserts MUST be serialized via a write mutex to prevent DuckDB concurrent-write errors.

#### Scenario: Successful insert
- **WHEN** `Store.Insert()` is called with a valid `Interaction`
- **THEN** a new row appears in the `interactions` table with all provided field values

#### Scenario: Concurrent inserts
- **WHEN** two `Store.Insert()` calls arrive simultaneously
- **THEN** both complete successfully — one waits for the mutex, then inserts after the first completes

### Requirement: Query interactions by date range
The system SHALL support querying interactions filtered by a date range (`from` and `to` timestamps), ordered by `ts` ascending. This is used by the daily reflection job.

#### Scenario: Query yesterday's interactions
- **WHEN** `Store.QueryByDateRange(from, to)` is called with yesterday's start and end timestamps
- **THEN** the system returns all interactions with `ts` between `from` and `to`, ordered by `ts` ascending

#### Scenario: No interactions in range
- **WHEN** `Store.QueryByDateRange(from, to)` is called for a date range with no data
- **THEN** the system returns an empty slice and no error

### Requirement: Upsert reflection
The system SHALL support upserting a reflection row keyed by `date`. If a reflection for the given date already exists, it SHALL be overwritten.

#### Scenario: First reflection for a date
- **WHEN** `Store.UpsertReflection()` is called for a date with no existing reflection
- **THEN** a new row is inserted into the `reflections` table

#### Scenario: Re-running reflection for same date
- **WHEN** `Store.UpsertReflection()` is called for a date that already has a reflection
- **THEN** the existing row is replaced with the new data

### Requirement: Prune old interactions
The system SHALL support deleting interactions older than a configurable number of days. This is used by the data lifecycle job.

#### Scenario: Prune interactions older than retention period
- **WHEN** `Store.PruneInteractions(retentionDays)` is called with a retention of 30 days
- **THEN** all rows in `interactions` with `ts` older than 30 days ago are deleted, and newer rows are untouched

#### Scenario: No interactions to prune
- **WHEN** `Store.PruneInteractions(retentionDays)` is called and all interactions are within the retention window
- **THEN** no rows are deleted
