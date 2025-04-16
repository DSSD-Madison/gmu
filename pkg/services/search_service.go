package services

import (
	"context"
	"fmt"
	"net/url"

	"github.com/DSSD-Madison/gmu/pkg/awskendra"
	"github.com/DSSD-Madison/gmu/pkg/db/generated"
	db_util "github.com/DSSD-Madison/gmu/pkg/db/util"
	"github.com/DSSD-Madison/gmu/pkg/logger"
)

type SearchService struct {
	log          logger.Logger
	kendraClient awskendra.Client
	dbQuerier    *db.Queries
}

func NewSearchService(log logger.Logger, kendra awskendra.Client, dbQuerier *db.Queries) *SearchService {
	return &SearchService{
		log:          log,
		kendraClient: kendra,
		dbQuerier:    dbQuerier,
	}
}

func (s *SearchService) SearchDocuments(ctx context.Context, query string, filters url.Values, pageNum int) (awskendra.KendraResults, error) {
	s.log.InfoContext(ctx, "Starting document search", "query", query, "page", pageNum)

	kendraFilterMap := convertURLValuesToKendraFilters(filters)
	s.log.DebugContext(ctx, "Converted filters for Kendra", "filter_map", kendraFilterMap)

	results, err := s.kendraClient.MakeQuery(ctx, query, kendraFilterMap, pageNum)
	if err != nil {
		s.log.ErrorContext(ctx, "Kendra MakeQuery failed", "query", query, "page", pageNum, "error", err)
		return awskendra.KendraResults{}, fmt.Errorf("failed to retrieve search results: %w", err)
	}
	s.log.DebugContext(ctx, "Received results from Kendra", "count", results.Count)

	if len(results.Results) > 0 {
		s.log.DebugContext(ctx, "Attempting to enrich results from database")

		err = db_util.AddImagesToResults(ctx, results, s.dbQuerier)
		if err != nil {
			s.log.WarnContext(ctx, "Failed to enrich results with DB data", "error", err)
		} else {
			s.log.DebugContext(ctx, "Successfully enriched results from database")
		}
	} else {
		s.log.DebugContext(ctx, "No results from Kendra to enrich")
	}

	s.log.InfoContext(ctx, "Document search completed", "query", query, "page", pageNum, "results_found", results.Count)
	return results, nil
}

// convertURLValuesToKendraFilters is a helper to transform filter format
func convertURLValuesToKendraFilters(values url.Values) map[string][]string {
	if values == nil {
		return nil
	}
	kendraFilters := make(map[string][]string)
	for k, v := range values {
		if len(v) > 0 {
			valsCopy := make([]string, len(v))
			copy(valsCopy, v)
			kendraFilters[k] = valsCopy
		}
	}
	if len(kendraFilters) == 0 {
		return nil
	}
	return kendraFilters
}
