package tui

import (
	"fmt"
	"testing"

	"github.com/ibrhr/qdoc/internal/llm"

	tea "charm.land/bubbletea/v2"
)

func TestHandleQueryKey_Quit_q(t *testing.T) {
	m := &Model{mode: modeQuery, setupStep: stepNone}
	_, cmd := m.handleQueryKey(tea.KeyPressMsg{Text: "q"})
	if cmd == nil {
		t.Error("expected quit command for q")
	}
	if m.scrollOffset != 0 {
		t.Errorf("scrollOffset should not change on quit")
	}
}

func TestHandleQueryKey_ScrollUp(t *testing.T) {
	m := &Model{
		mode:        modeQuery,
		setupStep:   stepNone,
		phase:       phaseQuerying,
		autoScroll:  true,
		scrollOffset: 5,
	}
	_, _ = m.handleQueryKey(tea.KeyPressMsg{Text: "up"})
	if m.scrollOffset != 4 {
		t.Errorf("scrollOffset = %d, want 4", m.scrollOffset)
	}
	if m.autoScroll {
		t.Error("autoScroll should be false after manual scroll")
	}
}

func TestHandleQueryKey_ScrollDown(t *testing.T) {
	m := &Model{
		mode:        modeQuery,
		setupStep:   stepNone,
		phase:       phaseQuerying,
		autoScroll:  true,
		scrollOffset: 5,
	}
	_, _ = m.handleQueryKey(tea.KeyPressMsg{Text: "down"})
	if m.scrollOffset != 6 {
		t.Errorf("scrollOffset = %d, want 6", m.scrollOffset)
	}
	if m.autoScroll {
		t.Error("autoScroll should be false after manual scroll")
	}
}

func TestHandleQueryKey_ScrollUpAtZero(t *testing.T) {
	m := &Model{
		mode:         modeQuery,
		setupStep:    stepNone,
		phase:        phaseQuerying,
		scrollOffset: 0,
	}
	_, _ = m.handleQueryKey(tea.KeyPressMsg{Text: "up"})
	if m.scrollOffset != 0 {
		t.Errorf("scrollOffset = %d, want 0 (clamped)", m.scrollOffset)
	}
}

func TestHandleQueryKey_PageUp(t *testing.T) {
	m := &Model{
		mode:         modeQuery,
		setupStep:    stepNone,
		phase:        phaseQuerying,
		autoScroll:   true,
		scrollOffset: 10,
	}
	_, _ = m.handleQueryKey(tea.KeyPressMsg{Text: "pgup"})
	if m.scrollOffset != 0 {
		t.Errorf("scrollOffset = %d, want 0", m.scrollOffset)
	}
}

func TestHandleQueryKey_PageDown(t *testing.T) {
	m := &Model{
		mode:         modeQuery,
		setupStep:    stepNone,
		phase:        phaseQuerying,
		scrollOffset: 5,
		displayLines: make([]displayLine, 20),
	}
	_, _ = m.handleQueryKey(tea.KeyPressMsg{Text: "pgdown"})
	if m.scrollOffset != 20 {
		t.Errorf("scrollOffset = %d, want 20", m.scrollOffset)
	}
}

func TestAddDisplay(t *testing.T) {
	m := &Model{}
	m.addDisplay(dlStep, "step 1")
	m.addDisplay(dlToolCall, "tool call 1")

	if len(m.displayLines) != 2 {
		t.Fatalf("expected 2 display lines, got %d", len(m.displayLines))
	}
	if m.displayLines[0].Type != dlStep {
		t.Errorf("line 0 type = %v, want dlStep", m.displayLines[0].Type)
	}
	if m.displayLines[0].Content != "step 1" {
		t.Errorf("line 0 content = %q", m.displayLines[0].Content)
	}
	if m.displayLines[1].Type != dlToolCall {
		t.Errorf("line 1 type = %v, want dlToolCall", m.displayLines[1].Type)
	}
}

