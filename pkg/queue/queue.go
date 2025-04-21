package queue

import "context"

type Job[P, R any] struct {
	Payload    P
	ResultChan chan<- R
	ctx        context.Context
}

func NewJob[P, R any](ctx context.Context, payload P, resultChan chan<- R) Job[P, R] {
	return Job[P, R]{
		Payload:    payload,
		ResultChan: resultChan,
		ctx:        ctx,
	}
}

type Queue[P, R any] interface {
	Enqueue(job Job[P, R]) bool
	Shutdown(ctx context.Context) error
}
