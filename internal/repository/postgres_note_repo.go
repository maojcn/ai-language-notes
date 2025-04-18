package repository

import (
	"ai-language-notes/internal/models"
	"ai-language-notes/internal/storage"
	"fmt"

	"github.com/google/uuid"
)

// NoteRepositoryImpl implements NoteRepository
type NoteRepositoryImpl struct {
	db *storage.PostgresStore
}

// NewNoteRepository creates a new NoteRepository
func NewNoteRepository(db *storage.PostgresStore) NoteRepository {
	return &NoteRepositoryImpl{db: db}
}

// CreateNote creates a new note in the database
func (r *NoteRepositoryImpl) CreateNote(note *models.Note) (*models.Note, error) {
	tx := r.db.GetDB().Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	if err := tx.Create(note).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	// Handle tags if present
	if len(note.Tags) > 0 {
		if err := tx.Model(note).Association("Tags").Replace(note.Tags); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to associate tags: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Reload note with tags
	if err := r.db.GetDB().Preload("Tags").First(note, note.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload note: %w", err)
	}

	return note, nil
}

// GetNoteByID retrieves a note by its ID
func (r *NoteRepositoryImpl) GetNoteByID(id uuid.UUID) (*models.Note, error) {
	var note models.Note
	if err := r.db.GetDB().Preload("Tags").First(&note, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("failed to get note by ID: %w", err)
	}
	return &note, nil
}

// UpdateNote updates an existing note
func (r *NoteRepositoryImpl) UpdateNote(note *models.Note) (*models.Note, error) {
	tx := r.db.GetDB().Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// Update main note fields
	if err := tx.Model(note).Updates(map[string]interface{}{
		"original_text":     note.OriginalText,
		"generated_content": note.GeneratedContent,
		"status":            note.Status,
		"error_message":     note.ErrorMessage,
		"updated_at":        note.UpdatedAt,
	}).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update note: %w", err)
	}

	// Update tags if present
	if note.Tags != nil {
		if err := tx.Model(note).Association("Tags").Replace(note.Tags); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update tags: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Reload note with tags
	if err := r.db.GetDB().Preload("Tags").First(note, note.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload note: %w", err)
	}

	return note, nil
}

// DeleteNote removes a note from the database
func (r *NoteRepositoryImpl) DeleteNote(id uuid.UUID) error {
	if err := r.db.GetDB().Delete(&models.Note{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}
	return nil
}

// GetNotesByUserID retrieves all notes for a specific user
func (r *NoteRepositoryImpl) GetNotesByUserID(userID uuid.UUID) ([]*models.Note, error) {
	var notes []*models.Note
	if err := r.db.GetDB().Preload("Tags").Where("user_id = ?", userID).Find(&notes).Error; err != nil {
		return nil, fmt.Errorf("failed to get notes by user ID: %w", err)
	}
	return notes, nil
}

// FindOrCreateTags finds existing tags or creates new ones
func (r *NoteRepositoryImpl) FindOrCreateTags(tagNames []string) ([]models.Tag, error) {
	var tags []models.Tag
	for _, name := range tagNames {
		var tag models.Tag
		err := r.db.GetDB().Where("name = ?", name).FirstOrCreate(&tag, models.Tag{Name: name}).Error
		if err != nil {
			return nil, fmt.Errorf("failed to find or create tag %s: %w", name, err)
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

// AddTagsToNote adds tags to a note
func (r *NoteRepositoryImpl) AddTagsToNote(noteID uuid.UUID, tags []models.Tag) error {
	note := &models.Note{ID: noteID}
	if err := r.db.GetDB().Model(note).Association("Tags").Append(tags); err != nil {
		return fmt.Errorf("failed to add tags to note: %w", err)
	}
	return nil
}
