package worker

import (
	"context"
	"fmt"
	"sync"

	"concurrent-downloader/database"
	"concurrent-downloader/downloader"
	"concurrent-downloader/models"
	"concurrent-downloader/parser"
)

// WorkerPool manages a pool of workers
type WorkerPool struct {
	workers    map[int]context.CancelFunc // Track each worker's cancel function
	mu         sync.Mutex
	globalCtx  context.Context
	numWorkers int
	jobs       <-chan models.Job
	db         *database.DB
	wg         *sync.WaitGroup
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(ctx context.Context, numWorkers int, jobs <-chan models.Job, db *database.DB, wg *sync.WaitGroup) *WorkerPool {
	return &WorkerPool{
		workers:    make(map[int]context.CancelFunc),
		globalCtx:  ctx,
		numWorkers: numWorkers,
		jobs:       jobs,
		db:         db,
		wg:         wg,
	}
}

// Start starts all workers
func (wp *WorkerPool) Start() {
	for i := 1; i <= wp.numWorkers; i++ {
		// Create per-worker context
		workerCtx, workerCancel := context.WithCancel(wp.globalCtx)

		// Store cancel function
		wp.mu.Lock()
		wp.workers[i] = workerCancel
		wp.mu.Unlock()

		wp.wg.Add(1)
		go wp.worker(workerCtx, i)
	}
}

// CancelWorker cancels a specific worker by ID
func (wp *WorkerPool) CancelWorker(workerID int) {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if cancel, exists := wp.workers[workerID]; exists {
		fmt.Printf("🛑 Cancelling worker %d\n", workerID)
		cancel()
		delete(wp.workers, workerID)
	}
}

// worker processes jobs from the channel
func (wp *WorkerPool) worker(ctx context.Context, id int) {
	defer wp.wg.Done()

	for job := range wp.jobs {
		// Check if this worker or global context is cancelled
		select {
		case <-ctx.Done():
			fmt.Printf("Worker %d: shutting down (cancelled)\n", id)
			return
		default:
		}

		fmt.Printf("Worker %d: processing job %d - %s\n", id, job.ID, job.URL)

		// Step 1: Download
		data, err := downloader.Download(ctx, job.URL)
		if err != nil {
			fmt.Printf("Worker %d: download failed for job %d: %v\n", id, job.ID, err)
			continue
		}

		// Step 2: Parse
		todo, err := parser.Parse(data)
		if err != nil {
			fmt.Printf("Worker %d: parse failed for job %d: %v\n", id, job.ID, err)
			continue
		}

		// Step 3: Insert to database
		err = wp.db.Insert(todo)
		if err != nil {
			fmt.Printf("Worker %d: database insert failed for job %d: %v\n", id, job.ID, err)
			continue
		}

		fmt.Printf("Worker %d: completed job %d ✓\n", id, job.ID)
	}

	fmt.Printf("Worker %d: finished\n", id)
}
