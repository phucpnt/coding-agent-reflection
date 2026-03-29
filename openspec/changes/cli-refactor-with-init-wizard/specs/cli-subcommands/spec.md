## ADDED Requirements

### Requirement: Single binary with cobra subcommands
The system SHALL provide a single `ai-collector` binary with subcommands: `init`, `start`, `stop`, `serve`, `status`, `reflect`, `query`, `setup`, `config`, `version`. Unknown subcommands SHALL print help text.

#### Scenario: Run with no subcommand
- **WHEN** the user runs `ai-collector` with no arguments
- **THEN** the system prints help text listing all available subcommands

#### Scenario: Run unknown subcommand
- **WHEN** the user runs `ai-collector unknown`
- **THEN** the system prints an error and help text

### Requirement: Start collector in background
The `start` subcommand SHALL launch the collector HTTP server as a background process, write the PID to `~/.local/share/ai-collector/collector.pid`, and redirect logs to `~/.local/share/ai-collector/collector.log`.

#### Scenario: Start when not running
- **WHEN** the user runs `ai-collector start` and no collector is running
- **THEN** the collector starts in the background, PID file is written, and the command prints the PID

#### Scenario: Start when already running
- **WHEN** the user runs `ai-collector start` and a collector is already running (valid PID in PID file)
- **THEN** the command prints a message indicating the collector is already running with its PID

### Requirement: Stop collector
The `stop` subcommand SHALL read the PID file, verify the process is running, and send SIGTERM.

#### Scenario: Stop running collector
- **WHEN** the user runs `ai-collector stop` and the collector is running
- **THEN** SIGTERM is sent, the process exits gracefully, and the PID file is removed

#### Scenario: Stop when not running
- **WHEN** the user runs `ai-collector stop` and no collector is running
- **THEN** the command prints a message indicating no collector is running

### Requirement: Serve in foreground
The `serve` subcommand SHALL run the collector HTTP server in the foreground (same behavior as the current `cmd/collector/main.go`). This is used internally by `start` and directly by users who want foreground operation.

#### Scenario: Serve in foreground
- **WHEN** the user runs `ai-collector serve`
- **THEN** the collector starts in the foreground, logging to stdout, and blocks until interrupted

### Requirement: Status subcommand
The `status` subcommand SHALL check if the collector is running, call the health endpoint, and display a summary of recent interactions.

#### Scenario: Status when running
- **WHEN** the user runs `ai-collector status` and the collector is running
- **THEN** the output shows collector health, port, and count of recent interactions

#### Scenario: Status when not running
- **WHEN** the user runs `ai-collector status` and the collector is not running
- **THEN** the output indicates the collector is not running

### Requirement: Reflect subcommand
The `reflect` subcommand SHALL trigger a reflection for today (default) or a specific date via `--date` flag by calling the collector's HTTP endpoint.

#### Scenario: Trigger reflection
- **WHEN** the user runs `ai-collector reflect`
- **THEN** a reflection is triggered for today and the summary is displayed

#### Scenario: Trigger reflection for specific date
- **WHEN** the user runs `ai-collector reflect --date 2026-03-27`
- **THEN** a reflection is triggered for the specified date

### Requirement: Query subcommand
The `query` subcommand SHALL query and display interactions. It SHALL accept subarguments: `today`, `week`, `all`, `reflections`, `stats`.

#### Scenario: Query today
- **WHEN** the user runs `ai-collector query today`
- **THEN** today's interactions are displayed

#### Scenario: Query with no argument
- **WHEN** the user runs `ai-collector query`
- **THEN** today's interactions are displayed (default)

### Requirement: Setup subcommand
The `setup` subcommand SHALL configure hooks for a specified provider (`claude`, `gemini`, `codex`). It SHALL accept `--global` and `--project` flags. If neither flag is provided, it SHALL prompt the user interactively.

#### Scenario: Setup claude globally
- **WHEN** the user runs `ai-collector setup claude --global`
- **THEN** Claude Code hooks are merged into `~/.claude/settings.json`

#### Scenario: Setup claude with prompt
- **WHEN** the user runs `ai-collector setup claude` with no flags
- **THEN** the user is prompted to choose global or project scope

### Requirement: Version subcommand
The `version` subcommand SHALL print the binary version, Go version, and build date.

#### Scenario: Show version
- **WHEN** the user runs `ai-collector version`
- **THEN** version information is printed
