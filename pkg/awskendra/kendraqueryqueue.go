package awskendra

import (
	"context"
	"fmt"

	"github.com/DSSD-Madison/gmu/pkg/logger"
	"github.com/aws/aws-sdk-go-v2/service/kendra"
)

// QueryExecutor defines the interface used for submitting and managing
// asynchronous query execution to AWS Kendra.
type QueryExecutor interface {
	// EnqueueQuery submits a kendra.QueryInput to the executor for asynchronous
	// processing. It returns a QueryResult after the query has finished.
	EnqueueQuery(ctx context.Context, query kendra.QueryInput) QueryResult

	// Shutdown initiates a graceful shutdown of the QueryExecutor.
	Shutdown(ctx context.Context) error
}

// QueryResult holds the Results of a query and an error, if one occurs.
type QueryResult struct {
	Results KendraResults
	Error   error
}

// kendraQueryExecutor implements the QueryExecutor interface to be used for queuing
// search requests to AWS Kendra.
type kendraQueryExecutor struct {
	queue Queue[kendra.QueryInput, QueryResult]
	log   logger.Logger
}

// NewKendraQueryQueue instantiates a KendraQueryExecutor with given parameters.
func NewKendraQueryQueue(
	awsClient *kendra.Client,
	log logger.Logger,
	workerCount int,
	bufferSize int,
) QueryExecutor {
	executorLogger := log.With("component", "KendraQueryExecutor")

	processorFunc := func(ctx context.Context, query kendra.QueryInput) QueryResult {
		executorLogger.DebugContext(ctx, "Processing Kendra query job", "query", *query.QueryText)

		output, err := awsClient.Query(ctx, &query)
		if err != nil {
			executorLogger.ErrorContext(ctx, "Kendra API query failed", "error", err)
			return QueryResult{Error: err}
		}

		results := queryOutputToResults(*output)
		executorLogger.DebugContext(ctx, "Finished processing Kendra query job", "result_count", results.Count)
		return QueryResult{Results: results, Error: nil}
	}

	genericQueue := NewGenericQueue(
		workerCount,
		bufferSize,
		executorLogger,
		processorFunc,
	)

	executorLogger.Info("Kendra query executor initialized")

	return &kendraQueryExecutor{
		queue: genericQueue,
		log:   executorLogger,
	}
}

// EnqueueQuery handles the asynchronous execution of a Kendra Query.
// The query will be stopped early if the given context is finished.
func (q *kendraQueryExecutor) EnqueueQuery(ctx context.Context, query kendra.QueryInput) QueryResult {
	resultChan := make(chan QueryResult, 1)

	job := Job[kendra.QueryInput, QueryResult]{
		Payload:    query,
		ResultChan: resultChan,
		ctx:        ctx,
	}

	q.log.DebugContext(ctx, "Attempting to enqueue Kendra query job", "query", *query.QueryText)

	if !q.queue.Enqueue(job) {
		err := fmt.Errorf("failed to enqueue query, queue may be full or stopped")
		q.log.ErrorContext(ctx, "Failed to enqueue Kendra query job", "query", *query.QueryText, "error", err)
	}

	q.log.DebugContext(ctx, "Kendra queue job enqueued successfully", "query", *query.QueryText)

	select {
	case result, ok := <-resultChan:
		if !ok {
			err := fmt.Errorf("result channel closed unexpectedly for query")
			q.log.ErrorContext(ctx, "Result channel closed unexpectedly", "query", *query.QueryText, "error", err)
		}
		q.log.DebugContext(ctx, "Received result for Kendra query job", "query", *query.QueryText, "has_error", result.Error != nil)
		return result
	case <-ctx.Done():
		err := fmt.Errorf("context cancelled while waiting for query result: %w", ctx.Err())
		q.log.WarnContext(ctx, "Context cancelled while waiting for Kendra query result", "query", *query.QueryText, "error", err)
		return QueryResult{Error: err}
	}
}

// Shutdown initiates graceful shutdown for the Queue.
func (q *kendraQueryExecutor) Shutdown(ctx context.Context) error {
	q.log.Info("Shutting down Kendra query executor...")
	return q.queue.Shutdown(ctx)
}
