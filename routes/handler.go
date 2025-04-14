package routes

import (
	"log/slog"

	"github.com/DSSD-Madison/gmu/pkg/awskendra"
	"github.com/DSSD-Madison/gmu/pkg/db"
)

type Handler struct {
	db     *db.Queries
	q      awskendra.QueryExecutor
	kendra *awskendra.KendraClient
	logger *slog.Logger
}

func NewHandler(db *db.Queries, q awskendra.QueryExecutor, k *awskendra.KendraClient, l *slog.Logger) Handler {
	return Handler{
		db:     db,
		q:      q,
		kendra: k,
		logger: l,
	}
}
