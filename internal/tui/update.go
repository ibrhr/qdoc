package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/ibrhr/qdoc/internal/config"
	"github.com/ibrhr/qdoc/internal/docsource"
	"github.com/ibrhr/qdoc/internal/llm"
	"github.com/ibrhr/qdoc/internal/provider"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.mode != modeQuery || m.setupStep != stepNone {
		return m.updateSetup(msg)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case progressTickMsg:
		m.spinnerIdx = msg.tick / 100
		if m.phase != phaseDone && m.phase != phaseError {
			return m, tickCmd()
		}
		return m, nil

	case docIndexMsg:
		return m.handleDocIndex(msg)

	case llm.StreamDelta:
		return m.handleStreamDelta(msg)

	case docContentMsg:
		return m.handleDocContent(msg)

	case docErrorMsg:
		m.err = msg.err
		m.phase = phaseError
		m.qState = qStateError
		m.addStep("Error", msg.err.Error())
		m.addDisplay(dlError, msg.err.Error())
		
		if msg.log != nil {
			msg.log.Log("ERROR: %v", msg.err)
			msg.log.Close()
		}
		if m.sessionLog != nil {
			m.sessionLog.Log("ERROR: %v", msg.err)
			m.sessionLog.Close()
			m.sessionLog = nil
		}
		return m, nil

	case queryCompleteMsg:
		m.answer = msg.answer
		m.steps = msg.steps
		m.phase = phaseDone
		m.qState = qStateDone
		rendered := renderMarkdown(msg.answer)
		if len(m.displayLines) > 0 {
			last := m.displayLines[len(m.displayLines)-1]
			if last.Type != dlDivider {
				m.addDisplay(dlDivider, "Answer")
			}
		}
		for _, line := range strings.Split(rendered, "\n") {
			m.addDisplay(dlAnswer, line)
		}
		m.autoScroll = true
		
		if m.sessionLog != nil {
			m.sessionLog.Raw("ANSWER", msg.answer)
			m.sessionLog.Log("Steps: %d total", len(msg.steps))
			for i, s := range msg.steps {
				m.sessionLog.Log("  Step %d: %s — %s", i+1, s.Phase, s.Detail)
			}
			m.sessionLog.Close()
			m.sessionLog = nil
		}

		return m, nil

	case setupCompleteMsg:
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "qdoc: reloading config: %v\n", err)
		} else {
			m.Cfg = cfg
		}
		m.phase = phaseInit
		m.qState = qStateInit
		m.displayLines = nil
		
		
		m.streamBuf = m.streamBuf[:0]
		m.reasoningBuf = m.reasoningBuf[:0]
		m.streamLive = m.streamLive[:0]
		m.reasoningLive = m.reasoningLive[:0]
		m.streaming = false
		m.readFiles = make(map[string]string)
		m.iteration = 0
		m.steps = make([]queryStep, 0)
		return m, tea.Batch(
			tickCmd(),
			startQuery(m.Source, m.Query),
		)

	case tea.KeyPressMsg:
		return m.handleQueryKey(msg)
	}

	return m, nil
}

func (m *Model) handleQueryKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc":
		if m.phase == phaseDone || m.phase == phaseError {
			return m, tea.Quit
		}
	case "up", "k":
		m.scrollOffset--
		if m.scrollOffset < 0 {
			m.scrollOffset = 0
		}
		m.autoScroll = false
	case "down", "j":
		m.scrollOffset++
		m.autoScroll = false
	case "pgup", "home":
		m.scrollOffset = 0
		m.autoScroll = false
	case "pgdown", "end":
		m.scrollOffset = len(m.displayLines)
		m.autoScroll = false
	}
	return m, nil
}

func (m *Model) handleDocIndex(msg docIndexMsg) (tea.Model, tea.Cmd) {
	m.entries = msg.entries
	m.sessionLog = msg.log
	m.phase = phaseQuerying
	m.qState = qStateIndexFetched
	m.addStep("Fetched index", fmt.Sprintf("%d pages", len(msg.entries)))

	client, err := m.resolveClient()
	if err != nil {
		m.mode = modeQuery
		m.setupStep = stepProvider
		m.cursor = 0
		m.inputBuffer = ""
		m.err = nil
		return m, tickCmd()
	}
	m.client = client

	m.addStep("Calling", fmt.Sprintf("%s via %s", client.ModelName(), client.ProviderName()))
	m.addDisplay(dlStep, fmt.Sprintf("● Calling %s on %s", client.ModelName(), client.ProviderName()))

	systemPrompt := llm.BuildSystemPrompt(m.Source, msg.entries, m.Query)
	m.llmMessages = []llm.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: m.Query},
	}

	if m.sessionLog != nil {
		m.sessionLog.Log("Provider: %s | Model: %s", client.ProviderName(), client.ModelName())
		m.sessionLog.Raw("SYSTEM PROMPT", systemPrompt)
		m.sessionLog.Log("User message: %s", m.Query)
	}

	return m, m.startStream()
}

