package config

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper"
)

// Config stores all configuration of the application.
// The values are read by viper from a config file or environment variables.
type Config struct {
	// Server settings
	HTTPPort string `mapstructure:"HTTP_PORT"`

	// Database settings
	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     string `mapstructure:"DB_PORT"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME"`
	DBSslMode  string `mapstructure:"DB_SSLMODE"`

	// Authentication settings
	JWTSecret         string        `mapstructure:"JWT_SECRET"`
	JWTExpirationTime time.Duration `mapstructure:"JWT_EXPIRATION_HOURS"`

	// Redis config
	RedisAddr     string `mapstructure:"REDIS_ADDR"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
	RedisDB       int    `mapstructure:"REDIS_DB"`

	// LLM API settings
	OpenAIAPIKey   string `mapstructure:"OPENAI_API_KEY"`
	DeepSeekAPIKey string `mapstructure:"DEEPSEEK_API_KEY"`
	LLMProvider    string `mapstructure:"LLM_PROVIDER"` // "openai" or "deepseek"

	// Feature flags
	EnableCache bool `mapstructure:"ENABLE_CACHE"`

	// Worker settings
	WorkerCount int `mapstructure:"WORKER_COUNT"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName(".env") // Look for .env file
	viper.SetConfigType("env")

	viper.AutomaticEnv() // Read environment variables

	// Set default values
	viper.SetDefault("HTTP_PORT", "8080")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "password")
	viper.SetDefault("DB_NAME", "lang_learn_db")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("REDIS_ADDR", "localhost:6379")
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("REDIS_DB", 0)
	viper.SetDefault("JWT_SECRET", "supersecretkey")
	viper.SetDefault("JWT_EXPIRATION_HOURS", "72h")
	viper.SetDefault("LLM_PROVIDER", "deepseek")
	viper.SetDefault("ENABLE_CACHE", true)
	viper.SetDefault("WORKER_COUNT", 3)

	err = viper.ReadInConfig()
	// Ignore error if config file is not found, rely on env vars/defaults
	if _, ok := err.(viper.ConfigFileNotFoundError); !ok && err != nil {
		log.Printf("Warning: Error reading config file: %v. Using defaults/env vars.", err)
	}

	// Extract hours as int and convert to duration
	if err = viper.Unmarshal(&config); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err = validateConfig(&config); err != nil {
		log.Printf("Warning: %v", err)
	}

	log.Println("Configuration loaded successfully")
	return
}

// validateConfig performs validation of the configuration
func validateConfig(cfg *Config) error {
	// Check for critical configuration issues
	if cfg.JWTSecret == "supersecretkey" || cfg.JWTSecret == "" {
		return fmt.Errorf("WARNING: JWT_SECRET is set to default or empty. Please configure a strong secret!")
	}

	// Check LLM provider configuration
	if cfg.LLMProvider == "openai" && cfg.OpenAIAPIKey == "" {
		return fmt.Errorf("WARNING: OpenAI is selected as LLM_PROVIDER but OPENAI_API_KEY is not set")
	}

	if cfg.LLMProvider == "deepseek" && cfg.DeepSeekAPIKey == "" {
		return fmt.Errorf("WARNING: DeepSeek is selected as LLM_PROVIDER but DEEPSEEK_API_KEY is not set")
	}

	// Database connection validation
	if cfg.DBUser == "" || cfg.DBPassword == "" || cfg.DBName == "" {
		return fmt.Errorf("WARNING: Database configuration incomplete")
	}

	return nil
}
