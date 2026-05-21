# CLI Reference

## Query

```bash
qdoc <source> <query>               # Interactive TUI
qdoc --no-tui <source> <query>      # Headless, markdown to stdout
qdoc --json <source> <query>        # Headless, JSON to stdout
```

Queries a documentation source. The LLM fetches the doc index, selects relevant pages, reads them, and synthesizes an answer — all internally. Up to 5 research turns before returning.

```bash
qdoc go "channels vs mutexes"
qdoc python "asyncio vs threading"
qdoc fastapi "dependency injection patterns"
qdoc react "useCallback dependencies"
qdoc nextjs "server components vs client components"
qdoc pydantic "BaseModel field validators"
qdoc ./docs "deployment guide"
qdoc --json go "generics" | jq '.answer'
```

### Flags

| Flag | Description |
|------|-------------|
| `--json` | Headless mode, JSON output with metadata (`answer`, `source`, `steps`) |
| `--no-tui` | Headless mode, raw markdown to stdout |
| `--version` | Print version and commit hash |
| `--help` | Print help |

`--json` implies `--no-tui`. You can use both together for clarity: `qdoc --no-tui --json go "query"`.

## Configuration Commands

| Command | Description |
|---------|-------------|
| `qdoc status` | Show current configuration (provider, models, keys masked) |
| `qdoc providers` | List all providers with model assignments and key status |
| `qdoc provider` | Interactive provider picker (TUI) |
| `qdoc model` | Interactive model picker across all providers (TUI) |
| `qdoc set key <provider> [key]` | Set API key. Prompts securely if key omitted |
| `qdoc set provider <name>` | Set default provider |
| `qdoc set model <provider> <model>` | Override model for a specific provider |

Prompt-safe key entry:

```bash
qdoc set key openai
# Enter API key: ████████ (input hidden, no terminal echo, no shell history)
```

## Info Commands

| Command | Description |
|---------|-------------|
| `qdoc sources` / `qdoc ls` | List available documentation sources |
| `qdoc --version` | Print version and commit hash |
| `qdoc --help` | Print usage |

## Environment Variables

| Variable | Overrides |
|----------|----------|
| `QDOC_PROVIDER` | Active provider |
| `QDOC_MODEL` | Active model (per-session) |
| `QDOC_BASE_URL` | API base URL (custom endpoint) |
| `OPENAI_API_KEY` | OpenAI API key |
| `DEEPSEEK_API_KEY` | DeepSeek API key |
| `OPENCODE_ZEN_API_KEY` | OpenCode Zen API key |
| `OPENCODE_GO_API_KEY` | OpenCode Go API key |

Priority: `QDOC_*` env vars → `config.json` → built-in defaults.

## Build-Time Variables

Embed version info via linker flags for custom builds:

```bash
go build -ldflags \
  "-X main.version=$(git describe --tags) \
   -X main.commit=$(git rev-parse --short HEAD)" \
  -o qdoc .
```

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success — answer generated |
| `1` | Error — missing API key, unknown source, API failure, rate limited, network error |

Always check exit codes in scripts. A non-zero exit means the query could not be completed and should be retried or escalated.
