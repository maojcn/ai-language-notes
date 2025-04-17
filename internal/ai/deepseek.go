package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// DeepseekService implements LLMService using Deepseek's API
type DeepseekService struct {
	apiKey     string
	apiURL     string
	httpClient *http.Client
}

// DeepseekChatCompletionRequest represents a request to the Deepseek Chat API
type DeepseekChatCompletionRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// DeepseekChatCompletionResponse represents a response from the Deepseek Chat API
type DeepseekChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

// NewDeepseekService creates a new Deepseek service
func NewDeepseekService(apiKey string) *DeepseekService {
	return &DeepseekService{
		apiKey: apiKey,
		apiURL: "https://api.deepseek.com/v1/chat/completions", // Update with actual Deepseek API endpoint
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProcessLanguageNote processes a language note using the Deepseek API
func (s *DeepseekService) ProcessLanguageNote(ctx context.Context, text, sourceLanguage, targetLanguage string) (string, error) {
	// Create the system prompt with language information
	systemPrompt := fmt.Sprintf(
		"You are a language learning assistant helping someone learn %s. Their native language is %s. "+
			"Analyze the provided text in %s and provide the following in your response:\n"+
			"1. Grammar corrections with explanations\n"+
			"2. Vocabulary suggestions and alternatives\n"+
			"3. Cultural context and nuance explanations\n"+
			"4. Example sentences using key phrases from the text\n"+
			"Format your response in clear sections with Markdown headings.",
		targetLanguage, sourceLanguage, targetLanguage,
	)

	// Create the request payload
	requestBody := DeepseekChatCompletionRequest{
		Model: "deepseek-chat", // Update with the appropriate Deepseek model name
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: text},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", s.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	// Send the request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
			return "", fmt.Errorf("API request failed with status %d", resp.StatusCode)
		}
		return "", fmt.Errorf("API request failed: %v", errorResponse)
	}

	// Parse the response
	var completionResponse DeepseekChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&completionResponse); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract the generated content
	if len(completionResponse.Choices) == 0 {
		return "", fmt.Errorf("no completion choices returned")
	}

	return completionResponse.Choices[0].Message.Content, nil
}
