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

## Providers & Models

Four providers are built in. See [Providers & Models](/reference/providers) for the full list of providers, models, and API endpoints.

### Setting a Provider

Interactive (recommended):

```bash
qdoc provider
```

Launches a TUI picker. Select with enter, it's saved immediately.

Manual:

```bash
qdoc set provider openai
```

### Choosing a Model

```bash
qdoc model
```

Interactive picker showing all models for all configured providers. The first item is always "use default" — select it to use the provider's built-in default. Any other selection is saved as an override in `config.json`.

To set a model for a specific provider:

```bash
qdoc set model openai gpt-5.4-mini
```

To override globally for the session:

```bash
export QDOC_MODEL=gpt-5.4-mini
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

If you prefer to skip the config file entirely, you can use environment variables:

```bash
export QDOC_PROVIDER=openai
export QDOC_MODEL=gpt-5.4-mini
export OPENAI_API_KEY=sk-...
qdoc --no-tui go "channels tutorial"
```

Provider-specific env vars are documented in [Providers & Models](/reference/providers).

## Custom API Base

Override the API endpoint for any provider:

```bash
export QDOC_BASE_URL=https://my-proxy.example.com/v1
```

Use this for proxies, self-hosted models (Ollama, vLLM), or any OpenAI-compatible API.

## Checking Configuration

```bash
qdoc status
```

Prints the current provider, configured keys (masked), model assignments, and config file path.

## Multiple Providers

You can set keys for multiple providers simultaneously. qdoc uses the one set as the default (`qdoc set provider <name>`) but you can switch at any time — the keys are preserved in config.
