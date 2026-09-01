package llm

import (
	"encoding/json"
	"os/exec"
	"testing"

	"github.com/ibrhr/qdoc/internal/auth"
)

func TestBuildPrompt(t *testing.T) {
	c := &CursorClient{}

	tests := []struct {
		name     string
		messages []ChatMessage
		want     string
	}{
		{
			name: "system only",
			messages: []ChatMessage{
				{Role: "system", Content: "You are a helpful assistant."},
			},
			want: "You are a helpful assistant.",
		},
		{
			name: "system and user",
			messages: []ChatMessage{
				{Role: "system", Content: "You are a helpful assistant."},
				{Role: "user", Content: "Hello!"},
			},
			want: "You are a helpful assistant.\n\nHello!",
		},
		{
			name: "multi-turn conversation",
			messages: []ChatMessage{
				{Role: "system", Content: "You are a docs assistant."},
				{Role: "user", Content: "How do I use X?"},
				{Role: "assistant", Content: "READ_FILE: /doc/x"},
				{Role: "user", Content: "[content of /doc/x]"},
			},
			want: "You are a docs assistant.\n\nHow do I use X?\n\nREAD_FILE: /doc/x\n\n[content of /doc/x]",
		},
		{
			name:     "empty messages",
			messages: []ChatMessage{},
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.buildPrompt(tt.messages)
			if got != tt.want {
				t.Errorf("buildPrompt() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseACPLine(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		wantErr bool
		check   func(*testing.T, *acpLineMsg)
	}{
		{
			name:    "invalid json",
			line:    "not json",
			wantErr: true,
		},
		{
			name:    "wrong jsonrpc version",
			line:    `{"jsonrpc":"1.0","id":1,"result":{}}`,
			wantErr: true,
		},
		{
			name:    "valid response",
			line:    `{"jsonrpc":"2.0","id":1,"result":{"sessionId":"abc"}}`,
			wantErr: false,
			check: func(t *testing.T, m *acpLineMsg) {
				if m.ID != 1 {
					t.Errorf("ID = %d, want 1", m.ID)
				}
				if !m.IsResponse() {
					t.Error("expected IsResponse() = true")
				}
			},
		},
		{
			name:    "valid notification",
			line:    `{"jsonrpc":"2.0","method":"session/update","params":{"sessionId":"abc"}}`,
			wantErr: false,
			check: func(t *testing.T, m *acpLineMsg) {
				if m.Method != "session/update" {
					t.Errorf("Method = %q, want session/update", m.Method)
				}
				if !m.IsNotification() {
					t.Error("expected IsNotification() = true")
				}
			},
		},
		{
			name:    "valid request from agent",
			line:    `{"jsonrpc":"2.0","id":5,"method":"session/request_permission","params":{}}`,
			wantErr: false,
			check: func(t *testing.T, m *acpLineMsg) {
				if m.ID != 5 {
					t.Errorf("ID = %d, want 5", m.ID)
				}
				if m.Method != "session/request_permission" {
					t.Errorf("Method = %q", m.Method)
				}
				if !m.IsRequest() {
					t.Error("expected IsRequest() = true")
				}
			},
		},
		{
			name:    "error response",
			line:    `{"jsonrpc":"2.0","id":2,"error":{"code":-32000,"message":"not authenticated"}}`,
			wantErr: false,
			check: func(t *testing.T, m *acpLineMsg) {
				if !m.IsResponse() {
					t.Error("expected IsResponse() = true")
				}
				if m.Error == nil {
					t.Fatal("expected Error to be non-nil")
				}
				if m.Error.Code != -32000 {
					t.Errorf("Error.Code = %d, want -32000", m.Error.Code)
				}
			},
		},
		{
			name:    "empty line",
			line:    ``,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := parseACPLine(tt.line)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, msg)
			}
		})
	}
}

func TestAcpLineMsg_IsMethods(t *testing.T) {
	tests := []struct {
		name           string
		msg            acpLineMsg
		isRequest      bool
		isNotification bool
		isResponse     bool
	}{
		{
			name:           "notification: method + no id",
			msg:            acpLineMsg{JSONRPC: "2.0", Method: "session/update"},
			isNotification: true,
		},
		{
			name:      "request: method + id",
			msg:       acpLineMsg{JSONRPC: "2.0", ID: 1, Method: "session/request_permission"},
			isRequest: true,
		},
		{
			name:       "response: id + result + no method",
			msg:        acpLineMsg{JSONRPC: "2.0", ID: 1, Result: json.RawMessage(`{}`)},
			isResponse: true,
		},
		{
			name:       "response: id + error + no method",
			msg:        acpLineMsg{JSONRPC: "2.0", ID: 1, Error: &jsonRPCError{Code: -1, Message: "err"}},
			isResponse: true,
		},
		{
			name: "none: id only (empty result, no method)",
			msg:  acpLineMsg{JSONRPC: "2.0", ID: 1},
		},
		{
			name: "cursor notification: cursor method + no id",
			msg:  acpLineMsg{JSONRPC: "2.0", Method: "cursor/update_todos"},
			isNotification: true,
		},
		{
			name:      "cursor request: cursor method + id",
			msg:       acpLineMsg{JSONRPC: "2.0", ID: 1, Method: "cursor/ask_question"},
			isRequest: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.msg.IsRequest(); got != tt.isRequest {
				t.Errorf("IsRequest() = %v, want %v", got, tt.isRequest)
			}
			if got := tt.msg.IsNotification(); got != tt.isNotification {
				t.Errorf("IsNotification() = %v, want %v", got, tt.isNotification)
			}
			if got := tt.msg.IsResponse(); got != tt.isResponse {
				t.Errorf("IsResponse() = %v, want %v", got, tt.isResponse)
			}
		})
	}
}

