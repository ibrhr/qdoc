---
name: qdoc
description: >-
  Query documentation using an LLM. Search Go docs, FastAPI docs, React, Next.js or local markdown directories.
  USE FOR: "how do I use X in Go?", "show me the FastAPI docs for Y", "what does the docs say about Z?", library docs, framework docs, API reference.
  DO NOT USE FOR: general coding questions not tied to specific documentation, or questions about proprietary/private codebases without docs.
---

# qdoc Skill — Query the Docs

> **AUTHORITATIVE GUIDANCE — MANDATORY COMPLIANCE**
> Use this skill to query official documentation instead of relying on potentially outdated training data or web searches. `qdoc` traverses docs autonomously and returns highly accurate, up-to-date API signatures, configurations, and patterns.

## Triggers

Activate this skill when:
- The user asks a question about a framework or library that has official documentation.
- You need to read official docs to answer accurately (e.g., API signatures, breaking changes, patterns).
- The user references a specific documentation source supported by qdoc (Go, FastAPI, Python, React, Next.js).
- You need to lookup documentation in a local project folder (`./docs`).

**DO NOT** use `qdoc` for:
- General coding concepts unrelated to specific libraries.
- Questions you can confidently answer from training data without verifying details.
- Searching the user's proprietary codebase source code (use standard grep/file search instead).

## Rules

1. **Verify Installation**: Before running a query, verify availability with `which qdoc`. If missing, install via `curl -fsSL https://qdoc.ibrhr.dev/install.sh | bash` or `npm install -g qdoc-agent`. On Windows, use WSL or Git Bash to run the install script. If installation fails, notify the user and stop.
2. **Mandatory Flags**: The TUI is for humans. You MUST run headlessly using `--no-tui` or `--json`.
3. **Execution Behavior**: `qdoc` executes multiple turns internally (up to 5 page fetches). Run the command and wait for the process to complete. Do not interrupt it.
4. **Error Handling**: Always check exit codes. 0 = Success. 1 = Error. If the process fails due to a missing API key, instruct the user to run `qdoc set key <provider>`.
5. **One-Shot Execution**: `qdoc` returns the final compiled answer and exits. It does not support follow-up questions within the same execution.

---

## Usage Syntax

### Agent Mode (Markdown Output)
Runs headlessly and prints the compiled markdown answer to `stdout`.
```bash
qdoc --no-tui <source> "<query>"
```

### JSON Mode (Structured Output)
Returns a JSON object containing the `answer`, `source`, and traversal `steps`.
```bash
qdoc --json <source> "<query>" | jq -r '.answer'
```

---

## Available Sources

| Source | Description |
|--------|-------------|
| `go` | Go standard library, toolchain, modules, tutorials (go.dev/doc) |
| `python` | Python 3 standard library and reference (docs.python.org/3) |
| `fastapi` | FastAPI framework documentation (fastapi.tiangolo.com) |
| `react` | React learn and API reference (react.dev) |
| `nextjs` | Next.js App/Pages Router and API reference (nextjs.org/docs) |
| `./path` | Local directory containing markdown, HTML, or text files |

*(Run `qdoc sources` to view all available built-in sources dynamically.)*

---

## Query Best Practices

To get the best results, formulate dense, highly specific queries:

- **Good**: `"error wrapping with fmt.Errorf %w, errors.Is, and errors.As in Go 1.24"`
- **Bad**: `"error handling"`
- **Good**: `"how APIRouter works with path operation decorators in FastAPI"`
- **Bad**: `"routing"`
- **Good**: `"os.ReadFile vs os.Open + bufio.Scanner for large files"`

Include exact function names, error codes, or specific module names in your query.

---

## Troubleshooting Reference

Use these CLI commands if you need to debug `qdoc` state:
- `qdoc status` — Check current provider and model configuration
- `qdoc providers` — List all providers and API key status
- `qdoc set key <prov>` — Prompt the user to run this if their API key is missing
