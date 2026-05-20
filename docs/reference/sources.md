# Documentation Sources

## Built-in Sources

### Go (`go`)

The official Go documentation at [go.dev/doc](https://go.dev/doc/).

```bash
qdoc go "effective go patterns"
qdoc go "how do modules work?"
```

Covers: tutorials, module system, database guides, release notes, effective Go, and all `/doc/` pages. The LLM also knows about `pkg.go.dev` for standard library references.

### FastAPI (`fastapi`)

The FastAPI framework documentation at [fastapi.tiangolo.com](https://fastapi.tiangolo.com/).

```bash
qdoc fastapi "dependency injection"
qdoc fastapi "background tasks"
```

## Local Directories

Point qdoc at any directory of markdown, HTML, or text files:

```bash
qdoc ./my-project-docs "deployment instructions"
qdoc ~/notes "project architecture"
```

qdoc recursively walks the directory, indexing all `.md`, `.mdx`, `.html`, `.rst`, `.txt`, and `.adoc` files. The LLM reads the most relevant ones.

## Listing Sources

```bash
qdoc sources
# or
qdoc ls
```

## Source Resolution

Source names are case-insensitive. `qdoc Go` and `qdoc go` are the same.

For a local path, use `./` or an absolute path. qdoc detects directories automatically.

## How Sources Work

Each source defines:

- **Index URL** — the page that lists all documentation pages
- **Link prefix** — used to extract relevant links from the index page
- **System prompt** — source-specific guidance for the LLM (e.g., where to find tutorials vs. reference)

When you query a source, qdoc:

1. Fetches the index
2. Gives the LLM the index + your query
3. LLM selects relevant URLs
4. qdoc fetches those pages, extracts main content
5. LLM reads the content and answers
