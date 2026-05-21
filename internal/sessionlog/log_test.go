package sessionlog

import (
	"os"
	"path/filepath"
	"testing"
)

func testLogDir(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	home := filepath.Join(tmp, "home")
	os.MkdirAll(home, 0755)
	logsDir := filepath.Join(home, ".config", "qdoc", "logs")
	os.MkdirAll(logsDir, 0755)
	t.Setenv("HOME", home)
	return logsDir
}

func TestSanitize_Alphanumeric(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"hello", "hello"},
		{"hello world", "hello_world"},
		{"hello-world", "hello-world"},
		{"hello_world", "hello_world"},
		{"GoLang-123", "GoLang-123"},
		{"query?with=params&more", "query_with_params_more"},
		{"hello!", "hello_"},
		{"hello@#$%", "hello____"},
		{" spaces ", "_spaces_"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := sanitize(tt.input); got != tt.want {
				t.Errorf("sanitize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSanitize_Truncation(t *testing.T) {
	long := ""
	for i := 0; i < 100; i++ {
		long += "abcdefghij"
	}
	result := sanitize(long)
	// sanitize() itself does not truncate; truncation happens in New()
	if len(result) != 1000 {
		t.Errorf("sanitize doesn't truncate, expected len %d, got len %d", 1000, len(result))
	}
}

func TestSanitize_Empty(t *testing.T) {
	if got := sanitize(""); got != "query" {
		t.Errorf("sanitize(\"\") = %q, want %q", got, "query")
	}
}

func TestNew_CreatesFile(t *testing.T) {
	testLogDir(t)
	l, err := New("go", "test query")
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if l.f == nil {
		t.Fatal("Logger.f is nil")
	}
	l.Close()
}

func TestLog(t *testing.T) {
	testLogDir(t)
	l, err := New("go", "test log message")
	if err != nil {
		t.Fatal(err)
	}
	l.Log("test: %s", "value")
	l.Close()

	data, err := os.ReadFile(l.f.Name())
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if len(content) == 0 {
		t.Error("log file is empty")
	}
}

func TestSection(t *testing.T) {
	testLogDir(t)
	l, err := New("go", "section test")
	if err != nil {
		t.Fatal(err)
	}
	l.Section("My Section")
	l.Close()

	data, err := os.ReadFile(l.f.Name())
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !contains(content, "─── My Section ───") {
		t.Errorf("section marker not found in: %s", content)
	}
}

func TestRaw(t *testing.T) {
	testLogDir(t)
	l, err := New("go", "raw test")
	if err != nil {
		t.Fatal(err)
	}
	l.Raw("RAW BLOCK", "raw content here")
	l.Close()

	data, err := os.ReadFile(l.f.Name())
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !contains(content, "raw content here") {
		t.Errorf("raw content not found in: %s", content)
	}
}

func TestRaw_AddsNewline(t *testing.T) {
	testLogDir(t)
	l, err := New("go", "newline test")
	if err != nil {
		t.Fatal(err)
	}
	l.Raw("BLOCK", "no trailing newline")
	l.Close()

	data, err := os.ReadFile(l.f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if data[len(data)-1] != '\n' {
		t.Error("raw content should end with newline")
	}
}

func TestClose_WritesEndMarker(t *testing.T) {
	testLogDir(t)
	l, err := New("go", "close test")
	if err != nil {
		t.Fatal(err)
	}
	if err := l.Close(); err != nil {
		t.Errorf("Close() error: %v", err)
	}

	data, err := os.ReadFile(l.f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if !contains(string(data), "=== END ===") {
		t.Error("end marker not found")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
