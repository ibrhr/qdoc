package llm

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/ibrhr/qdoc/internal/retry"
)

type CursorClient struct {
	agentPath string
	apiKey    string
	model     string
	provider  string
}

func (c *CursorClient) ModelName() string    { return c.model }
func (c *CursorClient) ProviderName() string { return c.provider }

func (c *CursorClient) Send(messages []ChatMessage) (string, error) {
	var lastErr error
	for attempt := 0; attempt < retry.LLMRetry.MaxAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(retry.BackoffDelay(retry.LLMRetry, attempt-1))
		}

		var sb strings.Builder
		err := c.runACP(messages, func(text, reasoning string) {
			sb.WriteString(text)
			sb.WriteString(reasoning)
		})
		if err != nil {
			if isCursorRetryableError(err) {
				lastErr = &retry.RetryableError{Msg: err.Error()}
				continue
			}
			return "", err
		}
		return sb.String(), nil
	}
	return "", fmt.Errorf("retry exhausted (%d attempts): %w", retry.LLMRetry.MaxAttempts, lastErr)
}

func (c *CursorClient) Stream(messages []ChatMessage, ch chan<- StreamDelta) {
	defer close(ch)

	var lastErr error
	for attempt := 0; attempt < retry.LLMRetry.MaxAttempts; attempt++ {
		if attempt > 0 {
			delay := retry.BackoffDelay(retry.LLMRetry, attempt-1)

			ch <- StreamDelta{
				Content: fmt.Sprintf("(retrying in %.1fs — attempt %d/%d)",
					delay.Seconds(), attempt+1, retry.LLMRetry.MaxAttempts),
				Retrying: true,
			}

			time.Sleep(delay)
		}

		err := c.runACP(messages, func(text, reasoning string) {
			if reasoning != "" {
				ch <- StreamDelta{Reasoning: reasoning}
			}
			if text != "" {
				ch <- StreamDelta{Content: text}
			}
		})
		if err != nil {
			if isCursorRetryableError(err) {
				lastErr = &retry.RetryableError{Msg: err.Error()}
				continue
			}
			ch <- StreamDelta{Err: err}
			return
		}
		ch <- StreamDelta{Done: true}
		return
	}
	ch <- StreamDelta{Err: fmt.Errorf("retry exhausted (%d attempts): %w", retry.LLMRetry.MaxAttempts, lastErr)}
}

type contentHandler func(text, reasoning string)

func (c *CursorClient) runACP(messages []ChatMessage, onContent contentHandler) error {
	cmd, stdin, stdout, err := c.startAgent()
	if err != nil {
		return err
	}
	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		cmd.Wait()
	}()

	sess := &acpSession{
		stdin:  stdin,
		stdout: bufio.NewScanner(stdout),
		nextID: 1,
		mu:     &sync.Mutex{},
	}
	sess.stdout.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	if err := sess.initialize(); err != nil {
		return err
	}

	if err := sess.authenticate(); err != nil {
		return err
	}

	if err := sess.newSession(); err != nil {
		return err
	}

	if err := sess.setAskMode(); err != nil {
		return err
	}

	prompt := c.buildPrompt(messages)
	promptBlocks := []acpContentBlock{{Type: "text", Text: prompt}}

	return sess.streamPrompt(promptBlocks, onContent)
}

