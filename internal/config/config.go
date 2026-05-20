package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Provider string            `json:"provider"`
	Keys     map[string]string `json:"keys"`
	Models   map[string]string `json:"models"`
}

func Default() Config {
	return Config{
		Provider: "openai",
		Keys:     make(map[string]string),
		Models:   make(map[string]string),
	}
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("no home dir: %w", err)
	}
	dir := filepath.Join(home, ".config", "qdoc")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("creating config dir: %w", err)
	}
	return dir, nil
}

func LogDir() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	logsDir := filepath.Join(dir, "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return "", fmt.Errorf("creating logs dir: %w", err)
	}
	return logsDir, nil
}

func Path() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func Load() (Config, error) {
	path, err := Path()
	if err != nil {
		return Default(), err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return Default(), err
	}

	cfg := Default()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Default(), err
	}

	if cfg.Keys == nil {
		cfg.Keys = make(map[string]string)
	}
	if cfg.Models == nil {
		cfg.Models = make(map[string]string)
	}

	return cfg, nil
}

func Save(cfg Config) error {
	path, err := Path()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}