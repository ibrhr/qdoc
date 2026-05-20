package provider

import (
	"fmt"
	"os"

	"github.com/ibrhr/qdoc/internal/config"
	"github.com/ibrhr/qdoc/internal/llm"
)

type Provider struct {
	Name         string
	BaseURL      string
	EnvKey       string
	DefaultModel string
	Description  string
	Models       []string
}

var Providers = []Provider{
	{
		Name:         "openai",
		BaseURL:      "https://api.openai.com/v1",
		EnvKey:       "OPENAI_API_KEY",
		DefaultModel: "gpt-4.1",
		Description:  "OpenAI API (gpt-4.1, gpt-4o, etc.)",
		Models:       []string{"gpt-4.1", "gpt-4.1-mini", "gpt-4o", "gpt-4o-mini", "gpt-4.1-nano", "gpt-5", "gpt-5-nano"},
	},
	{
		Name:         "openrouter",
		BaseURL:      "https://openrouter.ai/api/v1",
		EnvKey:       "OPENROUTER_API_KEY",
		DefaultModel: "openai/gpt-4.1",
		Description:  "OpenRouter - unified API for 300+ models",
		Models:       []string{"openai/gpt-4.1", "anthropic/claude-sonnet-4", "anthropic/claude-opus-4", "google/gemini-2.5-pro", "meta-llama/llama-4-maverick", "deepseek/deepseek-chat", "qwen/qwen3-coder"},
	},
	{
		Name:         "deepseek",
		BaseURL:      "https://api.deepseek.com/v1",
		EnvKey:       "DEEPSEEK_API_KEY",
		DefaultModel: "deepseek-chat",
		Description:  "DeepSeek API (deepseek-chat, deepseek-reasoner)",
		Models:       []string{"deepseek-chat", "deepseek-reasoner"},
	},
	{
		Name:         "opencode-zen",
		BaseURL:      "https://opencode.ai/zen/v1",
		EnvKey:       "OPENCODE_ZEN_API_KEY",
		DefaultModel: "gpt-5.4-mini",
		Description:  "OpenCode Zen - curated models tested for coding",
		Models: []string{
			"gpt-5.5", "gpt-5.5-pro",
			"gpt-5.4", "gpt-5.4-pro", "gpt-5.4-mini", "gpt-5.4-nano",
			"gpt-5.3-codex", "gpt-5.3-codex-spark",
			"gpt-5.2", "gpt-5.2-codex",
			"gpt-5.1", "gpt-5.1-codex", "gpt-5.1-codex-max", "gpt-5.1-codex-mini",
			"gpt-5", "gpt-5-codex", "gpt-5-nano",
			"claude-opus-4-7", "claude-opus-4-6", "claude-opus-4-5", "claude-opus-4-1",
			"claude-sonnet-4-6", "claude-sonnet-4-5", "claude-sonnet-4",
			"claude-haiku-4-5", "claude-3-5-haiku",
			"gemini-3.1-pro", "gemini-3-flash",
			"qwen3.6-plus", "qwen3.5-plus",
			"minimax-m2.7", "minimax-m2.5", "minimax-m2.5-free",
			"glm-5.1", "glm-5",
			"kimi-k2.6", "kimi-k2.5",
			"deepseek-v4-flash-free",
			"nemotron-3-super-free",
			"big-pickle",
		},
	},
	{
		Name:         "opencode-go",
		BaseURL:      "https://opencode.ai/zen/go/v1",
		EnvKey:       "OPENCODE_GO_API_KEY",
		DefaultModel: "deepseek-v4-flash-free",
		Description:  "OpenCode Go - low cost subscription for open coding models",
		Models: []string{
			"qwen3.6-plus", "qwen3.5-plus",
			"minimax-m2.7", "minimax-m2.5", "minimax-m2.5-free",
			"glm-5.1", "glm-5",
			"kimi-k2.6", "kimi-k2.5",
			"deepseek-v4-flash-free",
			"nemotron-3-super-free",
			"big-pickle",
		},
	},
}

func Find(name string) (Provider, bool) {
	for _, p := range Providers {
		if p.Name == name {
			return p, true
		}
	}
	return Provider{}, false
}

func BuildModelList(prov Provider) []string {
	items := make([]string, 0, len(prov.Models)+1)
	items = append(items, fmt.Sprintf("— use default (%s) —", prov.DefaultModel))
	items = append(items, prov.Models...)
	return items
}

func KeyExists(cfg config.Config, name string) bool {
	return keyExistsForProvider(cfg, name)
}

func AnyKeyConfigured(cfg config.Config) bool {
	for _, p := range Providers {
		if keyExistsForProvider(cfg, p.Name) {
			return true
		}
	}
	return false
}

func keyExistsForProvider(cfg config.Config, name string) bool {
	if cfg.Keys[name] != "" {
		return true
	}
	for _, p := range Providers {
		if p.Name == name {
			if os.Getenv(p.EnvKey) != "" {
				return true
			}
			break
		}
	}
	return false
}

func ResolveClient(cfg config.Config) (*llm.Client, error) {
	providerName := os.Getenv("QDOC_PROVIDER")
	if providerName == "" {
		providerName = cfg.Provider
	}

	prov, found := Find(providerName)
	if !found {
		return nil, ErrNoProvider{Name: providerName}
	}

	apiKey := os.Getenv(prov.EnvKey)
	if apiKey == "" {
		apiKey = cfg.Keys[prov.Name]
	}

	baseURL := os.Getenv("QDOC_BASE_URL")
	if baseURL == "" {
		baseURL = prov.BaseURL
	}

	model := os.Getenv("QDOC_MODEL")
	if model == "" {
		if m, ok := cfg.Models[prov.Name]; ok && m != "" {
			model = m
		} else {
			model = prov.DefaultModel
		}
	}

	if apiKey == "" {
		return nil, ErrNoKey{Provider: prov.Name, EnvKey: prov.EnvKey}
	}

	return &llm.Client{
		APIKey:   apiKey,
		BaseURL:  baseURL,
		Model:    model,
		Provider: prov.Name,
	}, nil
}

type ErrNoKey struct {
	Provider string
	EnvKey   string
}

func (e ErrNoKey) Error() string {
	return "no API key for " + e.Provider + ". set " + e.EnvKey + " or run: qdoc set key " + e.Provider
}

type ErrNoProvider struct {
	Name string
}

func (e ErrNoProvider) Error() string {
	return "unknown provider: " + e.Name
}