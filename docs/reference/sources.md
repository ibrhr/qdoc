# Documentation Sources

## Built-in Sources

### Go (`go`)

Official Go documentation at [go.dev/doc](https://go.dev/doc/). Covers the standard library, toolchain, module system, tutorials, release notes, and all `/doc/` pages.

```bash
qdoc go "effective Go patterns"
qdoc go "how do modules work?"
qdoc go "generics type constraints"
```

The LLM is given a system prompt mapping the `/doc/` directory structure (tutorials, modules, database guides, release notes, security docs, etc.) and knows about `pkg.go.dev` for standard library package references.

### FastAPI (`fastapi`)

FastAPI framework documentation at [fastapi.tiangolo.com](https://fastapi.tiangolo.com/). Covers the tutorial, advanced guide, deployment, and all reference pages.

```bash
qdoc fastapi "dependency injection patterns"
qdoc fastapi "background tasks and lifespan events"
qdoc fastapi "OAuth2 with JWT tokens"
```

The LLM gets a detailed prompt mapping the tutorial structure (path operations, validation, dependencies, security, databases) and advanced topics (middleware, WebSockets, templates, settings).

## Local Directories

Point qdoc at any directory of documentation files:

```bash
qdoc ./my-project-docs "deployment instructions"
qdoc ~/dev/project/docs "API authentication"
```

qdoc recursively walks the directory and indexes all files with these extensions:

- `.md`, `.mdx` — Markdown
- `.html`, `.htm` — HTML
- `.rst` — reStructuredText
- `.txt` — Plain text
- `.adoc` — AsciiDoc

The LLM sees the relative file paths as an index and selects the most relevant ones. qdoc reads them, extracts meaningful content, and passes it back to the LLM.

## Listing Sources

```bash
qdoc sources
# or
qdoc ls
```

## Source Resolution

Source names are case-insensitive. `qdoc Go` and `qdoc go` resolve to the same built-in source.

For local directories, prefix the path with `./` or use an absolute path. qdoc detects directories automatically and treats them as local doc sources.

## How Sources Work

Each built-in source defines:

| Component | Purpose |
|-----------|---------|
| **Index URL** | Page that lists all documentation pages (parsed for links) |
| **Link Prefix** | Filters links to stay within the documentation scope |
| **System Prompt** | Source-specific guidance for the LLM (directory structure, conventions) |

### Research Pipeline

```
Source.Query("how do channels work?")
  │
  ├─ 1. Fetch Index
  │     GET go.dev/doc/ → extract 245 page links
  │
  ├─ 2. LLM: Select Pages
  │     Prompt: "Here are 245 pages. User asks about channels.
  │              Which pages are most relevant?"
  │     Response: "effective_go, blog/pipelines, ref/spec"
  │
  ├─ 3. Parallel Fetch + Extract
  │     GET go.dev/doc/effective_go → extract main content (12k chars)
  │     GET go.dev/blog/pipelines   → extract main content
  │     GET go.dev/ref/spec         → extract main content
  │
  ├─ 4. LLM: Read + Decide
  │     Prompt: "Here's the content of 3 pages. Is this enough
  │              to answer about channels? If not, which pages?"
  │     Response: "I need one more: A Tour of Go, concurrency"
  │
  ├─ 5. Fetch More (up to 5 turns)
  │
  └─ 6. LLM: Synthesize Answer
        Response: "## Channels in Go\n\n..." (with citations)
```

## Content Extraction

For HTML pages, qdoc extracts the main content area using heuristics (`<main>`, `<article>`, content divs) and strips navigation, headers, footers, and sidebars. Content is truncated to 12,000 characters per page to stay within reasonable context windows while retaining substantive documentation.
