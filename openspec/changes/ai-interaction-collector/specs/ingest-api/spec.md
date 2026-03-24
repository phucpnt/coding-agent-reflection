## ADDED Requirements

### Requirement: Claude hook ingest endpoint
The system SHALL expose `POST /ingest/claude` that accepts JSON payloads from Claude Code HTTP hooks. The handler SHALL extract `session_id`, `cwd`, `user_prompt`, `agent_output`, and `tools_used` from the hook payload and normalize them into an `Interaction` struct. Unknown fields SHALL be ignored without error.

#### Scenario: Valid Claude hook payload
- **WHEN** a POST request arrives at `/ingest/claude` with a valid JSON body containing `session_id`, `cwd`, and transcript data
- **THEN** the system inserts one row into the `interactions` table with `provider` set to `"claude"`, populates all extractable fields, and responds with HTTP 200

#### Scenario: Claude payload with missing optional fields
- **WHEN** a POST request arrives at `/ingest/claude` with a valid JSON body but `tools_used` or `agent_output` is absent
- **THEN** the system inserts a row with those fields set to null and responds with HTTP 200

#### Scenario: Malformed Claude payload
- **WHEN** a POST request arrives at `/ingest/claude` with invalid JSON or missing required fields (`session_id`)
- **THEN** the system responds with HTTP 400 and a JSON error body describing the validation failure

### Requirement: Gemini hook ingest endpoint
The system SHALL expose `POST /ingest/gemini` that accepts JSON payloads from Gemini CLI HTTP hooks. The handler SHALL extract `session_id`, `cwd`, `user_prompt`, `agent_output`, and `tools_used` from the hook payload and normalize them into an `Interaction` struct. Unknown fields SHALL be ignored without error.

#### Scenario: Valid Gemini hook payload
- **WHEN** a POST request arrives at `/ingest/gemini` with a valid JSON body containing session info and final agent output
- **THEN** the system inserts one row into the `interactions` table with `provider` set to `"gemini"` and responds with HTTP 200

#### Scenario: Malformed Gemini payload
- **WHEN** a POST request arrives at `/ingest/gemini` with invalid JSON or missing required fields
- **THEN** the system responds with HTTP 400 and a JSON error body

### Requirement: Codex OTLP ingest endpoint
The system SHALL expose `POST /v1/traces` that accepts OTLP HTTP trace payloads (JSON-encoded). The handler SHALL iterate over resource spans, extract attributes (`user_prompt`, `completion`, `code.filepath`, `code.function`, token counts), and normalize each relevant span into an `Interaction` struct with `provider` set to `"codex"`. Unrecognized span types SHALL be logged at debug level and skipped.

#### Scenario: Valid OTLP trace with LLM span
- **WHEN** a POST request arrives at `/v1/traces` with an OTLP JSON body containing a span with `user_prompt` and `completion` attributes
- **THEN** the system inserts one row into the `interactions` table with `provider` set to `"codex"` and responds with HTTP 200

#### Scenario: OTLP trace with no relevant spans
- **WHEN** a POST request arrives at `/v1/traces` with spans that contain no recognized LLM attributes
- **THEN** the system skips all spans, inserts nothing, and responds with HTTP 200

#### Scenario: Malformed OTLP payload
- **WHEN** a POST request arrives at `/v1/traces` with a body that does not conform to OTLP JSON structure
- **THEN** the system responds with HTTP 400

### Requirement: Unified Interaction normalization
All ingest handlers SHALL produce a common `Interaction` struct with the following fields: `id` (UUID, generated), `ts` (timestamp, set at ingest time), `provider` (string), `session_id` (string), `project` (string, derived from cwd or repo path), `user_prompt` (string), `agent_output` (string), `context` (JSON string), `tokens_prompt` (int, nullable), `tokens_output` (int, nullable), `tools_used` (JSON string, nullable). The struct SHALL be passed to the storage layer for insertion.

#### Scenario: All providers produce identical struct shape
- **WHEN** interactions are ingested from Claude, Gemini, and Codex
- **THEN** all three produce `Interaction` structs with the same set of fields, differing only in `provider` value and field population

### Requirement: Health check endpoint
The system SHALL expose `GET /health` that returns HTTP 200 with `{"status": "ok"}` when the service is running and can reach DuckDB.

#### Scenario: Service is healthy
- **WHEN** a GET request arrives at `/health` and DuckDB is accessible
- **THEN** the system responds with HTTP 200 and `{"status": "ok"}`

#### Scenario: DuckDB is unreachable
- **WHEN** a GET request arrives at `/health` and DuckDB cannot be opened
- **THEN** the system responds with HTTP 503 and `{"status": "error", "message": "..."}`
