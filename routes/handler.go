package routes

import (
	"log/slog"

	"github.com/DSSD-Madison/gmu/db"
)

type Handler struct {
	Logger *slog.Logger
	db *db.Queries
}

func NewHandler(logger *slog.Logger, db *db.Queries) Handler {
	return Handler{Logger: logger, db: db}
}
