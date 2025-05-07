package application

import (
	"context"
	"fmt"

	"github.com/DSSD-Madison/gmu/internal/domain/search"
	"github.com/DSSD-Madison/gmu/internal/infra/aws/kendra"
	"github.com/DSSD-Madison/gmu/pkg/logger"
)

type suggestionService struct {
	log          logger.Logger
	kendraClient kendra.Client
}

func NewSuggestionService(log logger.Logger, kendraClient kendra.Client) Suggester {
	serviceLogger := log.With("Service", "Suggestion")
	return &suggestionService{
		log:          serviceLogger,
		kendraClient: kendraClient,
	}
}

func (s *suggestionService) GetSuggestions(ctx context.Context, query string) (search.Suggestions, error) {
	s.log.DebugContext(ctx, "Fetching suggestions", "query", query)

	suggestions, err := s.kendraClient.GetSuggestions(ctx, query)
	if err != nil {
		s.log.ErrorContext(ctx, "Kendra GetSuggestions failed", "query", query, "error", err)
		return search.Suggestions{}, fmt.Errorf("failed to retrieve suggestions: %w", err)
	}

	s.log.DebugContext(ctx, "Suggestions fetched successfully", "query", query, "count", len(suggestions.Suggestions))
	return suggestions, nil
}
