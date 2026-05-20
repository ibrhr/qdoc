# Providers & Models

qdoc uses the OpenAI-compatible chat completions API. Any endpoint that supports `/v1/chat/completions` with a `Bearer` token will work.

## Built-in Providers

### OpenAI

```bash
qdoc set key openai sk-your-key
```

| Field | Value |
|-------|-------|
| Default Model | `gpt-5.5` |
| API URL | `https://api.openai.com/v1` |
| Env Var | `OPENAI_API_KEY` |
| Models | `gpt-5.5`, `gpt-5.5-pro`, `gpt-5.4`, `gpt-5.4-pro`, `gpt-5.4-mini`, `gpt-5.4-nano` |

### DeepSeek

```bash
qdoc set key deepseek sk-your-key
```

| Field | Value |
|-------|-------|
| Default Model | `deepseek-v4-flash` |
| API URL | `https://api.deepseek.com/v1` |
| Env Var | `DEEPSEEK_API_KEY` |
| Models | `deepseek-v4-flash`, `deepseek-v4-pro` |

### OpenCode Zen

Curated models tested for coding. Includes frontier and open models.

```bash
qdoc set key opencode-zen sk-your-key
```

| Field | Value |
|-------|-------|
| Default Model | `gpt-5.4-mini` |
| API URL | `https://opencode.ai/zen/v1` |
| Env Var | `OPENCODE_ZEN_API_KEY` |

**Available models:**

| Category | Models |
|----------|--------|
| GPT | `gpt-5.5`, `gpt-5.5-pro`, `gpt-5.4`, `gpt-5.4-pro`, `gpt-5.4-mini`, `gpt-5.4-nano` |
| Claude | `claude-opus-4-7`, `claude-opus-4-6`, `claude-opus-4-5`, `claude-opus-4-1`, `claude-sonnet-4-6`, `claude-sonnet-4-5`, `claude-haiku-4-5` |
| Gemini | `gemini-3.1-pro`, `gemini-3-flash` |
| Qwen | `qwen3.6-plus`, `qwen3.5-plus` |
| MiniMax | `minimax-m2.7`, `minimax-m2.5`, `minimax-m2.5-free` |
| GLM | `glm-5.1` |
| Kimi | `kimi-k2.6`, `kimi-k2.5` |
| DeepSeek | `deepseek-v4-flash-free` |
| Other | `nemotron-3-super-free`, `big-pickle` |

### OpenCode Go

Low-cost subscription tier focused on open coding models.

```bash
qdoc set key opencode-go sk-your-key
```

| Field | Value |
|-------|-------|
| Default Model | `deepseek-v4-flash` |
| API URL | `https://opencode.ai/zen/go/v1` |
| Env Var | `OPENCODE_GO_API_KEY` |

**Available models:**

- `qwen3.6-plus`, `qwen3.5-plus`
- `minimax-m2.7`, `minimax-m2.5`
- `glm-5.1`
- `kimi-k2.6`, `kimi-k2.5`
- `deepseek-v4-flash`, `deepseek-v4-pro`
- `mimo-v2.5`, `mimo-v2.5-pro`

## Model Selection

### Interactive picker

```bash
qdoc model
```

Shows all available models across all configured providers. The first entry is "use default" — selecting it uses the provider's built-in default model. Any other selection is saved as an override in `config.json`.

### Per-provider override

```bash
qdoc set model openai gpt-5.4-mini
```

### Per-session override

```bash
export QDOC_MODEL=gpt-5.4-mini
```

### List current assignments

```bash
qdoc providers
```

Shows each provider's name, configured model, and whether an API key is set.

## Custom Providers

Any OpenAI-compatible API works. Set the base URL via environment variable:

```bash
export QDOC_BASE_URL=https://api.mistral.ai/v1
export QDOC_MODEL=mistral-large-latest
export QDOC_PROVIDER=custom
```

For persistent custom providers, modify `internal/provider/provider.go` and rebuild. See the [GitHub repo](https://github.com/ibrhr/qdoc) for contribution guidelines.
