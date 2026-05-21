# Documentation Sources

## Built-in Sources

### Go (`go`)

Official Go documentation at [go.dev/doc](https://go.dev/doc/). Covers the standard library, toolchain, module system, tutorials, release notes, and all `/doc/` pages.

```bash
qdoc go "effective Go patterns"
qdoc go "how do modules work?"
qdoc go "generics type constraints"
```

### Python (`python`)

Python 3 documentation at [docs.python.org/3/](https://docs.python.org/3/). Covers the full standard library, language reference, tutorial, HOWTOs, packaging, and what's new.

```bash
qdoc python "asyncio gather vs create_task"
qdoc python "dataclass field with default_factory"
qdoc python "subprocess run capture output"
```

### Next.js (`nextjs`)

Next.js documentation at [nextjs.org/docs](https://nextjs.org/docs/). Covers the full App Router and Pages Router, Getting Started, Guides, and API Reference.

```bash
qdoc nextjs "app router layouts and pages"
qdoc nextjs "server actions form mutations"
qdoc nextjs "revalidatePath vs revalidateTag"
```

### React (`react`)

React documentation at [react.dev](https://react.dev/). Covers the full Learn tutorial, API Reference, Rules of React, React Server Components, React Compiler, and ESLint plugin rules.

```bash
qdoc react "rules of hooks"
qdoc react "useCallback vs useMemo"
qdoc react "React Server Components patterns"
```

### FastAPI (`fastapi`)

FastAPI framework documentation at [fastapi.tiangolo.com](https://fastapi.tiangolo.com/). Covers the tutorial, advanced guide, deployment, and all reference pages.

```bash
qdoc fastapi "dependency injection patterns"
qdoc fastapi "background tasks and lifespan events"
qdoc fastapi "OAuth2 with JWT tokens"
```

### Pydantic (`pydantic`)

Pydantic documentation at [pydantic.dev](https://pydantic.dev/docs/validation/latest/get-started/). Covers the full V2 Validation library — models, fields, validators, serialization, types, configuration, and API reference.

```bash
qdoc pydantic "BaseModel with field validators"
qdoc pydantic "model_dump vs model_dump_json exclude fields"
qdoc pydantic "discriminated unions with Tag"
```

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

Each built-in source defines an index URL (parsed for page links), a link prefix (to stay within scope), and a system prompt (source-specific guidance for the LLM).

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

### Content Extraction

For HTML pages, qdoc extracts the main content area using heuristics (`<main>`, `<article>`, content divs) and strips navigation, headers, footers, and sidebars. Content is truncated to 12,000 characters per page to stay within reasonable context windows while retaining substantive documentation.
