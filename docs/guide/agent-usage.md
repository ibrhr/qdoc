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

qdoc ships with a skill file that teaches your agent when and how to invoke it — which sources are available, the correct CLI syntax, and that `--no-tui` is the mode to use. With the skill loaded, your agent knows qdoc is an option without you having to write custom instructions.

### Install via npx (recommended)

```bash
npx skills add ibrhr/qdoc
```

This downloads and installs the qdoc skill into your agent's skills directory automatically.

### Manual

Copy [SKILL.md](https://github.com/ibrhr/qdoc/blob/main/SKILL.md) from the repo into your agent's skills directory.

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
