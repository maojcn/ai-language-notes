package ai

import (
	"ai-language-notes/internal/api/dto"
	"context"
	"fmt"
	"time"
)

// OpenAIService implements LLMService using OpenAI's API
type OpenAIService struct {
	BaseClient *BaseLLMClient
}

// OpenAIChatCompletionRequest represents a request to OpenAI Chat API
type OpenAIChatCompletionRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// OpenAIChatCompletionResponse represents a response from OpenAI Chat API
type OpenAIChatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// NewOpenAIService creates a new OpenAI service
func NewOpenAIService(config LLMServiceConfig) *OpenAIService {
	metrics := NewLLMMetrics()

	retryConfig := DefaultRetryConfig()
	if config.MaxRetries > 0 {
		retryConfig.MaxRetries = config.MaxRetries
	}

	timeout := 30 * time.Second
	if config.Timeout > 0 {
		timeout = time.Duration(config.Timeout) * time.Second
	}

	modelName := "gpt-4"
	if config.ModelName != "" {
		modelName = config.ModelName
	}

	baseClient := &BaseLLMClient{
		APIKey:      config.APIKey,
		APIEndpoint: "https://api.openai.com/v1/chat/completions",
		ModelName:   modelName,
		Provider:    "openai",
		HTTPClient:  NewHTTPClient(timeout),
		Metrics:     metrics,
		RetryConfig: retryConfig,
	}

	return &OpenAIService{
		BaseClient: baseClient,
	}
}

// ProcessText implements LLMService.ProcessText
func (s *OpenAIService) ProcessText(ctx context.Context, text, sourceLanguage, targetLanguage string) (*dto.ProcessedContent, error) {
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
	request := OpenAIChatCompletionRequest{
		Model: s.BaseClient.ModelName,
		Messages: []Message{
			{Role: "system", Content: "You are a helpful language learning assistant that responds in JSON format."},
			{Role: "user", Content: prompt},
		},
	}

	// Send request
	var response OpenAIChatCompletionResponse
	if err := s.BaseClient.SendRequest(ctx, request, &response); err != nil {
		return nil, err
	}

	// Process response
	if len(response.Choices) == 0 {
		return nil, ErrInvalidResponse
	}

	content := response.Choices[0].Message.Content

	return ParseJSONContent(content)
}
