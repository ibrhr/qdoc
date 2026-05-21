package docsource

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestExtractLinks_Basic(t *testing.T) {
	html := `<html><body>
<a href="/doc/effective_go">Effective Go</a>
<a href="/doc/tutorial">Getting Started</a>
</body></html>`

	entries := extractLinks(html, "/doc/", "https://go.dev/doc")

	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(entries))
	}
	if entries[0].URL != "https://go.dev/doc/effective_go" {
		t.Errorf("URL = %q", entries[0].URL)
	}
	if entries[0].Title != "Effective Go" {
		t.Errorf("Title = %q", entries[0].Title)
	}
	if entries[1].URL != "https://go.dev/doc/tutorial" {
		t.Errorf("URL = %q", entries[1].URL)
	}
}

func TestExtractLinks_SkipBlocked(t *testing.T) {
	html := `<html><body>
<a href="/doc/manual.pdf">Manual PDF</a>
<a href="/doc/icon.png">Icon</a>
<a href="/doc/styles.css">Styles</a>
<a href="/doc/script.js">Script</a>
<a href="#section">Anchor</a>
<a href="mailto:test@test.com">Email</a>
<a href="http://external.com">External</a>
<a href="/doc/about-png-format">About PNG Format</a>
</body></html>`

	entries := extractLinks(html, "/doc/", "https://go.dev/doc")

	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2 (manual.pdf and about-png-format should NOT be blocked)", len(entries))
	}
	urls := make(map[string]bool)
	for _, e := range entries {
		urls[e.URL] = true
	}
	if !urls["https://go.dev/doc/manual.pdf"] {
		t.Error("manual.pdf should not be blocked")
	}
	if !urls["https://go.dev/doc/about-png-format"] {
		t.Error("about-png-format should not be blocked")
	}
}

func TestExtractLinks_Duplicates(t *testing.T) {
	html := `<html><body>
<a href="/doc/same">Link 1</a>
<a href="/doc/same">Link 2</a>
</body></html>`

	entries := extractLinks(html, "/doc/", "https://go.dev/doc")
	if len(entries) != 1 {
		t.Errorf("duplicates should be deduped, got %d entries", len(entries))
	}
}

func TestExtractLinks_EmptyTitle(t *testing.T) {
	html := `<html><body>
<a href="/doc/untitled"></a>
</body></html>`

	entries := extractLinks(html, "/doc/", "https://go.dev/doc")
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(entries))
	}
	if entries[0].Title != "untitled" {
		t.Errorf("Title = %q, want %q", entries[0].Title, "untitled")
	}
}

func TestExtractLinks_NoMatch(t *testing.T) {
	html := `<html><body>
<a href="/other/one">Other</a>
<a href="/other/two">Other</a>
</body></html>`

	entries := extractLinks(html, "/doc/", "https://go.dev/doc")
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestExtractText(t *testing.T) {
	doc, err := html.Parse(strings.NewReader("<span>hello <b>world</b></span>"))
	if err != nil {
		t.Fatal(err)
	}
	// find the span
	var span *html.Node
	var findSpan func(*html.Node)
	findSpan = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "span" {
			span = n
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findSpan(c)
		}
	}
	findSpan(doc)
	if span == nil {
		t.Fatal("span not found")
	}
	result := extractText(span)
	if result != "hello world" {
		t.Errorf("extractText = %q, want %q", result, "hello world")
	}
}

func TestTruncate(t *testing.T) {
	short := "short string"
	if truncate(short) != short {
		t.Error("short string should not be truncated")
	}

	long := strings.Repeat("a", maxContentChars+100)
	result := truncate(long)
	if len(result) != maxContentChars {
		t.Errorf("truncated length = %d, want %d", len(result), maxContentChars)
	}
}
