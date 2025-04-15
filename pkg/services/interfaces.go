package services

import (
	"context"
	"net/url"

	"github.com/DSSD-Madison/gmu/pkg/awskendra"
)

type Searcher interface {
	SearchDocuments(ctx context.Context, query string, filters url.Values, pageNum int) (awskendra.KendraResults, error)
}

type Suggester interface {
	GetSuggestions(ctx context.Context, query string) (awskendra.KendraSuggestions, error)
}
