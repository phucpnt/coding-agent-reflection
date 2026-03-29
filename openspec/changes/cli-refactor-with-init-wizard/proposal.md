## Why

The app currently requires users to set multiple env vars, run separate binaries (`collector`, `query`), and execute shell scripts for setup. This is friction-heavy for a tool that should be invisible after initial setup. A single binary with subcommands, a YAML config file, and an interactive init wizard brings the UX to the level users expect from modern CLI tools (`gh`, `docker`, `kubectl`).

## What Changes

- **Single `ai-collector` binary** — merge `cmd/collector/` and `cmd/query/` into one binary with cobra subcommands
- **Subcommands**: `init`, `start`, `stop`, `status`, `reflect`, `query`, `setup`, `config`
- **YAML config file** at `~/.config/ai-collector/config.yaml` — replaces env vars as primary config. Env vars still work as overrides.
- **Interactive `init` wizard** — first-run setup that asks for reflection CLI, output directory, hook scope, and installs the service
- **`setup` subcommand** — replaces shell scripts for hook configuration (`ai-collector setup claude`, `ai-collector setup gemini`, `ai-collector setup codex`)
- **`config` subcommand** — view and edit config (`ai-collector config`, `ai-collector config set reflection.cli gemini`)
- **`start`/`stop`** — manage the collector as a background process with PID file
- **Remove separate `cmd/query/` binary** — its functionality moves to `ai-collector query`
- **New dependency**: `cobra` for CLI framework, `viper` or manual YAML parsing for config

## Capabilities

### New Capabilities

- `cli-subcommands`: Single binary with cobra subcommands (init, start, stop, status, reflect, query, setup, config). PID-file based process management for start/stop.
- `yaml-config`: YAML config file at `~/.config/ai-collector/config.yaml`. Loaded on startup with env var overrides. `config` subcommand for viewing and editing.
- `init-wizard`: Interactive first-run setup that walks the user through choosing reflection CLI, output directory, hook installation scope (global/project), and starts the collector.

### Modified Capabilities

- `onboarding`: Makefile targets (`setup-claude`, `setup-gemini`, `setup-codex`) replaced by `ai-collector setup <provider>` subcommand. `make install` replaced by `go install` or binary download. `make run` replaced by `ai-collector start`.

## Impact

- **`cmd/collector/`** — replaced by `cmd/ai-collector/` with cobra root command
- **`cmd/query/`** — removed, merged into `ai-collector query` subcommand
- **`config/config.go`** — rewritten to load from YAML file with env var fallback
- **`scripts/setup-*.sh`** — logic moves into Go code under `internal/setup/`
- **`Makefile`** — simplified (just `install` and `test`), most targets become subcommands
- **New dependencies**: `github.com/spf13/cobra`
- **Data location** moves to `~/.local/share/ai-collector/` by default (XDG compliant)
