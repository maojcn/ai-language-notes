package handlers

import (
	"ai-language-notes/internal/api/dto"
	"ai-language-notes/internal/api/middleware"
	"ai-language-notes/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserHandler handles user profile related requests.
type UserHandler struct {
	UserService services.UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{UserService: userService}
}

// GetProfile retrieves the user's profile.
func (h *UserHandler) GetProfile(c *gin.Context) {
	// Extract user ID from context (set by auth middleware)
	userIDStr, exists := c.Get(middleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse UUID
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Fetch user using the service
	user, err := h.UserService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Return user profile
	c.JSON(http.StatusOK, user)
}

// UpdateProfile updates the user's profile.
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	// Extract user ID from context (set by auth middleware)
	userIDStr, exists := c.Get(middleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse UUID
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Parse update request
	var updateReq dto.UserProfileUpdateRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current user data
	user, err := h.UserService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Update user profile via service
	updatedUser, err := h.UserService.UpdateUserProfile(user, updateReq.NativeLanguage, updateReq.TargetLanguage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	// Return updated user profile
	c.JSON(http.StatusOK, updatedUser)
}
