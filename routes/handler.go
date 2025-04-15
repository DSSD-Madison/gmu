package routes

import (
	"log/slog"

	"github.com/DSSD-Madison/gmu/pkg/awskendra"
	"github.com/DSSD-Madison/gmu/pkg/db"
)

type Handler struct {
	db     *db.Queries
	q      awskendra.QueryExecutor
	client awskendra.Client
	logger *slog.Logger
}

func NewHandler(db *db.Queries, q awskendra.QueryExecutor, c awskendra.Client, l *slog.Logger) Handler {
	return Handler{
		db:     db,
		q:      q,
		client: c,
		logger: l,
	}
}
