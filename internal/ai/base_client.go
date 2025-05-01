package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// BaseLLMClient provides common functionality for LLM clients
type BaseLLMClient struct {
	APIKey      string
	APIEndpoint string
	ModelName   string
	Provider    string
	HTTPClient  HTTPClient
	Metrics     *LLMMetrics
	RetryConfig RetryConfig
}

// HTTPClient interface abstracts the HTTP client
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// RetryConfig defines retry behavior
type RetryConfig struct {
	MaxRetries  int
	InitialWait time.Duration
	MaxWait     time.Duration
}

// DefaultRetryConfig provides sensible defaults for retry behavior
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:  3,
		InitialWait: 500 * time.Millisecond,
		MaxWait:     10 * time.Second,
	}
}

// LLMMetrics holds Prometheus metrics for LLM operations
type LLMMetrics struct {
	RequestDuration *prometheus.HistogramVec
	RequestCounter  *prometheus.CounterVec
}

// NewHTTPClient creates a HTTP client with proper timeouts
func NewHTTPClient(timeout time.Duration) HTTPClient {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxConnsPerHost:     5,
			MaxIdleConnsPerHost: 5,
			IdleConnTimeout:     90 * time.Second,
		},
	}
}

// Message represents a message in a conversation with an LLM
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// SendRequest handles HTTP requests with metrics and retry logic
func (c *BaseLLMClient) SendRequest(ctx context.Context, requestBody interface{}, responseObj interface{}) error {
	var err error
	start := time.Now()
	status := "success"

	// Wrap the actual request in retry logic
	err = withRetry(ctx, c.RetryConfig, func() error {
		return c.doSendRequest(ctx, requestBody, responseObj)
	})

	// Record metrics
	duration := time.Since(start).Seconds()
	if err != nil {
		status = "error"
	}

	if c.Metrics != nil {
		c.Metrics.RequestDuration.WithLabelValues(c.Provider, status).Observe(duration)
		c.Metrics.RequestCounter.WithLabelValues(c.Provider, status).Inc()
	}

	return err
}

// doSendRequest performs the actual HTTP request without retries
func (c *BaseLLMClient) doSendRequest(ctx context.Context, requestBody interface{}, responseObj interface{}) error {
	// Marshal the request body to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", c.APIEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set common headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	// Send the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		// Check for context cancellation
		if ctx.Err() == context.DeadlineExceeded {
			return NewLLMError(http.StatusRequestTimeout, "request timed out", c.Provider, true)
		}
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for non-successful status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errorMsg := extractErrorMessage(bodyBytes)
		retryable := resp.StatusCode >= 500 || resp.StatusCode == 429

		return NewLLMError(resp.StatusCode, errorMsg, c.Provider, retryable)
	}

	// Parse the response into the provided object
	if err := json.Unmarshal(bodyBytes, responseObj); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

// withRetry implements retry logic with exponential backoff
func withRetry(ctx context.Context, config RetryConfig, fn func() error) error {
	var err error
	wait := config.InitialWait

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}

		// Only retry on retryable errors
		if !IsRetryableError(err) {
			return err
		}

		// Check if we should retry
		if attempt >= config.MaxRetries {
			break
		}

		// Wait with exponential backoff
		select {
		case <-time.After(wait):
			// Exponential backoff with jitter
			wait = time.Duration(float64(wait) * 1.5)
			if wait > config.MaxWait {
				wait = config.MaxWait
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return err
}

// extractErrorMessage tries to extract a meaningful error message from API response
func extractErrorMessage(bodyBytes []byte) string {
	var errorResponse map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &errorResponse); err != nil {
		return "unknown error occurred"
	}

	// Try various common error message formats
	if errObj, ok := errorResponse["error"].(map[string]interface{}); ok {
		if msg, ok := errObj["message"].(string); ok {
			return msg
		}
	}

	if errMsg, ok := errorResponse["error"].(string); ok {
		return errMsg
	}

	if msg, ok := errorResponse["message"].(string); ok {
		return msg
	}

	return "unknown error occurred"
}
