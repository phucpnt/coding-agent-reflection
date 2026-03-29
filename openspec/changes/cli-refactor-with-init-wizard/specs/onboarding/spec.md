## MODIFIED Requirements

### Requirement: Makefile install target
The system SHALL provide a `make install` target that builds the single `ai-collector` binary and installs it to `~/.local/bin/` (or `$GOPATH/bin` via `go install`).

#### Scenario: Fresh install
- **WHEN** a user runs `make install`
- **THEN** the `ai-collector` binary is built and placed in `~/.local/bin/`

### Requirement: README documentation
The system SHALL include a `README.md` documenting: prerequisites, installation via `go install` or `make install`, first-run setup via `ai-collector init`, subcommand reference, and configuration options.

#### Scenario: New user follows README
- **WHEN** a new user follows the README steps
- **THEN** they run `go install` or `make install`, then `ai-collector init`, and have a working setup

## REMOVED Requirements

### Requirement: Makefile setup-claude target
**Reason**: Replaced by `ai-collector setup claude` subcommand.
**Migration**: Run `ai-collector setup claude` instead of `make setup-claude`.

### Requirement: Makefile setup-gemini target
**Reason**: Replaced by `ai-collector setup gemini` subcommand.
**Migration**: Run `ai-collector setup gemini` instead of `make setup-gemini`.

### Requirement: Makefile setup-codex target
**Reason**: Replaced by `ai-collector setup codex` subcommand.
**Migration**: Run `ai-collector setup codex` instead of `make setup-codex`.

### Requirement: Makefile run target
**Reason**: Replaced by `ai-collector start` (background) and `ai-collector serve` (foreground).
**Migration**: Run `ai-collector start` or `ai-collector serve`.

### Requirement: Makefile verify target
**Reason**: Replaced by `ai-collector status` subcommand.
**Migration**: Run `ai-collector status`.

### Requirement: Makefile reflect target
**Reason**: Replaced by `ai-collector reflect` subcommand.
**Migration**: Run `ai-collector reflect`.

### Requirement: Makefile status target
**Reason**: Replaced by `ai-collector status` subcommand.
**Migration**: Run `ai-collector status`.
