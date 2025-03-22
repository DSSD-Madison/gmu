package routes

import (
	"github.com/DSSD-Madison/gmu/db"
	"github.com/DSSD-Madison/gmu/models"
)

type Handler struct {
	db *db.Queries
	kendra *models.KendraClient
}

func NewHandler(db *db.Queries, k *models.KendraClient) Handler {
	return Handler{
		db: db,
		kendra: k,
	}
}
