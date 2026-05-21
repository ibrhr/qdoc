---
description: Research a library/framework's entire documentation site, generate a comprehensive SystemPrompt for qdoc's agent, and wire it into the codebase
agent: build
---

Research the **$ARGUMENTS** documentation site thoroughly. This is a multi-phase task. Complete each phase before moving to the next.

## Phase 1: Discover the documentation site

Search the web to find the official documentation URL for $ARGUMENTS. Once found:
- Note the base URL (e.g. `https://docs.example.com`)
- Note the index page URL (the page listing all doc pages — usually the root or `/docs`)
- Note any link prefix that filters to doc-only pages (e.g. `/docs/`, `/latest/`)

Confirm the URL is accessible by fetching it.

## Phase 2: Traverse and map the documentation

Fetch the index page and study ALL available documentation pages. Your goal is to understand where every piece of information lives. Specifically:

1. Categories of docs (getting-started, guides, API reference, tutorials, advanced, deployment, etc.)
2. Sub-paths under each category and what they cover
3. Hidden/less-obvious sections (changelogs, migration guides, internal architecture, community pages, contributing guides, security policies, deprecation notices)
4. Special pages (glossary, FAQ, troubleshooting, common-errors, best-practices, style-guide)
5. Version-specific pages (e.g. `/v2/`, `/v3/`, `/next/`)
6. How the docs differentiate between major versions or editions (e.g. App Router vs Pages Router in Next.js, or Standard Library vs Language Reference in Python)
7. Non-obvious docs: CLI references, configuration file options, environment variables, editor/IDE setup, testing guides, benchmarking, deployment checklists

Be thorough. Read 5-10 actual pages to understand the content depth and style. Note:
- How deeply nested sections go
- Whether pages are self-contained or reference each other heavily
- Whether examples are runnable or illustrative
- Whether there are interactive playgrounds or sandboxes

## Phase 3: Analyze and take notes

Study the documentation structure thoroughly in your working context (no persisted file). Note:
- The full categorized layout of sections and sub-paths
- Hidden/non-obvious sections (changelogs, migration guides, security policies, deprecation notices, contributing guides, glossary, FAQ)
- Quirks & gotchas (URL overlaps, deprecated pages still indexed, redirects, sparse pages, JS-rendered content)
- Content quality: shallow vs deep coverage, example quality, cross-referencing habits

These notes inform the SystemPrompt you will write in Phase 4. Do NOT write a report file — just hold the key findings in context.

## Phase 4: Write the SystemPrompt for qdoc

Write the `SystemPrompt` field for a new `Source` entry. Follow the exact pattern of the existing sources in `internal/docsource/source.go`. The SystemPrompt must:

1. Start with: `{Name} documentation is at {BaseURL}. All URLs in the file list below are complete — use them as-is.`
2. Map out where every category of information lives using relative paths (the `LinkPrefix` is stripped from URLs)
3. Include routing guidance: "For X questions, start with Y section. For API specifics, see Z section."
4. Be written in lowercase-hyphen format for directory/file names EXCEPT when the actual doc site uses something else
5. End with: `Pick the most relevant pages and read them. Use the full URLs from the list below.`
6. Be comprehensive but not excessive — similar verbosity to the FastAPI prompt (~80 lines), not the Python prompt (~140 lines)

Use the existing prompts in `internal/docsource/source.go` as your style reference. Follow their conventions exactly:
- Indent directory paths consistently
- Use `—` em dashes for descriptions
- Group related sections
- Provide concrete path examples, not just categories
- Include routing hints at the bottom

## Phase 5: Wire into the codebase

### 5a. Add the Source to KnownSources

In `internal/docsource/source.go`:
- Add a new `Source` entry to the `KnownSources` slice
- Place it after `fastapi` (the last entry)
- Name: lowercase version of $ARGUMENTS (e.g. `flask`, `django`, `rust`)
- BaseURL: the site's base URL
- IndexURL: the page that lists all docs
- LinkPrefix: the prefix that filters links to doc-only pages
- SystemPrompt: the prompt from Phase 4
- Local: false

### 5b. Update docs/reference/sources.md

Add a new section for $ARGUMENTS following the pattern of the existing entries. Each source section needs:
- A `###` heading with the source name in title case and the key in backticks
- A brief description line with a link to the docs site
- 3 example `qdoc <name> "query"` commands
- A paragraph or bullet list describing what the LLM's system prompt covers

### 5c. Update docs/reference/cli.md

Add $ARGUMENTS to the examples list in the Query section (line 14-20 in the file).

### 5d. Update docs/guide/agent-usage.md

Add $ARGUMENTS to any source listings in this file (check all mentions of available sources like `go`, `python`, `fastapi`, `react`, `nextjs`).

## Phase 6: Verify

```bash
go build ./... && go vet ./...
```

Fix any compilation or vet errors. The SystemPrompt is a Go string literal — ensure:
- All double-quotes inside the string are escaped (`\"`)
- Backticks in Go code use raw string literals
- The string compiles correctly

After the build passes, do a final review:
- Re-read Phases 4-5 and verify every instruction was followed
- The SystemPrompt is placed correctly within the Source struct literal
- All doc updates use the correct markdown syntax and VitePress routing
- No extraneous report files were written to disk
