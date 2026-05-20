package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m *Model) View() tea.View {
	if m.mode != modeQuery || m.setupStep != stepNone {
		return tea.NewView(m.renderSetup())
	}

	var sb strings.Builder

	sb.WriteString(renderHeader(*m))
	sb.WriteString("\n")

	sb.WriteString(renderProgress(*m))
	sb.WriteString("\n")

	sb.WriteString(renderDivider(*m))
	sb.WriteString("\n")

	sb.WriteString(renderContentArea(m))
	sb.WriteString("\n")

	sb.WriteString(renderFooter(*m))

	v := tea.NewView(sb.String())
	v.AltScreen = true
	return v
}

func renderHeader(m Model) string {
	return titleStyle.Render(" qdoc ") + " " + dimStyle.Render("Query the Docs")
}

func renderDivider(m Model) string {
	line := strings.Repeat("-", m.Width)
	return separatorStyle.Render(line)
}

func renderProgress(m Model) string {
	var sb strings.Builder

	phaseColor := info
	switch m.phase {
	case phaseDone:
		phaseColor = success
	case phaseError:
		phaseColor = errC
	}

	spinner := spinnerFrame(dotFrames, m.spinnerIdx)

	dot := baseStyle.Foreground(phaseColor).Render("*")
	sb.WriteString(fmt.Sprintf("%s %s %s",
		dot,
		sectionStyle.Render(m.phase.String()),
		dimStyle.Render(spinner),
	))

	if m.Query != "" {
		sb.WriteString("  ")
		sb.WriteString(queryStyle.Render("query:"))
		sb.WriteString(" ")
		sb.WriteString(dimStyle.Render(m.Query))
	}

	if m.iteration > 0 {
		sb.WriteString(fmt.Sprintf("  %s %d/%d",
			dimStyle.Render("files:"),
			m.iteration,
			m.maxIters,
		))
	}

	if m.client != nil && m.streaming {
		sb.WriteString("  ")
		rLen := len(m.reasoningLive)
		cLen := len(m.streamLive)
		if rLen > 0 || cLen > 0 {
			sb.WriteString(toolCallStyle.Render(fmt.Sprintf("[R:%d C:%d]", rLen, cLen)))
		} else {
			sb.WriteString(toolCallStyle.Render("..."))
		}
	}

	return sb.String()
}

func renderContentArea(m *Model) string {
	contentH := m.Height - 5
	if contentH < 1 {
		contentH = 1
	}

	var lines []string

	for _, dl := range m.displayLines {
			lines = append(lines, renderDisplayLine(dl, m.Width))
	}

	if m.streaming && len(m.reasoningLive) > 0 {
		preview := string(m.reasoningLive)
		for _, line := range strings.Split(preview, "\n") {
			lines = append(lines, thoughtStyle.Render("  "+line))
		}
	}
	if m.streaming && len(m.streamLive) > 0 {
		preview := string(m.streamLive)
		for _, line := range strings.Split(preview, "\n") {
			lines = append(lines, thoughtStyle.Render("  "+line))
		}
	}

	if len(lines) == 0 {
		return dimStyle.Render("  Waiting...")
	}

	// Word-wrap lines that exceed terminal width (ANSI-aware)
	maxW := m.Width
	wrapped := make([]string, 0, len(lines)*2)
	for _, line := range lines {
		if lipgloss.Width(line) <= maxW {
			wrapped = append(wrapped, line)
		} else {
			parts := strings.Split(lipgloss.Wrap(line, maxW, " "), "\n")
			wrapped = append(wrapped, parts[0])
			for _, wl := range parts[1:] {
				wrapped = append(wrapped, "  "+wl)
			}
		}
	}
	lines = wrapped

	if m.autoScroll {
		m.scrollOffset = len(lines) - contentH
		if m.scrollOffset < 0 {
			m.scrollOffset = 0
		}
	}

	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
	maxOffset := len(lines) - contentH
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.scrollOffset > maxOffset {
		m.scrollOffset = maxOffset
	}

	end := m.scrollOffset + contentH
	if end > len(lines) {
		end = len(lines)
	}

	var sb strings.Builder
	for _, line := range lines[m.scrollOffset:end] {
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	shown := end - m.scrollOffset
	for i := shown; i < contentH; i++ {
		sb.WriteString("\n")
	}

	return sb.String()
}

func renderDisplayLine(dl displayLine, width int) string {
	switch dl.Type {
	case dlStep, dlToolCall:
		return toolCallStyle.Render(fmt.Sprintf("  %s", dl.Content))

	case dlAction:
		return actionStyle.Render(fmt.Sprintf("  %s", dl.Content))

	case dlThought:
		return thoughtStyle.Render(fmt.Sprintf("  %s", dl.Content))

	case dlFilePath:
		return successStyle.Render(fmt.Sprintf("  %s", dl.Content))

	case dlFileContent:
		return dimStyle.Render(fmt.Sprintf("  | %s", dl.Content))

	case dlAnswer:
		return answerStyle.Render(fmt.Sprintf("  %s", dl.Content))

	case dlDivider:
		return renderSectionDivider(dl.Content, width)

	case dlError:
		return errorStyle.Render(fmt.Sprintf("  %s", dl.Content))

	default:
		return fmt.Sprintf("  %s", dl.Content)
	}
}

func renderSectionDivider(label string, width int) string {
	const minLabelWidth = 36
	labelText := fmt.Sprintf(" ✦ %s ✦ ", label)
	labelLen := lipgloss.Width(labelText)

	totalWidth := width - 4
	if totalWidth < minLabelWidth {
		totalWidth = minLabelWidth
	}

	sideLen := (totalWidth - labelLen) / 2
	if sideLen < 2 {
		sideLen = 2
	}
	leftSide := strings.Repeat("─", sideLen)
	rightSideLen := totalWidth - labelLen - sideLen
	if rightSideLen < 2 {
		rightSideLen = 2
	}
	rightSide := strings.Repeat("─", rightSideLen)

	dividerLine := fmt.Sprintf("  %s%s%s",
		dividerSectionStyle.Render(leftSide),
		dividerLabelStyle.Render(labelText),
		dividerSectionStyle.Render(rightSide),
	)

	return dividerLine
}

func renderFooter(m Model) string {
	contentH := m.Height - 5
	if contentH < 1 {
		contentH = 1
	}

	var parts []string
	parts = append(parts, "q to quit")

	if m.scrollOffset > 0 {
		parts = append(parts, "↑")
	}
	lineCount := len(m.displayLines)
	if m.streaming {
		if len(m.reasoningLive) > 0 {
			lineCount += len(strings.Split(string(m.reasoningLive), "\n"))
		}
		if len(m.streamLive) > 0 {
			lineCount += len(strings.Split(string(m.streamLive), "\n"))
		}
	}
	if lineCount > 0 && m.scrollOffset+contentH < lineCount {
		parts = append(parts, "↓")
	}

	if m.phase == phaseDone || m.phase == phaseError {
		return footerStyle.Render("  " + strings.Join(parts, " ─ "))
	}

	prov := "?"
	if m.client != nil {
		prov = m.client.Provider
	} else {
		prov = m.Cfg.Provider
	}
	parts = append(parts, m.Source.Name, prov)
	return footerStyle.Render("  " + strings.Join(parts, " ─ "))
}