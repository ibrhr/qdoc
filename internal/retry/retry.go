package retry

import (
	"errors"
	"math"
	"math/rand"
	"net/http"
	"time"
)

type Config struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

var (
	LLMRetry = Config{
		MaxAttempts: 3,
		BaseDelay:   2 * time.Second,
		MaxDelay:    30 * time.Second,
	}

	FetchRetry = Config{
		MaxAttempts: 3,
		BaseDelay:   1 * time.Second,
		MaxDelay:    10 * time.Second,
	}
)

func BackoffDelay(rc Config, attempt int) time.Duration {
	delay := time.Duration(float64(rc.BaseDelay) * math.Pow(2, float64(attempt)))
	if delay > rc.MaxDelay {
		delay = rc.MaxDelay
	}
	return delay + time.Duration(float64(delay)*0.3*rand.Float64())
}

func IsRetryableHTTP(statusCode int) bool {
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

type RetryableError struct {
	StatusCode int
	Msg        string
}

func (e *RetryableError) Error() string {
	return e.Msg
}

func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}
	var retryableErr *RetryableError
	return errors.As(err, &retryableErr)
}