func (c *CursorClient) startAgent() (*exec.Cmd, io.WriteCloser, io.ReadCloser, error) {
	var args []string
	if c.apiKey != "" {
		args = []string{"--api-key", c.apiKey, "acp"}
	} else {
		args = []string{"acp"}
	}

	cmd := exec.Command(c.agentPath, args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("cursor: stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return nil, nil, nil, fmt.Errorf("cursor: stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		stdin.Close()
		stdout.Close()
		return nil, nil, nil, fmt.Errorf("cursor: start agent: %w", err)
	}

	return cmd, stdin, stdout, nil
}

func (c *CursorClient) buildPrompt(messages []ChatMessage) string {
	var sb strings.Builder
	for i, msg := range messages {
		switch msg.Role {
		case "system":
			sb.WriteString(msg.Content)
		case "user":
			if i > 0 {
				sb.WriteString("\n\n")
			}
			sb.WriteString(msg.Content)
		case "assistant":
			sb.WriteString("\n\n")
			sb.WriteString(msg.Content)
		}
	}
	return sb.String()
}

type acpSession struct {
	stdin      io.WriteCloser
	stdout     *bufio.Scanner
	nextID     int
	sessionID  string
	mu         *sync.Mutex
}

func (s *acpSession) initialize() error {
	params := map[string]interface{}{
		"protocolVersion": 1,
		"clientCapabilities": map[string]interface{}{
			"fs": map[string]bool{
				"readTextFile":  false,
				"writeTextFile": false,
			},
			"terminal": false,
		},
		"clientInfo": map[string]string{
			"name":    "qdoc",
			"version": "1.0.0",
		},
	}
	_, err := s.call("initialize", params)
	return err
}

func (s *acpSession) authenticate() error {
	params := map[string]string{
		"methodId": "cursor_login",
	}
	_, err := s.call("authenticate", params)
	return err
}

func (s *acpSession) newSession() error {
	params := map[string]interface{}{
		"cwd":        "/tmp",
		"mcpServers": []interface{}{},
	}
	result, err := s.call("session/new", params)
	if err != nil {
		return err
	}
	var parsed struct {
		SessionID string `json:"sessionId"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		return fmt.Errorf("cursor: parse session/new response: %w", err)
	}
	s.sessionID = parsed.SessionID
	return nil
}

func (s *acpSession) setAskMode() error {
	params := map[string]string{
		"sessionId": s.sessionID,
		"modeId":    "ask",
	}
	_, err := s.call("session/set_mode", params)
	if err != nil {
		if strings.Contains(err.Error(), "Method not found") {
			return nil
		}
		return err
	}
	return nil
}

func (s *acpSession) respond(id int, result interface{}) {
	resp := jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
	}
	if result != nil {
		data, _ := json.Marshal(result)
		resp.Result = data
	}
	data, _ := json.Marshal(resp)
	fmt.Fprintf(s.stdin, "%s\n", data)
}

func (s *acpSession) rejectPermission(id int) {
	s.respond(id, map[string]interface{}{
		"outcome": map[string]interface{}{
			"outcome":  "selected",
			"optionId": "reject-once",
		},
	})
}

func (s *acpSession) rejectAnyRequest(method string, id int) {
	if method == "session/request_permission" {
		s.rejectPermission(id)
		return
	}
	s.respondCursorExtension(id)
}

func (s *acpSession) respondCursorExtension(id int) {
	msg := map[string]interface{}{
		"outcome": map[string]string{"outcome": "cancelled"},
	}
	s.respond(id, msg)
}

func (s *acpSession) streamPrompt(
	prompt []acpContentBlock,
	onContent contentHandler,
) error {
	s.mu.Lock()
	id := s.nextID
	s.nextID++
	req := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  "session/prompt",
		Params: map[string]interface{}{
			"sessionId": s.sessionID,
			"prompt":    prompt,
		},
	}
	data, err := json.Marshal(req)
	s.mu.Unlock()
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(s.stdin, "%s\n", data); err != nil {
		return fmt.Errorf("cursor: write prompt: %w", err)
	}

	for s.stdout.Scan() {
		line := s.stdout.Text()
		if line == "" {
			continue
		}

		msg, err := parseACPLine(line)
		if err != nil {
			continue
		}

		if msg.IsRequest() {
			switch msg.Method {
			case "session/request_permission":
				s.rejectPermission(msg.ID)
			case "cursor/ask_question", "cursor/create_plan":
				s.respondCursorExtension(msg.ID)
			default:
				if strings.HasPrefix(msg.Method, "cursor/") {
					s.respondCursorExtension(msg.ID)
				}
			}
			continue
		}

		if msg.IsNotification() {
			if msg.Method == "session/update" {
				update := parseSessionUpdate(msg.Params)
				if update != nil {
					onContent(update.Text, update.Reasoning)
				}
			}
			continue
		}

		if msg.IsResponse() {
			if msg.ID == id {
				if msg.Error != nil {
					return fmt.Errorf("cursor: prompt error %d: %s", msg.Error.Code, msg.Error.Message)
				}
				return nil
			}
		}
	}

	if err := s.stdout.Err(); err != nil {
		return fmt.Errorf("cursor: read stream: %w", err)
	}

	return fmt.Errorf("cursor: agent exited unexpectedly")
}

type acpLineMsg struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      int              `json:"id"`
	Method  string           `json:"method"`
	Result  json.RawMessage  `json:"result,omitempty"`
	Error   *jsonRPCError    `json:"error,omitempty"`
	Params  json.RawMessage  `json:"params,omitempty"`
}

func (m *acpLineMsg) IsRequest() bool {
	return m.Method != "" && m.ID != 0
}

func (m *acpLineMsg) IsNotification() bool {
	return m.Method != "" && m.ID == 0
}

func (m *acpLineMsg) IsResponse() bool {
	return m.ID != 0 && m.Method == "" && (len(m.Result) > 0 || m.Error != nil)
}

type jsonRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonRPCError   `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type acpContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type sessionUpdate struct {
	Text      string
	Reasoning string
}

func (s *acpSession) call(method string, params interface{}) (json.RawMessage, error) {
	s.mu.Lock()
	id := s.nextID
	s.nextID++
	req := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}
	data, err := json.Marshal(req)
	s.mu.Unlock()
	if err != nil {
		return nil, err
	}

	if _, err := fmt.Fprintf(s.stdin, "%s\n", data); err != nil {
		return nil, fmt.Errorf("cursor: write %s: %w", method, err)
	}

	for s.stdout.Scan() {
		line := s.stdout.Text()
		if line == "" {
			continue
		}

		msg, err := parseACPLine(line)
		if err != nil {
			continue
		}

		if msg.IsResponse() && msg.ID == id {
			if msg.Error != nil {
				return nil, fmt.Errorf("cursor: %s error %d: %s", method, msg.Error.Code, msg.Error.Message)
			}
			return msg.Result, nil
		}

		if msg.IsRequest() {
			s.rejectAnyRequest(msg.Method, msg.ID)
			continue
		}

		if msg.IsNotification() {
			continue
		}
	}

	if err := s.stdout.Err(); err != nil {
		return nil, fmt.Errorf("cursor: read %s response: %w", method, err)
	}

	return nil, fmt.Errorf("cursor: agent exited during %s", method)
}

func isCursorRetryableError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	if strings.Contains(msg, "agent exited") {
		return true
	}
	if strings.Contains(msg, "write") {
		return true
	}
	if strings.Contains(msg, "read") {
		return true
	}
	return false
}

func parseACPLine(line string) (*acpLineMsg, error) {
	var msg acpLineMsg

	if err := json.Unmarshal([]byte(line), &msg); err != nil {
		return nil, err
	}

	if msg.JSONRPC != "2.0" {
		return nil, fmt.Errorf("invalid jsonrpc version")
	}

	return &msg, nil
}

func parseSessionUpdate(raw json.RawMessage) *sessionUpdate {
	var wrapper struct {
		Update struct {
			SessionUpdate string          `json:"sessionUpdate"`
			Content       json.RawMessage `json:"content,omitempty"`
		} `json:"update"`
	}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return nil
	}

	if wrapper.Update.SessionUpdate != "agent_message_chunk" {
		return nil
	}

	var content struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(wrapper.Update.Content, &content); err != nil {
		return nil
	}

	switch content.Type {
	case "text":
		return &sessionUpdate{Text: content.Text}
	case "reasoning":
		return &sessionUpdate{Reasoning: content.Text}
	}
	return nil
}

var execLookPath = exec.LookPath

func newCursorClient(cfg Config) (*CursorClient, error) {
	agentPath, err := execLookPath("agent")
	if err != nil {
		return nil, fmt.Errorf(
			"cursor: Cursor CLI not found. Install it from https://cursor.com, then run: agent login\n  (%w)",
			err,
		)
	}

	apiKey := cfg.Token.AccessToken

	model := cfg.Model
	if model == "" {
		model = "gpt-5.5"
	}

	return &CursorClient{
		agentPath: agentPath,
		apiKey:    apiKey,
		model:     model,
		provider:  cfg.Provider,
	}, nil
}


