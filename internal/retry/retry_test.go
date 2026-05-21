package retry

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestIsRetryableHTTP(t *testing.T) {
	tests := []struct {
		code int
		want bool
	}{
		{http.StatusTooManyRequests, true},
		{http.StatusInternalServerError, true},
		{http.StatusBadGateway, true},
		{http.StatusServiceUnavailable, true},
		{http.StatusGatewayTimeout, true},
		{http.StatusUnauthorized, false},
		{http.StatusForbidden, false},
		{http.StatusNotFound, false},
		{http.StatusBadRequest, false},
		{http.StatusOK, false},
		{http.StatusCreated, false},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("HTTP_%d", tt.code), func(t *testing.T) {
			if got := IsRetryableHTTP(tt.code); got != tt.want {
				t.Errorf("IsRetryableHTTP(%d) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"retryable", &RetryableError{Msg: "test"}, true},
		{"retryable with status", &RetryableError{StatusCode: 429, Msg: "test"}, true},
		{"plain error", errors.New("plain"), false},
		{"fmt wrapped", fmt.Errorf("wrapped: %w", &RetryableError{Msg: "base"}), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRetryableError(tt.err); got != tt.want {
				t.Errorf("IsRetryableError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestRetryableError_Error(t *testing.T) {
	e := &RetryableError{StatusCode: 429, Msg: "rate limited"}
	if got := e.Error(); got != "rate limited" {
		t.Errorf("Error() = %q, want %q", got, "rate limited")
	}
}

func TestBackoffDelay_Range(t *testing.T) {
	cfg := Config{MaxAttempts: 3, BaseDelay: 2 * time.Second, MaxDelay: 30 * time.Second}
	delay := BackoffDelay(cfg, 0)
	if delay < 2*time.Second || delay > 3*time.Second {
		t.Errorf("attempt 0 delay %v out of expected range [2s, 3s]", delay)
	}
}

func TestBackoffDelay_CapsAtMax(t *testing.T) {
	cfg := Config{MaxAttempts: 3, BaseDelay: 2 * time.Second, MaxDelay: 30 * time.Second}
	delay := BackoffDelay(cfg, 10)
	maxWithJitter := 30*time.Second + time.Duration(float64(30*time.Second)*0.3)
	if delay > maxWithJitter {
		t.Errorf("delay %v exceeds max with jitter %v", delay, maxWithJitter)
	}
}

func TestBackoffDelay_Increasing(t *testing.T) {
	cfg := Config{MaxAttempts: 3, BaseDelay: 1 * time.Second, MaxDelay: 60 * time.Second}
	var lasts [3]time.Duration
	for i := 0; i < 3; i++ {
		lasts[i] = BackoffDelay(cfg, i)
	}
	if !(lasts[0] < lasts[1] && lasts[1] < lasts[2]) {
		t.Errorf("delays not strictly increasing: %v, %v, %v", lasts[0], lasts[1], lasts[2])
	}
}

func TestLLMRetryConfig(t *testing.T) {
	if LLMRetry.MaxAttempts != 3 {
		t.Errorf("LLMRetry.MaxAttempts = %d, want 3", LLMRetry.MaxAttempts)
	}
	if LLMRetry.BaseDelay != 2*time.Second {
		t.Errorf("LLMRetry.BaseDelay = %v, want 2s", LLMRetry.BaseDelay)
	}
	if LLMRetry.MaxDelay != 30*time.Second {
		t.Errorf("LLMRetry.MaxDelay = %v, want 30s", LLMRetry.MaxDelay)
	}
}

func TestFetchRetryConfig(t *testing.T) {
	if FetchRetry.MaxAttempts != 3 {
		t.Errorf("FetchRetry.MaxAttempts = %d, want 3", FetchRetry.MaxAttempts)
	}
	if FetchRetry.BaseDelay != 1*time.Second {
		t.Errorf("FetchRetry.BaseDelay = %v, want 1s", FetchRetry.BaseDelay)
	}
	if FetchRetry.MaxDelay != 10*time.Second {
		t.Errorf("FetchRetry.MaxDelay = %v, want 10s", FetchRetry.MaxDelay)
	}
}
