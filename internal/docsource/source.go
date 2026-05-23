package docsource

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ibrhr/qdoc/internal/retry"
)

const maxContentChars = 12000

type Entry struct {
	URL   string
	Title string
}

type Source struct {
	Name         string
	BaseURL      string
	IndexURL     string
	LinkPrefix   string
	Local        bool
	SystemPrompt string
}

var KnownSources = []Source{
	{
		Name:       "pydantic",
		BaseURL:    "https://pydantic.dev/docs/validation/latest",
		IndexURL:   "https://pydantic.dev/docs/validation/latest/get-started/",
		LinkPrefix: "/docs/validation/latest/",
		SystemPrompt: `Pydantic documentation is at https://pydantic.dev/docs/validation/latest. All URLs in the file list below are complete — use them as-is.

Pydantic is a Python data validation library (v2). The docs live under /docs/validation/latest/.

Where things live:

  get-started/                  — Getting started
    get-started/                  Welcome — overview, examples, who uses Pydantic
    get-started/why/              Why Pydantic — type hints, speed, JSON Schema, strict/lax, ecosystem
    get-started/install/          Installation
    get-started/migration/        Migration Guide — V1 to V2 breaking changes (BaseModel, Field,
                                  validators, config, types, JSON schema, GenericModel, dataclasses)
    get-started/version-policy/   Version policy
    get-started/contributing/     Contributing guide
    get-started/changelog/        Changelog

  concepts/                     — User guide — where most "how do I" answers live
    concepts/models/              Models (BaseModel, data conversion, extra data, nested models,
                                  generic models, RootModel, dynamic creation, immutability,
                                  model_copy, model_validate, model_dump, model_construct)
    concepts/fields/              Fields (Field(), Annotated pattern, default values, aliases,
                                  constraints, strict, frozen, exclude, deprecated, computed_field,
                                  discriminator, validate_default)
    concepts/validators/          Validators (field: after/before/plain/wrap; model: after/before/wrap;
                                  ValidationInfo, validation context, raising errors)
    concepts/serialization/       Serialization (model_dump, model_dump_json, custom serializers,
                                  field_serializer, model_serializer, computed_field, include/exclude)
    concepts/types/               Types (standard lib types, pydantic types, constrained types,
                                  custom types with __get_pydantic_core_schema__)
    concepts/unions/              Unions (discriminated unions, Tag, union modes: smart/left_to_right)
    concepts/alias/               Aliases (alias, validation_alias, serialization_alias, alias_generator)
    concepts/config/              Configuration (ConfigDict — all options: extra, frozen, strict,
                                  from_attributes, validate_default, str_max_length, etc.)
    concepts/json_schema/         JSON Schema (generation, field-level/model-level customization)
    concepts/json/                JSON (parsing, serialization, JSON mode vs Python mode validation)
    concepts/dataclasses/         Dataclasses (pydantic dataclasses, stdlib dataclasses, config)
    concepts/forward_annotations/ Forward annotations, self-referencing/recursive models, model_rebuild
    concepts/strict_mode/         Strict mode (strict vs lax, per-field, per-call, per-model)
    concepts/type_adapter/        TypeAdapter — validate non-BaseModel types (replaces parse_obj_as)
    concepts/validation_decorator/ validate_call — validate function arguments (replaces validate_arguments)
    concepts/conversion_table/    Conversion table — all type coercion rules
    concepts/pydantic_settings/   Settings management (pydantic-settings, BaseSettings, env vars, .env)
    concepts/performance/         Performance tips (Sequence vs list, model_validate vs construct)
    concepts/experimental/        Experimental features

  api/                          — API reference
    api/pydantic/base_model/      BaseModel (all methods and attributes)
    api/pydantic/root_model/      RootModel
    api/pydantic/dataclasses/     Pydantic dataclass decorator
    api/pydantic/type_adapter/    TypeAdapter
    api/pydantic/validate_call/   validate_call decorator
    api/pydantic/fields/          Field(), FieldInfo, computed_field decorator
    api/pydantic/aliases/         Alias-related types
    api/pydantic/config/          ConfigDict — all config options documented
    api/pydantic/json_schema/     GenerateJsonSchema, JSON schema utilities
    api/pydantic/errors/          Error types
    api/pydantic/functional_validators/  AfterValidator, BeforeValidator, WrapValidator, PlainValidator, field_validator, model_validator
    api/pydantic/functional_serializers/ field_serializer, model_serializer
    api/pydantic/standard_library_types/ Standard library type behaviors and constraints
    api/pydantic/types/           Pydantic-specific types (EmailStr, UrlConstraints, etc.)
    api/pydantic/networks/        Network types (AnyUrl, HttpUrl, FileUrl, etc.)
    api/pydantic/version/         Version info
    api/pydantic/annotated_handlers/  GetCoreSchemaHandler, GetJsonSchemaHandler
    api/pydantic/experimental/    Experimental API
    api/pydantic-core/pydantic_core/  SchemaValidator, SchemaSerializer, ValidationError, PydanticCustomError
    api/pydantic-core/pydantic_core_schema/  Core schema TypedDict definitions
    api/pydantic_settings/        Pydantic Settings (BaseSettings, SettingsConfigDict)
    api/pydantic-extra-types/     Extra types: Color, Country, Payment, PhoneNumbers, Coordinate,
                                  MacAddress, ISBN, Pendulum, CurrencyCode, LanguageCode, ULID, etc.

  internals/                    — Internals (targeted at contributors)
    internals/architecture/       pydantic vs pydantic-core, core schema, validation/serialization pipeline
    internals/resolving_annotations/  Annotation resolution process

  examples/                     — Real-world usage examples
    examples/files/               Validating JSON, JSONL, CSV, TOML, YAML, XML, INI data
    examples/requests/            Web and API request validation
    examples/queues/              Queue data validation
    examples/orms/                Database/ORM integration (SQLAlchemy)
    examples/custom_validators/   Custom validator patterns
    examples/dynamic_models/      Dynamic model creation
    examples/pydantic_ai/         Pydantic AI agent integration

  errors/                       — Error handling
    errors/errors/                ValidationError, ErrorDetails, customizing error messages
    errors/validation_errors/     All validation error types
    errors/usage_errors/          All usage error types

  integrations/                 — Tool and platform integrations
    integrations/llms/            llms.txt and llms-full.txt for LLM consumption
    integrations/dev-tools/mypy/  Mypy plugin
    integrations/dev-tools/pyrefly/  Pyrefly type checker
    integrations/dev-tools/visual_studio_code/  VSCode integration
    integrations/dev-tools/datamodel_code_generator/  Code generation from JSON Schema
    integrations/dev-tools/rich/  Rich console integration
    integrations/dev-tools/linting/  Linting
    integrations/aws_lambda/      AWS Lambda integration

For "how do I create/use/configure a model" questions, start with concepts/models/ and concepts/fields/. For validation logic, see concepts/validators/. For serialization/export, see concepts/serialization/. For specific class/method signatures, see api/pydantic/. For V1 to V2 migration, see get-started/migration/. For file format validation, see examples/files/. For configuration options, see concepts/config/ or api/pydantic/config/.

Pick the most relevant pages and read them. Use the full URLs from the list below.`,
	},
	{
		Name:       "go",
		BaseURL:    "https://go.dev/doc",
		IndexURL:   "https://go.dev/doc/",
		LinkPrefix: "/doc/",
		SystemPrompt: `Go documentation is at https://go.dev/doc. All URLs in the file list below are complete — use them as-is.

Where things live under /doc/:
  tutorial/     — Step-by-step tutorials (getting-started, create-module, workspaces, web-service-gin, generics, etc.)
  modules/      — Module system (layout, dependencies, go.mod, publishing, versioning, etc.)
  database/     — Database access guides
  articles/     — In-depth articles (wiki, race_detector, go_command)
  codewalk/     — Guided code walkthroughs
  security/     — Security docs (fuzz, fips140)
  go1.X         — Release notes per Go version
  Top-level     — effective_go, code, install, faq, gc-guide, godebug, pgo, asm, gdb, editors, diagnostics, contribute, devel/release

Standard library packages: pkg.go.dev/<import-path>. Include the URL when relevant.

Pick the most relevant pages and read them. Use the full URLs from the list below.`,
	},
	{
		Name:       "nextjs",
		BaseURL:    "https://nextjs.org/docs",
		IndexURL:   "https://nextjs.org/docs",
		LinkPrefix: "/docs/",
		SystemPrompt: `Next.js documentation is at https://nextjs.org/docs. All URLs in the file list below are complete — use them as-is.

Next.js is a React framework with two routers:

  App Router (/docs/app/) — MODERN. Uses React Server Components, streaming, Server Actions.
                            Always prefer this for new questions unless the user says "Pages Router."
  Pages Router (/docs/pages/) — LEGACY. Traditional React SSR. Only use if explicitly asked.

Where things live under /docs/app/:

  getting-started/              — Core concepts (installation, project-structure, layouts-and-pages,
                                  linking-and-navigating, server-and-client-components,
                                  fetching-data, mutating-data, caching, revalidating,
                                  error-handling, css, images, fonts, metadata, route-handlers,
                                  proxy, deploying, upgrading)

  guides/                       — Task-focused tutorials
    authentication, authorization patterns, middleware for auth
    forms and mutations (Server Actions, useFormStatus, useActionState)
    database integration, data fetching patterns, data security
    rendering strategies (SSG/ISR/SSR/PPR), streaming, static exports
    caching, revalidation, stale-while-revalidate
    routing patterns (dynamic routes, parallel routes, intercepting routes, middleware)
    internationalization (i18n), SEO, metadata, OG images
    environment variables, draft mode, preview mode
    testing (Cypress, Jest, Playwright, Vitest)
    deployment (Vercel, self-hosting, Docker, static hosting)
    AI agents, OpenTelemetry, MDX, PWA, SPAs, Tailwind CSS, Sass, CSS-in-JS
    migrating (from CRA, Vite, Pages Router; codemods for v14→v15→v16)

  api-reference/                — Technical reference
    directives/                   'use client', 'use server', 'use cache'
    components/                   <Image>, <Link>, <Script>, <Form>, <Font>
    functions/                    cookies, headers, fetch, generateMetadata, generateStaticParams,
                                  redirect, notFound, revalidatePath, revalidateTag,
                                  useRouter, useParams, usePathname, useSearchParams,
                                  unstable_cache, after, draftMode, connection
    file-conventions/             page.js, layout.js, loading.js, error.js, not-found.js,
                                  route.js, template.js, default.js, forbidden.js, unauthorized.js
    metadata-files/               sitemap.xml, robots.txt, opengraph-image, manifest.json
    route-segment-config/         dynamic, revalidate, runtime, preferredRegion, maxDuration
    next-config-js/               all next.config.js options (basePath, redirects, rewrites,
                                  images, headers, turbopack, webpack, output, etc.)
    cli/                          create-next-app, next CLI commands

  architecture/                 — Internals (accessibility, Fast Refresh, compiler, supported browsers)
  community/                    — Contribution guide, Rspack

Pages Router structure (only when explicitly requested):
  /docs/pages/building-your-application/  — Routing, rendering, data fetching (getStaticProps, etc.)
  /docs/pages/api-reference/             — Components, functions, config for Pages Router

For "how do I X" questions, start with /docs/app/getting-started/ for fundamentals or /docs/app/guides/ for specific use cases. For API signatures, parameters, or config options, use /docs/app/api-reference/. Prefer App Router answers unless Pages Router is explicitly mentioned.

Pick the most relevant pages and read them. Use the full URLs from the list below.`,
	},
	{
		Name:       "python",
		BaseURL:    "https://docs.python.org/3",
		IndexURL:   "https://docs.python.org/3/",
		LinkPrefix: "",
		SystemPrompt: `Python 3 documentation is at https://docs.python.org/3/. All URLs in the file list below are complete — use them as-is.

Where things live under docs.python.org/3/:

  tutorial/                 — The Python Tutorial (classes, modules, errors, stdlib tour, etc.)
  library/                  — The Python Standard Library — THIS IS WHERE MOST ANSWERS LIVE
    library/functions.html    Built-in functions (print, len, range, zip, map, filter, sorted,
                              open, isinstance, hasattr, getattr, enumerate, any, all, etc.)
    library/stdtypes.html     Built-in types (str, int, float, list, dict, set, tuple, bool,
                              bytes, bytearray, memoryview, type, range, slice, frozenset)
    library/exceptions.html   Built-in exceptions (Exception, ValueError, TypeError, KeyError,
                              IndexError, OSError, RuntimeError, StopIteration, etc.)
    library/datetime.html     Date, time, datetime, timedelta, timezone
    library/collections.html  namedtuple, deque, Counter, defaultdict, OrderedDict, ChainMap
    library/collections.abc.html  Abstract base classes for containers
    library/math.html         Math (sqrt, ceil, floor, isclose, comb, perm, etc.) + cmath
    library/statistics.html   Statistics (mean, median, stdev, quantiles, correlation)
    library/random.html       Random numbers (random, randint, choice, shuffle, sample)
    library/re.html           Regular expressions (search, match, findall, sub, compile)
    library/json.html         JSON (loads, dumps, load, dump, JSONDecoder, JSONEncoder)
    library/os.html           OS interface (environ, chdir, listdir, mkdir, walk, stat, etc.)
    library/os.path.html      Path manipulation (join, split, exists, isfile, basename, etc.)
    library/pathlib.html      Object-oriented filesystem paths (Path, read_text, glob, iterdir)
    library/sys.html          System-specific (argv, path, version, stdin/stdout, exit, modules)
    library/subprocess.html   Subprocess management (run, Popen, CalledProcessError)
    library/shutil.html       High-level file ops (copy, move, rmtree, make_archive)
    library/io.html           I/O tools (open, TextIOWrapper, BytesIO, StringIO)
    library/logging.html      Logging (getLogger, handlers, formatters, filters)
    library/logging.handlers.html  Logging handlers (StreamHandler, FileHandler, RotatingFileHandler)
    library/argparse.html     CLI argument parsing (ArgumentParser, add_argument)
    library/configparser.html Config files (ConfigParser, read, sections, get)
    library/tempfile.html     Temp files and dirs (TemporaryFile, NamedTemporaryFile, mkdtemp)
    library/glob.html         File globbing (glob, iglob)
    library/fnmatch.html      Filename pattern matching
    library/unittest.html     Unit testing (TestCase, assertEqual, mock, setUp, tearDown)
    library/unittest.mock.html  Mock objects (Mock, MagicMock, patch, sentinel, call)
    library/doctest.html      Test interactive Python examples
    library/typing.html       Type hints (Any, Union, Optional, Callable, Protocol, TypedDict,
                              Sequence, Mapping, TypeVar, Generic, overload, cast, Final, Literal)
    library/dataclasses.html  Data classes (dataclass, field, asdict, astuple)
    library/enum.html         Enumerations (Enum, IntEnum, StrEnum, Flag, auto)
    library/itertools.html    Iterator tools (chain, cycle, product, combinations, permutations,
                              groupby, islice, zip_longest, accumulate, tee)
    library/functools.html    Higher-order functions (lru_cache, cache, partial, reduce, wraps,
                              singledispatch, total_ordering, cmp_to_key)
    library/operator.html     Standard operators as functions (itemgetter, attrgetter, methodcaller)
    library/contextlib.html   Context managers (contextmanager, suppress, redirect_stdout, ExitStack)
    library/threading.html    Thread-based parallelism (Thread, Lock, Event, Condition, Semaphore)
    library/multiprocessing.html  Process-based parallelism (Process, Pool, Queue, Pipe, Manager)
    library/concurrent.futures.html  Executor, ThreadPoolExecutor, ProcessPoolExecutor
    library/asyncio.html      Async I/O (run, create_task, gather, sleep, Queue, Streams, locks)
    library/asyncio-task.html Coroutines, Tasks, TaskGroup, shielding
    library/asyncio-stream.html  Streams (open_connection, start_server)
    library/asyncio-sync.html    Sync primitives (Lock, Event, Condition, Semaphore)
    library/asyncio-queue.html   Queues
    library/socket.html       Low-level networking (socket, connect, bind, listen, accept)
    library/ssl.html          TLS/SSL (SSLContext, wrap_socket)
    library/http.client.html  HTTP client (HTTPConnection, HTTPSConnection)
    library/http.server.html  HTTP server (HTTPServer, BaseHTTPRequestHandler)
    library/urllib.request.html  URL opening (urlopen, Request)
    library/urllib.parse.html    URL parsing (urlparse, urlencode, quote)
    library/xml.etree.elementtree.html  XML (Element, SubElement, parse, iterfind)
    library/xml.html           Other XML modules
    library/html.parser.html   HTML parser
    library/html.html          HTML escaping
    library/sqlite3.html       SQLite (connect, execute, executemany, Row, backup)
    library/csv.html            CSV (reader, writer, DictReader, DictWriter, Sniffer)
    library/pickle.html         Object serialization (dump, load, dumps, loads)
    library/shelve.html         Persistent dict
    library/copy.html           Shallow and deep copy
    library/decimal.html        Decimal fixed-point arithmetic
    library/fractions.html      Rational numbers
    library/hashlib.html        Secure hashes (md5, sha256, sha3, pbkdf2, scrypt)
    library/hmac.html           Keyed-Hashing for Message Authentication
    library/base64.html         Base64, Base32, Base16 encoding
    library/struct.html         Pack/unpack binary data (pack, unpack, calcsize)
    library/binascii.html       Binary/ASCII conversion
    library/textwrap.html       Text wrapping and filling
    library/difflib.html        Sequence comparison (unified_diff, context_diff, get_close_matches)
    library/gc.html             Garbage collection
    library/traceback.html      Print or retrieve stack traces
    library/inspect.html        Inspect live objects (signature, getmembers, getsource)
    library/abc.html            Abstract base classes
    library/weakref.html        Weak references
    library/types.html          Dynamic type creation (ModuleType, FunctionType, SimpleNamespace)
    library/codecs.html         Codec registry (encode, decode, open with encoding)
    library/pprint.html         Pretty-print (pprint, pformat)
    library/time.html           Time access (time, sleep, perf_counter, strftime, strptime)
    library/calendar.html       Calendar functions
    library/zoneinfo.html       IANA time zone support
    library/uuid.html           UUID objects
    library/gzip.html           gzip compression
    library/zipfile.html        ZIP archive handling
    library/tarfile.html        Tar archive handling
    library/zlib.html           zlib compression
    library/bz2.html            bzip2 compression
    library/lzma.html           LZMA compression
    library/text.html           String services overview
    library/string.html         String constants (ascii_letters, digits), Formatter, Template
    library/datatype.html       Data types overview
    library/numeric.html        Numeric and math overview
    library/filesys.html        File and directory overview
    library/concurrency.html    Concurrency overview
    library/netdata.html        Networking and internet overview
    library/ipaddress.html      IPv4/IPv6 manipulation
    library/email.html          Email (message, parser, MIME)
    library/imaplib.html        IMAP4 protocol client
    library/venv.html           Virtual environment creation
    library/ensurepip.html      Bootstrapping pip
    library/tkinter.html        Tk GUI toolkit
    library/pdb.html            Python debugger
    library/profile.html        Python profiler
    library/dis.html             Disassembler
    library/ast.html             Abstract Syntax Trees
    library/tokenize.html        Tokenizer
    library/importlib.html       Import machinery
    library/pkgutil.html         Package utilities
    library/warnings.html        Warning control
    library/signal.html          Signal handlers
    library/platform.html        Platform identification
    library/atexit.html          Exit handlers
    library/getpass.html         Portable password input
    library/readline.html        GNU readline interface
    library/cmd.html             Line-oriented command interpreters
    library/shlex.html           Shell-like syntax parsing

  reference/                — Language Reference
    reference/datamodel       Objects, values, types (special methods: __init__, __str__, __eq__,
                              __hash__, __call__, __getattr__, __enter__, __iter__, etc.)
    reference/lexical_analysis  Tokens, keywords, identifiers, literals, f-strings
    reference/expressions     Operators, comparisons, lambdas, await/async, yield, walrus (:=)
    reference/compound_stmts  if/while/for/try/with/match, class/function def, async for/with
    reference/simple_stmts    assignment, assert, pass, return, raise, break, continue, import
    reference/executionmodel  Naming, binding, scopes (LEGB), global/nonlocal
    reference/import          Import system, packages (__init__.py, __path__, namespace packages)

  using/                    — Setup and Usage (command line options, venv, Windows, macOS)
  howto/                    — HOWTO guides (logging, regex, sorting, unicode, functional, descriptors)
  installing/               — Installing Python packages (pip, requirements, wheels)
  distributing/             — Packaging and distributing (setup.py, pyproject.toml, setuptools)
  extending/                — C extension modules, embedding Python
  whatsnew/                 — What's New (3.14, 3.13, 3.12, ...)

For "how do I use X in Python", check library/ for the relevant standard library module. For language semantics ("what is a descriptor?", "how does __init__ work?"), see reference/datamodel or reference/. For packaging, see installing/ or distributing/. For C extensions, see extending/. For beginner tutorials, see tutorial/. For "what's new in Python 3.14", see whatsnew/.

Pick the most relevant pages and read them. Use the full URLs from the list below.`,
	},
	{
		Name:       "react",
		BaseURL:    "https://react.dev",
		IndexURL:   "https://react.dev/learn",
		LinkPrefix: "/",
		SystemPrompt: `React documentation is at https://react.dev. All URLs in the file list below are complete — use them as-is.

Where things live:
  /learn/                         — Main tutorial and conceptual guides
    /learn                        — Quick Start: 80% of daily React in one page
    /learn/tutorial-tic-tac-toe   — Build a tic-tac-toe game step by step
    /learn/thinking-in-react      — Thinking in React methodology
    /learn/installation           — Installing and setting up React (CRA, from scratch, existing project)
    /learn/setup                  — Editor setup, TypeScript, React DevTools
    /learn/describing-the-ui      — Components, JSX, props, conditional rendering, lists, purity, UI as tree
    /learn/adding-interactivity   — Events, state, render & commit, state as snapshot, queueing updates
    /learn/managing-state         — State structure, sharing, preserving, reducer, context, scaling up
    /learn/escape-hatches         — Refs, manipulating DOM, effects, custom hooks, separating events
    /learn/react-compiler        — React Compiler: intro, installation, incremental adoption, debugging

  /reference/react/               — React API Reference
    Hooks: useState, useEffect, useContext, useReducer, useCallback, useMemo,
           useRef, useImperativeHandle, useLayoutEffect, useDebugValue,
           useDeferredValue, useTransition, useId, useActionState, useOptimistic,
           useSyncExternalStore, useInsertionEffect, useEffectEvent
    Components: <Fragment> (<>), <Profiler>, <StrictMode>, <Suspense>, <Activity>, <ViewTransition>
    APIs: createContext, memo, lazy, startTransition, use, cache, cacheSignal,
          captureOwnerStack, act, addTransitionType

  /reference/react-dom/           — React DOM APIs (web/browser only)
    Hooks: useFormStatus
    Components: <form>, <input>, <select>, <textarea>, <option>, <progress>,
                <link>, <meta>, <script>, <style>, <title> (all HTML/SVG elements)
    Client APIs: createRoot, hydrateRoot
    Server APIs: renderToReadableStream, renderToPipeableStream, renderToString,
                 renderToStaticMarkup, resume, resumeToPipeableStream
    Static APIs: prerender, prerenderToNodeStream, resumeAndPrerender

  /reference/rules/               — Rules of React
    components-and-hooks-must-be-pure — Purity, idempotence, side-effect rules
    react-calls-components-and-hooks  — React owns the call schedule
    rules-of-hooks                     — Only call hooks at top level, from React functions

  /reference/rsc/                 — React Server Components
    server-components, server-functions, directives ('use client', 'use server')

  /reference/react-compiler/      — Compiler reference
    configuration, compilationMode, gating, logger, directives ('use memo', 'use no memo')

  /reference/eslint-plugin-react-hooks/ — Lint rules (exhaustive-deps, rules-of-hooks, purity, etc.)
  /reference/react/legacy         — Legacy APIs (Component, createElement, cloneElement, createRef, etc.)

For "how do I use X" questions, start with /learn/. For specific API signatures, parameters, return values, or edge-case behavior, see /reference/react/ or /reference/react-dom/. For server-side rendering, see /reference/react-dom/server. For React Server Components, see /reference/rsc/. For lint rules, see /reference/eslint-plugin-react-hooks/.

Pick the most relevant pages and read them. Use the full URLs from the list below.`,
	},
	{
		Name:       "fastapi",
		BaseURL:    "https://fastapi.tiangolo.com",
		IndexURL:   "https://fastapi.tiangolo.com/",
		LinkPrefix: "",
		SystemPrompt: `FastAPI documentation is at https://fastapi.tiangolo.com. All URLs in the file list below are complete — use them as-is.

Where things live:
  tutorial/                  — Tutorial - User Guide: step-by-step walkthrough
    tutorial/first-steps       creating a minimal app with path operations
    tutorial/path-params       declaring path parameters with type hints
    tutorial/query-params      query parameters, defaults, and optional params
    tutorial/body              request body with Pydantic models
    tutorial/*-validations     string/numeric validations for query, path, cookie, header params
    tutorial/query-param-models  grouping query params into Pydantic models
    tutorial/body-*            multiple body params, Field(), nested models
    tutorial/response-model    return type annotation and response shaping
    tutorial/extra-models      multiple models, Union types, model inheritance
    tutorial/response-status-code  setting HTTP status codes
    tutorial/request-forms     Form data, Form Models, file uploads, forms+files
    tutorial/handling-errors   HTTPException, custom error handlers
    tutorial/path-operation-configuration  tags, summary, description, operation_id
    tutorial/encoder           JSON-compatible encoder for non-Pydantic returns
    tutorial/body-updates      PATCH with partial updates (exclude_unset)
    tutorial/background-tasks  BackgroundTasks for post-response work
    tutorial/metadata          title, description, version, docs URLs, tags metadata
    tutorial/static-files      serving static files with StaticFiles
    tutorial/testing           writing tests with TestClient
    tutorial/debugging         debug mode
    tutorial/bigger-applications  structuring with APIRouter
    tutorial/middleware        adding custom middleware
    tutorial/cors              CORS configuration
    tutorial/sql-databases     SQL (SQLAlchemy) database integration
    tutorial/sse               Server-Sent Events (SSE)
    tutorial/dependencies/     Depends(), function deps, classes as deps, sub-deps, yield deps
    tutorial/security/         OAuth2, JWT, HTTP Basic, get-current-user, scopes
  advanced/                  — Advanced User Guide
    advanced/custom-response   HTMLResponse, RedirectResponse, StreamingResponse, FileResponse
    advanced/response-headers  setting custom response headers
    advanced/response-cookies  setting cookies in response
    advanced/additional-responses  multiple response models per status code
    advanced/additional-status-codes  additional status codes in OpenAPI
    advanced/websockets        WebSocket endpoints
    advanced/events           startup/shutdown lifespan events
    advanced/settings         Pydantic Settings for configuration
    advanced/templates        Jinja2 HTML templates
    advanced/middleware        advanced middleware techniques
    advanced/sub-applications  mounting sub-FastAPI apps, proxying
    advanced/behind-a-proxy   running behind nginx/traefik
    advanced/dataclasses       using dataclasses instead of Pydantic
    advanced/generate-clients  generating API clients from OpenAPI
    advanced/openapi-webhooks  OpenAPI webhook docs
    advanced/openapi-callbacks  OpenAPI callback docs
    advanced/using-request-directly  accessing the raw Request object
    advanced/wsgi              mounting WSGI apps (Flask, Django)
    advanced/async-tests       testing async code
    advanced/stream-data       streaming data
    advanced/custom-response   customizing response status codes and content type
    advanced/security/         HTTP Basic auth, OAuth2 scopes
  how-to/                    — How-To Guides
    how-to/graphql            adding GraphQL with Strawberry
    how-to/configure-swagger-ui  customizing Swagger UI
    how-to/extending-openapi   custom OpenAPI schema
    how-to/conditional-openapi  conditionally enabling OpenAPI docs
    how-to/migrate-from-pydantic-v1-to-pydantic-v2
    how-to/separate-openapi-schemas  separate input/output OpenAPI schemas
    how-to/custom-request-and-route  custom request classes and APIRoute
    how-to/testing-database    testing with a database
  deployment/                — Deployment
    deployment/docker          Docker deployment guide
    deployment/https           HTTPS setup
    deployment/concepts        deployment concepts
    deployment/server-workers  server workers (Gunicorn, Uvicorn)
    deployment/cloud           cloud deployment overview
  reference/                — API Reference (class and method docs)

For "how do I do X" questions, start with tutorial/. For advanced patterns, check advanced/. For specific class/method behavior, see reference/. For Docker, see deployment/docker.

Pick the most relevant pages and read them. Use the full URLs from the list below.`,
	},
}

