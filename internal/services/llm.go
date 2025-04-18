package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"ai-language-notes/internal/models"
)

type LLMService interface {
	ProcessText(text string, nativeLanguage, targetLanguage string) (*models.ProcessedContent, error)
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type DeepSeekService struct {
	APIKey     string
	ModelName  string
	HTTPClient HTTPClient
}

func NewDeepSeekService(apiKey string) *DeepSeekService {
	return &DeepSeekService{
		APIKey:     apiKey,
		ModelName:  "deepseek-chat",
		HTTPClient: &http.Client{},
	}
}

type chatCompletionRequest struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type processedContent struct {
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
}

func (s *DeepSeekService) ProcessText(text string, nativeLanguage, targetLanguage string) (*models.ProcessedContent, error) {
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
		nativeLanguage, targetLanguage, text)

	reqBody := chatCompletionRequest{
		Model: s.ModelName,
		Messages: []message{
			{Role: "system", Content: "You are a helpful language learning assistant that responds in JSON format."},
			{Role: "user", Content: prompt},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.deepseek.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.APIKey))

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var completionResponse chatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&completionResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Handle API errors
	if completionResponse.Error != nil {
		return nil, errors.New(completionResponse.Error.Message)
	}

	if len(completionResponse.Choices) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}

	// Parse the LLM response which should be JSON
	content := completionResponse.Choices[0].Message.Content

	// Strip markdown code blocks if present (handles ```json {...} ```)
	if strings.HasPrefix(content, "```") {
		// Find the position of the first opening brace
		startPos := strings.Index(content, "{")
		if startPos != -1 {
			// Find the position of the last closing brace
			endPos := strings.LastIndex(content, "}")
			if endPos != -1 && endPos > startPos {
				content = content[startPos : endPos+1]
			}
		}
	}

	// Use a map for flexible JSON structure instead of a strict struct
	var rawContent map[string]interface{}
	if err := json.Unmarshal([]byte(content), &rawContent); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	// Process the content field which could be a string or an object
	var finalContent string
	if contentVal, exists := rawContent["content"]; exists {
		switch c := contentVal.(type) {
		case string:
			finalContent = c
		case map[string]interface{}: // Handle case where content is an object
			contentBytes, err := json.MarshalIndent(c, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("failed to process content object: %w", err)
			}
			finalContent = string(contentBytes)
		default:
			return nil, fmt.Errorf("unexpected content format in LLM response")
		}
	} else {
		return nil, fmt.Errorf("content field missing in LLM response")
	}

	// Process tags
	var tags []string
	if tagsVal, exists := rawContent["tags"]; exists {
		if tagsSlice, ok := tagsVal.([]interface{}); ok {
			for _, tag := range tagsSlice {
				if tagStr, ok := tag.(string); ok {
					tags = append(tags, tagStr)
				}
			}
		}
	}

	return &models.ProcessedContent{
		Content: finalContent,
		Tags:    tags,
	}, nil
}
