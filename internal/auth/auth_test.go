package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestToken_IsZero(t *testing.T) {
	tok := Token{}
	if !tok.IsZero() {
		t.Error("empty token should be zero")
	}
	tok2 := Token{AccessToken: "x"}
	if tok2.IsZero() {
		t.Error("token with access_token should not be zero")
	}
}

func TestForProvider_APIKey(t *testing.T) {
	m, err := ForProvider(ProviderInfo{AuthType: "api_key"})
	if err != nil {
		t.Fatalf("ForProvider: %v", err)
	}
	if m.Type() != "api_key" {
		t.Errorf("Type() = %q, want api_key", m.Type())
	}
	if m.NeedsInteractiveAuth() {
		t.Error("api_key should not need interactive auth")
	}
}

func TestForProvider_DefaultsToAPIKey(t *testing.T) {
	m, err := ForProvider(ProviderInfo{})
	if err != nil {
		t.Fatalf("ForProvider with empty AuthType: %v", err)
	}
	if m.Type() != "api_key" {
		t.Errorf("Type() = %q, want api_key", m.Type())
	}
}

func TestForProvider_OAuthDevice(t *testing.T) {
	m, err := ForProvider(ProviderInfo{
		AuthType: "oauth_device",
		OAuthConfig: &OAuthConfig{
			DeviceAuthURL: "https://github.com/login/device/code",
			TokenURL:      "https://github.com/login/oauth/access_token",
			ClientID:      "test-client",
			Scope:         "read:user",
		},
	})
	if err != nil {
		t.Fatalf("ForProvider: %v", err)
	}
	if m.Type() != "oauth_device" {
		t.Errorf("Type() = %q, want oauth_device", m.Type())
	}
	if !m.NeedsInteractiveAuth() {
		t.Error("oauth_device should need interactive auth")
	}
}

func TestForProvider_OAuthDeviceMissingConfig(t *testing.T) {
	_, err := ForProvider(ProviderInfo{AuthType: "oauth_device"})
	if err == nil {
		t.Fatal("expected error when oauth_device has no OAuthConfig")
	}
}

func TestForProvider_Unsupported(t *testing.T) {
	_, err := ForProvider(ProviderInfo{AuthType: "unknown"})
	if err == nil {
		t.Fatal("expected error for unsupported auth type")
	}
}

func TestApiKeyAuth_ApplyAuth(t *testing.T) {
	a := &ApiKeyAuth{}
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	a.ApplyAuth(req, Token{AccessToken: "sk-test", TokenType: "bearer"})

	if got := req.Header.Get("Authorization"); got != "Bearer sk-test" {
		t.Errorf("Authorization = %q, want %q", got, "Bearer sk-test")
	}
}

func TestApiKeyAuth_Refresh(t *testing.T) {
	a := &ApiKeyAuth{}
	_, err := a.Refresh(ProviderInfo{}, Token{}, nil)
	if err != ErrNoRefresh {
		t.Errorf("expected ErrNoRefresh, got %v", err)
	}
}

func TestDeviceFlow_ApplyAuth(t *testing.T) {
	a := &DeviceFlowAuth{}
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	a.ApplyAuth(req, Token{AccessToken: "gho_test", TokenType: "bearer"})

	if got := req.Header.Get("Authorization"); got != "Bearer gho_test" {
		t.Errorf("Authorization = %q, want %q", got, "Bearer gho_test")
	}
}

