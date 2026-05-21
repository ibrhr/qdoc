package docsource

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestFind_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"go", true},
		{"GO", true},
		{"Go", true},
		{"gO", true},
		{"nextjs", true},
		{"NEXTJS", true},
		{"unknown", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, found := Find(tt.name)
			if found != tt.want {
				t.Errorf("Find(%q) = %v, want %v", tt.name, found, tt.want)
			}
		})
	}
}

func TestFind_LocalDir(t *testing.T) {
	dir := t.TempDir()
	abs, _ := filepath.Abs(dir)

	source, found := Find(dir)
	if !found {
		t.Fatal("expected local directory to be found")
	}
	if !source.Local {
		t.Error("expected Local = true for directory source")
	}
	if source.Name != filepath.Base(abs) {
		t.Errorf("Name = %q", source.Name)
	}
}

func TestFetchIndex_Local(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "sub"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "readme.md"), []byte("# Hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sub", "info.html"), []byte("<p>Info</p>"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "hidden.txt"), []byte("nope"), 0644); err != nil {
		t.Fatal(err)
	}

	source := Source{
		Name:     filepath.Base(dir),
		BaseURL:  "file://" + dir,
		IndexURL: "file://" + dir,
		Local:    true,
	}

	entries, err := source.FetchIndex()
	if err != nil {
		t.Fatalf("FetchIndex() error: %v", err)
	}

	if len(entries) != 3 {
		t.Fatalf("expected 3 entries (.md, .html, .txt), got %d", len(entries))
	}

	urls := make(map[string]bool)
	for _, e := range entries {
		urls[e.URL] = true
	}
	if !urls["readme.md"] {
		t.Error("expected readme.md entry")
	}
	if !urls["sub/info.html"] {
		t.Error("expected sub/info.html entry")
	}
}

func TestFetchContent_Local(t *testing.T) {
	dir := t.TempDir()
	content := "# Hello\nThis is a test document."
	if err := os.WriteFile(filepath.Join(dir, "readme.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	source := Source{
		Name:     filepath.Base(dir),
		BaseURL:  "file://" + dir,
		IndexURL: "file://" + dir,
		Local:    true,
	}

	result, err := source.FetchContent("readme.md")
	if err != nil {
		t.Fatalf("FetchContent() error: %v", err)
	}
	if !strings.Contains(result, "test document") {
		t.Errorf("content = %q, expected 'test document'", result)
	}
}

func TestFetchIndex_Remote(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html><body>
<a href="/docs/getting-started">Getting Started</a>
<a href="/docs/api">API Reference</a>
<a href="/docs/styles.css">CSS</a>
</body></html>`))
	}))
	defer srv.Close()

	source := Source{
		Name:       "test",
		BaseURL:    srv.URL + "/docs",
		IndexURL:   srv.URL + "/index",
		LinkPrefix: "/docs/",
	}
	entries, err := source.FetchIndex()
	if err != nil {
		t.Fatalf("FetchIndex() error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries (css blocked), got %d", len(entries))
	}
}

func TestFetchContent_Remote(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html><body>
<div class="content">
<p>Main content here with enough text to pass the minimum threshold of one hundred characters for extraction purposes.</p>
</div>
</body></html>`))
	}))
	defer srv.Close()

	source := Source{
		Name:       "test",
		BaseURL:    srv.URL,
		IndexURL:   srv.URL + "/index",
		LinkPrefix: "/docs/",
	}
	content, err := source.FetchContent(srv.URL + "/page")
	if err != nil {
		t.Fatalf("FetchContent() error: %v", err)
	}
	if !strings.Contains(content, "Main content here") {
		t.Errorf("content = %q", content)
	}
}

func TestFetchContent_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	source := Source{
		Name:       "test",
		BaseURL:    srv.URL,
		IndexURL:   srv.URL + "/index",
		LinkPrefix: "/docs/",
	}
	_, err := source.FetchContent(srv.URL + "/missing")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestFetchContent_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	source := Source{
		Name:       "test",
		BaseURL:    srv.URL,
		IndexURL:   srv.URL + "/index",
		LinkPrefix: "/docs/",
	}
	_, err := source.FetchContent(srv.URL + "/auth-required")
	if err == nil {
		t.Fatal("expected error for 401")
	}
}

