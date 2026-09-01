# AGENTS.md — qdoc

> **qdoc - Query the Docs** — qdoc is a tui/cli that you or your coding agent can use to get fast, accurate, relevant, and up-to-date information about any library or framework you're using!

## Build & verify

```bash
export PATH=$PATH:/usr/local/go/bin  # Go 1.26.3 lives here
go build ./...   # builds all packages
go vet ./...
go test ./...    # runs all tests
go test -race ./...
```

## Docs site

```bash
cd docs
npm install
npx vitepress dev    # http://localhost:5173
npx vitepress build  # output: docs/.vitepress/dist/
```

The docs site is built with VitePress v1.6+. Config at `docs/.vitepress/config.ts`.

## Architecture

Module path: `github.com/ibrhr/qdoc`

```
main.go                          # CLI routing + print commands
internal/
  config/
    config.go                    # Config struct, load, save, configPath
  auth/
    auth.go                      # Token, AuthMethod interface, ProviderInfo, AuthStatus, ForProvider
    token_store.go               # TokenStore — disk-backed token persistence, IsExpired
    device_flow.go               # DeviceFlowAuth — GitHub Copilot device flow OAuth
    pkce.go                      # PkceAuth — PKCE OAuth (OpenAI ChatGPT subscription)
  provider/
    provider.go                  # Provider, AccessMethod structs, registry, Find, ResolveClient, error types
    providers.json               # Provider registry data
  retry/
    retry.go                     # Shared retry Config, BackoffDelay, IsRetryableHTTP/Error
  llm/
    client.go                    # Client (OpenAI-compat), NewClient, Send(), Stream()
    codex_client.go              # CodexClient — OpenAI Codex Responses API (SSE stream)
    prompt.go                    # BuildSystemPrompt()
    parse.go                     # ParsedAction, ParseResponses()
    types.go                     # ChatMessage, StreamDelta, Streamer interface, Config
  docsource/
    source.go                    # Source, Entry, KnownSources, Find, fetch methods, retryableHTTPGet
    html.go                      # extractLinks, extractMainContent (unexported HTML helpers)
  runner/
    runner.go                    # Headless query runner (--no-tui, --json)
  tui/
    model.go                     # Model struct, phase/state types, constructors (New*), Init
    update.go                    # Update + query-mode handlers + streaming pipeline
    setup.go                     # Setup update handlers + setup view renderers
    view.go                      # View + query-mode rendering (header, progress, content, footer)
    styles.go                    # Theme colors + lipgloss styles + spinner
    render.go                    # renderMarkdown, renderInlineMarkdown, wordWrap
    messages.go                  # All tea.Msg types (docIndexMsg, streamDeltaMsg, etc.)
  sessionlog/
    log.go                       # Session logging to ~/.config/qdoc/logs/ (per-query .log files)
```

### Dependency graph (no cycles)

```
config (standalone)
  ↑
provider ←→ (imports config, llm for ResolveClient)
  ↑
llm ←── docsource (for BuildSystemPrompt)
  ↑         ↑
  │         │
  ├─── tui ─┘ (imports config, provider, llm, docsource)
  ├── runner  (imports config, provider, llm, docsource)
  └── main    (imports everything)

retry (standalone) ── imported by llm, docsource
```

### Key constructor functions

- `tui.NewQuery(sourceName, query, cfg)` → query mode
- `tui.NewProviderSelect(cfg)` → provider picker
- `tui.NewModelSelect(cfg)` → model picker (multi-provider)
- `tui.NewModelSelectSingle(cfg, prov)` → model picker (single provider)
- `tui.NewSetup(cfg)` → first-run setup
- `tui.ProvidersWithKeys(cfg)` → filter providers with keys

## Bubble Tea v2 conventions (do NOT use v1 patterns)

- **Import path**: `charm.land/bubbletea/v2` (NOT `github.com/charmbracelet/bubbletea`)
- **Lip gloss**: `charm.land/lipgloss/v2` (NOT `github.com/charmbracelet/lipgloss`)
- **Key presses**: use `msg.String()` for special keys (`"up"`, `"enter"`, `"backspace"`, `"ctrl+c"`), use `msg.Text` for printable input (NOT `msg.Rune` from v1)
- **View returns**: `tea.NewView(string)` (NOT plain `string`)
- **Tea.Program**: `tea.NewProgram(m)` with `p.Run()` (returns `(tea.Model, error)`)
- **Commands**: `func() tea.Msg`, composed with `tea.Batch`, `tea.Sequence`
- **Ticks**: `tea.Tick(duration, func(time.Time) tea.Msg)` returns `tea.Cmd`

## Critical pitfalls

