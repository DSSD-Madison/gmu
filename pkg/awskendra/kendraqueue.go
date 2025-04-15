package awskendra

import (
	"sync"
)

type KendraQueue[Payload, Result any] struct {
	jobs      []Job[Payload, Result]
	mu        sync.Mutex
	cond      *sync.Cond
	workers   []chan bool
	semaphore chan struct{}
}

func NewKendraQueue[Payload, Result any](workerCount int, maxItems int) *KendraQueue[Payload, Result] {
	q := &KendraQueue[Payload, Result]{
		semaphore: make(chan struct{}, maxItems),
	}
	q.cond = sync.NewCond(&q.mu)
	q.startWorkers(workerCount)
	return q
}

func (q *KendraQueue[Payload, Result]) Enqueue(job Job[Payload, Result]) {
	q.semaphore <- struct{}{}

	q.mu.Lock()
	defer q.mu.Unlock()
	q.jobs = append(q.jobs, job)
	q.cond.Signal()
}

func (q *KendraQueue[Payload, Result]) startWorkers(workerCount int) {
	q.workers = make([]chan bool, workerCount)

	for i := 0; i < workerCount; i++ {
		stopChan := make(chan bool)
		q.workers[i] = stopChan

		go func(workerID int, stopChan chan bool) {
			for {
				q.mu.Lock()

				for len(q.jobs) == 0 {
					q.cond.Wait()
				}

				job := q.jobs[0]
				q.jobs = q.jobs[1:]
				q.mu.Unlock()

				if job.Callback != nil {
					job.Callback(job.Payload)
				}

				<-q.semaphore

				select {
				case <-stopChan:
					return
				default:
				}
			}
		}(i, stopChan)
	}
}

func (q *KendraQueue[Payload, Result]) stopWorkers() {
	for _, stopChan := range q.workers {
		close(stopChan)
	}
}

type QueryResult struct {
	Results KendraResults
	Error   error
}
