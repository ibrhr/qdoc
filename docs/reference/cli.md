# CLI Reference

## Commands

### Query

```bash
qdoc <source> <query>
qdoc --no-tui <source> <query>    # headless, markdown output
qdoc --json <source> <query>      # headless, JSON output
```

Queries a documentation source. The LLM fetches docs, reads relevant pages, and answers.

Examples:

```bash
qdoc go "channels vs mutexes"
qdoc ./docs "deployment guide"
qdoc --json go "generics" | jq '.answer'
```

### Configuration Commands

| Command | Description |
|---|---|
| `qdoc status` | Show current configuration |
| `qdoc providers` | List all providers and their key/model status |
| `qdoc provider` | Interactive provider selection (TUI) |
| `qdoc model` | Interactive model selection (TUI) |
| `qdoc set key <provider> [key]` | Set API key (prompts if key omitted) |
| `qdoc set provider <name>` | Set default provider |
| `qdoc set model <provider> <model>` | Set model for a provider |

### Info Commands

| Command | Description |
|---|---|
| `qdoc sources` / `qdoc ls` | List available documentation sources |
| `qdoc version` / `qdoc --version` | Print version and commit hash |
| `qdoc help` | Print usage |

## Flags

| Flag | Description |
|---|---|
| `--json` | Headless mode with JSON output |
| `--no-tui` | Headless mode with markdown output |
| `--version` | Print version info |
| `--help` | Print help |

## Environment Variables

| Variable | Description |
|---|---|
| `QDOC_PROVIDER` | Override default provider |
| `QDOC_MODEL` | Override model for provider |
| `QDOC_BASE_URL` | Override API base URL |
| `OPENAI_API_KEY` | OpenAI API key |
| `DEEPSEEK_API_KEY` | DeepSeek API key |
| `OPENCODE_ZEN_API_KEY` | Opencode Zen API key |
| `OPENCODE_GO_API_KEY` | Opencode Go API key |

## Build-Time Variables

Set via `-ldflags`:

```bash
go build -ldflags "-X main.version=0.2.0 -X main.commit=$(git rev-parse --short HEAD)" -o qdoc .
```

## Exit Codes

| Code | Meaning |
|---|---|
| 0 | Success |
| 1 | Error (unknown source, missing key, API error, etc.) |
