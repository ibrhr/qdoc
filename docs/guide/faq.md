# FAQ

## What if my provider isn't listed?

Any OpenAI-compatible API works. Set the provider name, base URL, model, and API key:

```bash
export QDOC_PROVIDER=custom
export QDOC_BASE_URL=https://api.mistral.ai/v1
export QDOC_MODEL=mistral-large-latest
# Set your key via env var or config file
export MISTRAL_API_KEY=your-key
```

qdoc will use the custom endpoint for all LLM calls. See [Configuration](/guide/configuration) for details.

## Can I use local models (Ollama, vLLM, LM Studio)?

Yes. Point `QDOC_BASE_URL` at your local server:

```bash
export QDOC_PROVIDER=local
export QDOC_BASE_URL=http://localhost:11434/v1   # Ollama
export QDOC_MODEL=llama3.1
```

Local models work best when they support function calling or structured output — qdoc uses JSON action parsing internally. Smaller models may struggle with the page selection step.

## How accurate are the answers?

qdoc answers are grounded in actual documentation pages — not training data. The LLM reads real doc pages and synthesizes from them. Accuracy depends on:

- **Query specificity** — specific queries pick better pages
- **Model capability** — stronger models make better page selections and synthesize better answers
- **Doc quality** — if the source documentation is outdated or incomplete, the answer will be too

For API details, qdoc answers are generally more reliable than an agent's training data, which may be months out of date.

## What happens when rate limited?

qdoc has built-in retry with exponential backoff and jitter:

| Component | Attempts | Base Delay | Max Delay |
|-----------|----------|------------|-----------|
| LLM calls | 3 | 2s | 30s |
| Doc fetches | 3 | 1s | 10s |

Retryable errors: HTTP 429, 5xx, network timeouts, stream read failures.
Non-retryable: HTTP 400, 401, 403, 404.

In the TUI, retries show as `↻ (retrying in 4.2s — attempt 2/3)`. In headless mode, retries happen silently. If all attempts fail, qdoc exits with code 1.

## Can I add my own documentation sources?

Built-in sources are defined in code (`internal/docsource/source.go`). To add a custom source, modify the `KnownSources` slice and rebuild.

For one-off queries, use local directory mode:

```bash
qdoc ./my-project-docs "deployment instructions"
```

qdoc recursively walks the directory and indexes `.md`, `.mdx`, `.html`, `.rst`, `.txt`, and `.adoc` files.

## Does qdoc work without an API key?

No. qdoc requires an LLM API key to function — the research pipeline depends on LLM calls for page selection, content reading, and answer synthesis.

If you don't have a key, run `qdoc` and the interactive setup will prompt you to pick a provider and enter a key.

## Why not just let my agent browse the web?

Web browsing tools (like browser-use or Playwright) give agents raw HTML. The agent then has to:

1. Parse the HTML to find relevant links
2. Decide which pages to visit
3. Extract content from each page
4. Synthesize an answer

This is slow, token-heavy, and error-prone. qdoc does all of this internally with a purpose-built pipeline: index parsing, LLM-guided page selection, content extraction, and multi-turn research — all optimized for documentation specifically.

## How does qdoc handle long documentation pages?

Each page is truncated to 12,000 characters after content extraction. Navigation, headers, footers, and sidebars are stripped. If the LLM needs more content from a page, it can request additional pages in the next research turn.

## Can I use qdoc in CI/CD pipelines?

Yes. Headless mode (`--no-tui` or `--json`) is designed for automation:

```bash
qdoc --no-tui go "module versioning" > /tmp/answer.md
qdoc --json go "module versioning" | jq -r '.answer' > /tmp/answer.md
```

Set API keys via environment variables in your CI environment. Exit code 0 on success, 1 on failure.

## Where are logs stored?

Session logs are stored at `~/.config/qdoc/logs/` as per-query `.log` files. Each query gets its own log file with timestamps and research steps.
