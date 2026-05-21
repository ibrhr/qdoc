# Changelog

## v0.1.6

- Five new providers: xAI (Grok), Alibaba DashScope (Qwen), Google (Gemini), Zhipu AI (GLM), Moonshot (Kimi)
- Providers extracted from hardcoded Go to embedded `providers.json` — add a provider by editing JSON
- `Provider` struct expanded with `api_type`, `headers` — ready for non-OpenAI backends
- `llm.Client` interface + `NewClient` factory — pluggable API backends
- `Provider.Headers` (e.g. `anthropic-version`) flows through to HTTP requests
- Dynamic provider lists in CLI — error messages auto-derive from `providers.json`

## v0.1.5

- Fix npm publish workflow (`--allow-same-version`)

## v0.1.4

- Harden installers: checksum verification, Windows zip fix, smoke test, `go mod verify`
- Enhance versioning and releasing logic

## v0.1.3

- Complete docs rewrite — agent-first positioning
- Fix installation script

## v0.1.2

- Rewrite `install.sh`: install to `~/.qdoc/bin`, auto-PATH, version check, progress bar
- Fix binary naming in tarballs
- Add install script served at `qdoc.ibrhr.dev`

## v0.1.1

- Initial public release
- Publish binaries and npm package (`qdoc-agent`)
- Four providers: OpenAI, DeepSeek, OpenCode Zen, OpenCode Go
- Six built-in doc sources: Go, Python, Next.js, React, FastAPI, Pydantic
- Local directory support for custom docs
- TUI with live streaming, headless `--no-tui` and `--json` modes
- Built-in retry with exponential backoff

---

See [GitHub Releases](https://github.com/ibrhr/qdoc/releases) for full release artifacts.
