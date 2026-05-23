package provider

import (
	"testing"

	"github.com/ibrhr/qdoc/internal/config"
	"github.com/ibrhr/qdoc/internal/llm"
)

func TestFind_Exact(t *testing.T) {
	prov, found := Find("openai")
	if !found {
		t.Fatal("expected openai to be found")
	}
	if prov.Name != "openai" {
		t.Errorf("Name = %q, want %q", prov.Name, "openai")
	}
}

func TestFind_CaseSensitive(t *testing.T) {
	_, found := Find("OpenAI")
	if found {
		t.Error("Find should be case-sensitive (unlike docsource.Find)")
	}
}

func TestFind_NotExist(t *testing.T) {
	_, found := Find("nonexistent")
	if found {
		t.Error("Find should return false for unknown provider")
	}
}

func TestKeyExists_Configured(t *testing.T) {
	cfg := config.Config{
		Keys: map[string]string{"openai": "sk-test"},
	}
	if !KeyExists(cfg, "openai") {
		t.Error("KeyExists should return true when key is configured")
	}
}

func TestKeyExists_EnvVar(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-env-test")
	cfg := config.Config{
		Keys: map[string]string{},
	}
	if !KeyExists(cfg, "openai") {
		t.Error("KeyExists should return true when env var is set")
	}
}

func TestKeyExists_Neither(t *testing.T) {
	cfg := config.Config{
		Keys: map[string]string{},
	}
	if KeyExists(cfg, "openai") {
		t.Error("KeyExists should return false when no key is available")
	}
}