func TestAddStep(t *testing.T) {
	m := &Model{}
	m.addStep("Phase 1", "detail 1")
	m.addStep("Phase 2", "detail 2")

	if len(m.steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(m.steps))
	}
	if m.steps[0].Phase != "Phase 1" {
		t.Errorf("step 0 phase = %q", m.steps[0].Phase)
	}
	if m.steps[0].Detail != "detail 1" {
		t.Errorf("step 0 detail = %q", m.steps[0].Detail)
	}
	if m.steps[0].Step != 1 {
		t.Errorf("step 0 number = %d, want 1", m.steps[0].Step)
	}
	if m.steps[1].Step != 2 {
		t.Errorf("step 1 number = %d, want 2", m.steps[1].Step)
	}
}

func TestHandleStreamDelta_Content(t *testing.T) {
	m := &Model{
		mode:      modeQuery,
		setupStep: stepNone,
		streamCh:  nil,
	}
	delta := llm.StreamDelta{Content: "hello"}

	// We can't test the full pipeline without a channel, but we can test
	// that state is set correctly just before reading the channel
	model, _ := m.handleStreamDelta(delta)
	if model.(*Model).streaming != true {
		t.Error("streaming should be true after content delta")
	}
	streamBuf := string(model.(*Model).streamBuf)
	if streamBuf != "hello" {
		t.Errorf("streamBuf = %q, want %q", streamBuf, "hello")
	}
}

func TestHandleStreamDelta_Error(t *testing.T) {
	m := &Model{
		mode:      modeQuery,
		setupStep: stepNone,
	}
	delta := llm.StreamDelta{Err: fmt.Errorf("test error")}
	m.handleStreamDelta(delta)

	if m.phase != phaseError {
		t.Errorf("phase = %v, want phaseError", m.phase)
	}
	if m.qState != qStateError {
		t.Errorf("qState = %v, want qStateError", m.qState)
	}
	if m.err == nil {
		t.Error("expected error to be set")
	}
}

func TestHandleStreamDelta_Retrying(t *testing.T) {
	m := &Model{
		mode:      modeQuery,
		setupStep: stepNone,
		streamCh:  nil,
	}
	delta := llm.StreamDelta{Content: "retrying...", Retrying: true}
	m.handleStreamDelta(delta)

	if len(m.displayLines) != 1 {
		t.Fatalf("expected 1 display line, got %d", len(m.displayLines))
	}
	if m.displayLines[0].Type != dlToolCall {
		t.Errorf("display type = %v, want dlToolCall", m.displayLines[0].Type)
	}
}

func TestHandleDocContent_DecrementsPending(t *testing.T) {
	m := &Model{
		mode:     modeQuery,
		qState:   qStateFetchingFiles,
		readFiles: map[string]string{},
		filesPending: 2,
	}
	msg := docContentMsg{url: "https://example.com/doc", content: "content here"}
	m.handleDocContent(msg)

	if m.filesPending != 1 {
		t.Errorf("filesPending = %d, want 1", m.filesPending)
	}
	if m.readFiles[msg.url] != msg.content {
		t.Errorf("readFiles[%q] = %q", msg.url, m.readFiles[msg.url])
	}
}

func TestRemainingFetching_AllDone_GotAnswer(t *testing.T) {
	m := &Model{
		mode:         modeQuery,
		qState:       qStateFetchingFiles,
		filesPending: 0,
		gotAnswer:    true,
		readFiles:    map[string]string{"a": "content"},
		pendingReads: []string{"a"},
	}
	_, cmd := m.remainingFetching()
	if cmd == nil {
		t.Error("expected a command when answer ready and files done")
	}
}

func TestRemainingFetching_Pending(t *testing.T) {
	m := &Model{
		mode:         modeQuery,
		qState:       qStateFetchingFiles,
		filesPending: 3,
		readFiles:    map[string]string{},
		pendingReads: []string{"a", "b", "c"},
	}
	_, cmd := m.remainingFetching()
	if cmd != nil {
		t.Error("expected no command when files still pending")
	}
}

func TestStartFileFetches_SetsPending(t *testing.T) {
	m := &Model{
		pendingReads: []string{"https://a.com", "https://b.com", "https://c.com"},
	}
	m.startFileFetches()
	if m.filesPending != 3 {
		t.Errorf("filesPending = %d, want 3", m.filesPending)
	}
}
