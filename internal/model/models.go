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
