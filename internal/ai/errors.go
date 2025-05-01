package ai

import (
	"errors"
	"fmt"
)

// Common errors
var (
	ErrInvalidRequest  = errors.New("invalid request")
	ErrAPIRateLimited  = errors.New("API rate limited")
	ErrModelOverloaded = errors.New("model overloaded")
	ErrRequestTimeout  = errors.New("request timeout")
	ErrUnavailable     = errors.New("service unavailable")
	ErrInvalidResponse = errors.New("invalid response")
)

// LLMError represents an error from an LLM provider
type LLMError struct {
	StatusCode int
	Message    string
	Provider   string
	Retryable  bool
}

// Error implements the error interface
func (e *LLMError) Error() string {
	return fmt.Sprintf("%s API error (HTTP %d): %s", e.Provider, e.StatusCode, e.Message)
}

// NewLLMError creates a new LLM error
func NewLLMError(statusCode int, message, provider string, retryable bool) *LLMError {
	return &LLMError{
		StatusCode: statusCode,
		Message:    message,
		Provider:   provider,
		Retryable:  retryable,
	}
}

// IsRetryableError checks if an error should be retried
func IsRetryableError(err error) bool {
	var llmErr *LLMError
	if errors.As(err, &llmErr) {
		return llmErr.Retryable
	}
	return false
}
