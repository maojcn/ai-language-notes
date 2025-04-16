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

// --- DTOs (Data Transfer Objects) for API requests/responses ---
type AuthResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	UserID    uuid.UUID `json:"user_id"`
}

type RegisterRequest struct {
	Username       string `json:"username" binding:"required,min=3,max=50"`
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"required,min=6"`
	NativeLanguage string `json:"nativeLanguage" binding:"required,len=2"` // e.g., "en"
	TargetLanguage string `json:"targetLanguage" binding:"required,len=2"` // e.g., "de"
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type UserProfileUpdateRequest struct {
	NativeLanguage *string `json:"nativeLanguage,omitempty" binding:"omitempty,len=2"`
	TargetLanguage *string `json:"targetLanguage,omitempty" binding:"omitempty,len=2"`
}
