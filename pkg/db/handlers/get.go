package handlers

import (
	"context"

	db "github.com/DSSD-Madison/gmu/pkg/db/generated"
)

type DBDocumentQuerier interface {
	GetDocumentsByURIs(ctx context.Context, uris []string) ([]db.Document, error)
}

func GetDocuments(ctx context.Context, queries *db.Queries, uris []string) (map[string]db.GetDocumentsByURIsRow, error) {
	documents, err := queries.GetDocumentsByURIs(ctx, uris)
	if err != nil {
		return nil, err
	}

	documentMap := make(map[string]db.GetDocumentsByURIsRow)

	for _, document := range documents {
		documentMap[document.S3File] = document
	}

	return documentMap, nil
}
