package services

import (
	"ai-language-notes/internal/models"
	"ai-language-notes/internal/repository"

	"github.com/google/uuid"
)

// UserService defines the interface for user-related business logic
type UserService interface {
	GetUserByID(userID uuid.UUID) (*models.User, error)
	UpdateUserProfile(user *models.User, nativeLanguage *string, targetLanguage *string) (*models.User, error)
}

// userService implements the UserService interface
type userService struct {
	userRepo repository.UserRepository
}

// NewUserService creates a new UserService instance
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

// GetUserByID retrieves a user by their ID
func (s *userService) GetUserByID(userID uuid.UUID) (*models.User, error) {
	return s.userRepo.GetUserByID(userID)
}

// UpdateUserProfile updates user profile information
func (s *userService) UpdateUserProfile(user *models.User, nativeLanguage *string, targetLanguage *string) (*models.User, error) {
	if nativeLanguage != nil {
		user.NativeLanguage = *nativeLanguage
	}
	if targetLanguage != nil {
		user.TargetLanguage = *targetLanguage
	}

	return s.userRepo.UpdateUser(user)
}
