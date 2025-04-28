package queue

import (
	"context"
	"fmt"
	"sync"

	"github.com/DSSD-Madison/gmu/pkg/logger"
)

type ProcessorFunc[P, R any] func(ctx context.Context, payload P) R

type genericQueue[P, R any] struct {
	jobChan   chan Job[P, R]
	processor ProcessorFunc[P, R]
	log       logger.Logger
	wg        sync.WaitGroup
	stopChan  chan struct{}
	semaphore chan struct{}
}

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
		log:       log,
		stopChan:  make(chan struct{}),
	}

	q.startWorkers(workerCount)
	q.log.Info("Generic Queue started", "workers", workerCount, "buffer", bufferSize)
	return q
}

func (q *genericQueue[P, R]) Enqueue(ctx context.Context, payload P) (R, bool) {
	resultChan := make(chan R, 1)

	job := newJob(ctx, payload, resultChan)

	var result R
	select {
	case <-q.stopChan:
		q.log.WarnContext(job.ctx, "Enqueue failed: Queue is shutting down")
		return result, false
	default:
		select {
		case q.jobChan <- job:
			q.log.DebugContext(job.ctx, "Job enqueued successfully")
			result, ok := <-resultChan
			return result, ok
		case <-job.ctx.Done():
			q.log.WarnContext(job.ctx, "Enqueue failed: Job context cancelled before enqueueing")
			return result, false
		case <-q.stopChan:
			q.log.WarnContext(job.ctx, "Enqueu failed: Queue shut down during enqueue attempt")
			return result, false
		}
	}
}

func (q *genericQueue[P, R]) startWorkers(workerCount int) {
	q.wg.Add(workerCount)
	for i := range workerCount {
		go func(workerID int) {
			defer q.wg.Done()
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

func (q *genericQueue[P, R]) processJob(job Job[P, R], workerID int) {
	defer func() {
		if r := recover(); r != nil {
			q.log.ErrorContext(job.ctx, "Worker panicked while processing job", "panic", r, "worker_id", workerID)
			close(job.ResultChan)
		}
	}()

	if err := job.ctx.Err(); err != nil {
		q.log.WarnContext(job.ctx, "Job context cancelled before processing started", "error", err, "worker_id", workerID)
		close(job.ResultChan)
		return
	}

	q.log.DebugContext(job.ctx, "Worker processing job", "worker_id", workerID)
	result := q.processor(job.ctx, job.Payload)
	q.log.DebugContext(job.ctx, "Worker finished processing job", "worker_id", workerID)

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

	close(q.stopChan)

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
