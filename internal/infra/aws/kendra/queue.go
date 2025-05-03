package kendra

import (
	"context"
	"fmt"

	"github.com/DSSD-Madison/gmu/pkg/logger"
	"github.com/DSSD-Madison/gmu/pkg/queue"
	"github.com/aws/aws-sdk-go-v2/service/kendra"
	awskendra "github.com/aws/aws-sdk-go-v2/service/kendra"
)

type QueryExecutor interface {
	EnqueueQuery(ctx context.Context, query awskendra.QueryInput) QueryResult
	Shutdown(ctx context.Context) error
}

type QueryResult struct {
	Results awskendra.QueryOutput
	Error   error
}

type kendraQueryExecutor struct {
	queue queue.Queue[kendra.QueryInput, QueryResult]
	log   logger.Logger
}

func NewKendraQueryQueue(
	awsClient *awskendra.Client,
	log logger.Logger,
	workerCount int,
	bufferSize int,
) QueryExecutor {
	executorLogger := log.With("component", "KendraQueryExecutor")

	processorFunc := func(ctx context.Context, query awskendra.QueryInput) QueryResult {
		executorLogger.DebugContext(ctx, "Processing Kendra query job", "query", *query.QueryText)

		output, err := awsClient.Query(ctx, &query)
		if err != nil {
			executorLogger.ErrorContext(ctx, "Kendra API query failed", "error", err)
			return QueryResult{Error: err}
		}

		return QueryResult{Results: *output, Error: nil}
	}

	genericQueue := queue.NewGenericQueue(
		workerCount,
		bufferSize,
		executorLogger,
		processorFunc,
	)

	return &kendraQueryExecutor{
		queue: genericQueue,
		log:   executorLogger,
	}
}

func (q *kendraQueryExecutor) EnqueueQuery(ctx context.Context, query awskendra.QueryInput) QueryResult {
	resultChan := make(chan QueryResult, 1)

	job := queue.NewJob(ctx, query, resultChan)

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

func (q *kendraQueryExecutor) Shutdown(ctx context.Context) error {
	q.log.Info("Shutting down Kendra query executor...")
	return q.queue.Shutdown(ctx)
}
