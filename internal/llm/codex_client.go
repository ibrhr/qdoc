package llm

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ibrhr/qdoc/internal/auth"
	"github.com/ibrhr/qdoc/internal/retry"
)

type CodexClient struct {
	Auth     auth.AuthMethod
	Token    auth.Token
	BaseURL  string
	Model    string
	Provider string
	Headers  map[string]string
}

func (c *CodexClient) ModelName() string    { return c.Model }
func (c *CodexClient) ProviderName() string { return c.Provider }

type codexInput struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type codexRequest struct {
	Model        string       `json:"model"`
	Input        []codexInput `json:"input"`
	Instructions string       `json:"instructions,omitempty"`
	Stream       bool         `json:"stream"`
}

type codexDelta struct {
	Delta string `json:"delta"`
}

type codexChoiceChunk struct {
	Choices []struct {
		Delta struct {
			Content          string `json:"content"`
			ReasoningContent string `json:"reasoning_content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

func (c *CodexClient) setHeaders(req *http.Request) {
	c.Auth.ApplyAuth(req, c.Token)
	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}
	accountID := extractAccountID(c.Token)
	if accountID != "" {
		req.Header.Set("chatgpt-account-id", accountID)
	}
}

func (c *CodexClient) Send(messages []ChatMessage) (string, error) {
	body := c.buildRequest(messages, false)

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	url := strings.TrimSuffix(c.BaseURL, "/") + "/codex/responses"

	var lastErr error
	for attempt := 0; attempt < retry.LLMRetry.MaxAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(retry.BackoffDelay(retry.LLMRetry, attempt-1))
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
		if err != nil {
			return "", err
		}

		req.Header.Set("Content-Type", "application/json")
		c.setHeaders(req)

		resp, err := httpClient.Do(req)
		if err != nil {
			lastErr = &retry.RetryableError{Msg: fmt.Sprintf("codex request: %v", err)}
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = &retry.RetryableError{Msg: fmt.Sprintf("reading response: %v", err)}
			continue
		}

		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return "", fmt.Errorf("codex API error %d: %s", resp.StatusCode, string(respBody))
		}

		if resp.StatusCode != http.StatusOK {
			if retry.IsRetryableHTTP(resp.StatusCode) {
				lastErr = &retry.RetryableError{StatusCode: resp.StatusCode, Msg: fmt.Sprintf("codex API %d", resp.StatusCode)}
				continue
			}
			return "", fmt.Errorf("codex API error %d: %s", resp.StatusCode, string(respBody))
		}

		var parsed struct {
			Output []struct {
				Content []struct {
					Text string `json:"text"`
				} `json:"content"`
			} `json:"output"`
		}
		if err := json.Unmarshal(respBody, &parsed); err != nil {
			var chatResp struct {
				Choices []struct {
					Message ChatMessage `json:"message"`
				} `json:"choices"`
			}
			if err2 := json.Unmarshal(respBody, &chatResp); err2 == nil && len(chatResp.Choices) > 0 {
				return chatResp.Choices[0].Message.Content, nil
			}
			return "", fmt.Errorf("malformed response: %w", err)
		}

		var sb strings.Builder
		for _, out := range parsed.Output {
			for _, cnt := range out.Content {
				sb.WriteString(cnt.Text)
			}
		}
		return sb.String(), nil
	}

	return "", fmt.Errorf("retry exhausted (%d attempts): %w", retry.LLMRetry.MaxAttempts, lastErr)
}

func (c *CodexClient) Stream(messages []ChatMessage, ch chan<- StreamDelta) {
	defer close(ch)

	body := c.buildRequest(messages, true)

	jsonBody, err := json.Marshal(body)
	if err != nil {
		ch <- StreamDelta{Err: err}
		return
	}

	url := strings.TrimSuffix(c.BaseURL, "/") + "/codex/responses"

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

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
		if err != nil {
			ch <- StreamDelta{Err: err}
			return
		}

		req.Header.Set("Content-Type", "application/json")
		c.setHeaders(req)

		resp, err := httpClient.Do(req)
		if err != nil {
			lastErr = &retry.RetryableError{Msg: fmt.Sprintf("codex request: %v", err)}
			continue
		}

		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			ch <- StreamDelta{Err: fmt.Errorf("codex API error %d: %s", resp.StatusCode, string(body))}
			return
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if retry.IsRetryableHTTP(resp.StatusCode) {
				lastErr = &retry.RetryableError{StatusCode: resp.StatusCode, Msg: fmt.Sprintf("codex API %d: %s", resp.StatusCode, string(body))}
				continue
			}
			ch <- StreamDelta{Err: fmt.Errorf("codex API error %d: %s", resp.StatusCode, string(body))}
			return
		}

		timeout := time.AfterFunc(streamReadTimeout, func() {
			resp.Body.Close()
		})

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

		streamFailed := false
		currentEvent := ""
		for scanner.Scan() {
			timeout.Reset(streamReadTimeout)

			line := scanner.Text()

			if strings.HasPrefix(line, "event: ") {
				currentEvent = strings.TrimPrefix(line, "event: ")
				continue
			}

			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}

			content, reasoning := parseCodexChunk(data, currentEvent)
			if reasoning != "" {
				ch <- StreamDelta{Reasoning: reasoning}
			}
			if content != "" {
				ch <- StreamDelta{Content: content}
			}
			if currentEvent == "response.completed" {
				break
			}
		}

		timeout.Stop()
		resp.Body.Close()

		if err := scanner.Err(); err != nil {
			lastErr = &retry.RetryableError{Msg: fmt.Sprintf("reading stream: %v", err)}
			streamFailed = true
		}

		if !streamFailed {
			ch <- StreamDelta{Done: true}
			return
		}
	}

	ch <- StreamDelta{Err: fmt.Errorf("retry exhausted (%d attempts): %w", retry.LLMRetry.MaxAttempts, lastErr)}
}

func (c *CodexClient) buildRequest(messages []ChatMessage, stream bool) codexRequest {
	var instructions string
	var input []codexInput

	for _, msg := range messages {
		if msg.Role == "system" {
			instructions = msg.Content
		} else {
			input = append(input, codexInput{Role: msg.Role, Content: msg.Content})
		}
	}

	return codexRequest{
		Model:        c.Model,
		Input:        input,
		Instructions: instructions,
		Stream:       stream,
	}
}

func parseCodexChunk(data, event string) (content, reasoning string) {
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(data), &raw); err != nil {
		return "", ""
	}

	if d, ok := raw["delta"].(string); ok {
		return d, ""
	}

	if choices, ok := raw["choices"].([]interface{}); ok {
		for _, c := range choices {
			if cm, ok := c.(map[string]interface{}); ok {
				if delta, ok := cm["delta"].(map[string]interface{}); ok {
					if c, ok := delta["content"].(string); ok {
						content += c
					}
					if rc, ok := delta["reasoning_content"].(string); ok {
						reasoning += rc
					}
				}
				if fr, ok := cm["finish_reason"].(string); ok && fr != "" {
					return content, reasoning
				}
			}
		}
	}

	if output, ok := raw["output"].([]interface{}); ok {
		for _, o := range output {
			if om, ok := o.(map[string]interface{}); ok {
				if cnt, ok := om["content"].([]interface{}); ok {
					for _, cn := range cnt {
						if cm, ok := cn.(map[string]interface{}); ok {
							if t, ok := cm["text"].(string); ok {
								content += t
							}
						}
					}
				}
			}
		}
	}

	return content, reasoning
}

func extractAccountID(token auth.Token) string {
	if id, ok := token.Extra["account_id"]; ok && id != "" {
		return id
	}

	idToken := token.Extra["id_token"]
	if idToken == "" {
		return ""
	}

	parts := strings.Split(idToken, ".")
	if len(parts) < 2 {
		return ""
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ""
	}

	var claims struct {
		Sub string `json:"sub"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return ""
	}

	return claims.Sub
}