func (m *Model) handleStreamDelta(msg llm.StreamDelta) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		m.err = msg.Err
		m.phase = phaseError
		m.qState = qStateError
		m.addDisplay(dlError, msg.Err.Error())
		
		return m, nil
	}

	if msg.Retrying {
		m.addDisplay(dlToolCall, "↻ "+msg.Content)
		
		return m, readStreamChunk(m.streamCh)
	}

	if msg.Done {
		return m.handleStreamDone()
	}

	if msg.Reasoning != "" {
		m.reasoningBuf = append(m.reasoningBuf, msg.Reasoning...)
		m.reasoningLive = append(m.reasoningLive, msg.Reasoning...)
		m.streaming = true
		
		return m, readStreamChunk(m.streamCh)
	}

	m.streamBuf = append(m.streamBuf, msg.Content...)
	m.streamLive = append(m.streamLive, msg.Content...)
	m.streaming = true
	
	return m, readStreamChunk(m.streamCh)
}

func (m *Model) handleStreamDone() (tea.Model, tea.Cmd) {
	m.streaming = false
	m.qState = qStateParsing
	m.streamLive = m.streamLive[:0]
	m.reasoningLive = m.reasoningLive[:0]

	// Display reasoning tokens as thought
	if len(m.reasoningBuf) > 0 {
		for _, line := range strings.Split(strings.TrimSpace(string(m.reasoningBuf)), "\n") {
			if strings.TrimSpace(line) != "" {
				m.addDisplay(dlThought, line)
			}
		}
	}
	m.reasoningBuf = m.reasoningBuf[:0]

	full := string(m.streamBuf)

	if m.sessionLog != nil {
		m.sessionLog.Raw("RAW LLM RESPONSE", full)
	}
	actions := llm.ParseResponses(full)
	var urls []string
	hasAnswer := false

	for _, a := range actions {
		if a.Action == llm.ActionReadFile {
			urls = append(urls, a.URL)
		} else if a.Action == llm.ActionAnswer {
			hasAnswer = true
			m.gotAnswer = true
			m.answer = a.Content
		}
	}

	isJSONLine := make(map[int]bool)
	for i, line := range strings.Split(full, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") {
			var a struct {
				Action string `json:"action"`
			}
			if json.Unmarshal([]byte(trimmed), &a) == nil && (a.Action == "read_file" || a.Action == "answer") {
				isJSONLine[i] = true
			}
		}
	}

	answerLineSeen := false
	for i, line := range strings.Split(full, "\n") {
		if isJSONLine[i] {
			trimmed := strings.TrimSpace(line)
			if strings.Contains(trimmed, `"answer"`) {
				answerLineSeen = true
				m.addDisplay(dlDivider, "Answer")
				m.addDisplay(dlAction, "◆ ANSWER")
			} else if strings.Contains(trimmed, `"read_file"`) {
				var a struct {
					Action string `json:"action"`
					URL    string `json:"url"`
				}
				if json.Unmarshal([]byte(trimmed), &a) == nil {
					m.addDisplay(dlAction, "▶ READ  "+a.URL)
				}
			}
		} else if !answerLineSeen && strings.TrimSpace(line) != "" {
			m.addDisplay(dlThought, line)
		}
	}

	m.autoScroll = true
	m.streamBuf = m.streamBuf[:0]
	m.streamCh = nil

	if hasAnswer {
		if len(urls) > 0 {
			m.pendingReads = urls
			m.qState = qStateAnswering
			m.phase = phaseAnswering
			m.addDisplay(dlToolCall, "● Preparing answer...")
			m.addStep("Answer ready", "")

			return m, m.startFileFetches()
		}
		return m, func() tea.Msg {
			return queryCompleteMsg{
				answer: m.answer,
				source: m.Source.Name,
				steps:  m.steps,
			}
		}
	} else if len(urls) > 0 {
		m.phase = phaseReadingFile
		m.iteration++
		m.qState = qStateFetchingFiles
		m.pendingReads = urls
		m.addDisplay(dlToolCall, fmt.Sprintf("● Reading %d files...", len(urls)))

		return m, m.startFileFetches()
	}

	return m, func() tea.Msg {
		return queryCompleteMsg{
			answer: "(The model did not return a READ_FILE or ANSWER command.)",
			source: m.Source.Name,
			steps:  m.steps,
		}
	}
}