func Find(name string) (Source, bool) {
	for _, s := range KnownSources {
		if strings.EqualFold(s.Name, name) {
			return s, true
		}
	}

	info, err := os.Stat(name)
	if err == nil && info.IsDir() {
		abs, _ := filepath.Abs(name)
		return Source{
			Name:     filepath.Base(abs),
			BaseURL:  "file://" + filepath.ToSlash(abs),
			IndexURL: "file://" + filepath.ToSlash(abs),
			Local:    true,
		}, true
	}

	return Source{}, false
}

func (s Source) FetchIndex() ([]Entry, error) {
	if s.Local {
		return s.fetchLocalIndex()
	}

	resp, err := retryableHTTPGet(s.IndexURL)
	if err != nil {
		return nil, fmt.Errorf("fetching index: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading index: %w", err)
	}

	entries := extractLinks(string(body), s.LinkPrefix, s.BaseURL)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].URL < entries[j].URL
	})
	return entries, nil
}

func (s Source) FetchContent(rawURL string) (string, error) {
	if s.Local {
		return s.fetchLocalContent(rawURL)
	}

	resp, err := retryableHTTPGet(rawURL)
	if err != nil {
		return "", fmt.Errorf("fetching %s: %w", rawURL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", rawURL, err)
	}

	content := extractMainContent(string(body))
	if len(content) > maxContentChars {
		content = content[:maxContentChars]
	}
	return content, nil
}

func sourceRootDir(rawURL string) string {
	return filepath.FromSlash(strings.TrimPrefix(rawURL, "file://"))
}

func (s Source) fetchLocalIndex() ([]Entry, error) {
	rootDir := sourceRootDir(s.IndexURL)

	var entries []Entry
	seen := map[string]bool{}
	extensions := map[string]bool{
		".md": true, ".mdx": true, ".html": true,
		".rst": true, ".txt": true, ".adoc": true,
	}

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			base := filepath.Base(path)
			if strings.HasPrefix(base, ".") || base == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if !extensions[ext] {
			return nil
		}
		rel, _ := filepath.Rel(rootDir, path)
		relSlash := filepath.ToSlash(rel)
		if seen[relSlash] {
			return nil
		}
		seen[relSlash] = true
		title := strings.TrimSuffix(strings.TrimSuffix(relSlash, ext), "/index")
		title = strings.ReplaceAll(title, "/", " / ")
		entries = append(entries, Entry{URL: relSlash, Title: title})
		return nil
	})

	if err != nil {
		return entries, nil
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].URL < entries[j].URL
	})
	return entries, nil
}

