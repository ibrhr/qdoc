# Providers & Models

## Available Providers

### OpenAI

```bash
qdoc set key openai sk-your-key
```

| Field | Value |
|---|---|
| Default Model | `gpt-5.5` |
| API URL | `https://api.openai.com/v1` |
| Env Var | `OPENAI_API_KEY` |

### DeepSeek

```bash
qdoc set key deepseek sk-your-key
```

| Field | Value |
|---|---|
| Default Model | `deepseek-v4-flash` |
| API URL | `https://api.deepseek.com/v1` |
| Env Var | `DEEPSEEK_API_KEY` |

### Opencode Zen

```bash
qdoc set key opencode-zen sk-your-key
```

| Field | Value |
|---|---|
| Default Model | `gpt-5.4-mini` |
| API URL | `https://opencode.ai/zen/v1` |
| Env Var | `OPENCODE_ZEN_API_KEY` |

### Opencode Go

```bash
qdoc set key opencode-go sk-your-key
```

| Field | Value |
|---|---|
| Default Model | `deepseek-v4-flash-free` |
| API URL | `https://opencode.ai/zen/go/v1` |
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
