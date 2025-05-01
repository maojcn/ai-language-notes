package ai

import (
	"ai-language-notes/internal/api/dto"
	"context"
)

// LLMService defines the unified interface for LLM providers
type LLMService interface {
	// ProcessText processes text input and returns structured content
	ProcessText(ctx context.Context, text, sourceLanguage, targetLanguage string) (*dto.ProcessedContent, error)
}

// ProviderType represents the type of LLM provider
type ProviderType string

const (
	ProviderOpenAI   ProviderType = "openai"
	ProviderDeepseek ProviderType = "deepseek"
)

// LLMServiceConfig contains configuration for LLM services
type LLMServiceConfig struct {
	APIKey       string
	ModelName    string
	MaxRetries   int
	Timeout      int // in seconds
	ProviderType ProviderType
}
