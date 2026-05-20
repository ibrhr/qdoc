---
name: qdoc
description: Query documentation using an LLM. Search Go docs, FastAPI docs, or local markdown directories. USE FOR: "how do I use X in Go?", "show me the FastAPI docs for Y", "what does the docs say about Z?". DO NOT USE FOR: general coding questions not tied to specific documentation, or questions about proprietary/private codebases without docs.
---

# qdoc Skill — Agent for Agents

`qdoc` is a CLI that does documentation research for you. Pass it a source and a query; it fetches the doc index, has an LLM select the right pages, reads them, and returns a definitive answer. One call, one answer. No trial-and-error.

## When to Use

Use `qdoc` when:

- The user asks a question about a framework or library that has documentation
- You need to read official docs to answer accurately (API signatures, configuration, patterns)
- The user references a specific documentation source (Go, FastAPI, etc.)
- You need current documentation — more reliable than training data for API details

Do NOT use `qdoc` for:

- General coding questions unrelated to specific documentation
- Questions about the user's own codebase (use file search tools instead)
- Questions you can answer confidently from your training data

## Installation Check

Before using `qdoc`, verify it's installed:

```bash
which qdoc || curl -fsSL https://qdoc.ibrhr.dev/install.sh | bash
# Or via npm: npm install -g qdoc-agent
```

If `qdoc` is not installed and cannot be auto-installed, tell the user and stop. Do not proceed without it.

## Query Syntax

### Agent Mode (`--no-tui` — always use this)

```bash
qdoc --no-tui <source> <query>
```

Runs headlessly. The answer is printed as markdown to stdout. Always use `--no-tui` — the TUI is for human users.

```bash
qdoc --no-tui go "generics constraints and type inference"
qdoc --no-tui fastapi "OAuth2 password flow with JWT"
```

### JSON Mode (`--json` — for structured output)

```bash
qdoc --json <source> <query>
```

Returns JSON with `answer`, `source`, and `steps` fields.

```bash
qdoc --json go "generics" | jq -r '.answer'
qdoc --json go "generics" | jq '.steps[] | "\(.phase): \(.detail)"'
```

## Available Sources

```bash
qdoc sources
```

| Source | Description |
|--------|------------|
| `go` | Go standard library, toolchain, modules, tutorials — [go.dev/doc](https://go.dev/doc) |
| `fastapi` | FastAPI framework — [fastapi.tiangolo.com](https://fastapi.tiangolo.com) |
| `./path` | Any local directory of markdown, HTML, reStructuredText, or AsciiDoc files |

## Best Practices

### Write specific queries

| Weak | Strong |
|------|--------|
| `"error handling"` | `"error wrapping with fmt.Errorf, errors.Is, and errors.As in Go 1.24"` |
| `"routing"` | `"how APIRouter works with path operation decorators in FastAPI"` |

### Include function names and error messages

```bash
qdoc --no-tui go "os.ReadFile vs os.Open + bufio.Scanner for large files"
qdoc --no-tui fastapi "422 Unprocessable Entity with Pydantic v2 field validators"
```

### Query local project docs

```bash
qdoc --no-tui ./docs/api "POST /users endpoint request body schema"
```

## Important Behaviors

1. **One-shot**: qdoc always returns a single answer. No conversation, no follow-ups.
2. **Multi-turn internally**: Behind the scenes, qdoc may fetch pages across up to 5 turns. Wait for the result — `--no-tui` runs silently.
3. **Exit codes**: 0 = success (answer generated). 1 = error (missing key, unknown source, API failure). Always check.
4. **API key required**: The user must have configured at least one provider key. If missing, tell them to run `qdoc set key <provider>`.

## Troubleshooting

If `qdoc` fails:

```bash
qdoc status          # Check configuration
qdoc providers       # List providers and key status
qdoc set key openai  # Set or update API key
qdoc --version       # Check version
```

## Example

User asks: "how does error wrapping work in Go?"

```bash
qdoc --no-tui go "error wrapping with fmt.Errorf %w, errors.Is, errors.As"
```

Read the output. Present the answer to the user with any citations included.