1. **Value vs pointer receivers**: Bubble Tea's `Update()` is `(m Model)` (value). Modifications to `m` inside the method ARE persisted because Update returns `m`. But if a *method called from Update* has a value receiver, modifications to its local copy are LOST. When in doubt, use pointer receivers for methods that mutate state (`*Model`).

2. **Channel lifecycle for streaming**: `streamCh` must be set BEFORE the goroutine starts, and the goroutine closes it. The pattern: create channel → start goroutine → return `readStreamChunk(ch)`. Each chunk triggers another `readStreamChunk` in the Update handler. When channel closes, `StreamDelta{Done: true}` is sent.

3. **Don't use glamour**: It depends on lipgloss v1 which conflicts with lipgloss v2 at the `ansi.Style` level. The custom renderer in `tui/render.go` handles headers, code blocks, inline code, lists, and block quotes.

4. **Model select skip entry**: The first item in `modelsList` is always `"— use default (modelName) —"` (index 0). Cursor 0 + enter = skip (no model override saved). Only cursor > 0 saves a model override.

5. **Provider resolution order**: `QDOC_PROVIDER` env → `cfg.Provider` → not configured. Model: `QDOC_MODEL` env → `cfg.Models[provider]` → `prov.DefaultModel`. Key: env var → `cfg.Keys[provider]`.

6. **Error display in setup**: Errors render in a red `errorBox` above the setup content. They auto-clear on any user interaction (cursor move, input). If `saveConfig` fails during setup, the error is set and the user can retry.

7. **StreamDelta is now llm.StreamDelta**: Previously `streamDeltaMsg` was defined in `messages.go` as `struct{content string; done bool; err error; retrying bool}`. Now it's `llm.StreamDelta` with exported fields: `Content`, `Done`, `Err`, `Retrying`. The LLM client field on the model is `llm.Streamer` (interface), not `*llm.Client` — use `ModelName()` / `ProviderName()` methods, not `.Model` / `.Provider` fields.

8. **renderMarkdown blank line bug**: The loop already emits `\n` between every line (`if i > 0 { sb.WriteString("\n") }`). Do NOT add another `\n` for blank lines — it doubles them. Just `continue`.

9. **Word-wrap continuation indent**: After `lipgloss.Wrap(line, maxW, " ")`, continuation lines lose the `"  "` indentation prefix. You must prepend `"  "` to all parts past the first: `parts[0]` as-is, then `"  "+wl` for each `parts[1:]`.

10. **Don't use Faint(true) on thoughtStyle**: On dark terminal backgrounds, `Faint(true)` makes text effectively invisible. Use a dim foreground color with Italic, not Faint.

11. **System prompt design for LLM actions**:
    - Avoid capitalized warnings (`ONLY valid JSON`) — triggers format paranoia
    - Don't say `"exactly one { at start and one } at end"` — models worry about parser strictness
    - Don't give contradictory number ranges (`2-3` vs `up to 3` vs `a JSON action` singular)
    - Don't use prescriptive arrow rules (`"How do I X?" → /doc/foo`) — turns models into rule-matchers instead of readers

12. **`filesPending` counter in remainingFetching**: The TUI tracks pending file fetches via a per-batch `filesPending` counter (set in `startFileFetches`, decremented in `handleDocContent`). Do NOT use `len(m.pendingReads) - len(m.readFiles)` — `readFiles` is cumulative across iterations, so the subtraction gives incorrect results on iteration 2+.

13. **Retry logic lives in `internal/retry`**: Both `llm` and `docsource` import the shared `retry` package. Use `retry.LLMRetry` for LLM API calls (3 attempts, 2s base, 30s max) and `retry.FetchRetry` for doc fetches (3 attempts, 1s base, 10s max). Use `retry.IsRetryableError(err)` to check if an error warrants a retry.

14. **`IsExpired` must use the key that found the token**: When resolving a token via `store.Get(tsKey)` with fallback to `store.Get(prov.Name)`, track which key actually found the token and use that for `IsExpired`. Otherwise, `IsExpired(tsKey)` returns `false` for a non-existent composite key, silently bypassing expiration on fallback tokens.

15. **PKCE `Refresh` must set `ExpiresAt`**: The `PkceAuth.Refresh` method receives `tr.ExpiresIn` from the token response but must explicitly set `tkn.ExpiresAt` (matching what `exchangeCode` does). Without this, refreshed tokens always appear non-expired (`IsExpired` returns `false` for zero `ExpiresAt`).

16. **Token store uses composite keys for access-method-specific tokens**: PKCE tokens are stored under `"provider:method_id"` (e.g. `"openai:subscription"`). Device flow tokens use plain provider name (e.g. `"github-copilot"`). API keys are in `cfg.Keys[provider]`. Use `provider.tokenStoreKey(provName, methodID)` to build the correct key.

