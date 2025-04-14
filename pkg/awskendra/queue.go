package awskendra

type Job[Payload, Result any] struct {
	Payload    Payload
	ResultChan chan Result
	Callback   func(result Payload)
}

type Queue[Payload, Result any] interface {
	Enqueue(job Job[Payload, Result])
	startWorkers(workerCount int)
	stopWorkers()
}
