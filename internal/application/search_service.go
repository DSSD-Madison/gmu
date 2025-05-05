package application

import (
	"context"
	"fmt"
	"net/url"

	"github.com/DSSD-Madison/gmu/internal/domain/search"
	"github.com/DSSD-Madison/gmu/internal/infra/aws/kendra"
	db "github.com/DSSD-Madison/gmu/internal/infra/database/sqlc/generated"
	db_util "github.com/DSSD-Madison/gmu/internal/infra/database/util"
	"github.com/DSSD-Madison/gmu/pkg/logger"
)

type SearchService struct {
	log          logger.Logger
	kendraClient kendra.Client
	dbQuerier    *db.Queries
}

func NewSearchService(log logger.Logger, kendra kendra.Client, dbQuerier *db.Queries) *SearchService {
	serviceLogger := log.With("Service", "Search")
	return &SearchService{
		log:          serviceLogger,
		kendraClient: kendra,
		dbQuerier:    dbQuerier,
	}
}

func (s *SearchService) SearchDocuments(ctx context.Context, query string, filters url.Values, pageNum int) (search.Results, error) {
	s.log.DebugContext(ctx, "Starting document search",
		"query", query,
		"page", pageNum,
		"filter_count", len(filters))

	kendraFilterMap := convertURLValuesToKendraFilters(filters)

	results, err := s.kendraClient.MakeQuery(ctx, query, kendraFilterMap, pageNum)
	if err != nil {
		s.log.ErrorContext(ctx, "Kendra MakeQuery failed",
			"query", query,
			"page", pageNum,
			"filters", filters,
			"error", err)
		return search.Results{}, fmt.Errorf("failed to retrieve search results from Kendra: %w", err)
	}

	if len(results.Results) > 0 {
		s.log.DebugContext(ctx, "Attempting to enrich results from database")

		err = db_util.AddImagesToResults(ctx, results, s.dbQuerier)
		if err != nil {
			s.log.WarnContext(ctx, "Failed to enrich results with DB data",
				"query", query,
				"error", err)
		} else {
			s.log.DebugContext(ctx, "Successfully enriched results from database")
		}
	} else {
		s.log.DebugContext(ctx, "No results from Kendra to enrich")
	}

	s.log.DebugContext(ctx, "Document search completed",
		"query", query,
		"page", pageNum,
		"results_found", results.Count)
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
