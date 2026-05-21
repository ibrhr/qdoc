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

### Pydantic (`pydantic`)

Pydantic documentation at [pydantic.dev](https://pydantic.dev/docs/validation/latest/get-started/). Covers the full V2 Validation library — models, fields, validators, serialization, types, configuration, API reference, and more.

```bash
qdoc pydantic "BaseModel with field validators"
qdoc pydantic "model_dump vs model_dump_json exclude fields"
qdoc pydantic "discriminated unions with Tag"
```

The LLM receives a comprehensive system prompt covering:

- **Get Started** — Welcome, installation, why Pydantic, migration guide (V1 to V2), version policy, contributing, changelog
- **Concepts** — Models (BaseModel, generic models, RootModel, nested models), Fields (Field(), Annotated, aliases, constraints, computed fields), Validators (field and model: after/before/plain/wrap), Serialization (model_dump, custom serializers, computed fields), Types (standard lib, pydantic types, custom types), Unions (discriminated, Tag), Configuration (ConfigDict), JSON Schema, JSON, Dataclasses, Strict Mode, TypeAdapter, validate_call, Conversion Table, Settings, Performance, Experimental
- **API Reference** — BaseModel, RootModel, TypeAdapter, Field/FieldInfo, ConfigDict, functional validators/serializers, pydantic types, network types, pydantic-core (SchemaValidator, SchemaSerializer, ValidationError), pydantic-settings (BaseSettings), pydantic-extra-types (Color, Country, Payment, PhoneNumbers, etc.)
- **Internals** — Architecture (pydantic vs pydantic-core Rust backend), core schema, annotation resolution
- **Examples** — File validation (JSON, JSONL, CSV, TOML, YAML, XML, INI), web/API requests, queues, databases/ORMs, custom validators, dynamic models
- **Error Messages** — ValidationError and ErrorDetails, custom error messages, validation error types, usage error types
- **Integrations** — LLMs (llms.txt), Mypy, Pyrefly, VSCode, datamodel-code-generator, Rich, AWS Lambda

### FastAPI (`fastapi`)

FastAPI framework documentation at [fastapi.tiangolo.com](https://fastapi.tiangolo.com/). Covers the tutorial, advanced guide, deployment, and all reference pages.

```bash
qdoc fastapi "dependency injection patterns"
qdoc fastapi "background tasks and lifespan events"
qdoc fastapi "OAuth2 with JWT tokens"
```

The LLM gets a detailed prompt mapping the tutorial structure (path operations, validation, dependencies, security, databases) and advanced topics (middleware, WebSockets, templates, settings).

### Python (`python`)

Python 3 documentation at [docs.python.org/3/](https://docs.python.org/3/). Covers the full standard library, language reference, tutorial, HOWTOs, packaging, and what's new.

```bash
qdoc python "asyncio gather vs create_task"
qdoc python "dataclass field with default_factory"
qdoc python "subprocess run capture output"
```

The LLM receives a detailed system prompt covering:

- **Standard Library** — All built-in functions (`print`, `len`, `map`, `zip`, `isinstance`), built-in types (`str`, `list`, `dict`, `set`), exceptions, and every major module (`json`, `re`, `datetime`, `collections`, `pathlib`, `os`, `sys`, `subprocess`, `logging`, `argparse`, `unittest`, `typing`, `dataclasses`, `itertools`, `functools`, `asyncio`, `threading`, `multiprocessing`, `http`, `urllib`, `sqlite3`, `csv`, `pickle`, `hashlib`, `struct`, and many more)
- **Language Reference** — Data model (special methods like `__init__`, `__str__`, `__eq__`, `__enter__`, `__iter__`), lexical analysis, execution model (scopes LEGB), expressions (walrus operator, lambdas, f-strings), compound statements (`if`/`for`/`try`/`with`/`match`), imports
- **Tutorial** — Beginner walkthrough covering classes, modules, errors, stdlib tour
- **HOWTOs** — Logging, regex, sorting, unicode, functional programming, descriptors, enums, argparse
- **Packaging** — Installing (`pip`, `requirements.txt`), distributing (`pyproject.toml`, `setuptools`), virtual environments (`venv`)
- **C API** — Extending Python with C and embedding Python
- **What's New** — Release notes (3.14 through 3.0)

### Next.js (`nextjs`)

Next.js documentation at [nextjs.org/docs](https://nextjs.org/docs/). Covers the full App Router and Pages Router, Getting Started, Guides, and API Reference.

```bash
qdoc nextjs "app router layouts and pages"
qdoc nextjs "server actions form mutations"
qdoc nextjs "revalidatePath vs revalidateTag"
```

The LLM receives a comprehensive system prompt covering:

- **Getting Started** — Installation, project structure, layouts, linking, server/client components, data fetching, caching, revalidation, error handling, CSS, images, fonts, metadata, route handlers, deployment, upgrading
- **Guides** — Authentication, forms (Server Actions), database integration, rendering strategies (SSG/ISR/SSR/PPR), streaming, caching, internationalization, SEO, testing (Cypress, Jest, Playwright, Vitest), deployment (Vercel, self-hosting), migrations (CRA, Vite, Pages Router), AI agents, MDX, PWA, Tailwind
- **API Reference** — Directives (`'use client'`, `'use server'`, `'use cache'`), components (`<Image>`, `<Link>`, `<Script>`, `<Form>`, `<Font>`), functions (`cookies`, `headers`, `fetch`, `generateMetadata`, `generateStaticParams`, `redirect`, `revalidatePath`, `useRouter`, `useParams`), file conventions (`page.js`, `layout.js`, `loading.js`, `error.js`, `route.js`), config (`next.config.js`)
- **Pages Router** — Legacy router docs available but App Router is preferred

### React (`react`)

React documentation at [react.dev](https://react.dev/). Covers the full Learn tutorial, API Reference, Rules of React, React Server Components, React Compiler, and ESLint plugin rules.

```bash
qdoc react "rules of hooks"
qdoc react "useCallback vs useMemo"
qdoc react "React Server Components patterns"
```

The LLM receives a comprehensive system prompt covering:

- **Learn** — Quick Start, installation, describing the UI, adding interactivity, managing state, escape hatches
- **API Reference** — All hooks (`useState` through `useSyncExternalStore`), components (`<Fragment>`, `<Suspense>`, `<StrictMode>`), APIs (`createContext`, `memo`, `lazy`, `cache`, `use`), legacy APIs
- **React DOM** — Client APIs (`createRoot`, `hydrateRoot`), server APIs (`renderToReadableStream`, `renderToString`), static APIs (`prerender`), form components (`<form>`, `<input>`, `<select>`)
- **Rules of React** — Purity, component/Hook calling rules, Rules of Hooks
- **React Server Components** — Server Components, Server Functions, `'use client'` / `'use server'` directives
- **React Compiler** — Configuration, `'use memo'` / `'use no memo'` directives

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
