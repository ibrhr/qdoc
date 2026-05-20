# Getting Started

qdoc queries documentation sources using an LLM. It fetches the doc index, asks the LLM which pages are relevant, reads them, and answers your question.

## Quick Start

```bash
# 1. Install
curl -fsSL https://qdoc.ibrhr.dev/install.sh | sh

# 2. Set your API key
qdoc set key openai sk-your-key-here

# 3. Ask a question
qdoc go "how do I use generics?"
```

That's it. qdoc will fetch the Go documentation index, have the LLM pick relevant pages, read them, and answer.

## How It Works

1. **Fetch Index** — qdoc downloads the documentation index (a list of all pages)
2. **LLM Selects Pages** — The LLM picks the most relevant pages for your query
3. **Parallel Reads** — qdoc fetches those pages simultaneously
4. **Iterate** — The LLM may request more pages (up to 5 turns)
5. **Answer** — The LLM synthesizes a definitive answer from the content

## First-Run Setup

On first run with no API key configured, qdoc launches an interactive TUI:

```
qdoc    # launches setup wizard
```

You can also configure manually:

```bash
qdoc set key openai sk-abc123
qdoc set provider openai
qdoc model               # pick a model interactively
```

Check your config:

```bash
qdoc status
```

## Next Steps

- [Installation](/guide/installation) — all installation methods
- [Configuration](/guide/configuration) — providers, keys, models
- [Agent Usage](/guide/agent-usage) — CI/CD and AI agent integration
