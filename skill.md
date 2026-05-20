---
name: qdoc
description: Query documentation using an LLM. Search Go docs, FastAPI docs, or local markdown directories. USE FOR: "how do I use X in Go?", "show me the FastAPI docs for Y", "what does the docs say about Z?". DO NOT USE FOR: general coding questions not tied to specific documentation, or questions about proprietary/private codebases without docs.
---

# qdoc Skill

You are an AI coding agent with access to `qdoc`, a CLI that queries documentation sources using LLM-powered multi-turn research.

## When to Use

Use `qdoc` when:

- The user asks a question about a framework or library that has public documentation
- You need to read official docs to answer accurately
- The user references a specific documentation source (Go, FastAPI, etc.)
- You're asked "what does the docs say about X?"

Do NOT use `qdoc` for:

- General coding questions unrelated to specific documentation
- Questions about the user's own codebase (use file search tools instead)
- Questions you can answer confidently from your training data

## Installation Check

Before using `qdoc`, check if it's available:

```bash
which qdoc || (curl -sLO https://github.com/ibrhr/qdoc/releases/latest/download/qdoc_linux_amd64.tar.gz && tar xzf qdoc_linux_amd64.tar.gz && sudo mv qdoc /usr/local/bin/)
```

If not installed, guide the user to install it. Do not proceed without it.

## Query Syntax

### Standard Mode (TUI — great for users)

```bash
qdoc <source> <query>
```

Example:
```bash
qdoc go "how do generics work?"
```

This launches an interactive TUI with live streaming and scrollable output. Best for direct user interaction.

### Agent Mode (--no-tui — use this)

```bash
qdoc --no-tui <source> <query>
```

Example:
```bash
qdoc --no-tui go "generics constraints and type inference"
```

This runs headlessly and outputs the answer as markdown to stdout. Always use this mode — the TUI is for humans only.

### JSON Mode (--json — for structured parsing)

```bash
qdoc --json <source> <query>
```

```bash
qdoc --json go "generics" | jq -r '.answer'
```

## Available Sources

Run to list available sources:

```bash
qdoc sources
```

Common sources:

| Source | Description |
|---|---|
| `go` | Go standard library and toolchain docs |
| `fastapi` | FastAPI framework documentation |
| `./path/to/docs` | Any local directory of markdown files |

## Important Behaviors

1. **One-shot answers**: qdoc always returns a single, definitive answer. No conversation, no follow-ups.
2. **Multi-turn research**: Behind the scenes, qdoc may fetch multiple pages across up to 5 turns. Be patient — `--no-tui` runs silently.
3. **Exit code 0**: Answer was generated successfully. Exit code 1: Error (missing key, unknown source, API error).
4. **API key required**: The user must have configured at least one provider key (`qdoc set key <provider> <key>`).

## Troubleshooting

If `qdoc` fails:

1. Check configuration: `qdoc status`
2. List providers: `qdoc providers`
3. Set API key: `qdoc set key <provider> <key>`
4. Check version: `qdoc --version`

## Example Workflow

When a user asks "how does error wrapping work in Go?":

```bash
qdoc --no-tui go "error wrapping with fmt.Errorf and errors.Is"
```

Read the output and present it to the user. If the answer references specific URLs, include them as citations.

## Notes

- qdoc reads real documentation — it's more reliable than general LLM knowledge for current API details
- The answer may include code examples from the docs
- For very specific questions, make your query as precise as possible (include function names, error messages, etc.)