func TestAnyKeyConfigured(t *testing.T) {
	tests := []struct {
		name string
		keys map[string]string
		envs map[string]string
		want bool
	}{
		{"no keys", nil, nil, false},
		{"config key", map[string]string{"openai": "sk-test"}, nil, true},
		{"env key", nil, map[string]string{"OPENAI_API_KEY": "sk-env"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envs {
				t.Setenv(k, v)
			}
			cfg := config.Config{Keys: tt.keys}
			if got := AnyKeyConfigured(cfg); got != tt.want {
				t.Errorf("AnyKeyConfigured() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveClient_HappyPath(t *testing.T) {
	cfg := config.Config{
		Provider: "openai",
		Keys:     map[string]string{"openai": "sk-test"},
	}
	client, err := ResolveClient(cfg)
	if err != nil {
		t.Fatalf("ResolveClient() error: %v", err)
	}
	if client.ModelName() != "gpt-5.5" {
		t.Errorf("Model = %q, want %q", client.ModelName(), "gpt-5.5")
	}
	if client.ProviderName() != "openai" {
		t.Errorf("Provider = %q, want %q", client.ProviderName(), "openai")
	}
}

func TestResolveClient_NoKey(t *testing.T) {
	cfg := config.Config{
		Provider: "openai",
		Keys:     map[string]string{},
	}
	_, err := ResolveClient(cfg)
	if err == nil {
		t.Fatal("expected error when no API key")
	}
	noKey, ok := err.(ErrNoKey)
	if !ok {
		t.Fatalf("expected ErrNoKey, got %T: %v", err, err)
	}
	if noKey.Provider != "openai" {
		t.Errorf("ErrNoKey.Provider = %q", noKey.Provider)
	}
}

func TestResolveClient_NoProvider(t *testing.T) {
	cfg := config.Config{
		Provider: "nonexistent",
		Keys:     map[string]string{"nonexistent": "sk-test"},
	}
	_, err := ResolveClient(cfg)
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
	_, ok := err.(ErrNoProvider)
	if !ok {
		t.Fatalf("expected ErrNoProvider, got %T: %v", err, err)
	}
}

func TestResolveClient_EnvOverrideProvider(t *testing.T) {
	t.Setenv("QDOC_PROVIDER", "deepseek")
	t.Setenv("DEEPSEEK_API_KEY", "sk-deep")
	cfg := config.Config{
		Provider: "openai",
		Keys:     map[string]string{"openai": "sk-test"},
	}
	client, err := ResolveClient(cfg)
	if err != nil {
		t.Fatalf("ResolveClient() error: %v", err)
	}
	if client.ProviderName() != "deepseek" {
		t.Errorf("expected deepseek from env, got %q", client.ProviderName())
	}
}

func TestResolveClient_EnvOverrideModel(t *testing.T) {
	t.Setenv("QDOC_MODEL", "custom-model")
	cfg := config.Config{
		Provider: "openai",
		Keys:     map[string]string{"openai": "sk-test"},
	}
	client, err := ResolveClient(cfg)
	if err != nil {
		t.Fatalf("ResolveClient() error: %v", err)
	}
	if client.ModelName() != "custom-model" {
		t.Errorf("expected custom-model from env, got %q", client.ModelName())
	}
}

func TestResolveClient_EnvOverrideBaseURL(t *testing.T) {
	t.Setenv("QDOC_BASE_URL", "https://custom.api.com/v1")
	cfg := config.Config{
		Provider: "openai",
		Keys:     map[string]string{"openai": "sk-test"},
	}
	client, err := ResolveClient(cfg)
	if err != nil {
		t.Fatalf("ResolveClient() error: %v", err)
	}
	oc := client.(*llm.OpenAIClient)
	if oc.BaseURL != "https://custom.api.com/v1" {
		t.Errorf("BaseURL = %q, want %q", oc.BaseURL, "https://custom.api.com/v1")
	}
}

func TestResolveClient_ModelFromConfig(t *testing.T) {
	cfg := config.Config{
		Provider: "openai",
		Keys:     map[string]string{"openai": "sk-test"},
		Models:   map[string]string{"openai": "gpt-5.4-mini"},
	}
	client, err := ResolveClient(cfg)
	if err != nil {
		t.Fatalf("ResolveClient() error: %v", err)
	}
	if client.ModelName() != "gpt-5.4-mini" {
		t.Errorf("Model = %q, want %q", client.ModelName(), "gpt-5.4-mini")
	}
}

func TestResolveClient_KeyFromEnv(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-from-env")
	cfg := config.Config{
		Provider: "openai",
		Keys:     map[string]string{},
	}
	client, err := ResolveClient(cfg)
	if err != nil {
		t.Fatalf("ResolveClient() error: %v", err)
	}
	if client.ModelName() != "gpt-5.5" {
		t.Errorf("Model = %q, want %q", client.ModelName(), "gpt-5.5")
	}
}

func TestBuildModelList(t *testing.T) {
	prov := Provider{
		Name:         "openai",
		DefaultModel: "gpt-5.5",
		Models:       []string{"gpt-5.5", "gpt-5.5-pro"},
	}
	list := BuildModelList(prov)
	if len(list) != 3 {
		t.Fatalf("expected 3 items, got %d", len(list))
	}
	if list[0] != "— use default (gpt-5.5) —" {
		t.Errorf("first item = %q", list[0])
	}
	if list[1] != "gpt-5.5" {
		t.Errorf("second item = %q", list[1])
	}
	if list[2] != "gpt-5.5-pro" {
		t.Errorf("third item = %q", list[2])
	}
}

func TestErrNoKey_Error(t *testing.T) {
	e := ErrNoKey{Provider: "openai", EnvKey: "OPENAI_API_KEY"}
	msg := e.Error()
	if len(msg) == 0 {
		t.Error("error message should not be empty")
	}
}

func TestErrNoProvider_Error(t *testing.T) {
	e := ErrNoProvider{Name: "unknown"}
	msg := e.Error()
	if len(msg) == 0 {
		t.Error("error message should not be empty")
	}
}

func TestProvidersLoadedFromJSON(t *testing.T) {
	if len(Providers) < 10 {
		t.Fatalf("expected at least 10 providers from embedded JSON, got %d", len(Providers))
	}

	tests := []struct {
		name     string
		apiType  string
		authType string
		wantKey  string
	}{
		{"openai", "openai-compat", "api_key", "OPENAI_API_KEY"},
		{"deepseek", "openai-compat", "api_key", "DEEPSEEK_API_KEY"},
		{"xai", "openai-compat", "api_key", "XAI_API_KEY"},
		{"alibaba", "openai-compat", "api_key", "DASHSCOPE_API_KEY"},
		{"google", "openai-compat", "api_key", "GEMINI_API_KEY"},
		{"zhipu", "openai-compat", "api_key", "ZAI_API_KEY"},
		{"moonshot", "openai-compat", "api_key", "MOONSHOT_API_KEY"},
		{"github-copilot", "openai-compat", "oauth_device", "GITHUB_COPILOT_TOKEN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prov, found := Find(tt.name)
			if !found {
				t.Fatalf("%s not found in embedded providers", tt.name)
			}
			if prov.APIType != tt.apiType {
				t.Errorf("APIType = %q, want %q", prov.APIType, tt.apiType)
			}
			if prov.EnvKey != tt.wantKey {
				t.Errorf("EnvKey = %q, want %q", prov.EnvKey, tt.wantKey)
			}
			if got := prov.EffectiveAuthType(); got != tt.authType {
				t.Errorf("EffectiveAuthType() = %q, want %q", got, tt.authType)
			}
			if prov.DefaultModel == "" {
				t.Errorf("DefaultModel should not be empty for %s", tt.name)
			}
			if len(prov.Models) == 0 {
				t.Errorf("Models should not be empty for %s", tt.name)
			}
		})
	}
}

func TestProviderDefaultsToAPIKey(t *testing.T) {
	prov := Provider{}
	if got := prov.EffectiveAuthType(); got != "api_key" {
		t.Errorf("empty AuthType should default to api_key, got %q", got)
	}
	if prov.HasMultipleAuthOptions() {
		t.Error("provider with no auth_types should not have multiple options")
	}
}

func TestProviderMultipleAuthOptions(t *testing.T) {
	prov := Provider{AuthTypes: []string{"api_key", "oauth_pkce"}}
	if !prov.HasMultipleAuthOptions() {
		t.Error("provider with 2 auth_types should have multiple options")
	}
}

func TestGitHubCopilotOAuthConfig(t *testing.T) {
	prov, found := Find("github-copilot")
	if !found {
		t.Fatal("github-copilot not found")
	}
	if prov.AuthType != "oauth_device" {
		t.Errorf("AuthType = %q, want %q", prov.AuthType, "oauth_device")
	}
	if prov.OAuthConfig == nil {
		t.Fatal("OAuthConfig should not be nil for oauth_device provider")
	}
	if prov.OAuthConfig.DeviceAuthURL == "" {
		t.Error("DeviceAuthURL should not be empty")
	}
	if prov.OAuthConfig.TokenURL == "" {
		t.Error("TokenURL should not be empty")
	}
	if prov.OAuthConfig.Scope == "" {
		t.Error("Scope should not be empty")
	}
}

func TestOpenAIMultipleAuthTypes(t *testing.T) {
	prov, found := Find("openai")
	if !found {
		t.Fatal("openai not found")
	}
	if !prov.HasMultipleAuthOptions() {
		t.Error("openai should have multiple auth options (api_key + oauth_pkce)")
	}
}

