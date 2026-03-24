## 1. CLI Subprocess Completer

- [x] 1.1 Create `internal/reflection/cli.go` — `CLICompleter` struct implementing `LLMClient` interface. Uses `exec.CommandContext` to spawn the configured command, pipes prompt to stdin, captures stdout. 5-minute timeout via context.
- [x] 1.2 Update `config/config.go` — replace `LLMAPIKey`/`LLMModel` with `ReflectionCLI` (default: `claude --print`), `ReflectionSchedule` (default: `daily`), `ReflectionOutputDir` (default: `./data/reflections/`)
- [x] 1.3 Delete `internal/reflection/llm.go` — remove Claude API HTTP client
- [x] 1.4 Write tests for `CLICompleter` using a mock command (e.g., `echo`)

## 2. Reflection File Output

- [x] 2.1 Create `internal/reflection/filewriter.go` — `WriteReflectionFile(dir, targetDate, reflection)` that scans for existing `YYYYMMDD-*` files, determines next sequence number, writes `YYYYMMDD-NNN.md` with YAML frontmatter and structured sections
- [x] 2.2 Update `internal/reflection/job.go` — after `UpsertReflection`, call `WriteReflectionFile` to persist to disk
- [x] 2.3 Write tests for file writer — first file gets `001`, second gets `002`, content has frontmatter

## 3. Internal Reflection Scheduler

- [x] 3.1 Create `internal/reflection/scheduler.go` — background goroutine that ticks based on `ReflectionSchedule` config. For `daily`: compute next 02:00 local, sleep until then, trigger. For `hourly`: tick every hour. For `off`: no-op. Skips if reflection for target date already exists.
- [x] 3.2 Update `cmd/collector/main.go` — replace `ClaudeClient` with `CLICompleter`, start scheduler goroutine, auto-create `data/` and reflection output dir on startup
- [x] 3.3 Write tests for scheduler logic (next-tick calculation, skip-if-exists)

## 4. Setup Scripts

- [x] 4.1 Create `scripts/setup-claude.sh` — check for `jq`, back up existing settings, merge `UserPromptSubmit` and `Stop` hooks into `~/.claude/settings.json` using jq. Use absolute path to hook scripts.
- [x] 4.2 Create `scripts/setup-gemini.sh` — similar merge for Gemini CLI hook config
- [x] 4.3 Create `scripts/setup-codex.sh` — configure Codex OTel export to collector
- [x] 4.4 Create `scripts/verify.sh` — check collector health (`curl /health`), check configured CLI in PATH, check hook files contain collector URLs

## 5. Makefile

- [x] 5.1 Create `Makefile` with targets: `install`, `run`, `setup-claude`, `setup-gemini`, `setup-codex`, `reflect`, `verify`, `status`
- [x] 5.2 `install` target builds `collector` and `query` binaries
- [x] 5.3 `run` target starts collector in foreground
- [x] 5.4 `reflect` target calls `curl -X POST http://localhost:9000/jobs/daily-reflection`
- [x] 5.5 `status` target calls `curl http://localhost:9000/interactions` with jq formatting
- [x] 5.6 Setup targets delegate to respective scripts in `scripts/`
- [x] 5.7 `verify` target delegates to `scripts/verify.sh`

## 6. README

- [x] 6.1 Create `README.md` — prerequisites (Go, jq, at least one coding CLI), quick start (`make install && make setup-claude && make run`), configuration (env vars table), usage (query, reflect, verify), architecture overview

## 7. Cleanup and Testing

- [x] 7.1 Update `internal/reflection/handler.go` — ensure HTTP endpoint still works with `CLICompleter`, add file write after reflection
- [x] 7.2 Update existing reflection tests to use `CLICompleter` mock instead of API mock
- [x] 7.3 Run full test suite and fix any breakage from config/LLM client changes
