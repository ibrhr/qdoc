package llm

import (
	"strings"
	"testing"
)

func TestParseResponses_ReadFile(t *testing.T) {
	input := `{"action":"read_file","url":"https://go.dev/doc/effective_go"}`
	actions := ParseResponses(input)
	if len(actions) != 1 {
		t.Fatalf("got %d actions, want 1", len(actions))
	}
	if actions[0].Action != ActionReadFile {
		t.Errorf("Action = %q, want %q", actions[0].Action, ActionReadFile)
	}
	if actions[0].URL != "https://go.dev/doc/effective_go" {
		t.Errorf("URL = %q", actions[0].URL)
	}
}

func TestParseResponses_Answer(t *testing.T) {
	input := `{"action":"answer"}
This is the answer content.
Multiple lines.`
	actions := ParseResponses(input)
	if len(actions) != 1 {
		t.Fatalf("got %d actions, want 1", len(actions))
	}
	if actions[0].Action != ActionAnswer {
		t.Errorf("Action = %q, want %q", actions[0].Action, ActionAnswer)
	}
	if !strings.Contains(actions[0].Content, "This is the answer content.") {
		t.Errorf("Content missing expected text: %q", actions[0].Content)
	}
}

func TestParseResponses_MultipleReadsThenAnswer(t *testing.T) {
	input := `{"action":"read_file","url":"https://go.dev/doc/one"}
{"action":"read_file","url":"https://go.dev/doc/two"}
{"action":"answer"}
Final answer here.`
	actions := ParseResponses(input)
	if len(actions) != 3 {
		t.Fatalf("got %d actions, want 3", len(actions))
	}
	if actions[0].Action != ActionReadFile {
		t.Errorf("actions[0].Action = %q", actions[0].Action)
	}
	if actions[1].Action != ActionReadFile {
		t.Errorf("actions[1].Action = %q", actions[1].Action)
	}
	if actions[2].Action != ActionAnswer {
		t.Errorf("actions[2].Action = %q", actions[2].Action)
	}
}

func TestParseResponses_OnlyAnswer(t *testing.T) {
	input := `{"action":"answer"}
Just the answer.`
	actions := ParseResponses(input)
	if len(actions) != 1 {
		t.Fatalf("got %d actions, want 1", len(actions))
	}
	if actions[0].Action != ActionAnswer {
		t.Errorf("Action = %q, want %q", actions[0].Action, ActionAnswer)
	}
}

func TestParseResponses_Fallback(t *testing.T) {
	input := "This is not JSON at all. Just a plain text response."
	actions := ParseResponses(input)
	if len(actions) != 1 {
		t.Fatalf("got %d actions, want 1 (fallback)", len(actions))
	}
	if actions[0].Action != ActionAnswer {
		t.Errorf("fallback Action = %q, want %q", actions[0].Action, ActionAnswer)
	}
	if actions[0].Content != input {
		t.Errorf("fallback Content = %q, want %q", actions[0].Content, input)
	}
}

func TestParseResponses_TrailingGarbage(t *testing.T) {
	input := `{"action":"read_file","url":"https://example.com/doc"} some trailing text`
	actions := ParseResponses(input)
	if len(actions) != 1 {
		t.Fatalf("got %d actions, want 1", len(actions))
	}
	if actions[0].Action != ActionReadFile {
		t.Errorf("Action = %q", actions[0].Action)
	}
}

func TestParseResponses_EmptyURL(t *testing.T) {
	input := `{"action":"read_file","url":""}`
	actions := ParseResponses(input)
	if len(actions) != 1 {
		t.Fatalf("got %d actions, want 1 (fallback answer)", len(actions))
	}
	if actions[0].Action != ActionAnswer {
		t.Errorf("Action = %q, want %q (fallback)", actions[0].Action, ActionAnswer)
	}
}

func TestParseResponses_UnknownAction(t *testing.T) {
	input := `{"action":"unknown","url":"https://example.com"}`
	actions := ParseResponses(input)
	if len(actions) != 1 {
		t.Fatalf("got %d actions, want 1 (fallback answer)", len(actions))
	}
	if actions[0].Action != ActionAnswer {
		t.Errorf("Action = %q, want %q (fallback)", actions[0].Action, ActionAnswer)
	}
}

func TestParseResponses_NoAction(t *testing.T) {
	input := `{"action":"read_file","url":"https://example.com/doc1"}
{"action":"read_file","url":"https://example.com/doc2"}`
	actions := ParseResponses(input)
	if len(actions) != 2 {
		t.Fatalf("got %d actions, want 2 (reads only, no fallback)", len(actions))
	}
	if actions[0].Action != ActionReadFile {
		t.Errorf("actions[0].Action = %q", actions[0].Action)
	}
	if actions[1].Action != ActionReadFile {
		t.Errorf("actions[1].Action = %q", actions[1].Action)
	}
}

func TestParseResponses_MixedWithGarbage(t *testing.T) {
	input := `not json
{"action":"read_file","url":"https://example.com/doc1"}
some non-json commentary
{"action":"read_file","url":"https://example.com/doc2"}
{"action":"answer"}
answer text`
	actions := ParseResponses(input)
	if len(actions) != 3 {
		t.Fatalf("got %d actions, want 3", len(actions))
	}
	if actions[0].URL != "https://example.com/doc1" {
		t.Errorf("actions[0].URL = %q", actions[0].URL)
	}
	if actions[1].URL != "https://example.com/doc2" {
		t.Errorf("actions[1].URL = %q", actions[1].URL)
	}
	if actions[2].Action != ActionAnswer {
		t.Errorf("actions[2].Action = %q", actions[2].Action)
	}
}

func TestParseResponses_EmptyResponse(t *testing.T) {
	actions := ParseResponses("")
	if len(actions) != 1 {
		t.Fatalf("got %d actions, want 1 (fallback)", len(actions))
	}
	if actions[0].Action != ActionAnswer {
		t.Errorf("Action = %q, want %q", actions[0].Action, ActionAnswer)
	}
}

func TestParseResponses_ReadFileWithoutURL(t *testing.T) {
	input := `{"action":"read_file"}`
	actions := ParseResponses(input)
	if len(actions) != 1 {
		t.Fatalf("got %d actions, want 1 (fallback answer)", len(actions))
	}
	if actions[0].Action != ActionAnswer {
		t.Errorf("Action = %q, want %q (fallback)", actions[0].Action, ActionAnswer)
	}
}

func TestParseResponses_OnlyAnswerAfterReads(t *testing.T) {
	input := `{"action":"read_file","url":"https://example.com/a"}
{"action":"read_file","url":"https://example.com/b"}
{"action":"answer"}
The definitive answer.
With multiple lines.
And more content.`
	actions := ParseResponses(input)
	if len(actions) != 3 {
		t.Fatalf("got %d actions, want 3", len(actions))
	}
	if actions[2].Content != "The definitive answer.\nWith multiple lines.\nAnd more content." {
		t.Errorf("answer Content = %q", actions[2].Content)
	}
}
