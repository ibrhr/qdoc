# Why qdoc

AI coding agents are great at writing code. They are terrible at researching documentation.

## The Problem

When an agent needs to answer a question about a framework or library, it does this:

1. Fetches the documentation index (hundreds of links)
2. Guesses which pages look relevant
3. Fetches a handful, reads them
4. Realizes it picked the wrong pages, fetches more
5. Sometimes iterates a third or fourth time
6. Finally synthesizes an answer from whatever it collected

**Every step costs tokens.** Every wrong guess wastes inference. Every page fetched adds to the context window. A single documentation question can burn 10-15k tokens before the agent even starts coding.

And the answer is often wrong — because the agent skimmed the wrong pages or missed a critical detail buried in a page it didn't fetch.

## The Solution

qdoc delegates the entire research process to a dedicated LLM call:

```bash
qdoc --no-tui go "error wrapping patterns"
```

What happens internally:

| Turn | Action | Tokens |
|------|--------|--------|
| 1 | Fetch doc index (245 pages from go.dev/doc) | ~8k chars (HTTP, no LLM) |
| 2 | LLM scans index + query, picks relevant pages | ~400 output tokens |
| 3 | qdoc fetches selected pages in parallel | HTTP calls, no LLM |
| 4 | LLM reads content, decides if more pages needed | ~200 output tokens |
| 5 | qdoc fetches additional pages if needed | HTTP calls, no LLM |
| 6 | LLM synthesizes final answer with citations | ~800 output tokens |

**Result:** 3 LLM inferences, 4 page fetches, one definitive answer. The agent gets a clean markdown response it can use immediately — no context bloat, no wasted rounds.

## The Numbers

| Metric | Agent doing it manually | qdoc |
|--------|------------------------|------|
| LLM rounds | 6-8 | 3 |
| Page fetches | 10+ | 4-6 |
| Context tokens | 10-15k+ | 2-4k |
| Answer quality | Inconsistent (depends on which pages the agent found) | Consistent (LLM picks pages systematically) |
| Agent context pollution | High (all fetched page content sits in context) | None (only the final answer) |

## Who It's For

### AI Coding Agents

qdoc was built for agents first. Invoke it headlessly, get a markdown answer, move on. No TUI, no interaction, no prompt overhead. See [Agent Usage](/guide/agent-usage) for integration guides.

### Developers Who Want Better Answers

Sometimes you just want a good answer to a documentation question without reading through five pages yourself. The TUI lets you watch the research happen live, scroll through the answer, and switch providers or models on the fly.

### Teams with Local Documentation

Point qdoc at a `./docs` directory and it will index and query your own project's documentation — markdown, HTML, reStructuredText, AsciiDoc. Useful for onboarding, API reference questions, or deployment guides.

## Next Steps

- [Getting Started](/guide/getting-started) — install and run your first query
- [Agent Usage](/guide/agent-usage) — integrate qdoc with your coding agent
- [Configuration](/guide/configuration) — providers, keys, models
