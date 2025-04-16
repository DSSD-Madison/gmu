package awskendra

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/kendra"
)

type KendraSuggestionsQueue struct {
	queue  Queue[string, QueryResult[KendraSuggestions]]
	client *kendra.Client
}

func NewKendraSuggestionsQueue(c *kendra.Client) *KendraSuggestionsQueue {
	return &KendraSuggestionsQueue{
		queue:  NewKendraQueue[string, QueryResult[KendraSuggestions]](2, 4),
		client: c,
	}
}

func (q *KendraSuggestionsQueue) EnqueueQuery(query string, indexId *string) QueryResult[KendraSuggestions] {
	resultChan := make(chan QueryResult[KendraSuggestions])

	job := Job[string, QueryResult[KendraSuggestions]]{
		Payload:    query,
		ResultChan: resultChan,
		Callback: func(query string) {
			kendraQuery := kendra.GetQuerySuggestionsInput{
				IndexId:   indexId,
				QueryText: &query,
			}
			out, err := q.client.GetQuerySuggestions(context.TODO(), &kendraQuery)
			if err != nil {
				log.Printf("Kendra Query Failed %q", err)

				resultChan <- QueryResult[KendraSuggestions]{
					Results: KendraSuggestions{},
					Error:   err,
				}
			}

			suggestions := querySuggestionsOutputToSuggestions(*out)

			resultChan <- QueryResult[KendraSuggestions]{
				Results: suggestions,
				Error:   nil,
			}
		},
	}
	q.queue.Enqueue(job)

	return <-resultChan
}
