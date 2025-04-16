package handlers

import (
	"ai-language-notes/internal/auth"
	"ai-language-notes/internal/config"
	"ai-language-notes/internal/models"
	"ai-language-notes/internal/repository"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthHandler handles authentication-related requests.
type AuthHandler struct {
	userRepo repository.UserRepository
	Config   config.Config
}

// NewAuthHandler creates a new AuthHandler instance.
func NewAuthHandler(userRepo repository.UserRepository, cfg config.Config) *AuthHandler {
	return &AuthHandler{
		userRepo: userRepo,
		Config:   cfg,
	}
}

// Register handles user registration.
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	existingUser, err := h.userRepo.GetUserByUsername(req.Username)
	if err == nil && existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	existingUser, err = h.userRepo.GetUserByEmail(req.Email)
	if err == nil && existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already in use"})
		return
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}

	// Create user
	newUser := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		ID:           uuid.New(),
	}

	createdUser, err := h.userRepo.CreateUser(newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Generate token
	token, err := auth.GenerateToken(createdUser.ID, h.Config.JWTSecret, h.Config.JWTExpirationTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate authentication token"})
		return
	}

	c.JSON(http.StatusCreated, models.AuthResponse{
		Token:     token,
		ExpiresAt: time.Now().Add(h.Config.JWTExpirationTime),
		UserID:    createdUser.ID,
	})
}

// Login handles user login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Try to find user by username or email
	var user *models.User
	var err error

	// Check if input is email (contains @)
	if strings.Contains(req.Email, "@") {
		user, err = h.userRepo.GetUserByEmail(req.Email)
	} else {
		user, err = h.userRepo.GetUserByUsername(req.Email)
	}

	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Verify password
	if !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate token
	token, err := auth.GenerateToken(user.ID, h.Config.JWTSecret, h.Config.JWTExpirationTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate authentication token"})
		return
	}

	c.JSON(http.StatusOK, models.AuthResponse{
		Token:     token,
		ExpiresAt: time.Now().Add(h.Config.JWTExpirationTime),
		UserID:    user.ID,
	})
}
