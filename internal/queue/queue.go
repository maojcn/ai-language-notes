package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	ProcessingQueueKey = "queue:note_processing"
)

// LLMProcessingTask represents a task for LLM processing
type LLMProcessingTask struct {
	NoteID         uuid.UUID `json:"note_id"`
	OriginalText   string    `json:"original_text"`
	UserID         uuid.UUID `json:"user_id"`
	NativeLanguage string    `json:"native_language"`
	TargetLanguage string    `json:"target_language"`
	CreatedAt      time.Time `json:"created_at"`
}

// QueueService handles the task queue operations
type QueueService struct {
	redisClient *redis.Client
}

// NewQueueService creates a new queue service
func NewQueueService(redisClient *redis.Client) *QueueService {
	return &QueueService{
		redisClient: redisClient,
	}
}

// EnqueueTask adds a task to the processing queue
func (s *QueueService) EnqueueTask(ctx context.Context, task *LLMProcessingTask) error {
	taskBytes, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	// Add to the Redis list that serves as our queue
	err = s.redisClient.LPush(ctx, ProcessingQueueKey, taskBytes).Err()
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	return nil
}

// DequeueTask gets the next task from the queue
func (s *QueueService) DequeueTask(ctx context.Context) (*LLMProcessingTask, error) {
	// Use BRPOP to block until an item is available or timeout occurs
	result, err := s.redisClient.BRPop(ctx, 0, ProcessingQueueKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to dequeue task: %w", err)
	}

	// The result is a slice where index 0 is the key name and index 1 is the value
	if len(result) < 2 {
		return nil, fmt.Errorf("unexpected result format from Redis")
	}

	var task LLMProcessingTask
	err = json.Unmarshal([]byte(result[1]), &task)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	return &task, nil
}
