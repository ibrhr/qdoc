# Agent Usage

qdoc is designed to be used by AI coding agents (Claude Code, Cursor, Copilot, etc.). Two modes are available for headless/automated use.

## Headless Modes

### `--no-tui`: Plain Markdown Output

Skips the TUI entirely. The research loop runs in the background, and the answer is printed to stdout as markdown:

```bash
qdoc --no-tui go "how do generics work?"
```

Output is raw markdown — your agent can parse it directly.

### `--json`: Structured JSON Output

Same as `--no-tui`, but outputs JSON with metadata:

```bash
qdoc --json go "generics tutorial"
```

Output format:

```json
{
  "answer": "# Generics in Go\n\nGenerics allow you to...",
  "source": "go",
  "steps": [
    {"phase": "Fetched index", "detail": "245 pages"},
    {"phase": "Reading", "detail": "/doc/tutorial/generics"},
    {"phase": "Calling", "detail": "gpt-4.1 via openai"}
  ]
}
```

## CI/CD Pipeline Integration

```bash
# In any CI script — just set the env var and run
export OPENAI_API_KEY=${{ secrets.OPENAI_API_KEY }}
qdoc --json go "release notes for Go 1.22" | jq '.answer'
```

## Installing the Agent Skill

To add qdoc as a skill for your coding agent:

1. Copy <a href="https://github.com/ibrhr/qdoc/blob/main/skill.md">skill.md</a> to your agent's skills directory
2. The agent will now know how to use `qdoc` to query documentation

### opencode

Copy `skill.md` to `~/.agents/skills/qdoc/SKILL.md`

### Claude Code

Copy `skill.md` to `~/.claude/skills/qdoc.md`

### Cursor

Add `skill.md` as a custom instruction in Cursor settings.

### Manual

You can also just instruct your agent directly:

> When you need documentation about a framework or library, run: `qdoc <source> <query> --no-tui`. Sources: go, fastapi, or a local directory.

## Supported Sources

```bash
qdoc sources
```

| Source | Description |
|---|---|
| `go` | Go standard library and toolchain docs |
| `fastapi` | FastAPI documentation |
| `./path/to/docs` | Local directory of markdown files |
