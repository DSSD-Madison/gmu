package models

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/kendra"
)

type KendraQueryQueue struct {
	queue Queue[kendra.QueryInput, QueryResult]
	client *kendra.Client
}

func NewKendraQueryQueue() *KendraQueryQueue {
	return &KendraQueryQueue{
		queue: NewKendraQueue[kendra.QueryInput, QueryResult](2, 4),
		client: kendraClient(),
	}
}

// replace this with a handler for dependency injection
func (q *KendraQueryQueue) EnqueueQuery(query kendra.QueryInput) QueryResult {
	resultChan := make(chan QueryResult)

	job := Job[kendra.QueryInput, QueryResult] {
		Payload: query,
		ResultChan: resultChan,
		Callback: func(query kendra.QueryInput) {
			out, err := q.client.Query(context.TODO(), &query)
			if err != nil {
				log.Printf("Kendra Query Failed %q", err)

				resultChan <- QueryResult{
					Results: KendraResults{},
					Error: err,
				}
			}
			results := queryOutputToResults(*out)

			resultChan <- QueryResult{
				Results: results,
				Error: nil,
			}
		},
	}
	q.queue.Enqueue(job)

	return <- resultChan
}


