package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

var (
	mdH1Style = lipgloss.NewStyle().
			Foreground(fg).
			Bold(true).
			Padding(1, 0)

	mdH2Style = lipgloss.NewStyle().
			Foreground(accent).
			Bold(true).
			Padding(0, 1)

	mdH3Style = lipgloss.NewStyle().
			Foreground(accent)

	mdCodeBlockStyle = lipgloss.NewStyle().
				Padding(0, 1)

	mdInlineCodeStyle = lipgloss.NewStyle().
				Foreground(info)

	mdBoldStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(fg)

	mdItalicStyle = lipgloss.NewStyle().
			Italic(true).
			Foreground(fg)

	mdListStyle = lipgloss.NewStyle().
			Foreground(dim)

	mdQuoteStyle = lipgloss.NewStyle().
			Foreground(dim).
			Italic(true)
)

func renderMarkdown(md string) string {
	var sb strings.Builder
	lines := strings.Split(md, "\n")
	inCodeBlock := false
	var codeLang string

	for i, line := range lines {
		if i > 0 {
			sb.WriteString("\n")
		}

		if inCodeBlock {
			if strings.HasPrefix(strings.TrimSpace(line), "```") {
				inCodeBlock = false
				continue
			}
			sb.WriteString(mdCodeBlockStyle.Foreground(dim).Render("  " + strings.TrimRight(line, " \t")))
			continue
		}

		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = true
			codeLang = strings.TrimPrefix(trimmed, "```")
			if codeLang != "" {
				sb.WriteString(mdCodeBlockStyle.Foreground(success).Render(fmt.Sprintf("  ┌─ %s", codeLang)))
			}
			continue
		}

		if trimmed == "" {
			continue
		}

		if strings.HasPrefix(trimmed, "### ") {
			sb.WriteString(mdH3Style.Render("  " + trimmed[4:]))
		} else if strings.HasPrefix(trimmed, "## ") {
			sb.WriteString(mdH2Style.Render("  " + trimmed[3:]))
		} else if strings.HasPrefix(trimmed, "# ") {
			sb.WriteString(mdH1Style.Render("  " + trimmed[2:]))
		} else if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "+ ") {
			content := trimmed[2:]
			sb.WriteString(mdListStyle.Render("● " + renderInlineMarkdown(content)))
		} else if len(trimmed) > 2 && trimmed[0] >= '0' && trimmed[0] <= '9' && trimmed[1] == '.' {
			sb.WriteString(mdListStyle.Render("  " + trimmed))
		} else if strings.HasPrefix(trimmed, "> ") {
			content := strings.TrimPrefix(trimmed, "> ")
			sb.WriteString(mdQuoteStyle.Render("│ " + content))
		} else {
			sb.WriteString("  " + renderInlineMarkdown(trimmed))
		}
	}

	return sb.String()
}

func renderInlineMarkdown(text string) string {
	result := text

	var sb strings.Builder
	i := 0
	for i < len(result) {
		// Bold: **text**
		if i+1 < len(result) && result[i] == '*' && result[i+1] == '*' {
			end := strings.Index(result[i+2:], "**")
			if end >= 0 {
				end += i + 2
				content := result[i+2 : end]
				sb.WriteString(mdBoldStyle.Render(content))
				i = end + 2
				continue
			}
		}
		// Italic: *text* (must check after bold to avoid ambiguity)
		if result[i] == '*' {
			end := strings.IndexRune(result[i+1:], '*')
			if end >= 0 {
				end += i + 1
				content := result[i+1 : end]
				sb.WriteString(mdItalicStyle.Render(content))
				i = end + 1
				continue
			}
		}
		// Inline code: `text`
		if result[i] == '`' {
			end := strings.IndexRune(result[i+1:], '`')
			if end >= 0 {
				end += i + 1
				sb.WriteString(mdInlineCodeStyle.Render(result[i : end+1]))
				i = end + 1
				continue
			}
		}
		sb.WriteByte(result[i])
		i++
	}

	return sb.String()
}
