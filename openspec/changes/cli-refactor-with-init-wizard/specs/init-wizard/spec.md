## ADDED Requirements

### Requirement: Interactive init wizard
The `init` subcommand SHALL walk the user through first-run setup interactively. It SHALL prompt for: reflection CLI choice, reflection output directory, collector port, provider hook setup, and hook scope (global/project).

#### Scenario: First-run init
- **WHEN** the user runs `ai-collector init` with no existing config
- **THEN** the wizard prompts for each setting, writes the config file, configures hooks, creates data directories, and optionally starts the collector

#### Scenario: Init with existing config
- **WHEN** the user runs `ai-collector init` and a config file already exists
- **THEN** the wizard pre-fills current values and allows the user to modify them

### Requirement: Init writes config file
The init wizard SHALL write all user choices to `~/.config/ai-collector/config.yaml` upon completion.

#### Scenario: Config written after init
- **WHEN** the user completes the init wizard
- **THEN** `~/.config/ai-collector/config.yaml` exists with the chosen values

### Requirement: Init configures hooks
The init wizard SHALL offer to configure hooks for selected providers and execute the setup logic for each.

#### Scenario: Init with Claude hooks
- **WHEN** the user selects Claude Code during init and chooses global scope
- **THEN** Claude Code hooks are configured in `~/.claude/settings.json`

### Requirement: Init creates data directories
The init wizard SHALL create the data directory and reflection output directory if they do not exist.

#### Scenario: Directories created
- **WHEN** the user completes init
- **THEN** `~/.local/share/ai-collector/` and the reflection output directory exist

### Requirement: Init offers to start collector
After completing setup, the init wizard SHALL ask if the user wants to start the collector immediately.

#### Scenario: Start after init
- **WHEN** the user answers yes to "Start the collector now?"
- **THEN** the collector starts in the background and the PID is displayed

#### Scenario: Skip start after init
- **WHEN** the user answers no to "Start the collector now?"
- **THEN** the wizard exits with a message to run `ai-collector start` later

### Requirement: Non-interactive init
The `init` subcommand SHALL accept a `--defaults` flag that skips all prompts and uses default values.

#### Scenario: Init with defaults
- **WHEN** the user runs `ai-collector init --defaults`
- **THEN** config is written with all default values, no prompts are shown