func TestFetchContent_BadRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	source := Source{
		Name:       "test",
		BaseURL:    srv.URL,
		IndexURL:   srv.URL + "/index",
		LinkPrefix: "/docs/",
	}
	_, err := source.FetchContent(srv.URL + "/bad")
	if err == nil {
		t.Fatal("expected error for 400")
	}
}

func TestKnownSources(t *testing.T) {
	if len(KnownSources) != 6 {
		t.Errorf("expected 6 known sources, got %d", len(KnownSources))
	}
	names := []string{"go", "nextjs", "python", "react", "fastapi", "pydantic"}
	for _, name := range names {
		found := false
		for _, s := range KnownSources {
			if s.Name == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("known source %q not found", name)
		}
	}
}

func TestFetchLocalIndex_SkipsHidden(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".hidden"), 0755)
	os.MkdirAll(filepath.Join(dir, "node_modules", "pkg"), 0755)
	os.MkdirAll(filepath.Join(dir, "visible"), 0755)
	os.WriteFile(filepath.Join(dir, ".hidden", "secret.md"), []byte("secret"), 0644)
	os.WriteFile(filepath.Join(dir, "node_modules", "pkg", "readme.md"), []byte("pkg"), 0644)
	os.WriteFile(filepath.Join(dir, "visible", "doc.md"), []byte("doc"), 0644)

	source := Source{
		Name:     "test",
		BaseURL:  "file://" + dir,
		IndexURL: "file://" + dir,
		Local:    true,
	}

	entries, err := source.FetchIndex()
	if err != nil {
		t.Fatalf("FetchIndex() error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry (visible/doc.md), got %d", len(entries))
	}
	if entries[0].URL != "visible/doc.md" {
		t.Errorf("URL = %q", entries[0].URL)
	}
}

func TestFetchContent_LocalTruncation(t *testing.T) {
	dir := t.TempDir()
	longContent := strings.Repeat("x", maxContentChars+500)
	os.WriteFile(filepath.Join(dir, "long.md"), []byte(longContent), 0644)

	source := Source{
		Name:     "test",
		BaseURL:  "file://" + dir,
		IndexURL: "file://" + dir,
		Local:    true,
	}

	result, err := source.FetchContent("long.md")
	if err != nil {
		t.Fatalf("FetchContent() error: %v", err)
	}
	if len(result) > maxContentChars {
		t.Errorf("content length %d exceeds max %d", len(result), maxContentChars)
	}
}

func TestFetchLocalIndex_SortedByURL(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "zzz.md"), []byte("z"), 0644)
	os.WriteFile(filepath.Join(dir, "aaa.md"), []byte("a"), 0644)

	source := Source{
		Name:     "test",
		BaseURL:  "file://" + dir,
		IndexURL: "file://" + dir,
		Local:    true,
	}

	entries, err := source.FetchIndex()
	if err != nil {
		t.Fatalf("FetchIndex() error: %v", err)
	}
	if !sort.SliceIsSorted(entries, func(i, j int) bool {
		return entries[i].URL < entries[j].URL
	}) {
		t.Errorf("entries not sorted by URL: %v", entries)
	}
}

func TestSource_FetchContent_HTTPIndex(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fmt.Sprintf(`<html><body>
<a href="/doc/page1">Page 1</a>
<a href="/doc/page2">Page 2</a>
</body></html>`)))
	}))
	defer srv.Close()

	source := Source{
		Name:       "testsource",
		BaseURL:    srv.URL + "/doc",
		IndexURL:   srv.URL + "/index",
		LinkPrefix: "/doc/",
	}
	entries, err := source.FetchIndex()
	if err != nil {
		t.Fatalf("FetchIndex() error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Title != "Page 1" {
		t.Errorf("Title = %q", entries[0].Title)
	}
	if entries[1].Title != "Page 2" {
		t.Errorf("Title = %q", entries[1].Title)
	}
}
