package services

import (
	"context"
	"net/url"

	"github.com/DSSD-Madison/gmu/pkg/awskendra"
	db "github.com/DSSD-Madison/gmu/pkg/db/generated"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
)

type Searcher interface {
	SearchDocuments(ctx context.Context, query string, filters url.Values, pageNum int) (awskendra.KendraResults, error)
}

type Suggester interface {
	GetSuggestions(ctx context.Context, query string) (awskendra.KendraSuggestions, error)
}

type LoginManager interface {
	ValidateUser(ctx context.Context, username string) (db.User, error)
	ValidatePassword(user db.User, password string) error
	CreateSession(c echo.Context, user db.User) (*sessions.Session, error)
	Logout(c echo.Context) error
}

type FileManager interface {
	UploadFile()
	GetMetadata()
	UpdateMetadata()
}
