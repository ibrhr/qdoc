# Contributing

## Building from Source

Requires Go 1.26+.

```bash
git clone https://github.com/ibrhr/qdoc.git
cd qdoc
go build ./...
```

## Running Tests

191 tests across 9 packages. Table-driven, standard library `testing` only.

```bash
go test ./...           # run all tests
go test -race ./...     # with race detector
```

| Package | Coverage | Focus |
|---------|----------|-------|
| `retry` | 100% | Backoff, retryable errors |
| `provider` | 100% | Find, ResolveClient, env overrides |
| `sessionlog` | 93% | New, Log, sanitize, close |
| `llm` | 86% | ParseResponses, BuildSystemPrompt, Client.Send/Stream |
| `config` | 74% | Load, Save, Default, nil-map normalization |
| `docsource` | 72% | extractLinks, extractMainContent, FetchIndex/FetchContent |
| `tui` | 17% | renderMarkdown, handleStreamDelta, key handling |
| `runner` | 14% | Run(unknown source), Step/Result types |

## Linting

```bash
go vet ./...
```

## Running Locally

```bash
go run . go "how do channels work?"
go run . --no-tui python "asyncio vs threading"
```

## Adding a Documentation Source

Built-in sources are defined in `internal/docsource/source.go` in the `KnownSources` slice. Each source needs:

- **Name** — the CLI identifier (e.g., `"go"`, `"python"`)
- **URL** — the base URL of the documentation
- **Index URL** — the page that lists all documentation pages (parsed for links)
- **Link Prefix** — a filter to keep only relevant links
- **System Prompt** — source-specific guidance for the LLM (directory structure, conventions)

Example:

```go
{
    Name:      "myframework",
    URL:       "https://myframework.dev/docs",
    IndexURL:  "https://myframework.dev/docs",
    LinkPrefix: "/docs/",
    Prompt:    "myframework documentation covers...",
}
```

After adding, rebuild and test:

```bash
go build ./...
go run . myframework "your query"
```

## Adding a Provider

Providers are defined in `internal/provider/provider.go`. Each provider needs:

- **Name** — the CLI identifier (e.g., `"openai"`)
- **Default Model** — the model used when none is specified
- **API URL** — the OpenAI-compatible endpoint
- **Env Var** — the environment variable for the API key

## Pull Requests

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Write tests for new functionality
4. Run `go test ./...` and `go vet ./...`
5. Commit with a descriptive message
6. Push and open a PR

## Docs Site

The docs site is built with VitePress:

```bash
cd docs
npm install
npx vitepress dev    # http://localhost:5173
npx vitepress build  # output: docs/.vitepress/dist/
```

Config at `docs/.vitepress/config.ts`.
