package services

import (
	"ai-language-notes/internal/ai"
	"ai-language-notes/internal/models"
	"ai-language-notes/internal/queue"
	"ai-language-notes/internal/repository"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Common service errors
var (
	ErrNotAuthorized = errors.New("not authorized to access this resource")
	ErrNotFound      = errors.New("resource not found")
)

// NoteService defines the interface for note-related business logic
type NoteService interface {
	CreateNote(ctx context.Context, userID uuid.UUID, originalText string) (*models.Note, error)
	GetNoteByID(noteID uuid.UUID, userID uuid.UUID) (*models.Note, error)
	GetNotesByUserID(userID uuid.UUID) ([]*models.Note, error)
	DeleteNote(noteID uuid.UUID, userID uuid.UUID) error
}

// NoteServiceImpl implements the NoteService interface
type NoteServiceImpl struct {
	noteRepo     repository.NoteRepository
	userRepo     repository.UserRepository
	llmService   ai.LLMService
	queueService *queue.QueueService
}

// NewNoteService creates a new instance of NoteService
func NewNoteService(
	noteRepo repository.NoteRepository,
	userRepo repository.UserRepository,
	llmService ai.LLMService,
	queueService *queue.QueueService,
) NoteService {
	return &NoteServiceImpl{
		noteRepo:     noteRepo,
		userRepo:     userRepo,
		llmService:   llmService,
		queueService: queueService,
	}
}

// CreateNote handles the business logic for creating a new note
func (s *NoteServiceImpl) CreateNote(ctx context.Context, userID uuid.UUID, originalText string) (*models.Note, error) {
	// Get user's language preferences
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	// Create a new note with pending status
	newNote := &models.Note{
		ID:           uuid.New(),
		UserID:       userID,
		OriginalText: originalText,
		Status:       models.StatusPending,
	}

	// Save the note to the database
	savedNote, err := s.noteRepo.CreateNote(newNote)
	if err != nil {
		return nil, err
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
	if err = s.queueService.EnqueueTask(ctx, task); err != nil {
		return savedNote, err // Return note even if queueing fails
	}

	return savedNote, nil
}

// GetNoteByID retrieves a specific note and verifies ownership
func (s *NoteServiceImpl) GetNoteByID(noteID uuid.UUID, userID uuid.UUID) (*models.Note, error) {
	note, err := s.noteRepo.GetNoteByID(noteID)
	if err != nil {
		return nil, ErrNotFound
	}

	// Verify the note belongs to the user
	if note.UserID != userID {
		return nil, ErrNotAuthorized
	}

	return note, nil
}

// GetNotesByUserID retrieves all notes for a specific user
func (s *NoteServiceImpl) GetNotesByUserID(userID uuid.UUID) ([]*models.Note, error) {
	return s.noteRepo.GetNotesByUserID(userID)
}

// DeleteNote deletes a note after verifying ownership
func (s *NoteServiceImpl) DeleteNote(noteID uuid.UUID, userID uuid.UUID) error {
	note, err := s.noteRepo.GetNoteByID(noteID)
	if err != nil {
		return ErrNotFound
	}

	// Verify the note belongs to the user
	if note.UserID != userID {
		return ErrNotAuthorized
	}

	return s.noteRepo.DeleteNote(noteID)
}
