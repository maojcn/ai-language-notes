package services

import (
	"ai-language-notes/internal/api/dto"
	"ai-language-notes/internal/auth"
	"ai-language-notes/internal/config"
	"ai-language-notes/internal/models"
	"ai-language-notes/internal/repository"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Common authentication errors
var (
	ErrUsernameExists     = errors.New("username already exists")
	ErrEmailExists        = errors.New("email already in use")
	ErrPasswordProcessing = errors.New("failed to process password")
	ErrUserCreation       = errors.New("failed to create user")
	ErrTokenGeneration    = errors.New("failed to generate authentication token")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// AuthService handles authentication-related business logic
type AuthService struct {
	userRepo repository.UserRepository
	config   config.Config
}

// NewAuthService creates a new AuthService instance
func NewAuthService(userRepo repository.UserRepository, cfg config.Config) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		config:   cfg,
	}
}

// Register handles user registration logic
func (s *AuthService) Register(req dto.RegisterRequest) (*dto.AuthResponse, error) {
	// Start a transaction
	tx := s.userRepo.(*repository.UserRepositoryImpl).GetDB().Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// Check if user already exists within the transaction
	var existingUser *models.User
	if err := tx.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		tx.Rollback()
		return nil, ErrUsernameExists
	}

	if err := tx.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		tx.Rollback()
		return nil, ErrEmailExists
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		tx.Rollback()
		return nil, ErrPasswordProcessing
	}

	// Create user
	newUser := &models.User{
		Username:       req.Username,
		Email:          req.Email,
		PasswordHash:   hashedPassword,
		ID:             uuid.New(),
		NativeLanguage: req.NativeLanguage,
		TargetLanguage: req.TargetLanguage,
	}

	if err := tx.Create(newUser).Error; err != nil {
		tx.Rollback()
		return nil, ErrUserCreation
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Generate token
	token, err := auth.GenerateToken(newUser.ID, s.config.JWTSecret, s.config.JWTExpirationTime)
	if err != nil {
		return nil, ErrTokenGeneration
	}

	return &dto.AuthResponse{
		Token:     token,
		ExpiresAt: time.Now().Add(s.config.JWTExpirationTime),
		UserID:    newUser.ID,
	}, nil
}

// Login handles user login logic
func (s *AuthService) Login(req dto.LoginRequest) (*dto.AuthResponse, error) {
	// Try to find user by username or email
	var user *models.User
	var err error

	// Check if input is email (contains @)
	if strings.Contains(req.Email, "@") {
		user, err = s.userRepo.GetUserByEmail(req.Email)
	} else {
		user, err = s.userRepo.GetUserByUsername(req.Email)
	}

	if err != nil || user == nil {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	if !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	// Generate token
	token, err := auth.GenerateToken(user.ID, s.config.JWTSecret, s.config.JWTExpirationTime)
	if err != nil {
		return nil, ErrTokenGeneration
	}

	return &dto.AuthResponse{
		Token:     token,
		ExpiresAt: time.Now().Add(s.config.JWTExpirationTime),
		UserID:    user.ID,
	}, nil
}
