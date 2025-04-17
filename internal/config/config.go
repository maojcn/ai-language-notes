package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

// Config stores all configuration of the application.
// The values are read by viper from a config file or environment variables.
type Config struct {
	HTTPPort          string        `mapstructure:"HTTP_PORT"`
	DBHost            string        `mapstructure:"DB_HOST"`
	DBPort            string        `mapstructure:"DB_PORT"`
	DBUser            string        `mapstructure:"DB_USER"`
	DBPassword        string        `mapstructure:"DB_PASSWORD"`
	DBName            string        `mapstructure:"DB_NAME"`
	DBSslMode         string        `mapstructure:"DB_SSLMODE"`
	JWTSecret         string        `mapstructure:"JWT_SECRET"`
	JWTExpirationTime time.Duration `mapstructure:"JWT_EXPIRATION_HOURS"`

	// Redis Config
	RedisAddr     string `mapstructure:"REDIS_ADDR"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
	RedisDB       int    `mapstructure:"REDIS_DB"`

	OpenAIAPIKey   string `mapstructure:"OPENAI_API_KEY"`
	DeepSeekAPIKey string `mapstructure:"DEEPL_API_KEY"`
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
	viper.SetDefault("JWT_EXPIRATION_HOURS", 72)

	err = viper.ReadInConfig()
	// Ignore error if config file is not found, rely on env vars/defaults
	if _, ok := err.(viper.ConfigFileNotFoundError); !ok && err != nil {
		log.Printf("Warning: Error reading config file: %v. Using defaults/env vars.", err)
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return
	}

	// Convert hours to duration
	config.JWTExpirationTime = config.JWTExpirationTime * time.Hour

	// Basic validation (add more as needed)
	if config.JWTSecret == "supersecretkey" || config.JWTSecret == "" {
		log.Println("WARNING: JWT_SECRET is set to default or empty. Please configure a strong secret!")
	}

	log.Println("Configuration loaded successfully")
	return
}
