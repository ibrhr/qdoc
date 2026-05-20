# Agent Usage

qdoc was built to be invoked by AI coding agents. It replaces the multi-turn, trial-and-error doc research loop with a single CLI call.

## The Problem

When an AI coding agent needs to answer a question about a framework or library, it typically does this:

1. Skims the documentation index page (hundreds of links)
2. Guesses which pages look relevant
3. Fetches a handful, reads them
4. Realizes it needs different pages, fetches more
5. Sometimes iterates a third time
6. Finally synthesizes an answer

**This is expensive.** Every page fetch consumes context tokens. Every wrong guess wastes inference. A single doc question can burn 10-15k tokens before the agent even starts coding.

## The qdoc Solution

qdoc delegates the research to an LLM in one shot:

```
qdoc --no-tui go "error wrapping patterns"
```

What happens internally:

| Turn | Action | Cost |
|------|--------|------|
| 1 | Fetch doc index (245 pages from go.dev/doc) | 1 HTTP call, ~8k chars |
| 2 | LLM scans index + query, picks: `effective_go`, `blog/go1.13-errors` | 1 inference, ~400 output tokens |
| 3 | qdoc fetches 2 pages in parallel, extracts main content | 2 HTTP calls |
| 4 | LLM reads content, decides: need `errors` package docs too | 1 inference, ~200 output tokens |
| 5 | qdoc fetches 1 more page | 1 HTTP call |
| 6 | LLM synthesizes final answer with citations | 1 inference, ~800 output tokens |

**Result:** 3 LLM inferences, 4 page fetches, one definitive answer. No wasted context.

Compare to the agent doing it manually: 6-8 LLM rounds, 10+ page fetches, massive context bloat.

---

## Headless Modes

### `--no-tui`: Markdown to stdout

```bash
qdoc --no-tui <source> <query>
```

Runs the full research pipeline without the TUI. The answer is printed as raw markdown — your agent can parse it directly.

```bash
qdoc --no-tui go "generics type constraints"
```

Output goes to stdout. Use this in shell scripts, agent tool calls, or pipe it:

```bash
qdoc --no-tui fastapi "dependency injection" | less
```

### `--json`: Structured JSON output

```bash
qdoc --json <source> <query>
```

Identical behavior to `--no-tui`, but outputs JSON with metadata:

```json
{
  "answer": "# Generics in Go\n\nType parameters enable...",
  "source": "go",
  "steps": [
    { "phase": "Fetched index", "detail": "245 pages" },
    { "phase": "Reading", "detail": "/doc/tutorial/generics" },
    { "phase": "Reading", "detail": "/doc/effective_go" },
    { "phase": "Calling", "detail": "gpt-5.5 via openai" },
    { "phase": "Answering", "detail": "" }
  ]
}
```

Fields:

| Field | Type | Description |
|-------|------|-------------|
| `answer` | `string` | The synthesized answer in markdown |
| `source` | `string` | Doc source queried (`go`, `fastapi`, `./path`) |
| `steps` | `[]Step` | Research trace: each phase and what was done |

Parse with `jq`:

```bash
qdoc --json go "generics" | jq -r '.answer'
qdoc --json go "generics" | jq '.steps[] | "\(.phase): \(.detail)"'
```

---

## Agent Integration

### opencode

Copy the skill file to your opencode skills directory:

```bash
mkdir -p ~/.agents/skills/qdoc
cp skill.md ~/.agents/skills/qdoc/SKILL.md
```

opencode loads agent skills from `~/.agents/skills/`. After copying, opencode will know how to use `qdoc` automatically. Verify:

```bash
ls ~/.agents/skills/qdoc/SKILL.md
```

### Claude Code

```bash
cp skill.md ~/.claude/skills/qdoc.md
```

Claude Code reads skills from `~/.claude/skills/*.md`. The skill teaches Claude when to invoke qdoc, the correct CLI syntax, and which sources are available.

### Cursor

1. Open Cursor Settings → Rules
2. Add a new Project Rule or User Rule
3. Paste the contents of `skill.md`

Alternatively, add this as a custom instruction:

