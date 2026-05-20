# Providers & Models

## Available Providers

### OpenAI

```bash
qdoc set key openai sk-your-key
```

| Field | Value |
|---|---|
| Default Model | `gpt-4.1` |
| API URL | `https://api.openai.com/v1` |
| Env Var | `OPENAI_API_KEY` |

### OpenRouter

```bash
qdoc set key openrouter sk-or-v1-your-key
```

| Field | Value |
|---|---|
| Default Model | `openai/gpt-4.1` |
| API URL | `https://openrouter.ai/api/v1` |
| Env Var | `OPENROUTER_API_KEY` |

OpenRouter gives access to hundreds of models through a single API.

### DeepSeek

```bash
qdoc set key deepseek sk-your-key
```

| Field | Value |
|---|---|
| Default Model | `deepseek-chat` |
| API URL | `https://api.deepseek.com/v1` |
| Env Var | `DEEPSEEK_API_KEY` |

### Opencode Zen

```bash
qdoc set key opencode-zen sk-your-key
```

| Field | Value |
|---|---|
| Default Model | `gpt-5.4-mini` |
| API URL | `https://zen.opencode.ai/v1` |
| Env Var | `OPENCODE_ZEN_API_KEY` |

### Opencode Go

```bash
qdoc set key opencode-go sk-your-key
```

| Field | Value |
|---|---|
| Default Model | `deepseek-v4-flash-free` |
| API URL | `https://go.opencode.ai/v1` |
| Env Var | `OPENCODE_GO_API_KEY` |

## Model Selection

Use the interactive picker:

```bash
qdoc model
```

This shows all available models for all configured providers. Select one and it's saved to your config.

To list current model assignments:

```bash
qdoc providers
```

## Adding New Providers

qdoc supports any OpenAI-compatible API. To add a new provider, you'd need to modify `internal/provider/provider.go` and rebuild. See the [GitHub repo](https://github.com/ibrhr/qdoc) for details.
