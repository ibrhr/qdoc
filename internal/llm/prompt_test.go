package llm

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ibrhr/qdoc/internal/docsource"
)

func TestBuildSystemPrompt_Basic(t *testing.T) {
	source := docsource.Source{Name: "go", BaseURL: "https://go.dev/doc"}
	entries := []docsource.Entry{
		{URL: "https://go.dev/doc/effective_go", Title: "Effective Go"},
		{URL: "https://go.dev/doc/tutorial", Title: "Getting Started"},
	}
	query := "how do I use generics?"

	prompt := BuildSystemPrompt(source, entries, query)

	if !strings.Contains(prompt, "qdoc") {
		t.Error("prompt missing qdoc intro")
	}
	if !strings.Contains(prompt, "go documentation") {
		t.Error("prompt missing source name")
	}
	if !strings.Contains(prompt, "how do I use generics?") {
		t.Error("prompt missing query")
	}
	if !strings.Contains(prompt, "effective_go") {
		t.Error("prompt missing entry URL")
	}
	if !strings.Contains(prompt, "Effective Go") {
		t.Error("prompt missing entry title")
	}
	if !strings.Contains(prompt, "BASE URL") {
		t.Error("prompt missing base URL")
	}
}

func TestBuildSystemPrompt_Truncation(t *testing.T) {
	source := docsource.Source{Name: "go", BaseURL: "https://go.dev/doc"}
	entries := make([]docsource.Entry, 200)
	for i := range entries {
		entries[i] = docsource.Entry{
			URL:   fmt.Sprintf("https://go.dev/doc/page_%d", i),
			Title: fmt.Sprintf("Page %d", i),
		}
	}
	query := "test query"

	prompt := BuildSystemPrompt(source, entries, query)

	if !strings.Contains(prompt, "200 total") {
		t.Error("prompt should mention original total (200)")
	}
	if !strings.Contains(prompt, "showing first 120") {
		t.Error("prompt should mention showing first 120")
	}
}

func TestBuildSystemPrompt_EmptyEntries(t *testing.T) {
	source := docsource.Source{Name: "go", BaseURL: "https://go.dev/doc"}
	prompt := BuildSystemPrompt(source, nil, "query")
	if !strings.Contains(prompt, "QUERY: query") {
		t.Error("prompt should work with nil entries")
	}
}

func TestBuildSystemPrompt_MaxEntriesExact(t *testing.T) {
	source := docsource.Source{Name: "go", BaseURL: "https://go.dev/doc"}
	entries := make([]docsource.Entry, maxIndexEntries)
	for i := range entries {
		entries[i] = docsource.Entry{
			URL:   fmt.Sprintf("https://go.dev/doc/page_%d", i),
			Title: fmt.Sprintf("Page %d", i),
		}
	}
	prompt := BuildSystemPrompt(source, entries, "query")

	if strings.Contains(prompt, "... (") {
		t.Error("should not show truncation message when exactly at max")
	}
}

func TestBuildSystemPrompt_SystemPrompt(t *testing.T) {
	source := docsource.Source{
		Name:         "go",
		BaseURL:      "https://go.dev/doc",
		SystemPrompt: "Custom system instructions for Go.",
	}
	entries := []docsource.Entry{
		{URL: "https://go.dev/doc/one", Title: "One"},
	}
	prompt := BuildSystemPrompt(source, entries, "query")

	if !strings.Contains(prompt, "Custom system instructions for Go.") {
		t.Error("prompt missing system prompt from source")
	}
}

func TestBuildSystemPrompt_Structure(t *testing.T) {
	source := docsource.Source{Name: "go", BaseURL: "https://go.dev/doc"}
	entries := []docsource.Entry{
		{URL: "https://go.dev/doc/test", Title: "Test"},
	}
	prompt := BuildSystemPrompt(source, entries, "test query")

	if !strings.Contains(prompt, "HOW IT WORKS") {
		t.Error("missing HOW IT WORKS section")
	}
	if !strings.Contains(prompt, "AVAILABLE FILES") {
		t.Error("missing AVAILABLE FILES section")
	}
	if !strings.Contains(prompt, "QUERY:") {
		t.Error("missing QUERY section")
	}
	if !strings.Contains(prompt, `{"action":"read_file"`) {
		t.Error("missing action format example")
	}
	if !strings.Contains(prompt, `{"action":"answer"`) {
		t.Error("missing answer action format")
	}
}