func (m *Model) startFileFetches() tea.Cmd {
	m.filesPending = len(m.pendingReads)
	cmds := make([]tea.Cmd, 0, len(m.pendingReads))
	for _, u := range m.pendingReads {
		m.addStep("Reading", u)
		cmds = append(cmds, fetchDocContent(m.Source, u))
	}
	return tea.Batch(cmds...)
}

func (m *Model) handleDocContent(msg docContentMsg) (tea.Model, tea.Cmd) {
	m.readFiles[msg.url] = msg.content
	m.autoScroll = true
	m.filesPending--

	if m.sessionLog != nil {
		m.sessionLog.Log("Fetched %s (%d chars)", msg.url, len(msg.content))
	}

	// Feed content to LLM only — never display fetched file contents to user
	contentMsg := fmt.Sprintf("FILE: %s\n\n%s", msg.url, msg.content)
	m.llmMessages = append(m.llmMessages, llm.ChatMessage{
		Role:    "user",
		Content: contentMsg,
	})

	if m.sessionLog != nil {
		m.sessionLog.Raw("FEED TO LLM", contentMsg)
	}

	return m.remainingFetching()
}

func (m *Model) remainingFetching() (tea.Model, tea.Cmd) {
	if m.filesPending > 0 {
		return m, nil
	}

	m.qState = qStateFilesReady

	if m.gotAnswer || m.iteration >= m.maxIters {
		if m.gotAnswer {
			m.phase = phaseAnswering
			return m, func() tea.Msg {
				return queryCompleteMsg{
					answer: m.answer,
					source: m.Source.Name,
					steps:  m.steps,
				}
			}
		}
		m.llmMessages = append(m.llmMessages, llm.ChatMessage{
			Role:    "user",
			Content: "Maximum iterations reached. Output {\"action\":\"answer\"} on its own line, then your answer below.",
		})
	}

	m.streamBuf = m.streamBuf[:0]
	m.reasoningBuf = m.reasoningBuf[:0]
	m.streamLive = m.streamLive[:0]
	m.reasoningLive = m.reasoningLive[:0]
	m.streamCh = make(chan llm.StreamDelta, 256)
	m.streaming = true
	m.qState = qStateStreaming
	m.autoScroll = true

	if m.sessionLog != nil {
		m.sessionLog.Log("Starting stream (iteration %d/%d)", m.iteration, m.maxIters)
	}

	go m.client.Stream(m.llmMessages, m.streamCh)
	return m, readStreamChunk(m.streamCh)
}

func (m *Model) startStream() tea.Cmd {
	m.streamBuf = m.streamBuf[:0]
	m.reasoningBuf = m.reasoningBuf[:0]
	m.streamLive = m.streamLive[:0]
	m.reasoningLive = m.reasoningLive[:0]
	m.streamCh = make(chan llm.StreamDelta, 256)
	m.streaming = true
	m.qState = qStateStreaming
	m.autoScroll = true

	go m.client.Stream(m.llmMessages, m.streamCh)
	return readStreamChunk(m.streamCh)
}

func readStreamChunk(ch <-chan llm.StreamDelta) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return llm.StreamDelta{Done: true}
		}
		return msg
	}
}

func fetchDocContent(source docsource.Source, rawURL string) tea.Cmd {
	return func() tea.Msg {
		content, err := source.FetchContent(rawURL)
		if err != nil {
			return docErrorMsg{err: fmt.Errorf("fetching doc %s: %w", rawURL, err)}
		}
		return docContentMsg{
			url:     rawURL,
			content: content,
		}
	}
}

func (m *Model) addDisplay(typ displayLineType, content string) {
	m.displayLines = append(m.displayLines, displayLine{Type: typ, Content: content})
}

func (m *Model) addStep(phase, detail string) {
	m.steps = append(m.steps, queryStep{
		Step:   len(m.steps) + 1,
		Phase:  phase,
		Detail: detail,
	})
}

func (m *Model) resolveClient() (llm.Client, error) {
	if m.Cfg.AccessMethod != "" {
		return provider.ResolveClientWithMethod(m.Cfg, m.Cfg.AccessMethod)
	}
	return provider.ResolveClient(m.Cfg)
}