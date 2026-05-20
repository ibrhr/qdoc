# AGENTS.md — qdoc

## Build & verify

```bash
export PATH=$PATH:/usr/local/go/bin  # Go 1.26.3 lives here
go build ./...   # builds all packages
go vet ./...
```

## Architecture

Module path: `github.com/ibrhr/qdoc`

```
main.go                          # CLI routing + print commands
internal/
  config/
    config.go                    # Config struct, load, save, configPath
  provider/
    provider.go                  # Provider struct, registry, Find, ResolveClient, error types
  llm/
    client.go                    # Client struct, Send(), Stream() methods, retry logic
    prompt.go                    # BuildSystemPrompt()
    parse.go                     # ParsedAction, ParseResponses()
    types.go                     # ChatMessage, StreamDelta
  docsource/
    source.go                    # Source, Entry, KnownSources, Find, fetch methods, retryableHTTPGet
    html.go                      # extractLinks, extractMainContent (unexported HTML helpers)
  tui/
    model.go                     # Model struct, phase/state types, constructors (New*), Init
    update.go                    # Update + query-mode handlers + streaming pipeline
    setup.go                     # Setup update handlers + setup view renderers
    view.go                      # View + query-mode rendering (header, progress, content, footer)
    styles.go                    # Theme colors + lipgloss styles + spinner
    render.go                    # renderMarkdown, renderInlineMarkdown, wordWrap
    messages.go                  # All tea.Msg types (docIndexMsg, StreamDelta, etc.)
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
  └── tui ←─┘ (imports config, provider, llm, docsource)
       ↑
       │
     main (imports everything)
```

### Exported types mapping

| Old (package main) | New package | Exported as |
|---|---|---|
| `config` struct | `config` | `config.Config` |
| `provider` struct | `provider` | `provider.Provider` |
| `providers` var | `provider` | `provider.Providers` |
| `docSource` struct | `docsource` | `docsource.Source` |
| `docEntry` struct | `docsource` | `docsource.Entry` |
| `llmClient` struct | `llm` | `llm.Client` |
| `chatMessage` struct | `llm` | `llm.ChatMessage` |
| `parsedAction` struct | `llm` | `llm.ParsedAction` |
| `streamDeltaMsg` struct | `llm` | `llm.StreamDelta` |
| `model` struct | `tui` | `tui.Model` |

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

7. **StreamDelta is now llm.StreamDelta**: Previously `streamDeltaMsg` was defined in `messages.go` as `struct{content string; done bool; err error; retrying bool}`. Now it's `llm.StreamDelta` with exported fields: `Content`, `Done`, `Err`, `Retrying`.

8. **renderMarkdown blank line bug**: The loop already emits `\n` between every line (`if i > 0 { sb.WriteString("\n") }`). Do NOT add another `\n` for blank lines — it doubles them. Just `continue`.

9. **Word-wrap continuation indent**: After `lipgloss.Wrap(line, maxW, " ")`, continuation lines lose the `"  "` indentation prefix. You must prepend `"  "` to all parts past the first: `parts[0]` as-is, then `"  "+wl` for each `parts[1:]`.

10. **Don't use Faint(true) on thoughtStyle**: On dark terminal backgrounds, `Faint(true)` makes text effectively invisible. Use a dim foreground color with Italic, not Faint.

11. **System prompt design for LLM actions**:
    - Avoid capitalized warnings (`ONLY valid JSON`) — triggers format paranoia
    - Don't say `"exactly one { at start and one } at end"` — models worry about parser strictness
    - Don't give contradictory number ranges (`2-3` vs `up to 3` vs `a JSON action` singular)
    - Don't use prescriptive arrow rules (`"How do I X?" → /doc/foo`) — turns models into rule-matchers instead of readers
    - Keep it simple: here are the sections, here are the files, here's the query

## Streaming pipeline

```
docSource.FetchIndex → provider.ResolveClient → Client.Stream(goroutine)
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

```
qdoc go "query"        — query go docs (only source with system prompt)
qdoc ./dir "query"     — query local directory of docs
qdoc sources           — list documentation sources
qdoc providers         — list LLM providers + key/model status
qdoc status            — show current config
qdoc provider          — interactive provider picker (TUI)
qdoc model             — interactive model picker (TUI)
qdoc set key <p> [key] — set API key (prompts if key omitted, no terminal echo)
qdoc                   — first-run: interactive setup (no key configured); otherwise prints usage
```

## Config file

`~/.config/qdoc/config.json`:
```json
{
  "provider": "openai",
  "keys": {"openai": "sk-..."},
  "models": {"openai": "gpt-4.1"}
}
```
Emptied keys/models maps are re-initialized to non-nil on load.