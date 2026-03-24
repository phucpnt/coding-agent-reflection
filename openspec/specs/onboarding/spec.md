## ADDED Requirements

### Requirement: Makefile install target
The system SHALL provide a `make install` target that builds the `collector` and `query` binaries into the project root.

#### Scenario: Fresh install
- **WHEN** a user runs `make install` after cloning the repo
- **THEN** the `collector` and `query` binaries are built and placed in the project root

#### Scenario: Install with missing Go
- **WHEN** a user runs `make install` without Go installed
- **THEN** the build fails with a clear error message indicating Go is required

### Requirement: Makefile setup-claude target
The system SHALL provide a `make setup-claude` target that runs a script to inject collector hooks into the user's Claude Code settings. The script MUST back up the existing settings file before modifying it. The script MUST merge hooks without overwriting existing hook entries. The script SHALL require `jq` and exit with a clear message if it is not installed.

#### Scenario: Setup Claude with no existing hooks
- **WHEN** a user runs `make setup-claude` and their Claude settings have no existing hooks
- **THEN** the `UserPromptSubmit` and `Stop` hooks are added to the settings file

#### Scenario: Setup Claude with existing hooks
- **WHEN** a user runs `make setup-claude` and their Claude settings already have other hooks (e.g., agent-deck)
- **THEN** the collector hooks are merged alongside existing hooks without removing them

#### Scenario: Setup Claude creates backup
- **WHEN** a user runs `make setup-claude`
- **THEN** the original settings file is backed up to `settings.json.bak` before modification

#### Scenario: jq not installed
- **WHEN** a user runs `make setup-claude` without `jq` installed
- **THEN** the script exits with an error message and instructions to install `jq`

### Requirement: Makefile setup-gemini target
The system SHALL provide a `make setup-gemini` target that injects collector hooks into the Gemini CLI configuration.

#### Scenario: Setup Gemini hooks
- **WHEN** a user runs `make setup-gemini`
- **THEN** the Gemini CLI hook configuration is updated to send events to the collector

### Requirement: Makefile setup-codex target
The system SHALL provide a `make setup-codex` target that configures Codex OTel export to point at the collector.

#### Scenario: Setup Codex OTel
- **WHEN** a user runs `make setup-codex`
- **THEN** the Codex configuration is updated to export OTel traces to the collector

### Requirement: Makefile run target
The system SHALL provide a `make run` target that starts the collector in the foreground.

#### Scenario: Run collector
- **WHEN** a user runs `make run`
- **THEN** the collector starts on the configured port and logs to stdout

### Requirement: Makefile verify target
The system SHALL provide a `make verify` target that checks the collector is running, hooks are configured, and the reflection CLI is available in PATH.

#### Scenario: Verify healthy setup
- **WHEN** a user runs `make verify` with the collector running and hooks configured
- **THEN** the output shows all checks passing

#### Scenario: Verify collector not running
- **WHEN** a user runs `make verify` and the collector is not running
- **THEN** the output clearly indicates the collector health check failed

#### Scenario: Verify missing CLI
- **WHEN** a user runs `make verify` and the configured reflection CLI is not in PATH
- **THEN** the output warns that the reflection CLI is not found

### Requirement: Makefile reflect target
The system SHALL provide a `make reflect` target that triggers a reflection manually by calling `POST /jobs/daily-reflection` on the running collector.

#### Scenario: Manual reflection trigger
- **WHEN** a user runs `make reflect` with the collector running
- **THEN** the reflection job runs for today and the result is displayed

### Requirement: Makefile status target
The system SHALL provide a `make status` target that queries and displays recent interactions from the running collector.

#### Scenario: Show status
- **WHEN** a user runs `make status`
- **THEN** recent interactions are displayed in a summary format

### Requirement: Auto-create data directory
The collector SHALL create the `data/` directory and the reflection output directory on startup if they do not exist.

#### Scenario: First startup with no data directory
- **WHEN** the collector starts and `./data/` does not exist
- **THEN** the directory is created automatically before opening the database

### Requirement: README documentation
The system SHALL include a `README.md` documenting: prerequisites, installation, setup for each CLI, running the collector, verifying the setup, querying interactions, and triggering reflections.

#### Scenario: New user follows README
- **WHEN** a new user follows the README steps in order
- **THEN** they have a working collector with hooks configured for at least one CLI
