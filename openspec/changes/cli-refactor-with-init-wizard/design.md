## Context

The collector service works but has poor configuration UX: env vars only, two separate binaries, shell scripts for setup. Users expect a single binary with subcommands and a config file, similar to `gh`, `docker`, or `kubectl`.

Current codebase: `cmd/collector/main.go` (HTTP server), `cmd/query/main.go` (query tool), `config/config.go` (env var loader), `scripts/setup-*.sh` (hook installers).

## Goals / Non-Goals

**Goals:**
- Single `ai-collector` binary with cobra subcommands
- YAML config file as primary config, env vars as overrides
- Interactive init wizard for first-run
- Hook setup via Go code (replaces shell scripts)
- PID-file based start/stop for background process management
- XDG-compliant default paths (`~/.config/`, `~/.local/share/`)

**Non-Goals:**
- systemd/launchd service files (future improvement)
- Remote/networked operation
- Plugin system for custom providers
- GUI or TUI beyond the init wizard prompts

## Decisions

### 1. Cobra for CLI framework

**Choice**: Use `github.com/spf13/cobra` for subcommand routing and flag parsing.

**Rationale**: De facto standard for Go CLIs. Used by kubectl, gh, docker. Provides subcommands, flags, help text, shell completion for free. No need to build custom argument parsing.

**Alternatives considered**:
- `urfave/cli`: Also good, but less ecosystem adoption.
- stdlib `flag`: No subcommand support without manual routing.

### 2. YAML config with manual parsing (no viper)

**Choice**: Use `gopkg.in/yaml.v3` directly. Load from `~/.config/ai-collector/config.yaml`, override with env vars, override with CLI flags.

**Rationale**: Viper is heavy and pulls in many transitive dependencies for features we don't need (remote config, watching, multiple formats). A single YAML file with ~10 fields is trivially parsed with `yaml.v3`. The override chain (file → env → flags) is simple to implement manually.

**Alternatives considered**:
- Viper: Too heavy for 10 config fields. Adds 15+ transitive deps.
- TOML: Less familiar to most users than YAML.
- JSON: No comments, poor for config files.

### 3. Config file location — XDG compliant

**Choice**:
- Config: `~/.config/ai-collector/config.yaml`
- Data: `~/.local/share/ai-collector/` (DB, interactions files)
- Reflections: configurable, default `~/.local/share/ai-collector/reflections/`

**Rationale**: XDG Base Directory spec is the standard on Linux. On macOS, `~/.config/` works fine too. Keeps data out of the project directory, making the tool work across all projects.

**Alternatives considered**:
- `~/.ai-collector/`: Pollutes home directory.
- Project-local `./data/`: Doesn't work when used across multiple projects.

### 4. PID file for start/stop

**Choice**: `ai-collector start` forks the server process, writes PID to `~/.local/share/ai-collector/collector.pid`. `ai-collector stop` reads PID and sends SIGTERM.

**Rationale**: Simplest cross-platform approach. No dependency on systemd/launchd. The user can see the process, kill it manually, or use our subcommand.

Implementation: `start` runs `ai-collector serve` (the actual HTTP server) as a detached subprocess via `os/exec`, redirecting stdout/stderr to a log file.

**Alternatives considered**:
- Daemonize in-process: Complex, not idiomatic Go.
- systemd/launchd only: Not cross-platform, requires root for system-level.

### 5. Subcommand structure

```
ai-collector
├── init              — interactive first-run wizard
├── start             — start collector in background (writes PID)
├── stop              — stop background collector (reads PID, sends SIGTERM)
├── serve             — run collector in foreground (used by start internally)
├── status            — health check + recent interactions summary
├── reflect           — trigger reflection manually
│   └── --date        — specific date (default: today)
├── query             — query interactions
│   └── today|week|all|reflections|stats
├── setup             — configure hooks
│   └── claude|gemini|codex
│       └── --global/--project
├── config            — view/edit config
│   ├── (no args)     — show current config
│   └── set KEY VALUE — update a config value
└── version           — show version
```

`serve` is the internal command that runs the HTTP server — what `cmd/collector/main.go` does today. `start` is the user-facing command that launches `serve` in the background.

### 6. Init wizard flow

```
$ ai-collector init

Welcome to AI Interaction Collector!

Which CLI do you use for reflections?
  1) claude --print (default)
  2) codex --quiet
  3) gemini
  > 1

Where should reflections be saved?
  [~/.local/share/ai-collector/reflections]:

Collector port?
  [19321]:

Set up hooks for which providers?
  [x] Claude Code
  [ ] Gemini CLI
  [ ] Codex
  > 1

Install hooks globally or for this project?
  1) global (all sessions)
  2) project (this directory only)
  > 1

✓ Config written to ~/.config/ai-collector/config.yaml
✓ Claude Code hooks configured (global)
✓ Data directory created at ~/.local/share/ai-collector/

Start the collector now? [Y/n]: y
✓ Collector started (PID 12345)

Run `ai-collector status` to verify.
```

### 7. Project layout after refactor

```
cmd/ai-collector/
  main.go              — cobra root command
  cmd_init.go          — init wizard
  cmd_start.go         — start/stop/serve
  cmd_status.go        — status
  cmd_reflect.go       — reflect
  cmd_query.go         — query
  cmd_setup.go         — setup <provider>
  cmd_config.go        — config view/set
internal/
  config/              — YAML config loader (moved from config/)
    config.go
  ingest/              — unchanged
  model/               — unchanged
  reflection/          — unchanged
  setup/               — hook configuration logic (replaces scripts)
    claude.go
    gemini.go
    codex.go
  store/               — unchanged
```

## Risks / Trade-offs

- **[Breaking change]** Users with existing env-var configs need to run `ai-collector init` or manually create a YAML file. → Mitigation: env vars still work as overrides, so existing setups don't break immediately. Init wizard detects existing env vars and pre-fills.

- **[PID file stale]** If the collector crashes without cleanup, the PID file is stale. → Mitigation: `start` checks if the PID in the file is actually running before deciding to start a new instance.

- **[cobra dependency]** Adds ~5 transitive deps. → Mitigation: cobra is extremely stable and well-maintained. The trade-off is worth it for the CLI UX.

- **[Shell script removal]** Users who scripted against `setup-claude.sh` break. → Mitigation: Keep scripts as thin wrappers that call `ai-collector setup claude` for one release cycle, then remove.

## Open Questions

- Should `ai-collector` be the binary name or something shorter like `aic`? Lean towards `ai-collector` for clarity, with an alias option.
- Should the init wizard support non-interactive mode (`ai-collector init --defaults`) for CI/scripting?
