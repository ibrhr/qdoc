package provider

import (
	"encoding/json"
	"fmt"
	"os"
	_ "embed"

	"github.com/ibrhr/qdoc/internal/auth"
	"github.com/ibrhr/qdoc/internal/config"
	"github.com/ibrhr/qdoc/internal/llm"
)

type AccessMethod struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	BaseURL     string            `json:"base_url"`
	APIType     string            `json:"api_type"`
	AuthType    string            `json:"auth_type"`
	EnvKey      string            `json:"env_key,omitempty"`
	OAuthConfig *auth.OAuthConfig `json:"oauth_config,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Description string            `json:"description,omitempty"`
}

func (am AccessMethod) EffectiveAuthType() string {
	if am.AuthType != "" {
		return am.AuthType
	}
	return "api_key"
}

func (am AccessMethod) AuthInfo(provName string) auth.ProviderInfo {
	return auth.ProviderInfo{
		Name:        provName,
		AuthType:    am.EffectiveAuthType(),
		OAuthConfig: am.OAuthConfig,
		EnvKey:      am.EnvKey,
	}
}

type Provider struct {
	Name         string   `json:"name"`
	DefaultModel string   `json:"default_model"`
	Description  string   `json:"description"`
	Models       []string `json:"models"`

	AccessMethods []AccessMethod `json:"access_methods,omitempty"`

	BaseURL      string            `json:"base_url,omitempty"`
	APIType      string            `json:"api_type,omitempty"`
	AuthType     string            `json:"auth_type,omitempty"`
	AuthTypes    []string          `json:"auth_types,omitempty"`
	EnvKey       string            `json:"env_key,omitempty"`
	Headers      map[string]string `json:"headers,omitempty"`
	OAuthConfig  *auth.OAuthConfig `json:"oauth_config,omitempty"`
}

func (p Provider) EffectiveAuthType() string {
	if len(p.AccessMethods) > 0 {
		return p.AccessMethods[0].EffectiveAuthType()
	}
	if p.AuthType != "" {
		return p.AuthType
	}
	return "api_key"
}

func (p Provider) HasMultipleAuthOptions() bool {
	return len(p.AccessMethods) > 1
}

func (p Provider) HasAccessMethods() bool {
	return len(p.AccessMethods) > 0
}

func (p Provider) GetAccessMethod(id string) *AccessMethod {
	for i := range p.AccessMethods {
		if p.AccessMethods[i].ID == id {
			return &p.AccessMethods[i]
		}
	}
	return nil
}

func (p Provider) DefaultAccessMethod() *AccessMethod {
	if len(p.AccessMethods) > 0 {
		return &p.AccessMethods[0]
	}
	if p.BaseURL != "" {
		at := p.AuthType
		if at == "" {
			at = "api_key"
		}
		return &AccessMethod{
			ID:          "default",
			Name:        p.Name,
			BaseURL:     p.BaseURL,
			APIType:     p.APIType,
			AuthType:    at,
			EnvKey:      p.EnvKey,
			OAuthConfig: p.OAuthConfig,
			Headers:     p.Headers,
		}
	}
	return nil
}

func (p Provider) AccessMethodAuthInfo(method *AccessMethod) auth.ProviderInfo {
	if method != nil {
		return method.AuthInfo(p.Name)
	}
	return p.AuthInfo()
}

func (p Provider) AuthInfo() auth.ProviderInfo {
	if len(p.AccessMethods) > 0 {
		return p.AccessMethods[0].AuthInfo(p.Name)
	}
	if p.BaseURL != "" {
		am := p.DefaultAccessMethod()
		return am.AuthInfo(p.Name)
	}
	at := p.AuthType
	if at == "" {
		at = "api_key"
	}
	return auth.ProviderInfo{
		Name:        p.Name,
		AuthType:    at,
		OAuthConfig: p.OAuthConfig,
		EnvKey:      p.EnvKey,
	}
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
	return keyExistsForProvider(cfg, name, "")
}

func KeyExistsForMethod(cfg config.Config, provName, methodID string) bool {
	return keyExistsForProvider(cfg, provName, methodID)
}

func AnyKeyConfigured(cfg config.Config) bool {
	for _, p := range Providers {
		if keyExistsForProvider(cfg, p.Name, "") {
			return true
		}
	}
	return false
}

func tokenStoreKey(provName, methodID string) string {
	if methodID != "" {
		return provName + ":" + methodID
	}
	return provName
}

func keyExistsForProvider(cfg config.Config, name, methodID string) bool {
	if cfg.Keys[name] != "" {
		return true
	}
	if store, err := auth.LoadTokenStore(configDirPath()); err == nil {
		if tok, ok := store.Get(tokenStoreKey(name, methodID)); ok && !tok.IsZero() {
			return true
		}
		if methodID != "" {
			if tok, ok := store.Get(name); ok && !tok.IsZero() {
				return true
			}
		} else {
			for _, p := range Providers {
				if p.Name == name {
					for _, am := range p.AccessMethods {
						if tok, ok := store.Get(name + ":" + am.ID); ok && !tok.IsZero() {
							return true
						}
					}
					break
				}
			}
		}
	}
	for _, p := range Providers {
		if p.Name == name {
			if os.Getenv(p.EnvKey) != "" {
				return true
			}
			for _, am := range p.AccessMethods {
				if os.Getenv(am.EnvKey) != "" {
					return true
				}
			}
			break
		}
	}
	return false
}

func ResolveClient(cfg config.Config) (llm.Client, error) {
	return ResolveClientWithMethod(cfg, "")
}

func ResolveClientWithMethod(cfg config.Config, accessMethodID string) (llm.Client, error) {
	providerName := os.Getenv("QDOC_PROVIDER")
	if providerName == "" {
		providerName = cfg.Provider
	}

	prov, found := Find(providerName)
	if !found {
		return nil, ErrNoProvider{Name: providerName}
	}

	method := prov.DefaultAccessMethod()
	if accessMethodID != "" || cfg.AccessMethod != "" {
		lookupID := accessMethodID
		if lookupID == "" {
			lookupID = cfg.AccessMethod
		}
		if am := prov.GetAccessMethod(lookupID); am != nil {
			method = am
		}
	}
	if method == nil {
		return nil, fmt.Errorf("no access method available for %s", prov.Name)
	}

	info := method.AuthInfo(prov.Name)

	baseURL := os.Getenv("QDOC_BASE_URL")
	if baseURL == "" {
		baseURL = method.BaseURL
	}

	model := os.Getenv("QDOC_MODEL")
	if model == "" {
		if m, ok := cfg.Models[prov.Name]; ok && m != "" {
			model = m
		} else {
			model = prov.DefaultModel
		}
	}

	apiType := method.APIType
	if apiType == "" {
		apiType = "openai-compat"
	}

	authMethod, err := auth.ForProvider(info)
	if err != nil {
		return nil, err
	}

	tsKey := tokenStoreKey(prov.Name, method.ID)

	switch info.AuthType {
	case "api_key":
		token, err := resolveAPIKeyToken(method.EnvKey, prov.Name, cfg)
		if err != nil {
			return nil, err
		}
		return llm.NewClient(apiType, llm.Config{
			Auth:     authMethod,
			Token:    token,
			BaseURL:  baseURL,
			Model:    model,
			Provider: prov.Name,
			Headers:  mergeHeaders(method.Headers),
		})

	default:
		store, err := auth.LoadTokenStore(configDirPath())
		if err != nil {
			return nil, fmt.Errorf("loading token store: %w", err)
		}
		token, ok := store.Get(tsKey)
		foundKey := tsKey
		if !ok {
			token, ok = store.Get(prov.Name)
			foundKey = prov.Name
		}
		if !ok {
			return nil, ErrNoKey{Provider: prov.Name, EnvKey: method.EnvKey}
		}
		if store.IsExpired(foundKey) {
			return nil, fmt.Errorf("token for %s is expired — re-authenticate", prov.Name)
		}
		return llm.NewClient(apiType, llm.Config{
			Auth:     authMethod,
			Token:    token,
			BaseURL:  baseURL,
			Model:    model,
			Provider: prov.Name,
			Headers:  mergeHeaders(method.Headers),
		})
	}
}

func configDirPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home + "/.config/qdoc"
}

func resolveAPIKeyToken(envKey, provName string, cfg config.Config) (auth.Token, error) {
	apiKey := os.Getenv(envKey)
	if apiKey == "" {
		apiKey = cfg.Keys[provName]
	}
	if apiKey == "" {
		return auth.Token{}, ErrNoKey{Provider: provName, EnvKey: envKey}
	}
	return auth.Token{AccessToken: apiKey, TokenType: "bearer"}, nil
}

func mergeHeaders(methodHeaders map[string]string) map[string]string {
	if methodHeaders == nil {
		return nil
	}
	result := make(map[string]string, len(methodHeaders))
	for k, v := range methodHeaders {
		result[k] = v
	}
	return result
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
