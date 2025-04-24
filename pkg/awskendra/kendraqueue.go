package awskendra

import (
	"context"
	"fmt"
	"sync"

	"github.com/DSSD-Madison/gmu/pkg/logger"
)

// ProcessorFunc is a function type that defines the task executed
// by each worker in the queue. It takes a context and a payload
// of type P, and returns a result of type R.
type ProcessorFunc[P, R any] func(ctx context.Context, payload P) R

// genericQueue is an unexported implementation of the Queue interface.
// It manages a pool of workers that process jobs submitted via the
// Enqueue method. The queue has a configurable buffer size and
// processes jobs using a provided ProcessorFunc.
// It uses channels for job submission, worker control, and shutdown
// signaling. The sync.WaitGroup tracks active workers.
type genericQueue[P, R any] struct {
	jobChan   chan Job[P, R]      // jobChan is the buffered channel where jobs are sent for processing.
	processor ProcessorFunc[P, R] // processor is the function executed by workers for each job payload.
	log       logger.Logger       // log is the logger instance used for logging queue operations.
	wg        sync.WaitGroup      // wg tracks the number of active worker goroutines.
	stopChan  chan struct{}       // stopChan is closed to signal workers and the enqueue method that the queue is shutting down.
	semaphore chan struct{}       // NOTE: This field exists but isn't used in the provided methods. Document if it becomes active.
}

// NewGenericQueue creates and starts a new generic worker queue.
// It initializes a pool of `workerCount` goroutines that listen for
// jobs on an internal channel with a buffer capacity of `bufferSize`.
// The `processor` function is executed by workers for each job's payload.
// A `logger.Logger` is required for logging queue operations.
func NewGenericQueue[P, R any](
	workerCount int,
	bufferSize int,
	log logger.Logger,
	processor ProcessorFunc[P, R],
) Queue[P, R] {
	if workerCount <= 0 {
		workerCount = 1
	}
	if bufferSize < 0 {
		bufferSize = 0
	}

	q := &genericQueue[P, R]{
		jobChan:   make(chan Job[P, R], bufferSize),
		processor: processor,
		log:       log.With("component", "GenericQueue"), // add context to the logger
		stopChan:  make(chan struct{}),
		// semaphore: if semaphore is used, initialize it here.
	}

	q.startWorkers(workerCount)
	q.log.Info("Generic Queue started", "workers", workerCount, "buffer", bufferSize)
	return q
}

// Enqueue attempts to add a Job to the queue for processing.
// It returns true if the job was successfully enqueued before
// the queue started shutting down or the job's context was cancelled.
// It returns false if the queue is already shutting down, or if
// the job's context is cancelled while waiting to be enqueued
// into the potentially buffered channel.
//
// The method respects the job's context for cancellation during the
// enqueue attempt itself.
func (q *genericQueue[P, R]) Enqueue(job Job[P, R]) bool {
	select {
	case <-q.stopChan:
		// Queue is shutting down, do not accept new jobs.
		q.log.WarnContext(job.ctx, "Enqueue failed: Queue is shutting down")
		close(job.ResultChan)
		return false
	default:
		// Try to enqueue job or react to contex for shutdown or cancellation
		select {
		case q.jobChan <- job:
			q.log.DebugContext(job.ctx, "Job enqueued successfully")
			return true
		case <-job.ctx.Done():
			q.log.WarnContext(job.ctx, "Enqueue failed: Job context cancelled before enqueueing")
			close(job.ResultChan)
			return false
		case <-q.stopChan:
			q.log.WarnContext(job.ctx, "Enqueu failed: Queue shut down during enqueue attempt")
			close(job.ResultChan)
			return false
		}
	}
}

// startWorkers is an unexported helper method that launches the specified
// number of worker goroutines. Each worker reads jobs from the job channel
// and processes them using the configured processor function. Workers exit
// when the job channel is closed or the stop channel is signaled.
func (q *genericQueue[P, R]) startWorkers(workerCount int) {
	q.wg.Add(workerCount)
	for i := range workerCount {
		go func(workerID int) {
			defer q.wg.Done()
			q.log.Info("Worker started", "worker_id", workerID)
			for {
				select {
				case job, ok := <-q.jobChan:
					if !ok {
						q.log.Info("Worker stopping: job channel closed", "worker_id", workerID)
						return
					}

					q.processJob(job, workerID)

				case <-q.stopChan:
					q.log.Info("Worker stopping: shutdown signal received", "worker_id", workerID)
					return
				}
			}
		}(i)
	}
}

// processJob is an unexported helper method executed by workers to process
// a single job. It calls the configured ProcessorFunc with the job's context
// and payload, handles potential panics during processing, checks for context
// cancellation before and after processing, and attempts to send the result
// back on the job's ResultChan. It ensures the ResultChan is closed after
// processing is complete (or failed/cancelled).
func (q *genericQueue[P, R]) processJob(job Job[P, R], workerID int) {
	// Defer recovery to handle panics in the processor function
	defer func() {
		if r := recover(); r != nil {
			q.log.ErrorContext(job.ctx, "Worker panicked while processing job", "panic", r, "worker_id", workerID)
			close(job.ResultChan)
		}
	}()

	// Check if the job's context is already cancelled before processing
	if err := job.ctx.Err(); err != nil {
		q.log.WarnContext(job.ctx, "Job context cancelled before processing started", "error", err, "worker_id", workerID)
		close(job.ResultChan)
		return
	}

	q.log.DebugContext(job.ctx, "Worker processing job", "worker_id", workerID)

	// execute the actual processor function
	result := q.processor(job.ctx, job.Payload)

	q.log.DebugContext(job.ctx, "Worker finished processing job", "worker_id", workerID)

	// Attempt to send the result back
	// This may fail if the context is cancelled or the queue shuts down while sending
	select {
	case job.ResultChan <- result:
		q.log.DebugContext(job.ctx, "Worker sent result successfully", "worker_id", workerID)
	case <-job.ctx.Done():
		q.log.WarnContext(job.ctx, "Job context cancelled before worker could send result", "error", job.ctx.Err(), "worker_id", workerID)
	case <-q.stopChan:
		q.log.WarnContext(job.ctx, "Queue shutdown before worker could send result", "worker_id", workerID)
	}
	close(job.ResultChan)
}

func (q *genericQueue[P, R]) Shutdown(ctx context.Context) error {
	q.log.Info("Initiating queue shutdown...")

	// Signal to Enqueue and workers that shutdown has started
	close(q.stopChan)

	// Close the job channel. This signals to workers that no more jobs will be sent.
	// Workers receiving zero value from jobChan will exit their loop after this executes.
	close(q.jobChan)

	done := make(chan struct{})
	go func() {
		q.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		q.log.Info("Queue shutdown complete.")
		return nil
	case <-ctx.Done():
		q.log.Error("Queue shutdown timed out", "error", ctx.Err())
		return fmt.Errorf("queue shutdown timed out: %w", ctx.Err())
	}
}
