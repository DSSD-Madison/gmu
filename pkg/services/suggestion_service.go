package services

import (
	"context"
	"fmt"

	"github.com/DSSD-Madison/gmu/pkg/awskendra"
	"github.com/DSSD-Madison/gmu/pkg/logger"
)

type suggestionService struct {
	log          logger.Logger
	kendraClient awskendra.Client
}

func NewSuggestionService(log logger.Logger, kendraClient awskendra.Client) Suggester {
	return &suggestionService{
		log:          log.With("service", "SuggestionService"),
		kendraClient: kendraClient,
	}
}

func (s *suggestionService) GetSuggestions(ctx context.Context, query string) (awskendra.KendraSuggestions, error) {
	s.log.InfoContext(ctx, "Fetching suggestions", "query", query)

	suggestions, err := s.kendraClient.GetSuggestions(ctx, query)
	if err != nil {
		s.log.ErrorContext(ctx, "Kendra GetSuggestions failed", "query", query, "error", err)
		return awskendra.KendraSuggestions{}, fmt.Errorf("failed to retrieve suggestions: %w", err)
	}

	s.log.InfoContext(ctx, "Suggestions fetched successfully", "query", query, "count", len(suggestions.Suggestions))
	return suggestions, nil
}
