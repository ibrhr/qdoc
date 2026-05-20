# qdoc — Agent for Agents

Documentation research for AI coding agents. No trial-and-error. No wasted tokens. One call, one answer.

```
$ qdoc go "how do generics work in Go"

● Thinking ···
● Calling gpt-5.5 on openai
▶ READ: tutorial/generics
▶ READ: effective_go
▶ READ: go1.24

## Generics in Go

Type parameters enable...
```

---

## Why qdoc?

AI coding agents waste tokens when researching documentation. The typical agent loop:

```
Agent: searches docs index → finds 200+ links
Agent: fetches 5 pages, none are right
Agent: tries 5 more, close but incomplete
Agent: one more round, finally gets it
Agent: synthesizes answer

Total: ~10 page fetches, ~40% irrelevant content, 3 LLM rounds
```

**qdoc replaces this with a one-shot LLM call:**

```
qdoc: fetches doc index (1 HTTP call)
qdoc: passes index → query to LLM (1 inference)
LLM: returns "read these 3 specific pages"
qdoc: fetches them in parallel (3 HTTP calls)
qdoc: if needed, asks LLM for one more round (up to 5 turns)
LLM: synthesizes final answer with citations

Total: 1-5 page fetches, 100% LLM-selected relevance, 1-2 LLM rounds
```

The difference compounds: an agent making 10 doc queries per session saves 50+ unnecessary page fetches and 15-20 redundant LLM rounds.

---

## Quick start

```bash
# Linux, macOS (shell)
curl -fsSL https://qdoc.ibrhr.dev/install.sh | bash

# Any platform (npm)
npm install -g qdoc-agent
```

```bash
qdoc set key openai          # prompts for your API key (input hidden)
qdoc go "generics tutorial"  # ask your first question
```

---

## Usage

### As a human (interactive TUI)

```bash
qdoc go "how does context work?"
qdoc fastapi "dependency injection patterns"
qdoc ./my-docs "deployment guide"
```

Built with Bubble Tea v2. Live streaming, scrollable output, interactive provider/model pickers.

### As an agent (headless)

```bash
# Markdown to stdout — agent parses directly
qdoc --no-tui go "error wrapping with fmt.Errorf"

# Structured JSON — for programmatic consumption
qdoc --json go "generics constraints" | jq '.answer'

# In CI pipelines
qdoc --no-tui --json go "release notes for Go 1.24" > result.json
```

### Integration

Copy [`skill.md`](skill.md) to your agent's skills directory:

| Agent | Path |
|-------|------|
| **opencode** | `~/.agents/skills/qdoc/SKILL.md` |
| **Claude Code** | `~/.claude/skills/qdoc.md` |
| **Cursor** | Add as custom instruction |

Or instruct any agent manually:

> When you need documentation about a framework or library, run `qdoc --no-tui <source> <query>`. Available sources: go, fastapi, or any local docs directory.

---

## Providers

qdoc uses any OpenAI-compatible API.

| Provider | Default Model | Models Available |
|----------|--------------|-----------------|
| `openai` | `gpt-5.5` | gpt-5.5, gpt-5.5-pro, gpt-5.4, gpt-5.4-pro, gpt-5.4-mini, gpt-5.4-nano |
| `deepseek` | `deepseek-v4-flash` | deepseek-v4-flash, deepseek-v4-pro |
| `opencode-zen` | `gpt-5.4-mini` | GPT 5.5/5.4, Claude Opus 4.x, Claude Sonnet 4.x, Claude Haiku 4.5, Gemini 3.x, Qwen, MiniMax, GLM, Kimi, DeepSeek, Nemotron |
| `opencode-go` | `deepseek-v4-flash` | Qwen, MiniMax, GLM, Kimi, DeepSeek, MiMo (low-cost coding models) |

Set a provider interactively (`qdoc provider`), via config, or with env vars: `OPENAI_API_KEY`, `DEEPSEEK_API_KEY`, `OPENCODE_ZEN_API_KEY`, `OPENCODE_GO_API_KEY`.

---

## Documentation sources

| Source | Description |
|--------|------------|
| `go` | Go standard library, toolchain, modules, tutorials — [go.dev/doc](https://go.dev/doc) |
| `fastapi` | FastAPI framework — [fastapi.tiangolo.com](https://fastapi.tiangolo.com) |
| `./path` | Any local directory of markdown, HTML, RST, or AsciiDoc files |

```bash
qdoc sources    # list all available sources
```

---

## Configuration

```json
// ~/.config/qdoc/config.json
{
  "provider": "openai",
  "keys": { "openai": "sk-..." },
  "models": { "openai": "gpt-5.5" }
}
```

Resolution order: `QDOC_*` environment variables → config file → built-in defaults.

| Variable | Overrides |
|----------|----------|
| `QDOC_PROVIDER` | Active provider |
| `QDOC_MODEL` | Active model |
| `QDOC_BASE_URL` | API base URL |

---

## Commands

```bash
qdoc go "query"               # query Go docs
qdoc fastapi "query"          # query FastAPI docs
qdoc ./my-docs "query"        # query local docs directory

qdoc provider                 # interactive provider picker (TUI)
qdoc model                    # interactive model picker (TUI)
qdoc set key openai           # save an API key (prompts securely)
qdoc set key openai sk-...    # save an API key (inline)
qdoc set provider openai      # switch default provider
qdoc set model openai gpt-5.5 # set model for a provider

qdoc status                    # show current config
qdoc providers                 # list providers + key status
qdoc sources                   # list documentation sources
```

---

## Install

### Shell script (Linux, macOS)

```bash
curl -fsSL https://qdoc.ibrhr.dev/install.sh | bash
```

Installs to `~/.qdoc/bin` (no sudo). Adds itself to your shell config. Options:

```bash
curl -fsSL ... | bash -s -- --version 0.1.2     # specific version
curl -fsSL ... | bash -s -- --no-modify-path    # don't touch shell config
```

### npm (all platforms)

```bash
npm install -g qdoc-agent
```

### Manual download

Binaries for every platform on [GitHub Releases](https://github.com/ibrhr/qdoc/releases/latest).

### From source

```bash
git clone https://github.com/ibrhr/qdoc.git
cd qdoc
go build -ldflags "-X main.version=$(git describe --tags) -X main.commit=$(git rev-parse --short HEAD)" -o qdoc .
mv qdoc ~/.local/bin/
```

Requires Go 1.26+.

---

## Verify

```bash
qdoc --version
# qdoc 0.1.2 (abc1234)
```

---

## License

MIT
