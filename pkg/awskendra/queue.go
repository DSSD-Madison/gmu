package awskendra

import "context"

// Job represents a unit of work to be processed by a Queue.
// P is the type of the job's payload (input data).
// R is the type of the job's expected result.
type Job[P, R any] struct {
	// The input data for this struct
	Payload P

	// ResultChan is a channel where the result of processing this job
	// will be sent. It should be a send-only channel.
	// The channel will be closed after the result is sent, or if processing fails
	// or is cancelled before a result can be sent.
	ResultChan chan<- R

	// ctx is the context associated with a job, used for cancellation
	// signals or deadlines during processing and enqueuing.
	ctx context.Context
}

// Queue defines the interface for a generic job processing queue.
// It allows submitting jobs with a specific payload of type P
// and receiving results of type R.
// P is the type of the job's payload.
// R is the type of the job's result.
type Queue[P, R any] interface {
	// Enqueue attempts to add a job to the queue for processing.
	// It returns true if the job was successfully accepted by the queue
	// before the queue started shutting down or the job's context was cancelled.
	// It returns false otherwise.
	//
	// Implementations should respect the job's context
	// during the enqueue attempt itself.
	Enqueue(job Job[P, R]) bool

	// Shutdown initiates a graceful shutdown of the queue.
	// It signals the queue to stop accepting new jobs and waits
	// for currently enqueued or executing jobs to finish within the deadline
	// provided by the context.
	// It returns nil if the shutdown completes successfully within the context's deadline,
	// or an error if the context is cancelled or times out before shutdown is complete.
	// After Shutdown is cancelled, subsequent calls to Enqueue should fail.
	Shutdown(ctx context.Context) error
}
