package handlers

import (
	"ai-language-notes/internal/api/dto"
	"ai-language-notes/internal/api/middleware"
	"ai-language-notes/internal/models"
	"ai-language-notes/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// NoteHandler handles note-related requests
type NoteHandler struct {
	noteService services.NoteService
}

// NewNoteHandler creates a new NoteHandler
func NewNoteHandler(noteService services.NoteService) *NoteHandler {
	return &NoteHandler{
		noteService: noteService,
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
	var req dto.AddNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Use the service to create the note
	note, err := h.noteService.CreateNote(c.Request.Context(), userID, req.OriginalText)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save note"})
		return
	}

	// Return the note with pending status immediately
	c.JSON(http.StatusAccepted, convertNoteToResponse(note))
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
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Use the service to get the note
	note, err := h.noteService.GetNoteByID(noteID, userID)
	if err != nil {
		if err == services.ErrNotAuthorized {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to access this note"})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
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

	// Use the service to fetch notes
	notes, err := h.noteService.GetNotesByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve notes"})
		return
	}

	// Convert to response DTOs
	var responseNotes []dto.NoteResponse
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
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Use the service to delete the note
	err = h.noteService.DeleteNote(noteID, userID)
	if err != nil {
		if err == services.ErrNotAuthorized {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this note"})
			return
		}
		if err == services.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete note"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Note deleted successfully"})
}

// Helper function to convert Note model to NoteResponse DTO
func convertNoteToResponse(note *models.Note) dto.NoteResponse {
	tagNames := make([]string, len(note.Tags))
	for i, tag := range note.Tags {
		tagNames[i] = tag.Name
	}

	return dto.NoteResponse{
		ID:               note.ID,
		OriginalText:     note.OriginalText,
		GeneratedContent: note.GeneratedContent,
		Status:           note.Status,
		Tags:             tagNames,
		CreatedAt:        note.CreatedAt,
	}
}
