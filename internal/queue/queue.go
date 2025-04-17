package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"ai-language-notes/internal/models"

	"github.com/redis/go-redis/v9"
)

const (
	noteProcessingQueue = "note_processing_queue"
)

// RedisQueue implements queue functionality using Redis
type RedisQueue struct {
	client *redis.Client
}

// NewRedisQueue creates a new Redis queue
func NewRedisQueue(client *redis.Client) *RedisQueue {
	return &RedisQueue{client: client}
}

// EnqueueNoteProcessing adds a note to the processing queue
func (q *RedisQueue) EnqueueNoteProcessing(ctx context.Context, task *models.NoteProcessingTask) error {
	taskData, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	err = q.client.RPush(ctx, noteProcessingQueue, taskData).Err()
	if err != nil {
		return fmt.Errorf("failed to push task to queue: %w", err)
	}

	return nil
}

// DequeueNoteProcessing retrieves a note from the processing queue
func (q *RedisQueue) DequeueNoteProcessing(ctx context.Context) (*models.NoteProcessingTask, error) {
	// Use BLPOP for blocking pop with timeout
	result, err := q.client.BLPop(ctx, 0, noteProcessingQueue).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to pop task from queue: %w", err)
	}

	// BLPOP returns a slice where the first element is the key name and the second is the value
	if len(result) != 2 {
		return nil, fmt.Errorf("unexpected result format from BLPOP")
	}

	var task models.NoteProcessingTask
	err = json.Unmarshal([]byte(result[1]), &task)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	return &task, nil
}

// Close closes the Redis client
func (q *RedisQueue) Close() error {
	return q.client.Close()
}
