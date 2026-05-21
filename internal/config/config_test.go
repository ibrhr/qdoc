package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func testConfigDir(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	home := filepath.Join(tmp, "home")
	if err := os.MkdirAll(home, 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	return home
}

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.Provider != "openai" {
		t.Errorf("Provider = %q, want %q", cfg.Provider, "openai")
	}
	if cfg.Keys == nil {
		t.Error("Keys should not be nil")
	}
	if cfg.Models == nil {
		t.Error("Models should not be nil")
	}
}

func TestLoad_NotExist(t *testing.T) {
	testConfigDir(t)
	cfg, err := Load()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if cfg.Provider != "openai" {
		t.Errorf("Provider = %q, want %q", cfg.Provider, "openai")
	}
}

func TestLoad_Valid(t *testing.T) {
	testConfigDir(t)
	c := Default()
	c.Provider = "deepseek"
	c.Keys["deepseek"] = "sk-test"
	c.Models["deepseek"] = "gpt-4"
	data, _ := json.MarshalIndent(c, "", "  ")
	path, _ := Path()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Provider != "deepseek" {
		t.Errorf("Provider = %q, want %q", cfg.Provider, "deepseek")
	}
	if cfg.Keys["deepseek"] != "sk-test" {
		t.Errorf("Keys[deepseek] = %q, want %q", cfg.Keys["deepseek"], "sk-test")
	}
	if cfg.Models["deepseek"] != "gpt-4" {
		t.Errorf("Models[deepseek] = %q, want %q", cfg.Models["deepseek"], "gpt-4")
	}
}

func TestLoad_Corrupted(t *testing.T) {
	testConfigDir(t)
	path, _ := Path()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("not json"), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err == nil {
		t.Error("expected error for corrupted config, got nil")
	}
	if cfg.Provider != "openai" {
		t.Errorf("expected default provider on corrupt, got %q", cfg.Provider)
	}
}

func TestLoad_NilMaps(t *testing.T) {
	testConfigDir(t)
	path, _ := Path()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(path, []byte(`{"provider":"openai"}`), 0600)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Keys == nil {
		t.Error("Keys should be non-nil after load")
	}
	if cfg.Models == nil {
		t.Error("Models should be non-nil after load")
	}
}

func TestSave_Persists(t *testing.T) {
	testConfigDir(t)
	c := Default()
	c.Provider = "deepseek"
	c.Keys["deepseek"] = "sk-abc"

	if err := Save(c); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if loaded.Provider != c.Provider {
		t.Errorf("Provider = %q, want %q", loaded.Provider, c.Provider)
	}
	if loaded.Keys["deepseek"] != c.Keys["deepseek"] {
		t.Errorf("Keys[deepseek] = %q, want %q", loaded.Keys["deepseek"], c.Keys["deepseek"])
	}
}

func TestSave_CreatesDir(t *testing.T) {
	testConfigDir(t)
	c := Default()
	if err := Save(c); err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	path, _ := Path()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("config file not created: %v", err)
	}
}

func TestLogDir(t *testing.T) {
	testConfigDir(t)
	dir, err := LogDir()
	if err != nil {
		t.Fatalf("LogDir() error: %v", err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("logs dir not created: %v", err)
	}
}

func TestPath(t *testing.T) {
	testConfigDir(t)
	path, err := Path()
	if err != nil {
		t.Fatalf("Path() error: %v", err)
	}
	if path != filepath.Join(os.Getenv("HOME"), ".config", "qdoc", "config.json") {
		t.Errorf("Path() = %q, unexpected", path)
	}
}
