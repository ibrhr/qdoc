package provider

import (
	"encoding/json"
	"fmt"
	"os"
	_ "embed"

	"github.com/ibrhr/qdoc/internal/config"
	"github.com/ibrhr/qdoc/internal/llm"
)

type Provider struct {
	Name         string   `json:"name"`
	BaseURL      string   `json:"base_url"`
	APIType      string   `json:"api_type"`
	EnvKey       string   `json:"env_key"`
	DefaultModel string   `json:"default_model"`
	Description  string   `json:"description"`
	Models       []string `json:"models"`

	Headers map[string]string `json:"headers,omitempty"`
}

//go:embed providers.json
var providersJSON []byte

var Providers []Provider

func init() {
	if err := json.Unmarshal(providersJSON, &Providers); err != nil {
		panic("qdoc: failed to parse embedded providers.json: " + err.Error())
	}
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

func ResolveClient(cfg config.Config) (llm.Client, error) {
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

	apiType := prov.APIType
	if apiType == "" {
		apiType = "openai-compat"
	}

	return llm.NewClient(apiType, llm.Config{
		APIKey:   apiKey,
		BaseURL:  baseURL,
		Model:    model,
		Provider: prov.Name,
		Headers:  prov.Headers,
	})
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
