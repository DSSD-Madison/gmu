package services

import (
	db "github.com/DSSD-Madison/gmu/pkg/db/generated"
	"github.com/DSSD-Madison/gmu/pkg/logger"
)

type FilemanagerService struct {
	log logger.Logger
	db  db.Queries
}
