# Getting Started

qdoc queries documentation sources using LLM-powered multi-turn research. Pick your path:

- **You're an agent** → invoke qdoc headlessly, get a markdown answer, no TUI
- **You're a human** → launch the interactive TUI, watch the research live, scroll the result

Not sure why you need qdoc? See [Why qdoc](/guide/why).

---

## Quick Start

```bash
curl -fsSL https://qdoc.ibrhr.dev/install.sh | bash
```

```bash
npm install -g qdoc-agent
```

# Set your API key (prompts securely, input hidden)
qdoc set key openai

# Ask a question
qdoc go "how do generics work?"
qdoc python "typing Protocol vs ABC"
qdoc fastapi "dependency injection patterns"
qdoc react "useEffect cleanup"
qdoc nextjs "app router layouts"
```

That's it. qdoc fetches the Go documentation index, has the LLM pick relevant pages, reads them in parallel, and answers.

---

## For Agents (Headless)

```bash
qdoc --no-tui go "generics type constraints and type inference"
```

Prints a markdown answer to stdout. No TUI, no interaction, just the answer.

```bash
# JSON for structured parsing
qdoc --json go "error wrapping with fmt.Errorf" | jq '.answer'
```

For agent tool calls and any automated context. See [Agent Usage](/guide/agent-usage) for integration guides.

---

## For Humans (TUI)

```
$ qdoc go "how do channels work?"

  qdoc · go ▸ openai/gpt-5.5

  ● Fetching index · 245 pages

  ▶ READ: effective_go
  ▶ READ: blog/pipelines

  ● Answering ···

  ## Channels in Go
  Channels are typed conduits through which
  you send and receive values...

  ▸ ch := make(chan int)          — unbuffered
  ▸ ch := make(chan int, 10)      — buffered
  ▸ close(ch)                      — close a channel
```

Interactive features:
- **Live streaming** — watch the answer render in real time
- **Scrollable** — navigate long answers with arrow keys / j/k
- **Provider/model pickers** — `qdoc provider` and `qdoc model`

---

## First-Run Setup

No API key configured? qdoc launches an interactive setup wizard:

```
$ qdoc        # no args, no key

  qdoc · First Time Setup

  Pick a provider:
  ▸ openai
    deepseek
    opencode-zen
    opencode-go
```

You can also configure manually:

```bash
qdoc set key openai           # prompts for key (input hidden, no echo)
qdoc set key openai sk-...    # inline variant
qdoc set provider openai      # set default provider
qdoc model                    # pick a model interactively
```

Check your configuration at any time:

```bash
qdoc status
```

---

## Next Steps

- [Why qdoc](/guide/why) — the problem qdoc solves, and how it works
- [Agent Usage](/guide/agent-usage) — integrate qdoc with your coding agent
- [Installation](/guide/installation) — all installation methods, manual download, build from source
- [Configuration](/guide/configuration) — providers, keys, models, env vars
- [FAQ](/guide/faq) — common questions and answers
