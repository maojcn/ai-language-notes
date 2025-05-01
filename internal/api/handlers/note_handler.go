package handlers

import (
	"ai-language-notes/internal/ai"
	"ai-language-notes/internal/api/dto"
	"ai-language-notes/internal/api/middleware"
	"ai-language-notes/internal/models"
	"ai-language-notes/internal/queue"
	"ai-language-notes/internal/repository"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// NoteHandler handles note-related requests
type NoteHandler struct {
	noteRepo     repository.NoteRepository
	userRepo     repository.UserRepository
	llmService   ai.LLMService
	queueService *queue.QueueService
}

// NewNoteHandler creates a new NoteHandler
func NewNoteHandler(
	noteRepo repository.NoteRepository,
	userRepo repository.UserRepository,
	llmService ai.LLMService,
	queueService *queue.QueueService,
) *NoteHandler {
	return &NoteHandler{
		noteRepo:     noteRepo,
		userRepo:     userRepo,
		llmService:   llmService,
		queueService: queueService,
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
		Status:       models.StatusPending, // Start with pending status
	}

	// Save the note to the database with pending status
	savedNote, err := h.noteRepo.CreateNote(newNote)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save note"})
		return
	}

	// Create a task for async processing
	task := &queue.LLMProcessingTask{
		NoteID:         savedNote.ID,
		OriginalText:   savedNote.OriginalText,
		UserID:         userID,
		NativeLanguage: user.NativeLanguage,
		TargetLanguage: user.TargetLanguage,
		CreatedAt:      time.Now(),
	}

	// Enqueue the task for processing
	err = h.queueService.EnqueueTask(c.Request.Context(), task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to enqueue task for processing",
			"note":  convertNoteToResponse(savedNote),
		})
		return
	}

	// Return the note with pending status immediately
	c.JSON(http.StatusAccepted, convertNoteToResponse(savedNote))
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
