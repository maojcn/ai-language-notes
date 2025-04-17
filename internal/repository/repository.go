package repository

import (
	"ai-language-notes/internal/models"

	"github.com/google/uuid"
)

type UserRepository interface {
	CreateUser(user *models.User) (*models.User, error)
	GetUserByID(id uuid.UUID) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	UpdateUser(user *models.User) (*models.User, error)
	DeleteUser(id uuid.UUID) error
	GetAllUsers() ([]*models.User, error)
}

type NoteRepository interface {
	CreateNote(note *models.Note) (*models.Note, error)
	GetNoteByID(id uuid.UUID) (*models.Note, error)
	GetNotesByUserID(userID uuid.UUID) ([]*models.Note, error)
	UpdateNote(note *models.Note) (*models.Note, error)
	DeleteNote(id uuid.UUID) error
	UpdateNoteStatus(id uuid.UUID, status models.ProcessingStatus, content string, errorMsg string) error
}

type TagRepository interface {
	GetOrCreateTags(tagNames []string) ([]models.Tag, error)
	GetTagsByNoteID(noteID uuid.UUID) ([]models.Tag, error)
}
