package ai

import (
	"ai-language-notes/internal/models"
	"context"
	"fmt"
	"time"
)

// DeepseekService implements LLMService using Deepseek's API
type DeepseekService struct {
	BaseClient *BaseLLMClient
}

// DeepseekChatCompletionRequest represents a request to Deepseek Chat API
type DeepseekChatCompletionRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// DeepseekChatCompletionResponse represents a response from Deepseek Chat API
type DeepseekChatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// NewDeepseekService creates a new Deepseek service
func NewDeepseekService(config LLMServiceConfig) *DeepseekService {
	metrics := NewLLMMetrics()

	retryConfig := DefaultRetryConfig()
	if config.MaxRetries > 0 {
		retryConfig.MaxRetries = config.MaxRetries
	}

	timeout := 30 * time.Second
	if config.Timeout > 0 {
		timeout = time.Duration(config.Timeout) * time.Second
	}

	modelName := "deepseek-chat"
	if config.ModelName != "" {
		modelName = config.ModelName
	}

	baseClient := &BaseLLMClient{
		APIKey:      config.APIKey,
		APIEndpoint: "https://api.deepseek.com/v1/chat/completions",
		ModelName:   modelName,
		Provider:    "deepseek",
		HTTPClient:  NewHTTPClient(timeout),
		Metrics:     metrics,
		RetryConfig: retryConfig,
	}

	return &DeepseekService{
		BaseClient: baseClient,
	}
}

// ProcessText implements LLMService.ProcessText
func (s *DeepseekService) ProcessText(ctx context.Context, text, sourceLanguage, targetLanguage string) (*models.ProcessedContent, error) {
	prompt := fmt.Sprintf(
		`You are a language learning assistant. A user who speaks %s is learning %s.
They provided this text: "%s"

Please analyze this text and provide:
1. A breakdown of interesting vocabulary and grammar points
2. A brief explanation of cultural context if relevant
3. Alternative ways to express the same idea
4. Common mistakes learners might make with this phrase

Format your response as a JSON object with two fields:
- "content": detailed educational content about the text
- "tags": an array of 3-5 relevant tags (single words only) for categorizing this note

JSON response only, no additional text.`,
		sourceLanguage, targetLanguage, text)

	// Create request
	request := DeepseekChatCompletionRequest{
		Model: s.BaseClient.ModelName,
		Messages: []Message{
			{Role: "system", Content: "You are a helpful language learning assistant that responds in JSON format."},
			{Role: "user", Content: prompt},
		},
	}

	// Send request
	var response DeepseekChatCompletionResponse
	if err := s.BaseClient.SendRequest(ctx, request, &response); err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}

	// Process response
	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response choices returned from API")
	}

	content := response.Choices[0].Message.Content

	// Parse the content
	parsedContent, err := ParseJSONContent(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse content: %w", err)
	}

	return parsedContent, nil
}
