package awskendra

import "context"

type Job[P, R any] struct {
	Payload    P
	ResultChan chan<- R
	ctx        context.Context
}

type Queue[P, R any] interface {
	Enqueue(job Job[P, R]) bool
	Shutdown(ctx context.Context) error
}
