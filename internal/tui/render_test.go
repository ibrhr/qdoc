package tui

import (
	"strings"
	"testing"
)

func TestRenderMarkdown_Headers(t *testing.T) {
	input := `# H1 Title
## H2 Title
### H3 Title`

	result := renderMarkdown(input)

	if !strings.Contains(result, "H1 Title") {
		t.Error("missing H1")
	}
	if !strings.Contains(result, "H2 Title") {
		t.Error("missing H2")
	}
	if !strings.Contains(result, "H3 Title") {
		t.Error("missing H3")
	}
}

func TestRenderMarkdown_CodeBlocks(t *testing.T) {
	input := "```go\nfunc hello() {\n    fmt.Println(\"hi\")\n}\n```"
	result := renderMarkdown(input)

	if !strings.Contains(result, "func hello()") {
		t.Error("code block content missing")
	}
}

func TestRenderMarkdown_InlineCode(t *testing.T) {
	input := "Use the `fmt.Println` function"
	result := renderMarkdown(input)
	if !strings.Contains(result, "fmt.Println") {
		t.Error("inline code content missing")
	}
}

func TestRenderMarkdown_BoldItalic(t *testing.T) {
	input := "**bold text** and *italic text*"
	result := renderMarkdown(input)
	if !strings.Contains(result, "bold text") {
		t.Error("bold text missing")
	}
	if !strings.Contains(result, "italic text") {
		t.Error("italic text missing")
	}
}

func TestRenderMarkdown_ListItems(t *testing.T) {
	input := "- item one\n- item two\n- item three"
	result := renderMarkdown(input)
	if !strings.Contains(result, "item one") {
		t.Error("list item one missing")
	}
	if !strings.Contains(result, "item two") {
		t.Error("list item two missing")
	}
}

func TestRenderMarkdown_BlockQuotes(t *testing.T) {
	input := "> quoted text"
	result := renderMarkdown(input)
	if !strings.Contains(result, "quoted text") {
		t.Error("blockquote content missing")
	}
}

func TestRenderMarkdown_BlankLines(t *testing.T) {
	input := "line one\n\nline three"
	result := renderMarkdown(input)
	if !strings.Contains(result, "line one") {
		t.Error("first line missing")
	}
	if !strings.Contains(result, "line three") {
		t.Error("third line missing")
	}
	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) < 3 {
		t.Errorf("expected at least 3 lines (blank preserved), got %d", len(lines))
	}
}

func TestRenderMarkdown_Empty(t *testing.T) {
	result := renderMarkdown("")
	if result != "" {
		t.Errorf("expected empty result, got %q", result)
	}
}

func TestRenderMarkdown_MixedContent(t *testing.T) {
	input := `# Getting Started

This is **bold** and *italic* text.

` + "```go" + `
func main() {}
` + "```" + `

- First item
- Second item

> A quote

Use ` + "`inline`" + ` code.`

	result := renderMarkdown(input)

	checks := []string{
		"Getting Started",
		"bold",
		"italic",
		"func main()",
		"First item",
		"Second item",
		"A quote",
		"inline",
	}
	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("missing content: %q", check)
		}
	}
}

func TestRenderInlineMarkdown_Bold(t *testing.T) {
	input := "hello **world** there"
	result := renderInlineMarkdown(input)
	if !strings.Contains(result, "world") {
		t.Error("bold content missing")
	}
}

func TestRenderInlineMarkdown_Italic(t *testing.T) {
	input := "hello *world* there"
	result := renderInlineMarkdown(input)
	if !strings.Contains(result, "world") {
		t.Error("italic content missing")
	}
}

func TestRenderInlineMarkdown_Code(t *testing.T) {
	input := "use `code` here"
	result := renderInlineMarkdown(input)
	if !strings.Contains(result, "code") {
		t.Error("inline code missing")
	}
}

func TestRenderInlineMarkdown_NoFormatting(t *testing.T) {
	input := "plain text without formatting"
	result := renderInlineMarkdown(input)
	if !strings.Contains(result, "plain text") {
		t.Error("plain text missing")
	}
}

func TestRenderInlineMarkdown_Empty(t *testing.T) {
	result := renderInlineMarkdown("")
	if result != "" {
		t.Errorf("expected empty result, got %q", result)
	}
}
