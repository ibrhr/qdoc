package llm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

var httpClient = &http.Client{
	Timeout: 120 * time.Second,
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Client struct {
	APIKey   string
	BaseURL  string
	Model    string
	Provider string
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

type retryConfig struct {
	maxAttempts int
	baseDelay   time.Duration
	maxDelay    time.Duration
}

type retryableError struct {
	statusCode int
	msg        string
}

func (e *retryableError) Error() string {
	return e.msg
}

var defaultRetry = retryConfig{
	maxAttempts: 3,
	baseDelay:   2 * time.Second,
	maxDelay:    30 * time.Second,
}

const (
	streamReadTimeout = 60 * time.Second
)

func backoffDelay(rc retryConfig, attempt int) time.Duration {
	delay := rc.baseDelay
	for i := 0; i < attempt; i++ {
		delay *= 2
		if delay > rc.maxDelay {
			delay = rc.maxDelay
			break
		}
	}
	return delay + time.Duration(float64(delay)*0.3*rand.Float64())
}

func isRetryableHTTP(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}
	var retryableErr *retryableError
	return errors.As(err, &retryableErr)
}

func (c *Client) Send(messages []ChatMessage) (string, error) {
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
	for attempt := 0; attempt < defaultRetry.maxAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(backoffDelay(defaultRetry, attempt-1))
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
		if err != nil {
			return "", err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.APIKey)

		resp, err := httpClient.Do(req)
		if err != nil {
			lastErr = &retryableError{msg: fmt.Sprintf("llm request: %v", err)}
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = &retryableError{msg: fmt.Sprintf("reading response: %v", err)}
			continue
		}

		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return "", fmt.Errorf("llm API error %d: %s", resp.StatusCode, string(respBody))
		}

		if resp.StatusCode != http.StatusOK {
			if isRetryableHTTP(resp.StatusCode) {
				lastErr = &retryableError{statusCode: resp.StatusCode, msg: fmt.Sprintf("llm API %d", resp.StatusCode)}
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

	return "", fmt.Errorf("retry exhausted (%d attempts): %w", defaultRetry.maxAttempts, lastErr)
}

func (c *Client) Stream(messages []ChatMessage, ch chan<- StreamDelta) {
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
	for attempt := 0; attempt < defaultRetry.maxAttempts; attempt++ {
		if attempt > 0 {
			delay := backoffDelay(defaultRetry, attempt-1)

			ch <- StreamDelta{
				Content: fmt.Sprintf("(retrying in %.1fs — attempt %d/%d)",
					delay.Seconds(), attempt+1, defaultRetry.maxAttempts),
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
		req.Header.Set("Authorization", "Bearer "+c.APIKey)

		resp, err := httpClient.Do(req)
		if err != nil {
			lastErr = &retryableError{msg: fmt.Sprintf("llm request: %v", err)}
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
			if isRetryableHTTP(resp.StatusCode) {
				lastErr = &retryableError{statusCode: resp.StatusCode, msg: fmt.Sprintf("llm API %d: %s", resp.StatusCode, string(body))}
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
			lastErr = &retryableError{msg: fmt.Sprintf("reading stream: %v", err)}
			streamFailed = true
		}

		if !streamFailed {
			ch <- StreamDelta{Done: true}
			return
		}
	}

	ch <- StreamDelta{Err: fmt.Errorf("retry exhausted (%d attempts): %w", defaultRetry.maxAttempts, lastErr)}
}