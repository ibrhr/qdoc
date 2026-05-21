# Agent Usage

qdoc was built to be invoked by AI coding agents. It replaces the multi-turn, trial-and-error doc research loop with a single CLI call.

See [Why qdoc](/guide/why) for the full problem/solution breakdown and token cost comparison.

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

Identical behavior to `--no-tui`, but outputs JSON with `answer`, `source`, and `steps` fields.

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
> Available sources: `go`, `python`, `fastapi`, `react`, `nextjs`, `pydantic`, or any local docs directory (`./path`).
> Always use `--no-tui` for agent mode.

### GitHub Copilot

Add to your `.github/copilot-instructions.md`:

```markdown
## Documentation research

When you need documentation about a framework or library, run `qdoc --no-tui <source> <query>`.
Available sources: `go`, `fastapi`, `react`, `nextjs`, `pydantic`, or any local docs directory (`./path`).

Install if missing: `npm install -g qdoc-agent`
```

### Manual (any agent)

Add this to any agent's system prompt or custom instructions:

> When you need documentation about a framework or library and `qdoc` is available, use it directly:
> `qdoc --no-tui <source> <query>`
>
> Available sources: `go` (Go docs), `python` (Python docs), `fastapi` (FastAPI docs), `react` (React docs), `nextjs` (Next.js docs), `pydantic` (Pydantic docs), `./path` (local docs directory).
> This returns a single markdown answer — no conversation, no follow-ups.
> The answer is more reliable than your training data for current API details.

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

## Recommended Models for Agents

For agent use, prioritize speed and cost over maximum capability. Doc research doesn't require frontier models:

| Use Case | Recommended Model |
|----------|------------------|
| Fast, cheap research | `gpt-5.4-mini` (OpenAI) or `deepseek-v4-flash` |
| Complex API questions | `gpt-5.5` (OpenAI) or `gpt-5.4` (OpenCode Zen) |
| Deep code analysis | `deepseek-v4-pro` or `claude-opus-4-7` (OpenCode Zen) |

See [Configuration](/guide/configuration) for model setup.
