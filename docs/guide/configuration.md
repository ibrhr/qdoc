# Configuration

qdoc stores configuration in `~/.config/qdoc/config.json`. Create it manually or use the interactive commands — both work.

## Config File

```json
{
  "provider": "openai",
  "keys": {
    "openai": "sk-..."
  },
  "models": {
    "openai": "gpt-5.5"
  }
}
```

All fields are optional. qdoc resolves settings in this order:

```
QDOC_* env vars  →  config.json  →  built-in defaults
```

## Providers

Four providers are built in. Each uses the OpenAI-compatible API format with its own endpoint, default model, and environment variable.

| Provider | Default Model | API URL | Env Var |
|----------|--------------|---------|---------|
| `openai` | `gpt-5.5` | `api.openai.com/v1` | `OPENAI_API_KEY` |
| `deepseek` | `deepseek-v4-flash` | `api.deepseek.com/v1` | `DEEPSEEK_API_KEY` |
| `opencode-zen` | `gpt-5.4-mini` | `opencode.ai/zen/v1` | `OPENCODE_ZEN_API_KEY` |
| `opencode-go` | `deepseek-v4-flash` | `opencode.ai/zen/go/v1` | `OPENCODE_GO_API_KEY` |

**OpenCode Zen** includes GPT, Claude, Gemini, Qwen, MiniMax, GLM, Kimi, and DeepSeek models — a curated selection tested for coding. **OpenCode Go** is a low-cost subscription tier focused on open models.

## Setting a Provider

Interactive (recommended):

```bash
qdoc provider
```

Launches a TUI picker. Select with enter, it's saved immediately.

Manual:

```bash
qdoc set provider openai
```

## Setting an API Key

Keys are stored in `~/.config/qdoc/config.json`. Interactive mode hides your input:

```bash
qdoc set key openai
# Enter API key: ████████ (input hidden — no terminal echo)
```

Inline (use with caution — may appear in shell history):

```bash
qdoc set key openai sk-abc123yourkey
```

### Using Environment Variables

For CI pipelines or containerized environments, skip the config file entirely:

```bash
export QDOC_PROVIDER=openai
export QDOC_MODEL=gpt-5.4-mini
export OPENAI_API_KEY=sk-...
qdoc --no-tui go "channels tutorial"
```

Provider-specific env vars:
- `OPENAI_API_KEY`
- `DEEPSEEK_API_KEY`
- `OPENCODE_ZEN_API_KEY`
- `OPENCODE_GO_API_KEY`

## Choosing a Model

```bash
qdoc model
```

Interactive picker showing all models for all configured providers. The first item is always "use default" — select it to use the provider's built-in default.

To set a model for a specific provider:

```bash
qdoc set model openai gpt-5.4-mini
```

To override globally for the session:

```bash
export QDOC_MODEL=gpt-5.4-mini
```

### Recommended Models for Agents

For agent use, prioritize speed and cost over maximum capability. Doc research doesn't require frontier models:

| Use Case | Recommended Model |
|----------|------------------|
| Fast, cheap research | `gpt-5.4-mini` (OpenAI) or `deepseek-v4-flash` |
| Complex API questions | `gpt-5.5` (OpenAI) or `gpt-5.4` (OpenCode Zen) |
| Deep code analysis | `deepseek-v4-pro` or `claude-opus-4-7` (OpenCode Zen) |

## Checking Configuration

```bash
qdoc status
```

Prints the current provider, configured keys (masked), model assignments, and config file path.

## Custom API Base

Override the API endpoint for any provider:

```bash
export QDOC_BASE_URL=https://my-proxy.example.com/v1
```

Use this for proxies, self-hosted models (Ollama, vLLM), or any OpenAI-compatible API.

## Multiple Providers

You can set keys for multiple providers simultaneously. qdoc uses the one set as the default (`qdoc set provider <name>`) but you can switch at any time — the keys are preserved in config.
