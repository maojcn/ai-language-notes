package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system.
type User struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Username       string    `gorm:"varchar(255);uniqueIndex;not null" json:"username"`
	Email          string    `gorm:"varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash   string    `gorm:"varchar(255);not null" json:"-"` // Never expose hash
	NativeLanguage string    `gorm:"varchar(10);not null" json:"nativeLanguage"`
	TargetLanguage string    `gorm:"varchar(10);not null" json:"targetLanguage"`
	CreatedAt      time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

// Note represents a language learning note
type Note struct {
	ID               uuid.UUID        `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID           uuid.UUID        `gorm:"type:uuid;not null" json:"userId"`
	User             User             `gorm:"foreignKey:UserID" json:"-"`
	OriginalText     string           `gorm:"type:text;not null" json:"originalText"`
	GeneratedContent string           `gorm:"type:text" json:"generatedContent,omitempty"`
	Status           ProcessingStatus `gorm:"type:processing_status;not null;default:'pending'" json:"status"`
	ErrorMessage     string           `gorm:"type:text" json:"errorMessage,omitempty"`
	Tags             []Tag            `gorm:"many2many:note_tags;" json:"tags,omitempty"`
	CreatedAt        time.Time        `gorm:"default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt        time.Time        `gorm:"default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

// Tag represents a note tag
type Tag struct {
	ID    uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name  string    `gorm:"varchar(100);uniqueIndex;not null" json:"name"`
	Notes []Note    `gorm:"many2many:note_tags;" json:"-"`
}

// ProcessingStatus is the status of a note's AI processing
type ProcessingStatus string

const (
	StatusPending    ProcessingStatus = "pending"
	StatusProcessing ProcessingStatus = "processing"
	StatusCompleted  ProcessingStatus = "completed"
	StatusFailed     ProcessingStatus = "failed"
)
