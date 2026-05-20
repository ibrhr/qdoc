package tui

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/ibrhr/qdoc/internal/config"
	"github.com/ibrhr/qdoc/internal/docsource"
	"github.com/ibrhr/qdoc/internal/llm"
	"github.com/ibrhr/qdoc/internal/provider"
	"github.com/ibrhr/qdoc/internal/sessionlog"
)

type modeType string

const (
	modeQuery    modeType = "query"
	modeProvider modeType = "provider"
	modeModel    modeType = "model"
)

type setupStep string

const (
	stepNone     setupStep = ""
	stepProvider setupStep = "provider"
	stepKey      setupStep = "key"
	stepModel    setupStep = "model"
)

type phase int

const (
	phaseInit phase = iota
	phaseFetchingIndex
	phaseQuerying
	phaseReadingFile
	phaseAnswering
	phaseDone
	phaseError
)

func (p phase) String() string {
	switch p {
	case phaseInit:
		return "Initializing"
	case phaseFetchingIndex:
		return "Fetching docs index"
	case phaseQuerying:
		return "Thinking"
	case phaseReadingFile:
		return "Reading docs"
	case phaseAnswering:
		return "Answering"
	case phaseDone:
		return "Done"
	case phaseError:
		return "Error"
	default:
		return "Working"
	}
}

type queryState int

const (
	qStateInit queryState = iota
	qStateIndexFetched
	qStateStreaming
	qStateParsing
	qStateFetchingFiles
	qStateFilesReady
	qStateAnswering
	qStateDone
	qStateError
)

type displayLineType string

const (
	dlStep        displayLineType = "step"
	dlToolCall    displayLineType = "tool_call"
	dlAction      displayLineType = "action"
	dlThought     displayLineType = "thought"
	dlFilePath    displayLineType = "file_path"
	dlFileContent displayLineType = "file_content"
	dlAnswer      displayLineType = "answer"
	dlDivider     displayLineType = "divider"
	dlError       displayLineType = "error"
)

type displayLine struct {
	Type    displayLineType
	Content string
}

type queryStep struct {
	Step   int
	Phase  string
	Detail string
}

type Model struct {
	Width  int
	Height int

	Source docsource.Source
	Query  string
	Cfg    config.Config

	SelectedProvider string

	mode      modeType
	setupStep setupStep

	cursor         int
	inputBuffer    string
	providersList  []provider.Provider
	modelsList     []string

	phase      phase
	spinnerIdx int

	entries []docsource.Entry
	steps   []queryStep
	answer  string

	client      *llm.Client
	llmMessages []llm.ChatMessage
	readFiles   map[string]string
	iteration   int
	maxIters    int

	qState        queryState
	streamBuf     []byte
	reasoningBuf  []byte
	streaming     bool
	streamCh      chan llm.StreamDelta
	streamLive    []byte
	reasoningLive []byte
	pendingReads  []string
	gotAnswer   bool

	sessionLog   *sessionlog.Logger

	displayLines      []displayLine
	scrollOffset      int
	setupScrollOffset int
	autoScroll        bool

	err error
}

func NewQuery(sourceName, query string, cfg config.Config) Model {
	source, _ := docsource.Find(sourceName)
	providersList := append([]provider.Provider(nil), provider.Providers...)

	return Model{
		Source:           source,
		Query:            query,
		Cfg:              cfg,
		mode:             modeQuery,
		setupStep:        stepNone,
		cursor:           0,
		inputBuffer:      "",
		providersList:    providersList,
		modelsList:       nil,
		SelectedProvider: "",
		phase:            phaseInit,
		spinnerIdx:       0,
		readFiles:        make(map[string]string),
		iteration:        0,
		maxIters:         5,
		steps:            make([]queryStep, 0),
		displayLines:     make([]displayLine, 0),
		scrollOffset:     0,
		autoScroll:       true,
		qState:           qStateInit,
	}
}

func NewProviderSelect(cfg config.Config) Model {
	providersList := append([]provider.Provider(nil), provider.Providers...)
	return Model{
		Cfg:              cfg,
		mode:             modeProvider,
		setupStep:        stepProvider,
		cursor:           0,
		providersList:    providersList,
		SelectedProvider: "",
	}
}

func NewModelSelect(cfg config.Config) Model {
	return Model{
		Cfg:              cfg,
		mode:             modeModel,
		setupStep:        stepProvider,
		cursor:           0,
		providersList:    ProvidersWithKeys(cfg),
		SelectedProvider: "",
	}
}

func NewSetup(cfg config.Config) Model {
	providersList := append([]provider.Provider(nil), provider.Providers...)
	return Model{
		Cfg:              cfg,
		mode:             modeQuery,
		setupStep:        stepProvider,
		cursor:           0,
		providersList:    providersList,
		SelectedProvider: "",
	}
}

func NewModelSelectSingle(cfg config.Config, prov provider.Provider) Model {
	return Model{
		Cfg:              cfg,
		mode:             modeModel,
		setupStep:        stepModel,
		cursor:           0,
		SelectedProvider: prov.Name,
		modelsList:       provider.BuildModelList(prov),
		providersList:    []provider.Provider{prov},
	}
}

func ProvidersWithKeys(cfg config.Config) []provider.Provider {
	var result []provider.Provider
	for _, p := range provider.Providers {
		if provider.KeyExists(cfg, p.Name) {
			result = append(result, p)
		}
	}
	return result
}

func (m Model) Init() tea.Cmd {
	if m.mode != modeQuery {
		return tickCmd()
	}
	return tea.Batch(
		tickCmd(),
		startQuery(m.Source, m.Query),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return progressTickMsg{tick: int(t.UnixMilli())}
	})
}

func startQuery(source docsource.Source, query string) tea.Cmd {
	return func() tea.Msg {
		log, err := sessionlog.New(source.Name, query)
		if err != nil {
			// Non-fatal; log creation failure shouldn't block the query
			log = nil
		}

		entries, err := source.FetchIndex()
		if err != nil {
			if log != nil {
				log.Log("ERROR fetching index: %v", err)
				log.Close()
			}
			return docErrorMsg{err: fmt.Errorf("fetching index: %w", err), log: log}
		}
		if log != nil {
			log.Log("Index: %d entries found", len(entries))
		}
		return docIndexMsg{source: source.Name, entries: entries, log: log}
	}
}