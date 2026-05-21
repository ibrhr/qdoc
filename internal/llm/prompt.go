package llm

import (
	"fmt"
	"strings"

	"github.com/ibrhr/qdoc/internal/docsource"
)

const maxIndexEntries = 120

func BuildSystemPrompt(source docsource.Source, entries []docsource.Entry, query string) string {
	totalEntries := len(entries)
	maxEntries := maxIndexEntries
	if len(entries) > maxEntries {
		entries = entries[:maxEntries]
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("You are qdoc, an expert at navigating %s documentation. You answer exactly one question per session. This is not a conversation — there is no back-and-forth with the user. Give a definitive, self-contained answer. Do not ask follow-up questions, suggest the user try something else, or invite further discussion.\n\n", source.Name))

	if source.SystemPrompt != "" {
		sb.WriteString(source.SystemPrompt)
		sb.WriteString("\n\n")
	}

	sb.WriteString("HOW IT WORKS\n")
	sb.WriteString("You work in turns. Each turn, output one or more JSON actions (one per line):\n")
	sb.WriteString(`  {"action":"read_file","url":"https://..."}   Read a documentation page` + "\n")
	sb.WriteString(`  {"action":"answer"}                           Deliver your final answer` + "\n")
	sb.WriteString("After the answer line, write your response in markdown.\n\n")

	sb.WriteString("After each turn where you read files, those files' contents are provided back to you. You can then read more files or answer.\n\n")

	sb.WriteString("Research strategy:\n")
	sb.WriteString("- Read multiple files in one turn to fetch them in parallel.\n")
	sb.WriteString("- Spread reads across turns to drill deeper: read a page, see what it references, then read those next.\n")
	sb.WriteString("- Be thorough — read enough to give an informed answer before using {\"action\":\"answer\"}.\n")
	sb.WriteString("- Use code examples and specifics from the documentation in your answer.\n\n")

	sb.WriteString(fmt.Sprintf("BASE URL: %s\n\n", source.BaseURL))
	sb.WriteString("AVAILABLE FILES:\n")

	for _, e := range entries {
		sb.WriteString(fmt.Sprintf("  %s  (%s)\n", e.URL, e.Title))
	}

	if totalEntries > maxEntries {
		sb.WriteString(fmt.Sprintf("  ... (%d total, showing first %d)\n", totalEntries, maxEntries))
	}

	sb.WriteString(fmt.Sprintf("\nQUERY: %s\n", query))

	return sb.String()
}