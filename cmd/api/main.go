package main

import (
	"ai-language-notes/internal/ai"
	"ai-language-notes/internal/api"
	"ai-language-notes/internal/config"
	"ai-language-notes/internal/queue"
	"ai-language-notes/internal/repository"
	"ai-language-notes/internal/storage"
	"ai-language-notes/internal/worker"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database connection
	pgStore, err := storage.NewPostgresStorage(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Init Redis
	redisClient, err := storage.InitRedis(&cfg)
	if err != nil {
		log.Fatalf("FATAL: Could not initialize Redis: %v\n", err)
	}
	defer redisClient.Close()

	// Initialize repositories using the proper implementations
	userRepo := repository.NewUserRepository(pgStore)
	noteRepo := repository.NewNoteRepository(pgStore)

	// Initialize the AI service using factory
	var apiKey string
	if cfg.LLMProvider == "openai" {
		apiKey = cfg.OpenAIAPIKey
	} else {
		apiKey = cfg.DeepSeekAPIKey
	}

	llmService, err := ai.CreateLLMServiceFromConfig(cfg.LLMProvider, apiKey)
	if err != nil {
		log.Fatalf("Failed to initialize LLM service: %v", err)
	}

	// Initialize queue service
	queueService := queue.NewQueueService(redisClient)

	// Initialize worker service (with 3 worker goroutines)
	workerService := worker.NewWorker(
		queueService,
		noteRepo,
		userRepo,
		llmService,
		cfg.WorkerCount,
	)

	// Start the worker service
	workerService.Start()

	// Setup router with repositories and services
	router := api.SetupRouter(cfg, userRepo, noteRepo, llmService, queueService)

	// Configure HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Shutdown worker first
	workerService.Stop()

	// Create a deadline to wait for current operations to complete
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited gracefully")
}
