package queue

import "context"

type Result[R any] struct {
	Value R
	Error error
}

func newJob[P, R any](ctx context.Context, payload P, resultChan chan R) Job[P, R] {
	return Job[P, R]{
		ctx:        ctx,
		Payload:    payload,
		ResultChan: resultChan,
	}
}

type Job[P, R any] struct {
	Payload    P
	ResultChan chan<- R
	ctx        context.Context
}

type Queue[P, R any] interface {
	Enqueue(ctx context.Context, payload P) (R, bool)
	Shutdown(ctx context.Context) error
}
