package ai

import (
	"fmt"
)

// NewLLMService creates an LLM service based on the provided configuration
func NewLLMService(config LLMServiceConfig) (LLMService, error) {
	switch config.ProviderType {
	case ProviderOpenAI:
		if config.APIKey == "" {
			return nil, fmt.Errorf("OpenAI API key is required")
		}
		return NewOpenAIService(config), nil

	case ProviderDeepseek:
		if config.APIKey == "" {
			return nil, fmt.Errorf("Deepseek API key is required")
		}
		return NewDeepseekService(config), nil

	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", config.ProviderType)
	}
}

// CreateLLMServiceFromConfig creates an LLM service from provider name and API key
// This is a convenience function for simpler configuration
func CreateLLMServiceFromConfig(providerName string, apiKey string) (LLMService, error) {
	var providerType ProviderType

	switch providerName {
	case "openai":
		providerType = ProviderOpenAI
	case "deepseek":
		providerType = ProviderDeepseek
	default:
		return nil, fmt.Errorf("unsupported provider name: %s", providerName)
	}

	config := LLMServiceConfig{
		APIKey:       apiKey,
		ProviderType: providerType,
		MaxRetries:   3,
		Timeout:      30,
	}

	return NewLLMService(config)
}
