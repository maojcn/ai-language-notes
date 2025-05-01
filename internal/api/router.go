package api

import (
	"ai-language-notes/internal/ai"
	"ai-language-notes/internal/api/handlers"
	"ai-language-notes/internal/api/middleware"
	"ai-language-notes/internal/config"
	"ai-language-notes/internal/repository"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRouter configures the Gin router with all routes and middleware.
func SetupRouter(
	cfg config.Config,
	userRepo repository.UserRepository,
	noteRepo repository.NoteRepository,
	llmService ai.LLMService,
) *gin.Engine {

	// gin.SetMode(gin.ReleaseMode) // Uncomment for production
	r := gin.Default() // Includes Logger and Recovery middleware

	// CORS Middleware Configuration
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"*"} // Allow all origins (adjust for production)
	// Or specify allowed origins: corsConfig.AllowOrigins = []string{"http://localhost:3000", "https://yourapp.com"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	// corsConfig.AllowCredentials = true // Uncomment if using cookies/sessions with credentials
	r.Use(cors.New(corsConfig))

	// Health Check Endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	// --- API v1 Routes ---
	v1 := r.Group("/api/v1") // Or just use root path "/" if preferred

	// --- Authentication Routes ---
	authHandler := handlers.NewAuthHandler(userRepo, cfg)
	authRoutes := v1.Group("/auth")
	{
		authRoutes.POST("/register", authHandler.Register)
		authRoutes.POST("/login", authHandler.Login)
	}

	// Apply JWT Authentication Middleware to protected routes
	authMiddleware := middleware.AuthMiddleware(cfg)

	// --- User Routes ---
	userHandler := handlers.NewUserHandler(userRepo)
	userRoutes := v1.Group("/user")
	userRoutes.Use(authMiddleware) // Protect user routes
	{
		userRoutes.GET("/profile", userHandler.GetProfile)
		userRoutes.PUT("/profile", userHandler.UpdateProfile)
	}

	// --- Note Routes ---
	noteHandler := handlers.NewNoteHandler(noteRepo, userRepo, llmService)
	noteRoutes := v1.Group("/notes")
	noteRoutes.Use(authMiddleware) // Protect note routes
	{
		noteRoutes.POST("", noteHandler.CreateNote)
		noteRoutes.GET("", noteHandler.GetUserNotes)
		noteRoutes.GET("/:id", noteHandler.GetNote)
		noteRoutes.DELETE("/:id", noteHandler.DeleteNote)
	}

	// Handle Not Found routes
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
	})

	return r
}
