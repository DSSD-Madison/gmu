package handlers

import (
	"context"

	"github.com/labstack/echo/v4"

	"github.com/DSSD-Madison/gmu/db"
)

type DBDocumentQuerier interface {
	GetDocumentsByURIs(ctx context.Context, uris []string) ([]db.Document, error)
}

func GetDocuments(c echo.Context, queries *db.Queries, uris []string) (map[string]db.Document, error) {
    documents, err := queries.GetDocumentsByURIs(c.Request().Context(), uris)
    if err != nil {
        return nil, err
    }

    documentMap := make(map[string]db.Document)

    for _, document := range documents {
        documentMap[document.S3File] = document
    }

    return documentMap, nil
}
