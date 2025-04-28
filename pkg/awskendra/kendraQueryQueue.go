package awskendra

import (
	"context"
	"time"

	"github.com/DSSD-Madison/gmu/pkg/cache"
	"github.com/DSSD-Madison/gmu/pkg/logger"
	"github.com/DSSD-Madison/gmu/pkg/queue"
	"github.com/aws/aws-sdk-go-v2/service/kendra"
)

type KendraQueryQueue queue.Queue[kendra.QueryInput, queue.Result[KendraResults]]

func NewKendraQueryQueue(
	awsClient *kendra.Client,
	c cache.Cache[KendraResults],
	log logger.Logger,
	workerCount int,
	bufferSize int,
) KendraQueryQueue {
	processorFunc := func(ctx context.Context, query kendra.QueryInput) queue.Result[KendraResults] {
		log.DebugContext(ctx, "Processing Kendra query job", "query", *query.QueryText)
		pageNum := int(*query.PageNumber)

		cachedResults, exists := c.Get(*query.QueryText)
		if exists {
			return queue.Result[KendraResults]{Value: cachedResults, Error: nil}
		}

		output, err := awsClient.Query(ctx, &query)
		if err != nil {
			log.ErrorContext(ctx, "Kendra API query failed", "error", err)
			return queue.Result[KendraResults]{Error: err}
		}

		results := queryOutputToResults(*output)

		calculatedPages := (results.Count + 9) / 10
		totalPages := min(calculatedPages, 10)

		results.PageStatus = PageStatus{
			CurrentPage: pageNum,
			PrevPage:    pageNum - 1,
			NextPage:    pageNum + 1,
			HasPrev:     pageNum > 1,
			HasNext:     pageNum < totalPages,
			TotalPages:  totalPages,
		}

		results.Query = *query.QueryText

		log.DebugContext(ctx, "Finished processing Kendra query job", "result_count", results.Count)
		if !exists {
			c.Set(*query.QueryText, results, time.Hour)
		}
		return queue.Result[KendraResults]{Value: results, Error: nil}
	}

	q := queue.NewGenericQueue(
		workerCount,
		bufferSize,
		log,
		processorFunc,
	)

	return q
}
