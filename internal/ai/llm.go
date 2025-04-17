package ai

import (
	"context"
)

// LLMService defines the interface for LLM integration
type LLMService interface {
	ProcessLanguageNote(ctx context.Context, text, sourceLanguage, targetLanguage string) (string, error)
}

// Message represents a message in a chat API
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
