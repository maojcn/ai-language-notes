package storage

import (
	"ai-language-notes/internal/config"
	"ai-language-notes/internal/models"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// PostgresStore manages the database connection
type PostgresStore struct {
	db *gorm.DB
}

// NewPostgresStorage creates a new database connection
func NewPostgresStorage(cfg config.Config) (*PostgresStore, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSslMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Database connection established")

	// Auto Migration with explicit unique constraints
	log.Println("Running AutoMigration...")
	err = db.AutoMigrate(&models.User{}, &models.Note{}, &models.Tag{})
	if err != nil {
		log.Printf("AutoMigration failed: %v", err)
		return nil, fmt.Errorf("automigration failed: %w", err)
	}

	// Check and create unique constraints explicitly
	if !db.Migrator().HasIndex(&models.User{}, "idx_users_username") {
		err = db.Exec("CREATE UNIQUE INDEX idx_users_username ON users(username)").Error
		if err != nil {
			log.Printf("Failed to create unique index on username: %v", err)
			return nil, err
		}
	}

	if !db.Migrator().HasIndex(&models.User{}, "idx_users_email") {
		err = db.Exec("CREATE UNIQUE INDEX idx_users_email ON users(email)").Error
		if err != nil {
			log.Printf("Failed to create unique index on email: %v", err)
			return nil, err
		}
	}

	log.Println("AutoMigration completed.")

	return &PostgresStore{db: db}, nil
}

// GetDB returns the underlying GORM DB instance
func (s *PostgresStore) GetDB() *gorm.DB {
	return s.db
}
