package handlers

import (
	"ai-language-notes/internal/ai"
	"ai-language-notes/internal/api/middleware"
	"ai-language-notes/internal/models"
	"ai-language-notes/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// NoteHandler handles note-related requests
type NoteHandler struct {
	noteRepo   repository.NoteRepository
	userRepo   repository.UserRepository
	llmService ai.LLMService
}

// NewNoteHandler creates a new NoteHandler
func NewNoteHandler(
	noteRepo repository.NoteRepository,
	userRepo repository.UserRepository,
	llmService ai.LLMService,
) *NoteHandler {
	return &NoteHandler{
		noteRepo:   noteRepo,
		userRepo:   userRepo,
		llmService: llmService,
	}
}

// CreateNote handles creating a new note with LLM processing
func (h *NoteHandler) CreateNote(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDStr, exists := c.Get(middleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse UUID
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Parse request
	var req models.AddNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user's language preferences
	user, err := h.userRepo.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user information"})
		return
	}

	// Create a new note with pending status
	newNote := &models.Note{
		ID:           uuid.New(),
		UserID:       userID,
		OriginalText: req.OriginalText,
		Status:       models.StatusProcessing,
	}

	// Process the text with the LLM service
	processedContent, err := h.llmService.ProcessText(
		c.Request.Context(),
		req.OriginalText,
		user.NativeLanguage,
		user.TargetLanguage,
	)

	if err != nil {
		// Save note with error status
		newNote.Status = models.StatusFailed
		newNote.ErrorMessage = "Failed to process text: " + err.Error()

		savedNote, createErr := h.noteRepo.CreateNote(newNote)
		if createErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save note", "details": createErr.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process text with LLM",
			"note":  savedNote,
		})
		return
	}

	// Check that we have content and at least one tag
	if processedContent.Content == "" {
		newNote.Status = models.StatusFailed
		newNote.ErrorMessage = "LLM returned empty content"

		savedNote, createErr := h.noteRepo.CreateNote(newNote)
		if createErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save note", "details": createErr.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "LLM returned empty content",
			"note":  savedNote,
		})
		return
	}

	// Update note with processed content
	newNote.GeneratedContent = processedContent.Content
	newNote.Status = models.StatusCompleted

	// Add tags from processed content
	for _, tagName := range processedContent.Tags {
		newNote.Tags = append(newNote.Tags, models.Tag{
			ID:   uuid.New(),
			Name: tagName,
		})
	}

	// Save the note to the database
	savedNote, err := h.noteRepo.CreateNote(newNote)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save note"})
		return
	}

	// Convert to response DTO
	response := convertNoteToResponse(savedNote)

	c.JSON(http.StatusCreated, response)
}

// GetNote retrieves a specific note
func (h *NoteHandler) GetNote(c *gin.Context) {
	noteIDStr := c.Param("id")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid note ID format"})
		return
	}

	// Get user ID from context (set by auth middleware)
	userIDStr, exists := c.Get(middleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	// Fetch note
	note, err := h.noteRepo.GetNoteByID(noteID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	// Check that the note belongs to the authenticated user
	if note.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to access this note"})
		return
	}

	c.JSON(http.StatusOK, convertNoteToResponse(note))
}

// GetUserNotes retrieves all notes for the authenticated user
func (h *NoteHandler) GetUserNotes(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDStr, exists := c.Get(middleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse UUID
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Fetch notes
	notes, err := h.noteRepo.GetNotesByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve notes"})
		return
	}

	// Convert to response DTOs
	var responseNotes []models.NoteResponse
	for _, note := range notes {
		responseNotes = append(responseNotes, convertNoteToResponse(note))
	}

	c.JSON(http.StatusOK, responseNotes)
}

// DeleteNote deletes a note
func (h *NoteHandler) DeleteNote(c *gin.Context) {
	noteIDStr := c.Param("id")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid note ID format"})
		return
	}

	// Get user ID from context (set by auth middleware)
	userIDStr, exists := c.Get(middleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userID, _ := uuid.Parse(userIDStr.(string))

	// Fetch note to verify ownership
	note, err := h.noteRepo.GetNoteByID(noteID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	// Check that the note belongs to the authenticated user
	if note.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this note"})
		return
	}

	// Delete the note
	if err := h.noteRepo.DeleteNote(noteID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete note"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note deleted successfully"})
}

// Helper function to convert Note model to NoteResponse DTO
func convertNoteToResponse(note *models.Note) models.NoteResponse {
	tagNames := make([]string, len(note.Tags))
	for i, tag := range note.Tags {
		tagNames[i] = tag.Name
	}

	return models.NoteResponse{
		ID:               note.ID,
		OriginalText:     note.OriginalText,
		GeneratedContent: note.GeneratedContent,
		Status:           note.Status,
		Tags:             tagNames,
		CreatedAt:        note.CreatedAt,
	}
}