func (s Source) fetchLocalContent(path string) (string, error) {
	rootDir := sourceRootDir(s.BaseURL)
	fullPath := filepath.Join(rootDir, path)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", path, err)
	}

	content := string(data)
	if len(content) > maxContentChars {
		content = content[:maxContentChars]
	}
	return content, nil
}

func fetchLocal(rawURL string) (*http.Response, error) {
	filePath := sourceRootDir(rawURL)
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       f,
	}, nil
}

func retryableHTTPGet(rawURL string) (*http.Response, error) {
	if strings.HasPrefix(rawURL, "file://") {
		return fetchLocal(rawURL)
	}

	var lastErr error
	for attempt := 0; attempt < retry.FetchRetry.MaxAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(retry.BackoffDelay(retry.FetchRetry, attempt-1))
		}

		resp, err := http.Get(rawURL)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			resp.Body.Close()
			lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
			continue
		}

		if resp.StatusCode == http.StatusNotFound {
			resp.Body.Close()
			return nil, fmt.Errorf("not found: %s", rawURL)
		}

		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden ||
			resp.StatusCode == http.StatusBadRequest {
			resp.Body.Close()
			return nil, fmt.Errorf("HTTP %d fetching %s", resp.StatusCode, rawURL)
		}

		return resp, nil
	}

	return nil, fmt.Errorf("fetch retry exhausted (%d attempts): %w", retry.FetchRetry.MaxAttempts, lastErr)
}