package storage

import (
	"ai-language-notes/internal/config"
	"ai-language-notes/internal/models"
	"fmt"
	"log"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type PostgresStore struct {
	db *gorm.DB
}

func NewPostgresStorage(cfg config.Config) (*PostgresStore, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSslMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Log SQL queries
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Database connection established")

	// Auto Migration (Simple for development, use migration files for production)
	// TODO: Change this to use migration files(e.g., goose) in production
	log.Println("Running AutoMigration...")
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		log.Printf("AutoMigration failed: %v", err)
		return nil, fmt.Errorf("automigration failed: %w", err)
	}
	log.Println("AutoMigration completed.")

	return &PostgresStore{db: db}, nil
}

// CreateUser creates a new user in the database
func (s *PostgresStore) CreateUser(user *models.User) (*models.User, error) {
	if err := s.db.Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return user, nil
}

// GetUserByID retrieves a user by their ID
func (s *PostgresStore) GetUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return &user, nil
}

// GetUserByUsername retrieves a user by their username
func (s *PostgresStore) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, "username = ?", username).Error; err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by their email
func (s *PostgresStore) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, "email = ?", email).Error; err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &user, nil
}

// UpdateUser updates an existing user in the database
func (s *PostgresStore) UpdateUser(user *models.User) (*models.User, error) {
	if err := s.db.Save(user).Error; err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	return user, nil
}

// DeleteUser removes a user from the database by ID
func (s *PostgresStore) DeleteUser(id uuid.UUID) error {
	if err := s.db.Delete(&models.User{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// GetAllUsers retrieves all users from the database
func (s *PostgresStore) GetAllUsers() ([]*models.User, error) {
	var users []*models.User
	if err := s.db.Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}
	return users, nil
}

// GetDB returns the underlying GORM DB instance if needed elsewhere (use carefully)
func (s *PostgresStore) GetDB() *gorm.DB {
	return s.db
}
