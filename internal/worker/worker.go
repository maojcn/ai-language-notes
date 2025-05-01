package worker

import (
	"ai-language-notes/internal/ai"
	"ai-language-notes/internal/models"
	"ai-language-notes/internal/queue"
	"ai-language-notes/internal/repository"
	"context"
	"log"
	"sync"
	"time"
)

// Worker is responsible for processing LLM tasks
type Worker struct {
	queueService *queue.QueueService
	noteRepo     repository.NoteRepository
	userRepo     repository.UserRepository
	llmService   ai.LLMService
	workerCount  int
	stopCh       chan struct{}
	wg           sync.WaitGroup
}

// NewWorker creates a new worker
func NewWorker(
	queueService *queue.QueueService,
	noteRepo repository.NoteRepository,
	userRepo repository.UserRepository,
	llmService ai.LLMService,
	workerCount int,
) *Worker {
	return &Worker{
		queueService: queueService,
		noteRepo:     noteRepo,
		userRepo:     userRepo,
		llmService:   llmService,
		workerCount:  workerCount,
		stopCh:       make(chan struct{}),
	}
}

// Start begins the worker processing
func (w *Worker) Start() {
	for i := 0; i < w.workerCount; i++ {
		w.wg.Add(1)
		go w.processLoop(i)
	}
	log.Printf("Started %d workers for LLM processing", w.workerCount)
}

// Stop gracefully stops the worker
func (w *Worker) Stop() {
	close(w.stopCh)
	w.wg.Wait()
	log.Println("All workers stopped")
}

// processLoop is the main worker loop
func (w *Worker) processLoop(workerID int) {
	defer w.wg.Done()

	log.Printf("Worker %d started", workerID)

	for {
		select {
		case <-w.stopCh:
			log.Printf("Worker %d received stop signal", workerID)
			return
		default:
			// Create a timeout context for task processing
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

			// Try to get a task from the queue
			task, err := w.queueService.DequeueTask(ctx)
			cancel()

			if err != nil {
				log.Printf("Worker %d error dequeueing task: %v", workerID, err)
				// Small pause to prevent tight loop if there's a persistent error
				time.Sleep(1 * time.Second)
				continue
			}

			w.processTask(task, workerID)
		}
	}
}

// processTask processes a single LLM task
func (w *Worker) processTask(task *queue.LLMProcessingTask, workerID int) {
	log.Printf("Worker %d processing note %s", workerID, task.NoteID)

	// Get the note from the database
	note, err := w.noteRepo.GetNoteByID(task.NoteID)
	if err != nil {
		log.Printf("Worker %d failed to get note %s: %v", workerID, task.NoteID, err)
		return
	}

	// Update note status to processing
	note.Status = models.StatusProcessing
	_, err = w.noteRepo.UpdateNote(note)
	if err != nil {
		log.Printf("Worker %d failed to update note status to processing: %v", workerID, err)
		return
	}

	// Process the text with LLM service
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	processedContent, err := w.llmService.ProcessText(
		ctx,
		task.OriginalText,
		task.NativeLanguage,
		task.TargetLanguage,
	)

	if err != nil {
		log.Printf("Worker %d LLM processing failed for note %s: %v", workerID, task.NoteID, err)
		note.Status = models.StatusFailed
		note.ErrorMessage = "Failed to process text: " + err.Error()
		_, updateErr := w.noteRepo.UpdateNote(note)
		if updateErr != nil {
			log.Printf("Worker %d failed to update note with error status: %v", workerID, updateErr)
		}
		return
	}

	// Update note with processed content
	note.GeneratedContent = processedContent.Content
	note.Status = models.StatusCompleted

	// Add tags from processed content
	var tags []models.Tag
	for _, tagName := range processedContent.Tags {
		tags = append(tags, models.Tag{
			Name: tagName,
		})
	}
	note.Tags = tags

	// Save the updated note
	_, err = w.noteRepo.UpdateNote(note)
	if err != nil {
		log.Printf("Worker %d failed to save processed note: %v", workerID, err)
		return
	}

	log.Printf("Worker %d successfully processed note %s", workerID, task.NoteID)
}
