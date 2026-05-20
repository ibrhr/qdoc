# qdoc

An LLM-powered CLI tool that queries documentation sources and answers your questions.

## How it works

1. Fetches a documentation index from a source (e.g. go.dev/doc)
2. Sends the index + your question to an LLM
3. The LLM picks the most relevant pages and reads them
4. Returns a synthesized answer with proper references

```
$ qdoc go "how do generics work in Go"
● Thinking  ...
● Calling gpt-4.1 on openai
▶ READ: tutorial/generics
▶ READ: effective_go
● Answering ...
## Generics in Go ...
```

## Quick start

```bash
# Linux / macOS
curl -LO https://github.com/ibrhr/qdoc/releases/latest/download/qdoc_linux_amd64.tar.gz
tar xzf qdoc_linux_amd64.tar.gz && sudo mv qdoc /usr/local/bin/

# Windows (PowerShell)
curl -LO https://github.com/ibrhr/qdoc/releases/latest/download/qdoc_windows_amd64.zip
tar xzf qdoc_windows_amd64.zip && move qdoc.exe C:\Windows\System32\
```

See [all platforms & architectures](https://github.com/ibrhr/qdoc/releases/latest).

### Set up an API key

```bash
qdoc set key openai         # prompts for your key
qdoc provider               # interactive provider picker
qdoc model                  # pick a model
```

## Built-in providers

| Provider | Models |
|---|---|
| `openai` | gpt-5.5, gpt-5.4, gpt-5.4-mini, gpt-5.4-nano |
| `deepseek` | deepseek-v4-flash, deepseek-v4-pro |
| `opencode-zen` | GPT 5.5, GPT 5.4, Claude Opus 4.x, Gemini, Qwen, MiniMax, GLM, Kimi |
| `opencode-go` | deepseek-v4, qwen, minimax, GLM, Kimi, MiMo (low cost) |

Or use env vars: `OPENAI_API_KEY`, `DEEPSEEK_API_KEY`, `OPENCODE_ZEN_API_KEY`, `OPENCODE_GO_API_KEY`.

## Usage

```bash
# Query a doc source
qdoc go "generics tutorial"
qdoc go "how to use context"
qdoc ./my-local-docs "deployment guide"

# Manage configuration
qdoc provider                  # interactive provider picker
qdoc model                     # interactive model picker
qdoc set key openai           # save an API key
qdoc set provider openai       # switch default provider
qdoc set model openai gpt-5.5 # set model for a provider

# Inspect
qdoc status                    # show current config
qdoc sources                   # list doc sources
qdoc providers                 # list providers + key status
```

### Documentation sources

| Name | URL |
|---|---|
| `go` | go.dev/doc |

You can also point qdoc at any local directory of markdown, HTML, or reStructuredText files:

```bash
qdoc ./my-docs "query"
```

## Config

`~/.config/qdoc/config.json`:

```json
{
  "provider": "openai",
  "keys": { "openai": "sk-..." },
  "models": { "openai": "gpt-5.5" }
}
```

### Environment variables

| Variable | Overrides |
|---|---|
| `QDOC_PROVIDER` | Active provider |
| `QDOC_MODEL` | Active model |
| `QDOC_BASE_URL` | API base URL |

Resolution order: env vars → config file → built-in defaults.

## Build (for developers)

```bash
go build -ldflags "-X main.version=$(git describe --tags) -X main.commit=$(git rev-parse --short HEAD)" -o qdoc .
go vet ./...
```

Requires Go 1.26+. Dependencies: Bubble Tea v2, lipgloss v2, golang.org/x/net (HTML parser), golang.org/x/term.

## License

MIT
