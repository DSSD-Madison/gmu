package routes

import (
	"log/slog"

	"github.com/DSSD-Madison/gmu/pkg/awskendra"
	db "github.com/DSSD-Madison/gmu/pkg/db/generated"
)

type Handler struct {
	db      *db.Queries
	kendra  *awskendra.KendraClient
	bedrock *awskendra.BedrockClient
	logger  *slog.Logger
}

func NewHandler(db *db.Queries, k *awskendra.KendraClient, b *awskendra.BedrockClient, l *slog.Logger) Handler {
	return Handler{
		db:      db,
		kendra:  k,
		bedrock: b,
		logger:  l,
	}
}
