package llm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ibrhr/qdoc/internal/auth"
	"github.com/ibrhr/qdoc/internal/retry"
)

var httpClient = &http.Client{
	Timeout: 120 * time.Second,
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIClient struct {
	Auth     auth.AuthMethod
	Token    auth.Token
	BaseURL  string
	Model    string
	Provider string
	Headers  map[string]string
}

func (c *OpenAIClient) ModelName() string    { return c.Model }
func (c *OpenAIClient) ProviderName() string { return c.Provider }

func NewClient(apiType string, cfg Config) (Client, error) {
	if apiType == "" {
		apiType = "openai-compat"
	}
	switch apiType {
	case "openai-compat":
		return &OpenAIClient{
			Auth:     cfg.Auth,
			Token:    cfg.Token,
			BaseURL:  cfg.BaseURL,
			Model:    cfg.Model,
			Provider: cfg.Provider,
			Headers:  cfg.Headers,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported api type: %s", apiType)
	}
}

type streamRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type streamChunk struct {
	Choices []struct {
		Delta struct {
			Content          string `json:"content"`
			ReasoningContent string `json:"reasoning_content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

const (
	streamReadTimeout = 60 * time.Second
)

func (c *OpenAIClient) setHeaders(req *http.Request) {
	c.Auth.ApplyAuth(req, c.Token)
	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}
}

func (c *OpenAIClient) Send(messages []ChatMessage) (string, error) {
	body := streamRequest{
		Model:    c.Model,
		Messages: messages,
		Stream:   false,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	url := strings.TrimSuffix(c.BaseURL, "/") + "/chat/completions"

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
			lastErr = &retry.RetryableError{Msg: fmt.Sprintf("llm request: %v", err)}
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = &retry.RetryableError{Msg: fmt.Sprintf("reading response: %v", err)}
			continue
		}

		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return "", fmt.Errorf("llm API error %d: %s", resp.StatusCode, string(respBody))
		}

		if resp.StatusCode != http.StatusOK {
			if retry.IsRetryableHTTP(resp.StatusCode) {
				lastErr = &retry.RetryableError{StatusCode: resp.StatusCode, Msg: fmt.Sprintf("llm API %d", resp.StatusCode)}
				continue
			}
			return "", fmt.Errorf("llm API error %d: %s", resp.StatusCode, string(respBody))
		}

		var chatResp struct {
			Choices []struct {
				Message ChatMessage `json:"message"`
			} `json:"choices"`
		}
		if err := json.Unmarshal(respBody, &chatResp); err != nil {
			return "", fmt.Errorf("malformed response: %w", err)
		}

		if len(chatResp.Choices) == 0 {
			return "", fmt.Errorf("no response from LLM")
		}

		return chatResp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("retry exhausted (%d attempts): %w", retry.LLMRetry.MaxAttempts, lastErr)
}

func (c *OpenAIClient) Stream(messages []ChatMessage, ch chan<- StreamDelta) {
	defer close(ch)

	body := streamRequest{
		Model:    c.Model,
		Messages: messages,
		Stream:   true,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		ch <- StreamDelta{Err: err}
		return
	}

	url := strings.TrimSuffix(c.BaseURL, "/") + "/chat/completions"

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
			lastErr = &retry.RetryableError{Msg: fmt.Sprintf("llm request: %v", err)}
			continue
		}

		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			ch <- StreamDelta{Err: fmt.Errorf("llm API error %d: %s", resp.StatusCode, string(body))}
			return
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if retry.IsRetryableHTTP(resp.StatusCode) {
				lastErr = &retry.RetryableError{StatusCode: resp.StatusCode, Msg: fmt.Sprintf("llm API %d: %s", resp.StatusCode, string(body))}
				continue
			}
			ch <- StreamDelta{Err: fmt.Errorf("llm API error %d: %s", resp.StatusCode, string(body))}
			return
		}

		timeout := time.AfterFunc(streamReadTimeout, func() {
			resp.Body.Close()
		})

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

		streamFailed := false
		for scanner.Scan() {
			timeout.Reset(streamReadTimeout)

			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}

			var chunk streamChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}

			if len(chunk.Choices) > 0 {
				reasoning := chunk.Choices[0].Delta.ReasoningContent
				if reasoning != "" {
					ch <- StreamDelta{Reasoning: reasoning}
				}
				content := chunk.Choices[0].Delta.Content
				if content != "" {
					ch <- StreamDelta{Content: content}
				}
				if chunk.Choices[0].FinishReason != nil {
					break
				}
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
