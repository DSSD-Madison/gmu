package routes

import (
	"github.com/DSSD-Madison/gmu/pkg/awskendra"
	db "github.com/DSSD-Madison/gmu/pkg/db/generated"
	"github.com/DSSD-Madison/gmu/pkg/logger"
)

type Handler struct {
	db     *db.Queries
	client awskendra.Client
	logger logger.Logger
}

func NewHandler(db *db.Queries, c awskendra.Client, l logger.Logger) *Handler {
	return &Handler{
		db:     db,
		client: c,
		logger: l,
	}
}
