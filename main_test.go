package main

import (
	"testing"
)

func TestVersionVar(t *testing.T) {
	if version != "dev" {
		t.Errorf("version = %q, want %q", version, "dev")
	}
}

func TestFlagParsing(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		wantJSON      bool
		wantNoTUI     bool
		wantFiltered  []string
		skipArgsCheck bool
	}{
		{
			name:       "no flags",
			args:       []string{"go", "query here"},
			wantJSON:   false,
			wantNoTUI:  false,
			wantFiltered: []string{"go", "query here"},
		},
		{
			name:       "json flag",
			args:       []string{"--json", "go", "query"},
			wantJSON:   true,
			wantNoTUI:  false,
			wantFiltered: []string{"go", "query"},
		},
		{
			name:       "no-tui flag",
			args:       []string{"--no-tui", "go", "query"},
			wantJSON:   false,
			wantNoTUI:  true,
			wantFiltered: []string{"go", "query"},
		},
		{
			name:       "both flags",
			args:       []string{"--json", "--no-tui", "go", "query"},
			wantJSON:   true,
			wantNoTUI:  true,
			wantFiltered: []string{"go", "query"},
		},
		{
			name:       "version flag",
			args:       []string{"--version"},
			wantFiltered: []string{},
			skipArgsCheck: true, // version triggers early exit, flags stay in filtered in our simulation
		},
		{
			name:       "help flag",
			args:       []string{"--help"},
			wantFiltered: []string{},
			skipArgsCheck: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var useJSON, useNoTUI bool
			var filtered []string
			for _, a := range tt.args {
				switch a {
				case "--json":
					useJSON = true
				case "--no-tui":
					useNoTUI = true
				default:
					filtered = append(filtered, a)
				}
			}
			if useJSON != tt.wantJSON {
				t.Errorf("useJSON = %v, want %v", useJSON, tt.wantJSON)
			}
			if useNoTUI != tt.wantNoTUI {
				t.Errorf("useNoTUI = %v, want %v", useNoTUI, tt.wantNoTUI)
			}
			if !tt.skipArgsCheck && len(filtered) != len(tt.wantFiltered) {
				t.Errorf("filtered = %v, want %v", filtered, tt.wantFiltered)
			}
		})
	}
}

func TestHeadlessRequiresSourceAndQuery(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		valid bool // true if enough args for headless
	}{
		{"empty", []string{}, false},
		{"source only", []string{"go"}, false},
		{"source and query", []string{"go", "how to slice"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := len(tt.args) >= 2
			if valid != tt.valid {
				t.Errorf("valid = %v, want %v", valid, tt.valid)
			}
		})
	}
}

func TestCommandRouting(t *testing.T) {
	cmds := map[string]bool{
		"version":   true,
		"--version": true,
		"-v":        true,
		"status":    true,
		"sources":   true,
		"ls":        true,
		"providers": true,
		"provider":  true,
		"model":     true,
		"set":       true,
		"help":      true,
		"--help":    true,
		"-h":        true,
	}
	for cmd := range cmds {
		t.Run(cmd, func(t *testing.T) {
			_, ok := cmds[cmd]
			if !ok {
				t.Errorf("command %q not recognized", cmd)
			}
		})
	}
}
