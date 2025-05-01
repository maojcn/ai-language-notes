package handlers

import (
	"ai-language-notes/internal/api/dto"
	"ai-language-notes/internal/config"
	"ai-language-notes/internal/repository"
	"ai-language-notes/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication-related requests.
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler creates a new AuthHandler instance.
func NewAuthHandler(userRepo repository.UserRepository, cfg config.Config) *AuthHandler {
	return &AuthHandler{
		authService: services.NewAuthService(userRepo, cfg),
	}
}

// Register handles user registration.
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.authService.Register(req)
	if err != nil {
		status := http.StatusInternalServerError

		switch err {
		case services.ErrUsernameExists, services.ErrEmailExists:
			status = http.StatusConflict
		case services.ErrPasswordProcessing, services.ErrUserCreation, services.ErrTokenGeneration:
			status = http.StatusInternalServerError
		}

		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// Login handles user login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.authService.Login(req)
	if err != nil {
		status := http.StatusInternalServerError

		switch err {
		case services.ErrInvalidCredentials:
			status = http.StatusUnauthorized
		case services.ErrTokenGeneration:
			status = http.StatusInternalServerError
		}

		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
