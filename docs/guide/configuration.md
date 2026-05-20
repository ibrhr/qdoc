# Configuration

qdoc stores configuration in `~/.config/qdoc/config.json`.

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

## Providers

qdoc supports multiple LLM providers. Each has its own API key and default model:

| Provider | Default Model | Env Var |
|---|---|---|
| `openai` | `gpt-5.5` | `OPENAI_API_KEY` |
| `deepseek` | `deepseek-v4-flash` | `DEEPSEEK_API_KEY` |
| `opencode-zen` | `gpt-5.4-mini` | `OPENCODE_ZEN_API_KEY` |
| `opencode-go` | `deepseek-v4-flash-free` | `OPENCODE_GO_API_KEY` |

## Setting a Provider

Interactive (recommended):

```bash
qdoc provider
```

Manual:

```bash
qdoc set provider openai
```

## Setting an API Key

Interactive:

```bash
qdoc set key openai
# Prompts for key (input hidden)
```

Inline:

```bash
qdoc set key openai sk-abc123yourkey
```

### Using Environment Variables

You can skip config and use environment variables instead:

```bash
export QDOC_PROVIDER=openai
export QDOC_MODEL=gpt-5.5
export OPENAI_API_KEY=sk-...
qdoc go "channels tutorial"
```

Priority order: `QDOC_*` env vars → config file → defaults.

## Choosing a Model

```bash
qdoc model
```

This launches an interactive model picker for all configured providers. The default model is selected automatically.

To override for a specific provider:

```bash
qdoc set model openai gpt-5.5
```

## Checking Configuration

```bash
qdoc status
```
