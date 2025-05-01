package dto

import (
	"ai-language-notes/internal/models"
	"time"

	"github.com/google/uuid"
)

// AddNoteRequest represents the request to add a new language note
type AddNoteRequest struct {
	OriginalText string   `json:"originalText" binding:"required"`
	Tags         []string `json:"tags,omitempty"`
}

// NoteResponse represents the response for note operations
type NoteResponse struct {
	ID               uuid.UUID               `json:"id"`
	OriginalText     string                  `json:"originalText"`
	GeneratedContent string                  `json:"generatedContent,omitempty"`
	Status           models.ProcessingStatus `json:"status"`
	Tags             []string                `json:"tags,omitempty"`
	CreatedAt        time.Time               `json:"createdAt"`
}

// ProcessedContent represents the structured content from LLM processing
type ProcessedContent struct {
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
}
