## ADDED Requirements

### Requirement: YAML config file
The system SHALL load configuration from `~/.config/ai-collector/config.yaml`. If the file does not exist, the system SHALL use built-in defaults. Environment variables SHALL override config file values. CLI flags SHALL override both.

#### Scenario: Config file exists
- **WHEN** the collector starts and `~/.config/ai-collector/config.yaml` exists
- **THEN** configuration is loaded from the file

#### Scenario: Config file missing
- **WHEN** the collector starts and no config file exists
- **THEN** built-in defaults are used (port 19321, reflection CLI `claude --print`, etc.)

#### Scenario: Env var overrides config file
- **WHEN** the config file sets `port: 19321` and `COLLECTOR_PORT=8080` is set
- **THEN** the collector uses port 8080

### Requirement: Config view subcommand
The `config` subcommand with no arguments SHALL print the current effective configuration (merged from file + env + defaults).

#### Scenario: View config
- **WHEN** the user runs `ai-collector config`
- **THEN** the current configuration is printed in YAML format

### Requirement: Config set subcommand
The `config set` subcommand SHALL update a single config value in the YAML file. It SHALL create the config file and directory if they do not exist.

#### Scenario: Set a config value
- **WHEN** the user runs `ai-collector config set reflection.cli "gemini"`
- **THEN** the config file is updated with the new value

#### Scenario: Set creates config file
- **WHEN** the user runs `ai-collector config set port 8080` and no config file exists
- **THEN** the config file is created at `~/.config/ai-collector/config.yaml` with the specified value

### Requirement: XDG-compliant default paths
The system SHALL use XDG Base Directory paths by default: config in `~/.config/ai-collector/`, data in `~/.local/share/ai-collector/`. These paths SHALL be overridable via config or env vars.

#### Scenario: Default data directory
- **WHEN** no custom paths are configured
- **THEN** the database is stored at `~/.local/share/ai-collector/ai_interactions.db` and reflections at `~/.local/share/ai-collector/reflections/`
