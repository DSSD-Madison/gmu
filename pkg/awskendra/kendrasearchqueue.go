package awskendra

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/kendra"
)

type KendraSearchQueue struct {
	queue  Queue[kendra.QueryInput, QueryResult[KendraResults]]
	client *kendra.Client
}

func NewKendraSearchQueue(c *kendra.Client) *KendraSearchQueue {
	return &KendraSearchQueue{
		queue:  NewKendraQueue[kendra.QueryInput, QueryResult[KendraResults]](2, 4),
		client: c,
	}
}

// replace this with a handler for dependency injection
func (q *KendraSearchQueue) EnqueueQuery(query kendra.QueryInput) QueryResult[KendraResults] {
	resultChan := make(chan QueryResult[KendraResults])

	job := Job[kendra.QueryInput, QueryResult[KendraResults]]{
		Payload:    query,
		ResultChan: resultChan,
		Callback: func(query kendra.QueryInput) {
			out, err := q.client.Query(context.TODO(), &query)
			if err != nil {
				log.Printf("Kendra Query Failed %q", err)

				resultChan <- QueryResult[KendraResults]{
					Results: KendraResults{},
					Error:   err,
				}
			}
			results := queryOutputToResults(*out)

			resultChan <- QueryResult[KendraResults]{
				Results: results,
				Error:   nil,
			}
		},
	}
	q.queue.Enqueue(job)

	return <-resultChan
}
