package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type tokenData struct {
	Tokens map[string]Token `json:"tokens"`
}

type TokenStore struct {
	path string
	mu   sync.RWMutex
	data tokenData
}

func LoadTokenStore(dir string) (*TokenStore, error) {
	ts := &TokenStore{
		path: filepath.Join(dir, "tokens.json"),
		data: tokenData{Tokens: make(map[string]Token)},
	}

	data, err := os.ReadFile(ts.path)
	if err != nil {
		if os.IsNotExist(err) {
			return ts, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, &ts.data); err != nil {
		return nil, err
	}
	if ts.data.Tokens == nil {
		ts.data.Tokens = make(map[string]Token)
	}
	return ts, nil
}

func (ts *TokenStore) Get(provider string) (Token, bool) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	t, ok := ts.data.Tokens[provider]
	return t, ok
}

func (ts *TokenStore) Set(provider string, token Token) error {
	ts.mu.Lock()
	ts.data.Tokens[provider] = token
	ts.mu.Unlock()
	return ts.save()
}

func (ts *TokenStore) Delete(provider string) error {
	ts.mu.Lock()
	delete(ts.data.Tokens, provider)
	ts.mu.Unlock()
	return ts.save()
}

func (ts *TokenStore) IsExpired(provider string) bool {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	t, ok := ts.data.Tokens[provider]
	if !ok {
		return false
	}
	if t.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(t.ExpiresAt)
}

func (ts *TokenStore) save() error {
	data, err := json.MarshalIndent(ts.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ts.path, data, 0600)
}
