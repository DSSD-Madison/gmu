package awskendra

import (
	"context"

	"github.com/DSSD-Madison/gmu/pkg/logger"
	"github.com/DSSD-Madison/gmu/pkg/queue"
	"github.com/aws/aws-sdk-go-v2/service/kendra"
)

type KendraSuggestionQueue queue.Queue[kendra.GetQuerySuggestionsInput, queue.Result[KendraSuggestions]]

func NewKendraSuggestionsQueue(
	awsClient *kendra.Client,
	log logger.Logger,
	workerCount int,
	bufferSize int,
) KendraSuggestionQueue {
	processorFunc := func(ctx context.Context, query kendra.GetQuerySuggestionsInput) queue.Result[KendraSuggestions] {
		log.DebugContext(ctx, "Processing Kendra query job", "query", *query.QueryText)

		output, err := awsClient.GetQuerySuggestions(ctx, &query)
		if err != nil {
			log.ErrorContext(ctx, "Kendra API query failed", "error", err)
			return queue.Result[KendraSuggestions]{Error: err}
		}

		results := querySuggestionsOutputToSuggestions(*output)
		return queue.Result[KendraSuggestions]{Value: results, Error: nil}
	}

	q := queue.NewGenericQueue(
		workerCount,
		bufferSize,
		log,
		processorFunc,
	)

	return q
}
