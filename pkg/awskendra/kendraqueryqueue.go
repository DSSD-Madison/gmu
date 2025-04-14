package awskendra

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/kendra"
)

type KendraQueryQueue struct {
	queue  Queue[kendra.QueryInput, QueryResult]
	client *KendraClient
}

func NewKendraQueryQueue(client *KendraClient, workerCount, maxItems int) *KendraQueryQueue {
	return &KendraQueryQueue{
		queue:  NewKendraQueue[kendra.QueryInput, QueryResult](workerCount, maxItems),
		client: client,
	}
}

// replace this with a handler for dependency injection
func (q *KendraQueryQueue) EnqueueQuery(query kendra.QueryInput) QueryResult {
	resultChan := make(chan QueryResult)

	job := Job[kendra.QueryInput, QueryResult]{
		Payload:    query,
		ResultChan: resultChan,
		Callback: func(query kendra.QueryInput) {
			out, err := q.client.client.Query(context.TODO(), &query)
			if err != nil {
				log.Printf("Kendra Query Failed %q", err)

				resultChan <- QueryResult{
					Results: KendraResults{},
					Error:   err,
				}
			}
			results := queryOutputToResults(*out)

			resultChan <- QueryResult{
				Results: results,
				Error:   nil,
			}
		},
	}
	q.queue.Enqueue(job)

	return <-resultChan
}
