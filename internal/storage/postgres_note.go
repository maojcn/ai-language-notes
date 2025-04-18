package storage

import (
	"ai-language-notes/internal/models"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateNote creates a new note in the database
func (s *PostgresStore) CreateNote(note *models.Note) (*models.Note, error) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// Create any new tags that don't exist yet
	if note.Tags != nil {
		for i := range note.Tags {
			var existingTag models.Tag
			if err := tx.Where("name = ?", note.Tags[i].Name).First(&existingTag).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					// Tag doesn't exist, create it
					if err := tx.Create(&note.Tags[i]).Error; err != nil {
						tx.Rollback()
						return nil, fmt.Errorf("failed to create tag: %w", err)
					}
				} else {
					tx.Rollback()
					return nil, fmt.Errorf("error checking for existing tag: %w", err)
				}
			} else {
				// Tag exists, use the existing tag
				note.Tags[i] = existingTag
			}
		}
	}

	if err := tx.Create(note).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return note, nil
}

// GetNoteByID retrieves a note by its ID
func (s *PostgresStore) GetNoteByID(id uuid.UUID) (*models.Note, error) {
	var note models.Note
	if err := s.db.Preload("Tags").First(&note, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("failed to get note by ID: %w", err)
	}
	return &note, nil
}

// GetNotesByUserID retrieves all notes for a specific user
func (s *PostgresStore) GetNotesByUserID(userID uuid.UUID) ([]*models.Note, error) {
	var notes []*models.Note
	if err := s.db.Preload("Tags").Where("user_id = ?", userID).Find(&notes).Error; err != nil {
		return nil, fmt.Errorf("failed to get notes by user ID: %w", err)
	}
	return notes, nil
}

// UpdateNote updates an existing note
func (s *PostgresStore) UpdateNote(note *models.Note) (*models.Note, error) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// Get current note to compare tags
	var currentNote models.Note
	if err := tx.Preload("Tags").First(&currentNote, "id = ?", note.ID).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to get current note: %w", err)
	}

	// Update the note fields but not associations yet
	if err := tx.Model(&currentNote).Updates(map[string]interface{}{
		"original_text":     note.OriginalText,
		"generated_content": note.GeneratedContent,
		"status":            note.Status,
		"error_message":     note.ErrorMessage,
	}).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update note: %w", err)
	}

	// Handle tags (if changing)
	if note.Tags != nil {
		// Clear existing tag associations
		if err := tx.Model(&currentNote).Association("Tags").Clear(); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to clear tags: %w", err)
		}

		// Add new tags
		for i := range note.Tags {
			var existingTag models.Tag
			if err := tx.Where("name = ?", note.Tags[i].Name).First(&existingTag).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					// Tag doesn't exist, create it
					if err := tx.Create(&note.Tags[i]).Error; err != nil {
						tx.Rollback()
						return nil, fmt.Errorf("failed to create tag: %w", err)
					}

					// Associate the new tag
					if err := tx.Model(&currentNote).Association("Tags").Append(&note.Tags[i]); err != nil {
						tx.Rollback()
						return nil, fmt.Errorf("failed to associate new tag: %w", err)
					}
				} else {
					tx.Rollback()
					return nil, fmt.Errorf("error checking for existing tag: %w", err)
				}
			} else {
				// Tag exists, use the existing tag
				if err := tx.Model(&currentNote).Association("Tags").Append(&existingTag); err != nil {
					tx.Rollback()
					return nil, fmt.Errorf("failed to associate existing tag: %w", err)
				}
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Reload the note with tags
	var updatedNote models.Note
	if err := s.db.Preload("Tags").First(&updatedNote, "id = ?", note.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload updated note: %w", err)
	}

	return &updatedNote, nil
}

// DeleteNote removes a note from the database
func (s *PostgresStore) DeleteNote(id uuid.UUID) error {
	if err := s.db.Delete(&models.Note{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}
	return nil
}