func TestParseSessionUpdate(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want *sessionUpdate
	}{
		{
			name: "agent_message_chunk with text",
			raw:  `{"sessionId":"abc","update":{"sessionUpdate":"agent_message_chunk","content":{"type":"text","text":"Hello world"}}}`,
			want: &sessionUpdate{Text: "Hello world"},
		},
		{
			name: "non-agent_message_chunk",
			raw:  `{"sessionId":"abc","update":{"sessionUpdate":"tool_call"}}`,
			want: nil,
		},
		{
			name: "non-text content",
			raw:  `{"sessionId":"abc","update":{"sessionUpdate":"agent_message_chunk","content":{"type":"image","text":"ignored"}}}`,
			want: nil,
		},
		{
			name: "invalid json",
			raw:  `not json`,
			want: nil,
		},
		{
			name: "empty raw",
			raw:  `{}`,
			want: nil,
		},
		{
			name: "agent_message_chunk with reasoning",
			raw:  `{"sessionId":"abc","update":{"sessionUpdate":"agent_message_chunk","content":{"type":"text","text":"Let me think..."}}}`,
			want: &sessionUpdate{Text: "Let me think..."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSessionUpdate(json.RawMessage(tt.raw))
			if tt.want == nil {
				if got != nil {
					t.Errorf("expected nil, got %+v", got)
				}
				return
			}
			if got == nil {
				t.Fatal("expected non-nil")
			}
			if got.Text != tt.want.Text {
				t.Errorf("Text = %q, want %q", got.Text, tt.want.Text)
			}
		})
	}
}

func TestNewCursorClientBinaryNotFound(t *testing.T) {
	origLookPath := execLookPath
	execLookPath = func(name string) (string, error) {
		return "", &exec.Error{Name: name, Err: exec.ErrNotFound}
	}
	defer func() { execLookPath = origLookPath }()

	cfg := Config{
		Token:    auth.Token{AccessToken: "test-key"},
		Model:    "gpt-5.5",
		Provider: "cursor",
	}

	_, err := newCursorClient(cfg)
	if err == nil {
		t.Fatal("expected error when agent binary not found")
	}
}

func TestNewCursorClientSuccess(t *testing.T) {
	origLookPath := execLookPath
	execLookPath = func(name string) (string, error) {
		return "/usr/local/bin/agent", nil
	}
	defer func() { execLookPath = origLookPath }()

	tests := []struct {
		name      string
		cfg       Config
		wantModel string
		wantKey   string
	}{
		{
			name: "with api key",
			cfg: Config{
				Token:    auth.Token{AccessToken: "crsr_test123"},
				Model:    "claude-4.7-opus",
				Provider: "cursor",
			},
			wantModel: "claude-4.7-opus",
			wantKey:   "crsr_test123",
		},
		{
			name: "default model when empty",
			cfg: Config{
				Token:    auth.Token{},
				Model:    "",
				Provider: "cursor",
			},
			wantModel: "gpt-5.5",
			wantKey:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := newCursorClient(tt.cfg)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if c.model != tt.wantModel {
				t.Errorf("model = %q, want %q", c.model, tt.wantModel)
			}
			if c.apiKey != tt.wantKey {
				t.Errorf("apiKey = %q, want %q", c.apiKey, tt.wantKey)
			}
			if c.agentPath != "/usr/local/bin/agent" {
				t.Errorf("agentPath = %q, want /usr/local/bin/agent", c.agentPath)
			}
			if c.provider != "cursor" {
				t.Errorf("provider = %q, want cursor", c.provider)
			}
		})
	}
}

func TestCursorClientImplementsClient(t *testing.T) {
	origLookPath := execLookPath
	execLookPath = func(name string) (string, error) {
		return "/usr/local/bin/agent", nil
	}
	defer func() { execLookPath = origLookPath }()

	cfg := Config{
		Token:    auth.Token{AccessToken: "key"},
		Model:    "gpt-5.5",
		Provider: "cursor",
	}

	c, err := newCursorClient(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.ModelName() != "gpt-5.5" {
		t.Errorf("ModelName() = %q, want gpt-5.5", c.ModelName())
	}
	if c.ProviderName() != "cursor" {
		t.Errorf("ProviderName() = %q, want cursor", c.ProviderName())
	}

	var client Client = c
	if client == nil {
		t.Error("CursorClient does not implement Client")
	}
}

func TestNewClientCursorAcp(t *testing.T) {
	origLookPath := execLookPath
	execLookPath = func(name string) (string, error) {
		return "/usr/local/bin/agent", nil
	}
	defer func() { execLookPath = origLookPath }()

	cfg := Config{
		Token:    auth.Token{AccessToken: "test-key"},
		Model:    "claude-4.7-opus",
		Provider: "cursor",
	}

	client, err := NewClient("cursor-acp", cfg)
	if err != nil {
		t.Fatalf("NewClient(cursor-acp) error: %v", err)
	}
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.ModelName() != "claude-4.7-opus" {
		t.Errorf("ModelName() = %q, want claude-4.7-opus", client.ModelName())
	}
	if client.ProviderName() != "cursor" {
		t.Errorf("ProviderName() = %q, want cursor", client.ProviderName())
	}
}

func TestNewClientCursorAcpUnknown(t *testing.T) {
	client, err := NewClient("unknown-type", Config{})
	if err == nil {
		t.Error("expected error for unknown api type")
	}
	if client != nil {
		t.Error("expected nil client for unknown api type")
	}
}