> When you need documentation about a framework or library, run:
> `qdoc --no-tui <source> <query>`
> Available sources: `go`, `fastapi`, or any local docs directory (`./path`).
> Always use `--no-tui` for agent mode.

### GitHub Copilot

Add to your `.github/copilot-instructions.md`:

```markdown
## Documentation research

When you need documentation about a framework or library, run `qdoc --no-tui <source> <query>`.
Available sources: `go`, `fastapi`, or any local docs directory (`./path`).

Install if missing: `npm install -g qdoc-agent`
```

### Manual (any agent)

Add this to any agent's system prompt or custom instructions:

> When you need documentation about a framework or library and `qdoc` is available, use it directly:
> `qdoc --no-tui <source> <query>`
>
> Available sources: `go` (Go docs), `fastapi` (FastAPI docs), `./path` (local docs directory).
> This returns a single markdown answer — no conversation, no follow-ups.
> The answer is more reliable than your training data for current API details.

---

## CI/CD Pipelines

### GitHub Actions

```yaml
name: Doc Check
on: [push]
jobs:
  verify:
    runs-on: ubuntu-latest
    steps:
      - name: Install qdoc
        run: curl -fsSL https://qdoc.ibrhr.dev/install.sh | bash

      - name: Query docs
        env:
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
        run: |
          qdoc --json --no-tui go "breaking changes in Go 1.24" > go124.json
          cat go124.json | jq '.answer'
```

### Makefile target

```makefile
docs-check:
	@qdoc --version > /dev/null 2>&1 || npm install -g qdoc-agent
	@qdoc --no-tui ./docs "API reference for authentication endpoints"
```

---

## How It Works (Internals)

```
Query → Fetch Index → [LLM: Pick Pages] → Parallel Fetch → Extract Content
                                                    ↓
                                              [LLM: Read + Decide]
                                                    ↓
                                         Need more? →← Done?
                                              ↓         ↓
                                         Next turn    Synthesize Answer
                                        (up to 5)         ↓
                                                      Output
```

### Retry Logic

Both the LLM calls and doc fetches have built-in retry:

| Component | Attempts | Base Delay | Max Delay | Strategy |
|-----------|----------|------------|-----------|----------|
| LLM inference | 3 | 2s | 30s | Exponential + 30% jitter |
| Doc fetches | 3 | 1s | 10s | Exponential + 30% jitter |

Retryable errors: HTTP 429, 5xx, network timeouts, stream read failures.
Non-retryable: HTTP 400, 401, 403, 404.

### Content Extraction

qdoc extracts the main content from HTML documentation pages, stripping navigation, headers, footers, and sidebars. Each page is truncated to 12,000 characters to keep context manageable while retaining substantive content.

---

## Best Practices for Agent Queries

### Be specific

| Weak | Strong |
|------|--------|
| `"error handling"` | `"error wrapping with fmt.Errorf, errors.Is, and errors.As"` |
| `"routing"` | `"how APIRouter works with multiple path operations in FastAPI"` |
| `"context"` | `"context.Context cancellation and deadline propagation"` |

Specific queries help the LLM select the right pages on the first pass, reducing research turns.

### Include function names

```bash
# Weak
qdoc --no-tui go "reading files"

# Strong
qdoc --no-tui go "os.ReadFile vs os.Open + bufio.Scanner"
```

### Reference specific errors

```bash
qdoc --no-tui fastapi "422 Unprocessable Entity validation error response model"
```

### Query local docs for your project

```bash
qdoc --no-tui ./docs/api "POST /users endpoint request body schema"
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Answer generated successfully |
| 1 | Error (missing key, unknown source, API failure, rate limited) |

Agents should check `$?` after each call. Non-zero exit code = retry or escalate.

---

## Environment Variables for CI

```bash
# Set these in CI secrets, not in code
export QDOC_PROVIDER=openai
export QDOC_MODEL=gpt-5.4-mini
export OPENAI_API_KEY=$OPENAI_API_KEY
```

`gpt-5.4-mini` is a good default for agent use — fast, cheap, and sufficient for doc research.