17. **`resolveClient()` in TUI checks `cfg.AccessMethod`**: In `update.go`, the model's `resolveClient()` method calls `provider.ResolveClientWithMethod` when `cfg.AccessMethod` is set, otherwise falls back to `provider.ResolveClient`. Always use `m.resolveClient()` instead of calling `provider.ResolveClient` directly from the TUI.

## Streaming pipeline

```
docSource.FetchIndex → m.resolveClient() → Client.Stream(goroutine)
  → readStreamChunk → handleStreamDelta (accumulate, show live preview)
  → handleStreamDone (parse READ_FILE/ANSWER)
  → startFileFetches (tea.Batch)
  → handleDocContent → remainingFetching
  → Client.Stream (next round) OR queryCompleteMsg (done)
```

## Retry config

- **LLM stream/send**: 3 attempts, 2s base, 30s max, exponential + 30% jitter
- **Doc fetches**: 3 attempts, 1s base, 10s max
- **Retryable**: HTTP 429, 5xx, network errors, stream read errors
- **Non-retryable**: HTTP 401, 403, 400, 404
- **UI feedback**: `llm.StreamDelta{Retrying: true}` shows `↻ (retrying in 4.2s — attempt 2/3)`

## CLI commands

```bash
# Query docs
qdoc go "query"              # query Go docs (go.dev/doc)
qdoc python "query"          # query Python docs (docs.python.org/3)
qdoc fastapi "query"         # query FastAPI docs (fastapi.tiangolo.com)
qdoc react "query"           # query React docs (react.dev)
qdoc nextjs "query"          # query Next.js docs (nextjs.org/docs)
qdoc ./dir "query"           # query local directory of docs

# Headless / agent modes
qdoc --no-tui go "query"     # markdown to stdout, no TUI
qdoc --json go "query"       # JSON to stdout, with metadata

# Configuration
qdoc provider                # interactive provider picker (TUI)
qdoc model                   # interactive model picker (TUI)
qdoc set key <p> [key]       # set API key (prompts if key omitted, no terminal echo)
qdoc set provider <name>     # set default provider
qdoc set model <p> <model>   # set model for a provider
qdoc set access <p> <m>      # set access method (e.g. "openai subscription")

# Inspection
qdoc status                  # show current config
qdoc providers               # list LLM providers + key/model status
qdoc sources                 # list documentation sources

# First run
qdoc                         # interactive setup (no key configured); otherwise prints usage
```

## Documentation sources

Built-in sources defined in `internal/docsource/source.go`. Run `qdoc sources` for the live list. Supports `./path` for local directories recursively parsing `.md`, `.mdx`, `.html`, `.rst`, `.txt`, `.adoc`.

## Providers

Defined in `internal/provider/providers.json`. All use OpenAI-compatible `/v1/chat/completions` API by default, except providers with custom access methods (e.g. `openai` subscription using Codex Responses API). Custom base URL via `QDOC_BASE_URL` env var.

### Access methods

Providers can support multiple access methods (e.g. API key vs OAuth subscription). The `AccessMethod` struct in `internal/provider/provider.go` defines: `ID`, `Name`, `BaseURL`, `APIType`, `AuthType`, `EnvKey`, `OAuthConfig`, `Headers`, `Description`.

- `provider.KeyExistsForMethod(cfg, provName, methodID)` checks if credentials are available for a specific method
- `provider.ResolveClientWithMethod(cfg, methodID)` resolves with a specific access method
- `cfg.AccessMethod` stores the user's preferred access method per provider

## Config file

`~/.config/qdoc/config.json`:

```json
{
  "provider": "openai",
  "access_method": "subscription",
  "keys": {"openai": "sk-..."},
  "models": {"openai": "gpt-5.5"}
}
```

## Install script

`docs/public/install.sh` — POSIX bootstrap re-execs with bash if needed. Installs binary to `~/.qdoc/bin`, adds to shell config. Supports Linux, macOS, and Windows (WSL, Git Bash, MSYS2, Cygwin). Options: `--version`, `--no-modify-path`.

## npm package

Package name: `qdoc-agent` (defined in `package.json`). `bin/qdoc.js` proxies to the platform binary. `install.js` downloads the binary during `postinstall`.

## Release workflow

`.github/workflows/release.yml` — triggered by git tag `v*`. Builds Go binaries for all platforms, attaches to GitHub Release, publishes to npm (`qdoc-agent`).

## Docs site deployment

`.github/workflows/docs.yml` — builds VitePress docs and verifies artifacts on push and PRs. Paired with Cloudflare Pages project `qdoc` (configured via `wrangler.jsonc`, `_headers`, and `_redirects`).
