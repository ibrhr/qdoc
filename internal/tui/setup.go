package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/ibrhr/qdoc/internal/config"
	"github.com/ibrhr/qdoc/internal/provider"
)

func (m *Model) updateSetup(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

	case progressTickMsg:
		m.spinnerIdx = msg.tick / 100
		return m, tickCmd()

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "q":
			if m.mode != modeQuery {
				return m, tea.Quit
			}
		}

		switch m.setupStep {
		case stepProvider:
			return m.handleProviderSelect(msg)
		case stepKey:
			return m.handleKeyInput(msg)
		case stepModel:
			return m.handleModelSelect(msg)
		}

	case tea.PasteMsg:
		if m.setupStep == stepKey {
			return m.handleKeyPaste(msg)
		}
	}

	return m, nil
}

func (m *Model) handleProviderSelect(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	m.err = nil

	switch msg.String() {
	case "up", "k":
		m.cursor--
		if m.cursor < 0 {
			m.cursor = len(m.providersList) - 1
		}
	case "down", "j":
		m.cursor++
		if m.cursor >= len(m.providersList) {
			m.cursor = 0
		}
	case "enter":
		m.SelectedProvider = m.providersList[m.cursor].Name
		m.cursor = 0
		m.inputBuffer = ""

		if provider.KeyExists(m.Cfg, m.SelectedProvider) {
			prov, _ := provider.Find(m.SelectedProvider)
			m.modelsList = provider.BuildModelList(prov)
			m.setupStep = stepModel
			return m, nil
		}

		m.setupStep = stepKey
		return m, nil
	}
	return m, nil
}

func (m *Model) handleKeyInput(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	m.err = nil

	switch msg.String() {
	case "up", "down", "left", "right":
		return m, nil
	case "enter":
		if m.inputBuffer == "" {
			m.err = fmt.Errorf("no key provided")
			return m, nil
		}
		m.Cfg.Keys[m.SelectedProvider] = m.inputBuffer
		if err := config.Save(m.Cfg); err != nil {
			m.err = err
			m.setupStep = stepNone
			return m, nil
		}

		m.setupStep = stepModel
		m.cursor = 0
		prov, _ := provider.Find(m.SelectedProvider)
		m.modelsList = provider.BuildModelList(prov)
		return m, nil

	case "backspace":
		if len(m.inputBuffer) > 0 {
			m.inputBuffer = m.inputBuffer[:len(m.inputBuffer)-1]
		}
	default:
		if len(msg.Text) > 0 {
			m.inputBuffer += msg.Text
		}
	}
	return m, nil
}

func (m *Model) handleKeyPaste(msg tea.PasteMsg) (tea.Model, tea.Cmd) {
	m.err = nil
	m.inputBuffer = msg.Content
	return m, nil
}

func (m *Model) handleModelSelect(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	m.err = nil

	switch msg.String() {
	case "up", "k":
		if len(m.modelsList) > 0 {
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.modelsList) - 1
			}
		}
	case "down", "j":
		if len(m.modelsList) > 0 {
			m.cursor++
			if m.cursor >= len(m.modelsList) {
				m.cursor = 0
			}
		}
	case "enter":
		if m.cursor > 0 && m.cursor < len(m.modelsList) {
			if m.Cfg.Models == nil {
				m.Cfg.Models = make(map[string]string)
			}
			m.Cfg.Models[m.SelectedProvider] = m.modelsList[m.cursor]
		}
		m.Cfg.Provider = m.SelectedProvider
		config.Save(m.Cfg)

		if m.mode == modeModel {
			return m, tea.Quit
		}

		m.setupStep = stepNone
		m.mode = modeQuery
		if m.Query == "" {
			return m, tea.Quit
		}
		return m, func() tea.Msg {
			return setupCompleteMsg{}
		}

	case "backspace", "q":
		return m, nil
	case "c":
		m.inputBuffer = ""
		m.setupStep = stepKey
		return m, nil
	default:
		return m, nil
	}
	return m, nil
}

func (m *Model) renderSetup() string {
	var sb strings.Builder

	if m.mode == modeQuery && m.setupStep != stepNone && m.Query != "" {
		sb.WriteString(dimStyle.Render("  ⏸ Query paused — let's set up your provider first"))
		sb.WriteString("\n\n")
	}

	if m.err != nil {
		sb.WriteString(errorBox.Render("Error: " + m.err.Error()))
		sb.WriteString("\n\n")
	}

	switch m.setupStep {
	case stepProvider:
		return sb.String() + m.renderProviderSelect() + "\n"
	case stepKey:
		return sb.String() + m.renderKeyInput() + "\n"
	case stepModel:
		return sb.String() + m.renderModelSelect() + "\n"
	default:
		return sb.String() + m.renderStandaloneProvider() + "\n"
	}
}

func (m Model) renderStandaloneProvider() string {
	var sb strings.Builder
	sb.WriteString(renderSetupHeader("Select Provider"))
	sb.WriteString("\n")
	sb.WriteString(renderProviderList(m.providersList, m.cursor, m.Cfg))
	sb.WriteString("\n\n")
	sb.WriteString(renderSetupFooter())
	return sb.String()
}

