## 1. Dependencies and Project Structure

- [x] 1.1 Add `cobra` and `yaml.v3` dependencies: `go get github.com/spf13/cobra gopkg.in/yaml.v3`
- [x] 1.2 Create `cmd/ai-collector/` directory structure
- [x] 1.3 Move `config/config.go` to `internal/config/config.go`, rewrite to load YAML with env var fallback and XDG default paths

## 2. YAML Config

- [x] 2.1 Define config struct with YAML tags and default values. Fields: port, db_path, retention_days, reflection.cli, reflection.schedule, reflection.output_dir, reflection.prompt
- [x] 2.2 Implement `Load()` — read YAML file from `~/.config/ai-collector/config.yaml`, apply env var overrides
- [x] 2.3 Implement `Save(cfg)` — write config struct to YAML file, creating directory if needed
- [x] 2.4 Implement `ConfigDir()`, `DataDir()` helpers using XDG paths with `$HOME` fallback
- [x] 2.5 Write tests for config load/save/override

## 3. Cobra Root and Subcommands

- [x] 3.1 Create `cmd/ai-collector/main.go` — cobra root command with version flag
- [x] 3.2 Create `cmd/ai-collector/cmd_serve.go` — foreground HTTP server (migrate from `cmd/collector/main.go`)
- [x] 3.3 Create `cmd/ai-collector/cmd_start.go` — launch `serve` as background process, write PID file, redirect logs
- [x] 3.4 Create `cmd/ai-collector/cmd_stop.go` — read PID file, verify process, send SIGTERM, remove PID file
- [x] 3.5 Create `cmd/ai-collector/cmd_status.go` — check PID file, call `/health`, show recent interaction count
- [x] 3.6 Create `cmd/ai-collector/cmd_reflect.go` — POST to `/jobs/daily-reflection` with optional `--date` flag
- [x] 3.7 Create `cmd/ai-collector/cmd_query.go` — merge `cmd/query/main.go` logic, accept `today|week|all|reflections|stats`
- [x] 3.8 Create `cmd/ai-collector/cmd_config.go` — `config` shows current config, `config set KEY VALUE` updates YAML file
- [x] 3.9 Create `cmd/ai-collector/cmd_version.go` — print version, Go version, build date using ldflags

## 4. Setup Subcommand

- [x] 4.1 Create `internal/setup/claude.go` — port shell script logic to Go: read settings.json, merge hooks with `encoding/json`, backup original
- [x] 4.2 Create `internal/setup/gemini.go` — same pattern for Gemini CLI config
- [x] 4.3 Create `internal/setup/codex.go` — same pattern for Codex config
- [x] 4.4 Create `cmd/ai-collector/cmd_setup.go` — `setup claude|gemini|codex` with `--global`/`--project` flags, interactive prompt if neither flag given
- [x] 4.5 Write tests for setup logic using temp directories

## 5. Init Wizard

- [x] 5.1 Create `cmd/ai-collector/cmd_init.go` — interactive wizard: prompt for reflection CLI, output dir, port, providers, hook scope
- [x] 5.2 Implement `--defaults` flag for non-interactive mode
- [x] 5.3 Wire init to call config.Save(), setup logic, directory creation, and optionally `start`
- [x] 5.4 Handle existing config: pre-fill current values, allow modification

## 6. Cleanup

- [x] 6.1 Remove `cmd/collector/` and `cmd/query/` directories
- [x] 6.2 Remove old `config/` package (now at `internal/config/`)
- [x] 6.3 Update `Makefile` — keep `install` (builds `ai-collector` to `~/.local/bin/`) and `test` targets only
- [x] 6.4 Update `README.md` — new installation flow, subcommand reference, config file docs
- [x] 6.5 Update `.github/workflows/release.yml` — build `ai-collector` instead of `collector`+`query`
- [x] 6.6 Update hook scripts to use new default paths and port
- [x] 6.7 Run full test suite, fix any breakage
