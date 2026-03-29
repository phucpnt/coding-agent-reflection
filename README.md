# AI Interaction Collector

Capture interactions from coding agents (Claude Code, Gemini CLI, Codex) into SQLite and generate daily self-reflections.

## Quick Start

```bash
# Install
go install github.com/phuc/coding-agent-reflection/cmd/ai-collector@latest

# First-run setup (interactive wizard)
ai-collector init

# Or with all defaults
ai-collector init --defaults
```

The init wizard will:
1. Ask which CLI to use for reflections (claude/gemini/codex)
2. Configure reflection output directory and port
3. Set up hooks for your coding CLIs
4. Optionally start the collector

## Commands

```bash
ai-collector init                    # Interactive setup wizard
ai-collector start                   # Start collector in background
ai-collector stop                    # Stop background collector
ai-collector serve                   # Run collector in foreground (dev)
ai-collector status                  # Health check + recent interactions
ai-collector reflect                 # Trigger reflection for today
ai-collector reflect --date 2026-03-27
ai-collector query                   # Show today's interactions
ai-collector query week              # Last 7 days
ai-collector query stats             # Summary with counts
ai-collector query reflections       # Recent reflections
ai-collector setup claude            # Configure Claude Code hooks
ai-collector setup claude --global   # Skip scope prompt
ai-collector setup gemini            # Configure Gemini CLI hooks
ai-collector setup codex             # Configure Codex OTel export
ai-collector config                  # View current config
ai-collector config set PORT 8080    # Update a config value
ai-collector version                 # Show version info
```

## Configuration

Config file: `~/.config/ai-collector/config.yaml`

```yaml
port: 19321
db_path: ~/.local/share/ai-collector/ai_interactions.db
retention_days: 0

reflection:
  cli: claude --print
  schedule: daily
  output_dir: ~/.local/share/ai-collector/reflections
  prompt: ""  # path to custom prompt template
```

Environment variables override the config file:

| Variable | Config Key | Default |
|----------|-----------|---------|
| `COLLECTOR_PORT` | `port` | `19321` |
| `COLLECTOR_DB_PATH` | `db_path` | `~/.local/share/ai-collector/ai_interactions.db` |
| `COLLECTOR_RETENTION_DAYS` | `retention_days` | `0` (keep all) |
| `COLLECTOR_REFLECTION_CLI` | `reflection.cli` | `claude --print` |
| `COLLECTOR_REFLECTION_SCHEDULE` | `reflection.schedule` | `daily` |
| `COLLECTOR_REFLECTION_DIR` | `reflection.output_dir` | `~/.local/share/ai-collector/reflections` |
| `COLLECTOR_REFLECTION_PROMPT` | `reflection.prompt` | built-in template |

### Custom Reflection Prompt

Override the reflection prompt with your own template:

```bash
ai-collector config set reflection.prompt /path/to/my-prompt.md
```

The template uses `{{INTERACTIONS_FILE}}` as a placeholder for the file path containing the day's interactions. See `prompts/reflection.md` for the default.

## How It Works

1. **Hooks** fire on each coding CLI interaction and POST data to the collector
2. **Collector** normalizes payloads and stores them in SQLite
3. **Scheduler** checks on startup + every hour if yesterday's reflection is missing
4. **Reflection** writes interactions to a file, pipes a prompt to your CLI, saves the result as `YYYYMMDD-NNN.md`

## File Locations

| What | Path |
|------|------|
| Config | `~/.config/ai-collector/config.yaml` |
| Database | `~/.local/share/ai-collector/ai_interactions.db` |
| Reflections | `~/.local/share/ai-collector/reflections/` |
| Logs | `~/.local/share/ai-collector/collector.log` |
| PID file | `~/.local/share/ai-collector/collector.pid` |

## Building from Source

```bash
git clone https://github.com/phucpnt/coding-agent-reflection
cd coding-agent-reflection
make install    # builds and installs to ~/.local/bin/
make test       # run tests
```
