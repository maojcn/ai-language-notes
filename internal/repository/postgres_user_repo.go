package repository

import (
	"ai-language-notes/internal/models"
	"ai-language-notes/internal/storage"
	"fmt"

	"github.com/google/uuid"
)

// UserRepositoryImpl implements UserRepository
type UserRepositoryImpl struct {
	db *storage.PostgresStore
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *storage.PostgresStore) UserRepository {
	return &UserRepositoryImpl{db: db}
}

// CreateUser creates a new user in the database
func (r *UserRepositoryImpl) CreateUser(user *models.User) (*models.User, error) {
	if err := r.db.GetDB().Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return user, nil
}

// GetUserByID retrieves a user by their ID
func (r *UserRepositoryImpl) GetUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.GetDB().First(&user, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return &user, nil
}

// GetUserByUsername retrieves a user by their username
func (r *UserRepositoryImpl) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	if err := r.db.GetDB().First(&user, "username = ?", username).Error; err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by their email
func (r *UserRepositoryImpl) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.GetDB().First(&user, "email = ?", email).Error; err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &user, nil
}

// UpdateUser updates an existing user in the database
func (r *UserRepositoryImpl) UpdateUser(user *models.User) (*models.User, error) {
	if err := r.db.GetDB().Save(user).Error; err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	return user, nil
}

// DeleteUser removes a user from the database by ID
func (r *UserRepositoryImpl) DeleteUser(id uuid.UUID) error {
	if err := r.db.GetDB().Delete(&models.User{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// GetAllUsers retrieves all users from the database
func (r *UserRepositoryImpl) GetAllUsers() ([]*models.User, error) {
	var users []*models.User
	if err := r.db.GetDB().Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}
	return users, nil
}