func TestDeviceFlow_Authenticate_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login/device/code" {
			json.NewEncoder(w).Encode(deviceCodeResp{
				DeviceCode:      "dev-123",
				UserCode:        "ABCD-1234",
				VerificationURI: "https://github.com/login/device",
				ExpiresIn:       900,
				Interval:        1,
			})
			return
		}
		if r.URL.Path == "/login/oauth/access_token" {
			json.NewEncoder(w).Encode(tokenResp{
				AccessToken: "gho_test_token",
				TokenType:   "bearer",
				Scope:       "read:user",
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	a := &DeviceFlowAuth{
		DeviceAuthURL: srv.URL + "/login/device/code",
		TokenURL:      srv.URL + "/login/oauth/access_token",
		ClientID:      "test-client",
		Scope:         "read:user",
	}

	ch := a.Authenticate(ProviderInfo{Name: "test-provider"}, nil)

	s1 := <-ch
	if s1.Stage != StageAwaitingUser {
		t.Errorf("first stage = %q, want %q", s1.Stage, StageAwaitingUser)
	}
	if s1.UserCode != "ABCD-1234" {
		t.Errorf("UserCode = %q", s1.UserCode)
	}

	s2 := <-ch
	if s2.Stage != StageAuthorized {
		t.Errorf("second stage = %q (err=%v), want %q", s2.Stage, s2.Err, StageAuthorized)
	}
	if s2.Token == nil || s2.Token.AccessToken != "gho_test_token" {
		t.Errorf("Token = %v", s2.Token)
	}
}

func TestDeviceFlow_Authenticate_PendingThenSuccess(t *testing.T) {
	pollCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login/device/code" {
			json.NewEncoder(w).Encode(deviceCodeResp{
				DeviceCode:      "dev-456",
				UserCode:        "EFGH-5678",
				VerificationURI: "https://github.com/login/device",
				ExpiresIn:       900,
				Interval:        1,
			})
			return
		}
		if r.URL.Path == "/login/oauth/access_token" {
			pollCount++
			if pollCount < 3 {
				json.NewEncoder(w).Encode(tokenResp{
					Error:     "authorization_pending",
					ErrorDesc: "The authorization request is still pending",
				})
				return
			}
			json.NewEncoder(w).Encode(tokenResp{
				AccessToken: "gho_delayed",
				TokenType:   "bearer",
				Scope:       "read:user",
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	a := &DeviceFlowAuth{
		DeviceAuthURL: srv.URL + "/login/device/code",
		TokenURL:      srv.URL + "/login/oauth/access_token",
		ClientID:      "test-client",
		Scope:         "read:user",
	}

	ch := a.Authenticate(ProviderInfo{Name: "test-provider"}, nil)

	s1 := <-ch
	if s1.Stage != StageAwaitingUser {
		t.Errorf("first stage = %q", s1.Stage)
	}

	for s := range ch {
		if s.Stage == StagePolling {
			continue
		}
		if s.Stage == StageAuthorized {
			if s.Token.AccessToken != "gho_delayed" {
				t.Errorf("Token = %q", s.Token.AccessToken)
			}
			return
		}
		t.Errorf("unexpected stage: %q (err=%v)", s.Stage, s.Err)
	}
}

func TestDeviceFlow_Authenticate_Denied(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login/device/code" {
			json.NewEncoder(w).Encode(deviceCodeResp{
				DeviceCode:      "dev-789",
				UserCode:        "IJKL-9012",
				VerificationURI: "https://github.com/login/device",
				ExpiresIn:       900,
				Interval:        1,
			})
			return
		}
		if r.URL.Path == "/login/oauth/access_token" {
			json.NewEncoder(w).Encode(tokenResp{
				Error:     "access_denied",
				ErrorDesc: "User denied authorization",
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	a := &DeviceFlowAuth{
		DeviceAuthURL: srv.URL + "/login/device/code",
		TokenURL:      srv.URL + "/login/oauth/access_token",
		ClientID:      "test-client",
		Scope:         "read:user",
	}

	ch := a.Authenticate(ProviderInfo{Name: "test-provider"}, nil)

	<-ch // awaiting user
	s2 := <-ch
	if s2.Stage != StageError {
		t.Errorf("expected StageError, got %q", s2.Stage)
	}
	if s2.Err == nil {
		t.Error("expected error for access_denied")
	}
}

func TestDeviceFlow_EnvVarOverride(t *testing.T) {
	t.Setenv("QDOC_GITHUB_CLIENT_ID", "env-client-id")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login/device/code" {
			r.ParseForm()
			if got := r.Form.Get("client_id"); got != "env-client-id" {
				t.Errorf("client_id = %q, want env-client-id", got)
			}
			json.NewEncoder(w).Encode(deviceCodeResp{
				DeviceCode:      "dev-env",
				UserCode:        "ENVV-AR01",
				VerificationURI: "https://github.com/login/device",
				ExpiresIn:       900,
				Interval:        1,
			})
			return
		}
		if r.URL.Path == "/login/oauth/access_token" {
			json.NewEncoder(w).Encode(tokenResp{
				AccessToken: "gho_env_token",
				TokenType:   "bearer",
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	a := &DeviceFlowAuth{
		DeviceAuthURL: srv.URL + "/login/device/code",
		TokenURL:      srv.URL + "/login/oauth/access_token",
		ClientID:      "embedded-client",
		Scope:         "read:user",
	}

	ch := a.Authenticate(ProviderInfo{Name: "test-provider"}, nil)
	<-ch
	s2 := <-ch
	if s2.Stage != StageAuthorized {
		t.Errorf("stage = %q (err=%v)", s2.Stage, s2.Err)
	}
}

func TestTokenStore_Set_Get_Delete(t *testing.T) {
	dir := t.TempDir()
	store, err := LoadTokenStore(dir)
	if err != nil {
		t.Fatalf("LoadTokenStore: %v", err)
	}

	_, ok := store.Get("test-provider")
	if ok {
		t.Error("expected no token initially")
	}

	tok := Token{AccessToken: "test-token", TokenType: "bearer"}
	if err := store.Set("test-provider", tok); err != nil {
		t.Fatalf("Set: %v", err)
	}

	got, ok := store.Get("test-provider")
	if !ok {
		t.Fatal("expected token to be present")
	}
	if got.AccessToken != "test-token" {
		t.Errorf("AccessToken = %q", got.AccessToken)
	}

	if err := store.Delete("test-provider"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, ok = store.Get("test-provider")
	if ok {
		t.Error("expected token to be deleted")
	}
}

func TestTokenStore_IsExpired(t *testing.T) {
	dir := t.TempDir()
	store, err := LoadTokenStore(dir)
	if err != nil {
		t.Fatalf("LoadTokenStore: %v", err)
	}

	if store.IsExpired("nonexistent") {
		t.Error("nonexistent should not be expired")
	}

	store.Set("no-expiry", Token{AccessToken: "test", TokenType: "bearer"})
	if store.IsExpired("no-expiry") {
		t.Error("token with zero expiry should not be expired")
	}
}

func TestTokenStore_Persistence(t *testing.T) {
	dir := t.TempDir()
	store, err := LoadTokenStore(dir)
	if err != nil {
		t.Fatalf("LoadTokenStore: %v", err)
	}
	store.Set("p1", Token{AccessToken: "t1", TokenType: "bearer"})

	store2, err := LoadTokenStore(dir)
	if err != nil {
		t.Fatalf("LoadTokenStore (reload): %v", err)
	}
	tok, ok := store2.Get("p1")
	if !ok || tok.AccessToken != "t1" {
		t.Errorf("token not persisted: ok=%v tok=%v", ok, tok)
	}
}

func TestTokenStore_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tokens.json")
	os.WriteFile(path, []byte(`{}`), 0600)

	store, err := LoadTokenStore(dir)
	if err != nil {
		t.Fatalf("LoadTokenStore: %v", err)
	}
	_, ok := store.Get("anything")
	if ok {
		t.Error("expected no tokens from empty file")
	}
}
