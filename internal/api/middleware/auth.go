package middleware

import (
	"net/http"
	"strings"

	"ai-language-notes/internal/auth"
	"ai-language-notes/internal/config"

	"github.com/gin-gonic/gin"
)

const (
	AuthorizationHeaderKey  = "Authorization"
	AuthorizationTypeBearer = "Bearer"
	UserIDKey               = "userID" // Key to store userID in Gin context
)

// AuthMiddleware creates a Gin middleware for JWT authentication.
func AuthMiddleware(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeaderKey)
		if len(authHeader) == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			return
		}

		fields := strings.Fields(authHeader)
		if len(fields) < 2 || !strings.EqualFold(fields[0], AuthorizationTypeBearer) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			return
		}

		accessToken := fields[1]
		claims, err := auth.ValidateToken(accessToken, cfg.JWTSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()}) // Provide specific error from validation
			return
		}

		// Set user ID in context for downstream handlers
		c.Set(UserIDKey, claims.UserID.String()) // Store as string for easier retrieval

		c.Next() // Proceed to the next handler
	}
}
