# qdoc - Query the Docs

qdoc is a tui/cli that you or your coding agent can use to get fast, accurate, relevant, and up-to-date information about any library or framework you're using!

Under the hood, qdoc is an agent equipped with specialized system prompts for each documentation, when you ask about something, we load its specific guide, and qdoc will traverse their docs until it finds an accurate answer to your question.

## Why use qdoc?

Because when you let your main coding agent search for something on the internet, it fills its context with alot of irrelevant information and adds up to your expenses, making it more expensive AND degrades your agent's performance. But when you use qdoc, your agent asks a question and gets a comprehensive informed answer without rotting its context.

---

## Quick start

```bash
curl -fsSL https://qdoc.ibrhr.dev/install.sh | bash
```

```bash
npm install -g qdoc-agent
```

```bash
qdoc set key openai          # prompts for your API key (input hidden)
qdoc go "generics tutorial"   # ask your first question
qdoc python "type hints"      # Python stdlib docs
qdoc react "hooks rules"      # React docs via react.dev
qdoc nextjs "app router"      # Next.js docs via nextjs.org
```

---

## Usage

### As a human (interactive TUI)

```bash
qdoc go "how does context work?"
qdoc python "asyncio gather vs create_task"
qdoc fastapi "dependency injection patterns"
qdoc react "useEffect cleanup vs useLayoutEffect"
qdoc nextjs "server actions vs api routes"
qdoc ./my-docs "deployment guide"
```

Built with Bubble Tea v2. Live streaming, scrollable output, interactive provider/model pickers.

### As an agent (headless)

```bash
# Markdown to stdout — agent parses directly
qdoc --no-tui go "error wrapping with fmt.Errorf"

# Structured JSON — for programmatic consumption
qdoc --json go "generics constraints" | jq '.answer'
```

## Give Your AI Agent the qdoc skill

You can give your favorite coding agent (OpenCode, Claude Code, Cursor, Windsurf, etc.) the ability to autonomously read official documentation using `qdoc`.

When you install the **qdoc skill**, your agent will automatically use the CLI under the hood whenever you ask library-specific questions, ensuring its answers are accurate, grounded, and up-to-date!

There are two ways to teach your agent how to use `qdoc`:

### Method 1: Using the Skills CLI (Recommended)

The easiest way to install the skill is using the universal [Agent Skills](https://skills.sh/) package manager. Simply run:

```bash
npx skills add ibrhr/qdoc
```
*This command automatically detects your AI agent and seamlessly links the qdoc capabilities to your workspace.*

### Method 2: Manual Installation

If you prefer to install it manually, or if you're using an agent that relies on local workspace rules, you can download the configuration file directly into your project.

1. Download the skill file to your workspace:
   ```bash
   curl -O https://raw.githubusercontent.com/ibrhr/qdoc/main/SKILL.md
   ```

2. Move it to the appropriate directory based on your agent:
   - **OpenCode**: `.opencode/skills/qdoc/SKILL.md`
   - **Claude Code**: `.claude/skills/qdoc/SKILL.md`
   - **Cursor**: Rename it to `.cursor/rules/qdoc.mdc`
   - **Windsurf**: Rename it to `.windsurfrules`

**That's it!** Now, just tell your agent: *"How do I do X in FastAPI?"* and watch it trigger `qdoc` to fetch the exact, updated answer.

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
| `python` | Python 3 standard library, tutorial, reference — [docs.python.org/3](https://docs.python.org/3) |
| `nextjs` | Next.js App Router + Pages Router — [nextjs.org/docs](https://nextjs.org/docs) |
| `fastapi` | FastAPI framework — [fastapi.tiangolo.com](https://fastapi.tiangolo.com) |
| `react` | React docs (learn, reference, hooks) — [react.dev](https://react.dev) |
| `pydantic` | Pydantic validation library — [pydantic.dev](https://pydantic.dev) |
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
qdoc python "query"           # query Python docs
qdoc fastapi "query"          # query FastAPI docs
qdoc react "query"            # query React docs
qdoc nextjs "query"           # query Next.js docs
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

### Shell script (all platforms)

```bash
curl -fsSL https://qdoc.ibrhr.dev/install.sh | bash
```

Installs to `~/.qdoc/bin` (no sudo). Adds itself to your shell config. Options:

```bash
curl -fsSL ... | bash -s -- --version X.Y.Z    # specific version
curl -fsSL ... | bash -s -- --no-modify-path    # don't touch shell config
```

### npm (all platforms)

```bash
npm install -g qdoc-agent
```

Pin a specific version with `qdoc-agent@0.1.5`.

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

### Development

```bash
go build ./...       # build all packages
go vet ./...         # static analysis
go test ./...        # run 191 tests across 9 packages
go test -race ./...  # with race detector
go test -cover ./... # with coverage
```

Tests use the standard library `testing` package (no third-party deps). HTTP-dependent tests use `httptest.NewServer`.

---

## Verify

```bash
qdoc --version
# qdoc 0.1.5 (abc1234)
```

---

## License

MIT
