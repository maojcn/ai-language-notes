package dto

import (
	"time"

	"github.com/google/uuid"
)

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
