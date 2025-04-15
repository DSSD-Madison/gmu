package routes

import (
	"log/slog"

	"github.com/DSSD-Madison/gmu/pkg/awskendra"
	db "github.com/DSSD-Madison/gmu/pkg/db/generated"
)

type Handler struct {
	db *db.Queries
	kendra *awskendra.KendraClient
	logger *slog.Logger
}

func NewHandler(db *db.Queries, k *awskendra.KendraClient, l *slog.Logger) Handler {
	return Handler{
		db: db,
		kendra: k,
		logger: l,
	}
}