func (m Model) renderProviderSelect() string {
	var sb strings.Builder
	sb.WriteString(renderSetupHeader("Setup — Choose Provider"))
	sb.WriteString("\n")
	sb.WriteString(dimStyle.Render("  No API key configured. Pick a provider to get started."))
	sb.WriteString("\n\n")
	sb.WriteString(renderProviderList(m.providersList, m.cursor, m.Cfg))
	sb.WriteString("\n\n")
	sb.WriteString(renderSetupFooter())
	return sb.String()
}

func (m Model) renderKeyInput() string {
	prov, _ := provider.Find(m.SelectedProvider)

	var sb strings.Builder
	sb.WriteString(renderSetupHeader("Setup — API Key"))
	sb.WriteString("\n")
	sb.WriteString(dimStyle.Render(fmt.Sprintf("  Provider: %s", prov.Name)))
	sb.WriteString("\n")
	sb.WriteString(dimStyle.Render(fmt.Sprintf("  Env var: %s", prov.EnvKey)))
	sb.WriteString("\n\n")

	display := m.inputBuffer
	if display == "" {
		display = dimStyle.Render("paste or type your key...")
	}

	sb.WriteString("  ")
	sb.WriteString(accentStyle.Render("API KEY > "))
	sb.WriteString(display)
	sb.WriteString("\n\n")
	sb.WriteString(dimStyle.Render("  Press enter to save and continue."))
	sb.WriteString("\n")
	return sb.String()
}

func (m *Model) renderModelSelect() string {
	var sb strings.Builder
	sb.WriteString(renderSetupHeader("Setup — Choose Model"))
	sb.WriteString("\n")
	sb.WriteString(dimStyle.Render(fmt.Sprintf("  Provider: %s", m.SelectedProvider)))
	sb.WriteString("\n")
	sb.WriteString(dimStyle.Render("  Press enter on a model to select it, or on the default to skip."))
	sb.WriteString("\n\n")

	fixedLines := 7 // header + blank + provider + instruction + blank + blank + footer
	if m.err != nil {
		fixedLines += 2 // error line + blank
	}
	if m.mode == modeQuery && m.setupStep != stepNone && m.Query != "" {
		fixedLines += 2 // paused message + blank
	}

	visible := m.Height - fixedLines
	if visible < 1 {
		visible = 1
	}
	total := len(m.modelsList)

	if m.cursor < m.setupScrollOffset {
		m.setupScrollOffset = m.cursor
	}
	if m.cursor >= m.setupScrollOffset+visible {
		m.setupScrollOffset = m.cursor - visible + 1
	}
	if m.setupScrollOffset < 0 {
		m.setupScrollOffset = 0
	}
	if maxOff := total - visible; m.setupScrollOffset > maxOff {
		if maxOff < 0 {
			m.setupScrollOffset = 0
		} else {
			m.setupScrollOffset = maxOff
		}
	}

	if m.setupScrollOffset > 0 {
		sb.WriteString(dimStyle.Render(fmt.Sprintf("  ↑ %d more above", m.setupScrollOffset)))
		sb.WriteString("\n")
	}

	end := m.setupScrollOffset + visible
	if end > total {
		end = total
	}

	for i := m.setupScrollOffset; i < end; i++ {
		modelName := m.modelsList[i]
		cursorMark := "  "
		if i == m.cursor {
			cursorMark = cursorStyle.Render(">")
		}

		line := fmt.Sprintf("%s %s", cursorMark, modelName)
		if i == m.cursor {
			sb.WriteString(cursorStyle.Render(line))
		} else if i == 0 {
			sb.WriteString(dimStyle.Render(line))
		} else {
			sb.WriteString(dimStyle.Render(line))
		}
		sb.WriteString("\n")
	}

	if end < total {
		sb.WriteString(dimStyle.Render(fmt.Sprintf("  ↓ %d more below", total-end)))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(dimStyle.Render("  up/down or j/k navigate  ─  enter select  ─  c change API key  ─  esc quit"))
	return sb.String()
}

func renderSetupHeader(title string) string {
	return headerBox.Width(72).Render(
		titleStyle.Render(" qdoc ") + "  " + sectionStyle.Render(title),
	)
}

func renderProviderList(provList []provider.Provider, cursor int, cfg config.Config) string {
	var sb strings.Builder
	for i, p := range provList {
		cursorMark := "  "
		if i == cursor {
			cursorMark = cursorStyle.Render(">")
		}

		status := dimStyle.Render(" (no key)")
		if cfg.Keys[p.Name] != "" {
			status = successStyle.Render(" (key set)")
		}

		activeMark := ""
		if cfg.Provider == p.Name {
			activeMark = sectionStyle.Render(" <- active")
		}

		line := fmt.Sprintf("%s %s%s%s", cursorMark, p.Name, status, activeMark)
		if i == cursor {
			sb.WriteString(cursorStyle.Render(line))
		} else {
			sb.WriteString(dimStyle.Render(line))
		}
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("     %s\n", dimStyle.Render(p.Description)))
	}
	return sb.String()
}

func renderSetupFooter() string {
	return dimStyle.Render("  up/down or j/k navigate  ─  enter select  ─  esc quit")
}